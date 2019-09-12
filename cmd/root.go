package cmd

import (
	"github.com/Masterminds/semver"
	"github.com/devspace-cloud/devspace/cmd/add"
	"github.com/devspace-cloud/devspace/cmd/cleanup"
	"github.com/devspace-cloud/devspace/cmd/connect"
	"github.com/devspace-cloud/devspace/cmd/create"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/cmd/list"
	"github.com/devspace-cloud/devspace/cmd/remove"
	"github.com/devspace-cloud/devspace/cmd/reset"
	"github.com/devspace-cloud/devspace/cmd/set"
	"github.com/devspace-cloud/devspace/cmd/status"
	"github.com/devspace-cloud/devspace/cmd/update"
	"github.com/devspace-cloud/devspace/cmd/use"
	"github.com/devspace-cloud/devspace/pkg/devspace/upgrade"
	"github.com/devspace-cloud/devspace/pkg/util/analytics/cloudanalytics"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "devspace",
	SilenceUsage:  true,
	SilenceErrors: true,
	Short:         "Welcome to the DevSpace!",
	PersistentPreRun: func(cobraCmd *cobra.Command, args []string) {
		if globalFlags.Silent {
			log.GetInstance().SetLevel(logrus.FatalLevel)
		}

		// Get version of current binary
		version := upgrade.GetVersion()

		if version != "" {
			latestStableVersion, err := upgrade.CheckForNewerVersion()
			if err == nil { // Check versions only if newest version could be determined without errors
				semverVersion, err := semver.NewVersion(version)
				if err == nil { // Only compare version if version can be parsed
					semverLatestStableVersion, err := semver.NewVersion(latestStableVersion)
					if err == nil { // Only compare version if latestStableVersion can be parsed
						// If latestStableVersion > version
						if semverLatestStableVersion.Compare(semverVersion) == 1 {
							log.Warnf("There is a newer version of DevSpace: v%s. Run `devspace upgrade` to upgrade to the newest version.\n", latestStableVersion)
						}
					}
				}
			}
		}
	},
	Long: `DevSpace accelerates developing, deploying and debugging applications with Docker and Kubernetes. Get started by running the init command in one of your projects:

	devspace init`,
}

var globalFlags *flags.GlobalFlags

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Set version for --version flag
	rootCmd.Version = upgrade.GetVersion()

	// Report any panics
	defer cloudanalytics.ReportPanics()

	// Execute command
	err := rootCmd.Execute()
	cloudanalytics.SendCommandEvent(err)
	if err != nil {
		if globalFlags.Debug {
			log.Fatalf("%+v", err)
		} else {
			log.Fatal(err)
		}
	}
}

func init() {
	persistentFlags := rootCmd.PersistentFlags()
	globalFlags = flags.SetGlobalFlags(persistentFlags)

	// Add sub commands
	rootCmd.AddCommand(add.NewAddCmd(globalFlags))
	rootCmd.AddCommand(cleanup.NewCleanupCmd(globalFlags))
	rootCmd.AddCommand(connect.NewConnectCmd())
	rootCmd.AddCommand(create.NewCreateCmd())
	rootCmd.AddCommand(list.NewListCmd(globalFlags))
	rootCmd.AddCommand(remove.NewRemoveCmd())
	rootCmd.AddCommand(reset.NewResetCmd())
	rootCmd.AddCommand(set.NewSetCmd())
	rootCmd.AddCommand(status.NewStatusCmd())
	rootCmd.AddCommand(use.NewUseCmd())
	rootCmd.AddCommand(update.NewUpdateCmd(globalFlags))

	// Add main commands
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewDevCmd(globalFlags))
	rootCmd.AddCommand(NewBuildCmd(globalFlags))
	rootCmd.AddCommand(NewSyncCmd(globalFlags))
	rootCmd.AddCommand(NewPurgeCmd(globalFlags))
	rootCmd.AddCommand(NewUpgradeCmd())
	rootCmd.AddCommand(NewDeployCmd(globalFlags))
	rootCmd.AddCommand(NewEnterCmd(globalFlags))
	rootCmd.AddCommand(NewLoginCmd())
	rootCmd.AddCommand(NewAnalyzeCmd(globalFlags))
	rootCmd.AddCommand(NewLogsCmd(globalFlags))
	rootCmd.AddCommand(NewOpenCmd(globalFlags))
	rootCmd.AddCommand(NewUICmd())
	rootCmd.AddCommand(NewRunCmd(globalFlags))
	rootCmd.AddCommand(NewAttachCmd(globalFlags))

	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Panic(err)
		}

		// Search config in home directory with name ".devspace" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".devspace")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}
