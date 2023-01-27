// SPDX-FileCopyrightText: 2023 Kalle Fagerberg
// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jilleJr/rootless-personio/pkg/config"
	"github.com/jilleJr/rootless-personio/pkg/util"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg config.Config
var cfgFileFlag string

var rootFlags = struct {
	config   string
	showHelp bool
	verbose  int
	quiet    bool
}{}

var rootCmd = &cobra.Command{
	Use:   "rootless-personio",
	Short: "Access Personio as employee from the command-line",
	Long: `Access Personio via your own employee credentials,
instead of obtaining admin/root API credentials.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(defaultConfig config.Config) {
	cfg = defaultConfig
	initLogger() // set up logging first using default config

	rootCmd.PersistentFlags().String("url", "", "Base URL used to access Personio")
	rootCmd.PersistentFlags().String("auth.email", "", "Email used when logging in")
	// Using pflag.Var here instead of pflag.String to get enum validation.
	rootCmd.PersistentFlags().VarP(&cfg.Output, "output", "o", "Sets the output format")
	rootCmd.PersistentFlags().Var(&cfg.Log.Level, "log.level", "Sets the logging level")
	rootCmd.PersistentFlags().Var(&cfg.Log.Format, "log.format", "Sets the logging format")
	viper.BindPFlags(rootCmd.PersistentFlags())

	err := rootCmd.Execute()
	if err != nil {
		log.Error().Msgf("Failed: %s", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&rootFlags.config, "config", rootFlags.config, "Config file (default is $HOME/.rootless-personio.yaml)")
	rootCmd.PersistentFlags().BoolP("help", "h", false, "Show this help text")
	rootCmd.PersistentFlags().CountVarP(&rootFlags.verbose, "verbose", "v", `Shows verbose logging (-v=info, -vv=debug, -vvv=trace)`)
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.quiet, "quiet", "q", false, `Disables logging (same as "--log.level disabled")`)
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PERSONIO") // implicit underscore delimiter
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("personio")
	viper.SetConfigType("yaml")

	files := []string{"/etc/rootless-personio/personio.yaml"}

	if homePath, err := os.UserHomeDir(); err == nil {
		files = append(files, filepath.Join(homePath, ".personio.yaml"))
	}
	if cfgPath, err := os.UserConfigDir(); err == nil {
		files = append(files, filepath.Join(cfgPath, "personio.yaml"))
	}

	files = append(files, ".personio.yaml")

	if cfgFileFlag != "" {
		files = append(files, cfgFileFlag)
	}

	filesLoaded, err := mergeInConfigFiles(files)
	if err != nil {
		log.Error().Msgf("Failed decoding config file:\n%s", err)
		os.Exit(1)
	}

	// Set up logger last time, now that we've read in the new config
	initLogger()

	for _, file := range filesLoaded {
		log.Debug().
			Str("file", util.PrettyPath(file)).
			Msg("Loaded configuration.")
	}
}

func mergeInConfigFiles(files []string) ([]string, error) {
	var filesLoaded []string

	for _, file := range files {
		viper.SetConfigFile(file)
		if err := viper.MergeInConfig(); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		filesLoaded = append(filesLoaded, file)
	}

	if err := viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(), // default hook
		mapstructure.StringToSliceHookFunc(","),     // default hook
	))); err != nil {
		return nil, fmt.Errorf("unmarshalling config: %w", err)
	}

	return filesLoaded, nil
}

func initLogger() {
	overrideLoggerSettings()

	if cfg.Log.Format == config.LogFormatJSON {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "Jan-02 15:04",
		})
	}
	log.Logger = log.Level(zerolog.Level(cfg.Log.Level))
}

func overrideLoggerSettings() {
	switch rootFlags.verbose {
	case 0:
		break
	case 1:
		cfg.Log.Level = config.LogLevel(zerolog.InfoLevel)
	case 2:
		cfg.Log.Level = config.LogLevel(zerolog.DebugLevel)
	default:
		cfg.Log.Level = config.LogLevel(zerolog.TraceLevel)
	}

	if rootFlags.quiet {
		cfg.Log.Level = config.LogLevel(zerolog.Disabled)
	}
}
