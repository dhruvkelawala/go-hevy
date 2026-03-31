package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/dhruvkelawala/go-hevy/internal/api"
	appconfig "github.com/dhruvkelawala/go-hevy/internal/config"
	"github.com/spf13/cobra"
)

var version = "dev"

type outputMode string

const (
	outputTable   outputMode = "table"
	outputJSON    outputMode = "json"
	outputCompact outputMode = "compact"
)

type appContext struct {
	config     *appconfig.Config
	outputMode outputMode
	page       int
	pageSize   int
}

var app appContext

var rootCmd = &cobra.Command{
	Use:   "hevy",
	Short: "CLI for the Hevy workout API",
	Long:  "go-hevy is a fast, scriptable CLI for workouts, routines, exercises, and user data from Hevy.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := appconfig.Load()
		if err != nil {
			return err
		}
		app.config = cfg
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output JSON")
	rootCmd.PersistentFlags().Bool("compact", false, "Output compact one-line summaries")
	rootCmd.PersistentFlags().IntVar(&app.page, "page", 1, "Page number for list commands")
	rootCmd.PersistentFlags().IntVar(&app.pageSize, "page-size", 5, "Page size for list commands")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := appconfig.Load()
		if err != nil {
			return err
		}
		app.config = cfg

		jsonFlag, err := cmd.Flags().GetBool("json")
		if err != nil {
			return err
		}
		compactFlag, err := cmd.Flags().GetBool("compact")
		if err != nil {
			return err
		}
		if jsonFlag && compactFlag {
			return fmt.Errorf("--json and --compact cannot be used together")
		}
		app.outputMode = outputTable
		if jsonFlag {
			app.outputMode = outputJSON
		}
		if compactFlag {
			app.outputMode = outputCompact
		}
		return nil
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(workoutsCmd)
	rootCmd.AddCommand(workoutCmd)
	rootCmd.AddCommand(countCmd)
	rootCmd.AddCommand(routinesCmd)
	rootCmd.AddCommand(routineCmd)
	rootCmd.AddCommand(exercisesCmd)
	rootCmd.AddCommand(exerciseCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(meCmd)
}

func clientFromConfig() (*api.Client, error) {
	if app.config == nil {
		return nil, fmt.Errorf("config is not loaded")
	}
	return api.NewClient(app.config.EffectiveAPIKey())
}

func contextForCommand(cmd *cobra.Command) context.Context {
	return cmd.Context()
}

func printLine(format string, values ...any) {
	fmt.Fprintf(os.Stdout, format+"\n", values...)
}
