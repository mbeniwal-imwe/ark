package backup

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/core/crypto"
	awsfeat "github.com/mbeniwal-imwe/ark/internal/features/aws"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

var (
	profileName string
	bucketName  string
	prefix      string
)

var BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create and restore encrypted backups to S3",
}

var configureCmd = &cobra.Command{
	Use:   "configure <bucket> [prefix]",
	Short: "Configure S3 bucket for backups",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]
		pref := "ark/backup/"
		if len(args) == 2 {
			pref = ensureSlash(args[1])
		}
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		cfg.Backup.S3Bucket = bucket
		cfg.Backup.S3Prefix = pref
		cfg.UpdatedAt = time.Now()
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("✅ Backup target set to s3://%s/%s\n", bucket, pref)
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create encrypted backup and upload to S3",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		if cfg.Backup.S3Bucket == "" {
			return fmt.Errorf("backup not configured. Run 'ark backup configure <bucket> [prefix]'")
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, cfg.MasterKey)
		if err != nil {
			return err
		}
		defer db.Close()

		// Create DB backup bytes
		data, err := db.Backup()
		if err != nil {
			return err
		}
		// Encrypt client-side using master key
		enc, err := crypto.NewEncryptor(cfg.MasterKey)
		if err != nil {
			return err
		}
		blob, err := enc.Encrypt(data)
		if err != nil {
			return err
		}

		// Build S3 client
		prof := profileName
		if prof == "" {
			svc := awsfeat.Service{DB: db}
			prof, _ = svc.GetDefaultProfile()
		}
		if prof == "" {
			return fmt.Errorf("no profile specified or default set")
		}
		s3svc, err := awsfeat.NewS3Service(context.Background(), db, prof)
		if err != nil {
			return err
		}

		// Upload with timestamp name
		key := fmt.Sprintf("%sark-backup-%s.bin", ensureSlash(cfg.Backup.S3Prefix), time.Now().UTC().Format("20060102-150405"))
		_, err = s3svc.S3.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String(cfg.Backup.S3Bucket),
			Key:    aws.String(key),
			Body:   strings.NewReader(hex.EncodeToString(blob)),
		})
		if err != nil {
			return err
		}
		fmt.Printf("✅ Backup uploaded to s3://%s/%s\n", cfg.Backup.S3Bucket, key)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List backups in configured S3 bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		if cfg.Backup.S3Bucket == "" {
			return fmt.Errorf("backup not configured")
		}
		// Build S3 client
		db, err := storage.NewDatabase(cfg.DatabasePath, cfg.MasterKey)
		if err != nil {
			return err
		}
		defer db.Close()
		prof := profileName
		if prof == "" {
			svc := awsfeat.Service{DB: db}
			prof, _ = svc.GetDefaultProfile()
		}
		s3svc, err := awsfeat.NewS3Service(context.Background(), db, prof)
		if err != nil {
			return err
		}
		objs, err := s3svc.ListObjects(context.Background(), cfg.Backup.S3Bucket, ensureSlash(cfg.Backup.S3Prefix))
		if err != nil {
			return err
		}
		if len(objs) == 0 {
			fmt.Println("No backups found.")
			return nil
		}
		for _, o := range objs {
			fmt.Printf("%s\t%d\t%s\n", aws.ToString(o.Key), o.Size, o.LastModified.Format("2006-01-02 15:04:05"))
		}
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore <s3key>",
	Short: "Restore from a backup key in S3",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		if cfg.Backup.S3Bucket == "" {
			return fmt.Errorf("backup not configured")
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, cfg.MasterKey)
		if err != nil {
			return err
		}
		defer db.Close()
		prof := profileName
		if prof == "" {
			svc := awsfeat.Service{DB: db}
			prof, _ = svc.GetDefaultProfile()
		}
		s3svc, err := awsfeat.NewS3Service(context.Background(), db, prof)
		if err != nil {
			return err
		}
		// Download
		tmp := filepath.Join(cfg.ConfigDir, "backup", "restore.tmp")
		if err := s3svc.DownloadFile(context.Background(), cfg.Backup.S3Bucket, key, tmp); err != nil {
			return err
		}
		// Decode hex and decrypt
		hexData, err := os.ReadFile(tmp)
		if err != nil {
			return err
		}
		blob, err := hex.DecodeString(string(hexData))
		if err != nil {
			return err
		}
		enc, err := crypto.NewEncryptor(cfg.MasterKey)
		if err != nil {
			return err
		}
		plain, err := enc.Decrypt(blob)
		if err != nil {
			return err
		}
		if err := db.Restore(plain); err != nil {
			return err
		}
		fmt.Println("✅ Restore complete")
		return nil
	},
}

func init() {
	BackupCmd.AddCommand(configureCmd)
	BackupCmd.AddCommand(createCmd)
	BackupCmd.AddCommand(listCmd)
	BackupCmd.AddCommand(restoreCmd)
	for _, c := range []*cobra.Command{createCmd, listCmd, restoreCmd, configureCmd} {
		c.Flags().StringVarP(&profileName, "profile", "p", "", "AWS profile to use")
	}
}

func ensureSlash(p string) string {
	if p == "" {
		return p
	}
	if strings.HasSuffix(p, "/") {
		return p
	}
	return p + "/"
}
