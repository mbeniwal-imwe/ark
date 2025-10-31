package aws

import (
	"context"
	"fmt"
	"os/user"
	"strings"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	awsfeat "github.com/mbeniwal-imwe/ark/internal/features/aws"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "aws",
	Short: "AWS configuration and profile management",
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import profiles from ~/.aws",
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
		svc := awsfeat.Service{DB: db}
		u, _ := user.Current()
		n, err := svc.ImportFromAWSDir(u.HomeDir)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Imported %d profile(s) from ~/.aws\n", n)
		return nil
	},
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List stored AWS profiles",
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
		svc := awsfeat.Service{DB: db}
		list, err := svc.ListProfiles()
		if err != nil {
			return err
		}
		if len(list) == 0 {
			fmt.Println("No profiles found. Use 'ark aws import'.")
			return nil
		}
		for _, p := range list {
			fmt.Printf("%s\t%s\t%s\n", p.Name, p.Region, maskKey(p.AccessKeyID))
		}
		def, _ := svc.GetDefaultProfile()
		if def != "" {
			fmt.Printf("Default: %s\n", def)
		}
		return nil
	},
}

var selectCmd = &cobra.Command{
	Use:   "select <profile>",
	Short: "Set default AWS profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
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
		svc := awsfeat.Service{DB: db}
		if err := svc.SetDefaultProfile(name); err != nil {
			return err
		}
		fmt.Printf("✅ Default profile set to %s\n", name)
		return nil
	},
}

var testCmd = &cobra.Command{
	Use:   "test [profile]",
	Short: "Test connection for a profile",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check prerequisites first
		if err := awsfeat.CheckPrerequisites(); err != nil {
			return err
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
		svc := awsfeat.Service{DB: db}
		prof := ""
		if len(args) == 1 {
			prof = args[0]
		} else {
			prof, _ = svc.GetDefaultProfile()
		}
		if prof == "" {
			return fmt.Errorf("no profile specified or default set")
		}
		out, err := svc.TestConnection(context.Background(), prof)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

var prereqCmd = &cobra.Command{
	Use:   "prereq",
	Short: "Check AWS prerequisites",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := awsfeat.CheckPrerequisites(); err != nil {
			return err
		}
		fmt.Println("✅ All AWS prerequisites are met")
		return awsfeat.TestAWSCLI()
	},
}

func init() {
	Cmd.AddCommand(importCmd)
	Cmd.AddCommand(profilesCmd)
	Cmd.AddCommand(selectCmd)
	Cmd.AddCommand(testCmd)
	Cmd.AddCommand(prereqCmd)
}

func maskKey(k string) string {
	if len(k) <= 4 {
		return k
	}
	return strings.Repeat("*", len(k)-4) + k[len(k)-4:]
}
