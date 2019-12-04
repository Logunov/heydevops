/*
Copyright © 2019 Ilya V. Logounov <ilya@logounov.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package cmd

import (
	"github.com/Logunov/heydevops/clone"
	"os"
	"strings"

	"github.com/Logunov/heydevops/helpers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log = logrus.New()

	flagConfig             = "config"
	flagToken              = "token"
	flagGitlabURL          = "gitlab-url"
	flagGitlabAPIURL       = "gitlab-api-url"
	flagDryRun             = "dry-run"
	flagExpandBranches     = "expand-branches"
	flagLogLevel           = "log-level"
	flagListOptionsPerPage = "list-options-per-page"

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "heydevops",
		Short: "Hey, DevOps!",
		Long:  "heydevops clones group from GitLab to local directory",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig()
			initLogger()

			var coreConfig = clone.ConfigStruct{
				Logger:             log,
				DryRun:             viper.GetBool(flagDryRun),
				ExpandBranches:     viper.GetBool(flagExpandBranches),
				GitLabURL:          viper.GetString(flagGitlabURL),
				GitLabAPIURL:       viper.GetString(flagGitlabAPIURL),
				Token:              viper.GetString(flagToken),
				ListOptionsPerPage: viper.GetInt(flagListOptionsPerPage),
				Repos: clone.SkipCloneStringsStruct{
					Clone: viper.GetStringSlice("repos.clone"),
					Skip:  viper.GetStringSlice("repos.skip"),
				},
				Branches: clone.BranchesStruct{
					Prefix: viper.GetString("branches.prefix"),
					Suffix: viper.GetString("branches.suffix"),
					Slash:  viper.GetString("branches.slash"),
					SkipCloneStringsStruct: clone.SkipCloneStringsStruct{
						Clone: viper.GetStringSlice("branches.clone"),
						Skip:  viper.GetStringSlice("branches.skip"),
					},
				},
			}
			log.Trace("Core config: ", coreConfig)

			clone.Init(&coreConfig)
			clone.Clone()
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	helpers.SetLogger(log)
	err := rootCmd.Execute()
	helpers.CheckPanic(err)
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.WarnLevel)

	rootCmd.PersistentFlags().StringP(flagConfig, "c", "heydevops.yaml", "config file")
	rootCmd.PersistentFlags().BoolP(flagDryRun, "n", false, "If true, only print without changing anything")
	rootCmd.PersistentFlags().BoolP(flagExpandBranches, "b", false, "If true, branches will be expanded into subdirectories")
	rootCmd.PersistentFlags().StringP(flagGitlabAPIURL, "a", "", "GitLab API address if it is located at non-default path")
	rootCmd.PersistentFlags().StringP(flagGitlabURL, "u", "", "GitLab address")
	rootCmd.PersistentFlags().Int(flagListOptionsPerPage, 10, "For paginated result sets, the number of results to include per page")
	rootCmd.PersistentFlags().StringP(flagLogLevel, "l", "warn", "Level of logging:\n  0 - panic\n  1 - fatal\n  2 - error\n  3 - warn (warning)\n  4 - info\n  5 - debug\n  6 - trace")
	rootCmd.PersistentFlags().StringP(flagToken, "t", "", "GitLab token from http://<gitlab>/profile/personal_access_tokens page")

	var err error

	err = viper.BindPFlag(flagConfig, rootCmd.PersistentFlags().Lookup(flagConfig))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagDryRun, rootCmd.PersistentFlags().Lookup(flagDryRun))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagExpandBranches, rootCmd.PersistentFlags().Lookup(flagExpandBranches))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagGitlabAPIURL, rootCmd.PersistentFlags().Lookup(flagGitlabAPIURL))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagGitlabURL, rootCmd.PersistentFlags().Lookup(flagGitlabURL))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagListOptionsPerPage, rootCmd.PersistentFlags().Lookup(flagListOptionsPerPage))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagLogLevel, rootCmd.PersistentFlags().Lookup(flagLogLevel))
	helpers.CheckDebug(err)

	err = viper.BindPFlag(flagToken, rootCmd.PersistentFlags().Lookup(flagToken))
	helpers.CheckDebug(err)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Use config file from the flag.
	viper.SetConfigFile(viper.GetString("config"))

	// read in environment variables that match
	viper.SetEnvPrefix("HEYDEVOPS")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	helpers.CheckError(err)
	log.Debug("Using config file: ", viper.ConfigFileUsed())
}

func initLogger() {
	logLevelString := viper.GetString(flagLogLevel)
	logLevel, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		log.Error("Failed to parse log level string \"" + logLevelString + "\"\n" + err.Error())
		return
	}
	log.SetLevel(logLevel)
}
