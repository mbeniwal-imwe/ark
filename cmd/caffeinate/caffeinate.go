package caffeinate

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/features/caffeinate"
	"github.com/spf13/cobra"
)

var (
	interval int
	mode     string
)

// CaffeinateCmd root group
var CaffeinateCmd = &cobra.Command{
	Use:   "caffeinate",
	Short: "Keep the device awake by periodic activity",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start caffeinate background process",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, _ := cmd.Root().PersistentFlags().GetString("config-dir")
		r := &caffeinate.Runner{ConfigDir: cfgDir, Interval: secondsToDuration(interval), Mode: caffeinate.Mode(mode)}
		if r.Mode == "" {
			r.Mode = caffeinate.ModeWiggle
		}
		if err := r.Start(); err != nil {
			return err
		}
		fmt.Println("✅ Caffeinate started")
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop caffeinate background process",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, _ := cmd.Root().PersistentFlags().GetString("config-dir")
		r := &caffeinate.Runner{ConfigDir: cfgDir}
		if err := r.Stop(); err != nil {
			return err
		}
		fmt.Println("🛑 Caffeinate stopped")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show caffeinate status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir, _ := cmd.Root().PersistentFlags().GetString("config-dir")
		r := &caffeinate.Runner{ConfigDir: cfgDir}
		s, err := r.Status()
		if err != nil {
			return err
		}
		fmt.Println(s)
		return nil
	},
}

// Internal run loop command (not for users)
var internalRunCmd = &cobra.Command{
	Use:    "_run",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if interval <= 0 {
			interval = 30
		}
		return caffeinate.RunLoop(interval, caffeinate.Mode(mode))
	},
}

func init() {
	CaffeinateCmd.AddCommand(startCmd)
	CaffeinateCmd.AddCommand(stopCmd)
	CaffeinateCmd.AddCommand(statusCmd)
	CaffeinateCmd.AddCommand(internalRunCmd)

	startCmd.Flags().IntVarP(&interval, "interval", "i", 30, "Interval seconds between actions")
	startCmd.Flags().StringVarP(&mode, "mode", "m", string(caffeinate.ModeWiggle), "Mode: wiggle|caffeinate")

	internalRunCmd.Flags().IntVar(&interval, "interval", 30, "interval seconds")
	internalRunCmd.Flags().StringVar(&mode, "mode", string(caffeinate.ModeWiggle), "mode")
}

func secondsToDuration(s int) time.Duration {
	d, _ := time.ParseDuration(strconv.Itoa(s) + "s")
	return d
}
