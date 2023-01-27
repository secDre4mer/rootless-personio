// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootFlags = struct {
	config   string
	showHelp bool
}{}

var rootCmd = &cobra.Command{
	Use:   "rootless-personio",
	Short: "Access Personio as employee from the command-line",
	Long: `Access Personio via your own employee credentials,
instead of obtaining admin/root API credentials.`,
	Run: func(cmd *cobra.Command, args []string) {},
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

	rootCmd.PersistentFlags().StringVar(&rootFlags.config, "config", rootFlags.config, "Config file (default is $HOME/.rootless-personio.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.showHelp, "help", "h", false, "Show this help text")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if rootFlags.config != "" {
		// Use config file from the flag.
		viper.SetConfigType("yaml")
		viper.SetConfigFile(rootFlags.config)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".rootless-personio" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".rootless-personio.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if rootFlags.config == "" &&
			(os.IsNotExist(err) ||
				errors.As(err, &viper.ConfigFileNotFoundError{})) {
			return
		}
		cobra.CheckErr(fmt.Errorf("read config: %w", err))
	}
	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
}
