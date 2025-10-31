package lock

import (
	"fmt"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/core/password"
	"github.com/mbeniwal-imwe/ark/internal/features/dirlock"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

var (
	useMaster bool
	hideDir   bool
	passOpt   string
)

var LockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Lock/unlock directories",
}

var addCmd = &cobra.Command{
	Use:   "add <directory>",
	Short: "Lock a directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		// Open DB
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()

		svc := &dirlock.Service{DB: db}
		passwordValue := ""
		if !useMaster {
			if passOpt == "" {
				p, err := password.GetPasswordWithConfirmation("Set directory password: ", "Confirm password: ")
				if err != nil {
					return err
				}
				passwordValue = p
			} else {
				passwordValue = passOpt
			}
		}
		if err := svc.Lock(dir, useMaster, passwordValue, hideDir); err != nil {
			return err
		}
		fmt.Println("✅ Directory locked")
		return nil
	},
}

var unlockCmd = &cobra.Command{
	Use:   "unlock <directory>",
	Short: "Unlock a directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		// Open DB
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()

		svc := &dirlock.Service{DB: db}

		// Determine which password to prompt
		// Optimistically ask for master; if wrong, ask for custom
		master, err := password.GetMasterPassword()
		if err != nil {
			return err
		}
		provided := master
		if err := svc.Unlock(dir, master, provided); err != nil {
			// Try custom
			p, err2 := password.GetMasterPassword()
			if err2 != nil {
				return err
			}
			if err3 := svc.Unlock(dir, master, p); err3 != nil {
				return err
			}
		}
		fmt.Println("🔓 Directory unlocked")
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List locked directories",
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
		svc := &dirlock.Service{DB: db}
		recs, err := svc.List()
		if err != nil {
			return err
		}
		if len(recs) == 0 {
			fmt.Println("No locked directories.")
			return nil
		}
		for _, r := range recs {
			fmt.Printf("%s\tmaster=%v\thidden=%v\tlocked=%v\n", r.Path, r.UseMaster, r.Hidden, r.Encrypted)
		}
		return nil
	},
}

func init() {
	LockCmd.AddCommand(addCmd)
	LockCmd.AddCommand(unlockCmd)
	LockCmd.AddCommand(listCmd)

	addCmd.Flags().BoolVar(&useMaster, "use-master", false, "Use Ark master password")
	addCmd.Flags().BoolVar(&hideDir, "hide", false, "Hide directory (macOS)")
	addCmd.Flags().StringVar(&passOpt, "password", "", "Set a custom password (non-interactive)")
}
