package s3cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/mbeniwal-imwe/ark/internal/core/config"
	awsfeat "github.com/mbeniwal-imwe/ark/internal/features/aws"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

var (
	profileName string
)

var S3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Manage S3 buckets and objects",
}

var bucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "List S3 buckets",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()
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
		buckets, err := s3svc.ListBuckets(context.Background())
		if err != nil {
			return err
		}
		if len(buckets) == 0 {
			fmt.Println("No buckets found.")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCREATED")
		for _, b := range buckets {
			fmt.Fprintf(w, "%s\t%s\n", aws.ToString(b.Name), b.CreationDate.Format("2006-01-02 15:04:05"))
		}
		return w.Flush()
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls <bucket> [prefix]",
	Short: "List objects in a bucket/prefix",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]
		prefix := ""
		if len(args) == 2 {
			prefix = args[1]
		}
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()
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
		objs, err := s3svc.ListObjects(context.Background(), bucket, prefix)
		if err != nil {
			return err
		}
		if len(objs) == 0 {
			fmt.Println("No objects.")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tSIZE\tLAST MODIFIED")
		for _, o := range objs {
			fmt.Fprintf(w, "%s\t%d\t%s\n", aws.ToString(o.Key), o.Size, o.LastModified.Format("2006-01-02 15:04:05"))
		}
		return w.Flush()
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload <localPath> <bucket> <key>",
	Short: "Upload a file to S3",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		local := args[0]
		bucket := args[1]
		key := args[2]
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()
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
		if err := s3svc.UploadFile(context.Background(), local, bucket, key); err != nil {
			return err
		}
		fmt.Printf("✅ Uploaded %s to s3://%s/%s\n", filepath.Base(local), bucket, key)
		return nil
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download <bucket> <key> <localPath>",
	Short: "Download an S3 object",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]
		key := args[1]
		local := args[2]
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()
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
		if err := s3svc.DownloadFile(context.Background(), bucket, key, local); err != nil {
			return err
		}
		fmt.Printf("✅ Downloaded s3://%s/%s to %s\n", bucket, key, local)
		return nil
	},
}

func init() {
	S3Cmd.AddCommand(bucketsCmd)
	S3Cmd.AddCommand(lsCmd)
	S3Cmd.AddCommand(uploadCmd)
	S3Cmd.AddCommand(downloadCmd)
	for _, c := range []*cobra.Command{bucketsCmd, lsCmd, uploadCmd, downloadCmd} {
		c.Flags().StringVarP(&profileName, "profile", "p", "", "AWS profile to use")
	}
}
