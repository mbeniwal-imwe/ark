package ec2

import (
	"context"
	"fmt"
	"os/exec"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/mbeniwal-imwe/ark/internal/core/config"
	awsfeat "github.com/mbeniwal-imwe/ark/internal/features/aws"
	"github.com/mbeniwal-imwe/ark/internal/storage"
	"github.com/spf13/cobra"
)

var (
	profileName string
	sshKeyPath  string
	sshUser     string
)

var EC2Cmd = &cobra.Command{
	Use:   "ec2",
	Short: "Manage EC2 instances",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List EC2 instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		// Get profile
		masterKey, err := cfg.GetMasterKey()
		if err != nil {
			return err
		}
		db, err := storage.NewDatabase(cfg.DatabasePath, masterKey)
		if err != nil {
			return err
		}
		defer db.Close()

		profile := profileName
		if profile == "" {
			svc := awsfeat.Service{DB: db}
			profile, _ = svc.GetDefaultProfile()
		}
		if profile == "" {
			return fmt.Errorf("no profile specified or default set")
		}

		// List instances
		ec2Svc, err := awsfeat.NewEC2Service(context.Background(), db, profile)
		if err != nil {
			return err
		}

		instances, err := ec2Svc.ListInstances(context.Background())
		if err != nil {
			return err
		}

		if len(instances) == 0 {
			fmt.Println("No instances found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "INSTANCE ID\tSTATE\tTYPE\tPUBLIC IP\tPRIVATE IP\tNAME")
		for _, inst := range instances {
			name := "N/A"
			for _, tag := range inst.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				aws.ToString(inst.InstanceId),
				inst.State.Name,
				inst.InstanceType,
				getString(inst.PublicIpAddress),
				getString(inst.PrivateIpAddress),
				name,
			)
		}
		return w.Flush()
	},
}

var registerCmd = &cobra.Command{
	Use:   "register <name> <instance-id>",
	Short: "Register an EC2 instance with a custom name",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		instanceID := args[1]

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

		profile := profileName
		if profile == "" {
			svc := awsfeat.Service{DB: db}
			profile, _ = svc.GetDefaultProfile()
		}
		if profile == "" {
			return fmt.Errorf("no profile specified or default set")
		}

		ec2Svc, err := awsfeat.NewEC2Service(context.Background(), db, profile)
		if err != nil {
			return err
		}

		if err := ec2Svc.RegisterInstance(context.Background(), name, instanceID, sshKeyPath, sshUser); err != nil {
			return err
		}

		fmt.Printf("✅ Instance registered as '%s'\n", name)
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start <name|instance-id>",
	Short: "Start an EC2 instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]

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

		profile := profileName
		if profile == "" {
			svc := awsfeat.Service{DB: db}
			profile, _ = svc.GetDefaultProfile()
		}
		if profile == "" {
			return fmt.Errorf("no profile specified or default set")
		}

		ec2Svc, err := awsfeat.NewEC2Service(context.Background(), db, profile)
		if err != nil {
			return err
		}

		// Try registered name first, then assume it's an instance ID
		var instanceID string
		registered, err := ec2Svc.GetRegisteredInstance(identifier)
		if err == nil {
			instanceID = registered.InstanceID
		} else {
			instanceID = identifier
		}

		if err := ec2Svc.StartInstance(context.Background(), instanceID); err != nil {
			return err
		}

		fmt.Printf("✅ Starting instance %s\n", instanceID)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <name|instance-id>",
	Short: "Stop an EC2 instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]

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

		profile := profileName
		if profile == "" {
			svc := awsfeat.Service{DB: db}
			profile, _ = svc.GetDefaultProfile()
		}
		if profile == "" {
			return fmt.Errorf("no profile specified or default set")
		}

		ec2Svc, err := awsfeat.NewEC2Service(context.Background(), db, profile)
		if err != nil {
			return err
		}

		// Try registered name first, then assume it's an instance ID
		var instanceID string
		registered, err := ec2Svc.GetRegisteredInstance(identifier)
		if err == nil {
			instanceID = registered.InstanceID
		} else {
			instanceID = identifier
		}

		if err := ec2Svc.StopInstance(context.Background(), instanceID); err != nil {
			return err
		}

		fmt.Printf("✅ Stopping instance %s\n", instanceID)
		return nil
	},
}

var metricsCmd = &cobra.Command{
	Use:   "metrics <name|instance-id>",
	Short: "Show metrics for an EC2 instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]

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

		profile := profileName
		if profile == "" {
			svc := awsfeat.Service{DB: db}
			profile, _ = svc.GetDefaultProfile()
		}
		if profile == "" {
			return fmt.Errorf("no profile specified or default set")
		}

		ec2Svc, err := awsfeat.NewEC2Service(context.Background(), db, profile)
		if err != nil {
			return err
		}

		// Try registered name first, then assume it's an instance ID
		var instanceID string
		registered, err := ec2Svc.GetRegisteredInstance(identifier)
		if err == nil {
			instanceID = registered.InstanceID
		} else {
			instanceID = identifier
		}

		metrics, err := ec2Svc.GetInstanceMetrics(context.Background(), instanceID)
		if err != nil {
			return err
		}

		fmt.Println(metrics)
		return nil
	},
}

var sshCmd = &cobra.Command{
	Use:   "ssh <name>",
	Short: "SSH to an EC2 instance",
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

		profile := profileName
		if profile == "" {
			svc := awsfeat.Service{DB: db}
			profile, _ = svc.GetDefaultProfile()
		}
		if profile == "" {
			return fmt.Errorf("no profile specified or default set")
		}

		ec2Svc, err := awsfeat.NewEC2Service(context.Background(), db, profile)
		if err != nil {
			return err
		}

		registered, err := ec2Svc.GetRegisteredInstance(name)
		if err != nil {
			return fmt.Errorf("registered instance not found: %s. Use 'ark ec2 register' first", name)
		}

		sshCmd := awsfeat.BuildSSHCommand(registered)
		if sshCmd == "" {
			return fmt.Errorf("SSH configuration incomplete. Register with --ssh-key flag")
		}

		fmt.Printf("Running: %s\n", sshCmd)
		execCmd := exec.Command("sh", "-c", sshCmd)
		execCmd.Stdin = cmd.InOrStdin()
		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.ErrOrStderr()
		return execCmd.Run()
	},
}

func init() {
	EC2Cmd.AddCommand(listCmd)
	EC2Cmd.AddCommand(registerCmd)
	EC2Cmd.AddCommand(startCmd)
	EC2Cmd.AddCommand(stopCmd)
	EC2Cmd.AddCommand(metricsCmd)
	EC2Cmd.AddCommand(sshCmd)

	// Global flags
	for _, c := range []*cobra.Command{listCmd, registerCmd, startCmd, stopCmd, metricsCmd} {
		c.Flags().StringVarP(&profileName, "profile", "p", "", "AWS profile to use")
	}

	registerCmd.Flags().StringVar(&sshKeyPath, "ssh-key", "", "Path to SSH private key")
	registerCmd.Flags().StringVar(&sshUser, "ssh-user", "ec2-user", "SSH username")
}

func getString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}
