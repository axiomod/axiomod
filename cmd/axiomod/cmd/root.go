/*
Copyright © 2025 Enterprise Go
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Import command packages
	"axiomod/cmd/axiomod/cmd/core"
	"axiomod/cmd/axiomod/cmd/generate"
	"axiomod/cmd/axiomod/cmd/migrate"
	"axiomod/cmd/axiomod/cmd/plugin"
	"axiomod/cmd/axiomod/cmd/validator"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "axiomod",
	Short: "A CLI tool for managing the Axiomod Go Macroservice Framework",
	Long: `Axiomod is a CLI tool designed to streamline the development,
management, and deployment of services built with the Axiomod framework.

It provides commands for:
- Initializing new projects
- Generating code (modules, handlers, services)
- Managing database migrations
- Running validators (architecture, naming, static analysis, etc.)
- Building and deploying services
- Managing plugins
- And more...
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.axiomod.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Add commands from sub-packages
	rootCmd.AddCommand(core.NewInitCmd())
	rootCmd.AddCommand(generate.NewGenerateCmd()) // Parent generate command
	rootCmd.AddCommand(migrate.NewMigrateCmd())   // Parent migrate command
	rootCmd.AddCommand(core.NewConfigCmd())       // Parent config command
	rootCmd.AddCommand(core.NewTestCmd())
	rootCmd.AddCommand(core.NewLintCmd())
	rootCmd.AddCommand(core.NewFmtCmd())
	rootCmd.AddCommand(core.NewBuildCmd())
	rootCmd.AddCommand(core.NewDockerizeCmd())
	rootCmd.AddCommand(core.NewDeployCmd())
	rootCmd.AddCommand(core.NewStatusCmd())
	rootCmd.AddCommand(core.NewLogsCmd())
	rootCmd.AddCommand(core.NewHealthcheckCmd())
	rootCmd.AddCommand(plugin.NewPluginCmd()) // Parent plugin command
	rootCmd.AddCommand(core.NewInteractiveCmd())
	rootCmd.AddCommand(core.NewVersionCmd())
	rootCmd.AddCommand(validator.NewValidatorCmd()) // Parent validator command

	// Note: Subcommands like generate service, migrate create, plugin install
	// are added within their respective parent command packages (generate.go, migrate.go, plugin.go)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".axiomod" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".axiomod")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
