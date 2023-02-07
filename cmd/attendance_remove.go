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
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var attendanceRemoveFlags = struct {
}{}

var attendanceRemoveCmd = &cobra.Command{
	Use:     "remove <YYYY-MM-DD>",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"rm", "delete", "del"},
	Short:   "Clears attendance periods",
	Long: `Clears (deletes) attendance periods for a specific day.

Provide the date in format YYYY-MM-DD, e.g 2023-01-23 for Jan 1, 2023.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		date, err := time.Parse(time.DateOnly, args[0])

		client, err := newLoggedInClient()
		if err != nil {
			return err
		}

		err = client.DeleteAttendance(date)
		if err != nil {
			return err
		}
		log.Info().
			Str("day", date.Format(time.DateOnly)).
			Msg("Successfully deleted attendance periods for day.")

		return printOutputJSONOrYAML(map[string]any{
			"date": date.Format(time.DateOnly),
		})
	},
}

func init() {
	attendanceCmd.AddCommand(attendanceRemoveCmd)
}
