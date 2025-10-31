package logs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mbeniwal-imwe/ark/internal/core/config"
	"github.com/mbeniwal-imwe/ark/internal/core/logger"
	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View and manage application logs",
}

var viewCmd = &cobra.Command{
	Use:   "view [feature]",
	Short: "View logs for a specific feature or all features",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		feature := ""
		if len(args) > 0 {
			feature = args[0]
		}

		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		_, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		// Initialize logger
		logConfig := logger.LogConfig{
			Enabled:  true,
			MaxDays:  30,
			MaxSize:  100,
			Compress: true,
			LogDir:   cfgDir + "/logs",
		}
		loggerInstance, err := logger.NewLogger(logConfig)
		if err != nil {
			return err
		}
		defer loggerInstance.Close()

		// Get logs
		limit := 50
		if limitStr, _ := cmd.Flags().GetString("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
		}

		logs, err := loggerInstance.GetLogs(feature, limit)
		if err != nil {
			return err
		}

		if len(logs) == 0 {
			fmt.Println("No logs found.")
			return nil
		}

		// Display logs
		for _, log := range logs {
			timestamp := log.Timestamp.Format("2006-01-02 15:04:05")
			level := log.Level.String()
			feat := log.Feature
			message := log.Message

			// Color coding
			var color string
			switch log.Level {
			case logger.DEBUG:
				color = "\033[36m" // Cyan
			case logger.INFO:
				color = "\033[32m" // Green
			case logger.WARN:
				color = "\033[33m" // Yellow
			case logger.ERROR:
				color = "\033[31m" // Red
			}
			reset := "\033[0m"

			fmt.Printf("%s[%s] %s %s: %s%s\n",
				color, timestamp, level, feat, message, reset)
		}

		return nil
	},
}

var tailCmd = &cobra.Command{
	Use:   "tail [feature]",
	Short: "Tail logs in real-time",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		feature := ""
		if len(args) > 0 {
			feature = args[0]
		}

		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		_, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		// Initialize logger
		logConfig := logger.LogConfig{
			Enabled:  true,
			MaxDays:  30,
			MaxSize:  100,
			Compress: true,
			LogDir:   cfgDir + "/logs",
		}
		loggerInstance, err := logger.NewLogger(logConfig)
		if err != nil {
			return err
		}
		defer loggerInstance.Close()

		fmt.Printf("Tailing logs for feature: %s (Press Ctrl+C to stop)\n",
			func() string {
				if feature == "" {
					return "all"
				}
				return feature
			}())

		// Simple tail implementation - in production, use file watching
		for {
			logs, err := loggerInstance.GetLogs(feature, 10)
			if err != nil {
				return err
			}

			for _, log := range logs {
				if log.Timestamp.After(time.Now().Add(-5 * time.Second)) {
					timestamp := log.Timestamp.Format("15:04:05")
					level := log.Level.String()
					feat := log.Feature
					message := log.Message

					// Color coding
					var color string
					switch log.Level {
					case logger.DEBUG:
						color = "\033[36m"
					case logger.INFO:
						color = "\033[32m"
					case logger.WARN:
						color = "\033[33m"
					case logger.ERROR:
						color = "\033[31m"
					}
					reset := "\033[0m"

					fmt.Printf("%s[%s] %s %s: %s%s\n",
						color, timestamp, level, feat, message, reset)
				}
			}

			time.Sleep(2 * time.Second)
		}
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgDir := cmd.Root().PersistentFlags().Lookup("config-dir").Value.String()
		_, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		// Initialize logger
		logConfig := logger.LogConfig{
			Enabled:  true,
			MaxDays:  30,
			MaxSize:  100,
			Compress: true,
			LogDir:   cfgDir + "/logs",
		}
		loggerInstance, err := logger.NewLogger(logConfig)
		if err != nil {
			return err
		}
		defer loggerInstance.Close()

		// Confirm deletion
		fmt.Print("Are you sure you want to clear all logs? (yes/no): ")
		var confirmation string
		fmt.Scanln(&confirmation)
		if strings.ToLower(confirmation) != "yes" {
			fmt.Println("Log clearing cancelled.")
			return nil
		}

		if err := loggerInstance.ClearLogs(); err != nil {
			return err
		}

		fmt.Println("✅ All logs cleared.")
		return nil
	},
}

func init() {
	LogsCmd.AddCommand(viewCmd)
	LogsCmd.AddCommand(tailCmd)
	LogsCmd.AddCommand(clearCmd)

	viewCmd.Flags().StringP("limit", "l", "50", "Number of log entries to show")
}
