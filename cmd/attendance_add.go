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
	"errors"
	"fmt"
	"time"

	"github.com/applejag/rootless-personio/pkg/config"
	"github.com/applejag/rootless-personio/pkg/personio"
	"github.com/spf13/cobra"
)

var attendanceAddFlags = struct {
	startTime string
}{}

var attendanceAddCmd = &cobra.Command{
	Use:     "add <YYYY-MM-DD> <project> <duration>",
	Short:   "Add attendance periods",
	Long:    `Adds an attendance period.`,
	Example: `add 2023-01-25 "Project X" 4h`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("expected 3 arguments, got %d", len(args))
		}
		date, err := time.Parse(time.DateOnly, args[0])
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		projectName := args[1]
		duration, err := parseDuration(args[2])
		if err != nil {
			return fmt.Errorf("invalid duration format: %w", err)
		}
		if duration <= 0 {
			return errors.New("duration must be positive")
		}
		client, err := newLoggedInClient()
		if err != nil {
			return err
		}

		var projectID *int
		if projectName != "none" {
			projectId, err := client.GetProjectID(projectName)
			if err != nil {
				return fmt.Errorf("failed to get project ID: %w", err)
			}
			projectID = &projectId
		}

		calendar, err := client.GetMyAttendanceCalendar(date, date)
		if err != nil {
			return fmt.Errorf("failed to get attendance calendar: %w", err)
		}
		currentDay := calendar[0]
		var startTime time.Time
		if attendanceAddFlags.startTime != "" {
			startTime, err = parseTime(date, attendanceAddFlags.startTime)
			if err != nil {
				return fmt.Errorf("invalid start time format: %w", err)
			}
		} else if len(currentDay.Periods) == 0 {
			if cfg.StandardStartTime == "" {
				return errors.New("no start time provided and no default start time configured")
			}

			startTime, err = parseTime(date, cfg.StandardStartTime)
			if err != nil {
				return fmt.Errorf("invalid start time format: %w", err)
			}
		} else {
			for _, period := range currentDay.Periods {
				if period.End.After(startTime) {
					startTime = period.End.Time
				}
			}
		}
		endTime := startTime.Add(duration)
		currentDay.Periods = append(currentDay.Periods, personio.Period{
			Start:     personio.PersonioTime{startTime},
			End:       personio.PersonioTime{endTime},
			ProjectID: projectID,
			Type:      personio.PeriodTypeWork,
		})

		err = client.SetAttendance(date, currentDay.Periods)
		if err != nil {
			return fmt.Errorf("failed to set attendance: %w", err)
		}

		if cfg.Output == config.OutFormatPretty {
			projectTimes := map[string]time.Duration{}
			for _, period := range currentDay.Periods {
				var projectName = "<none>"
				if period.ProjectID != nil {
					projectName, err = client.GetProjectName(*period.ProjectID)
					if err != nil {
						return fmt.Errorf("failed to get project name: %w", err)
					}
				}
				projectTimes[projectName] += period.End.Time.Sub(period.Start.Time)
			}
			fmt.Printf("Attendance overview for %s:\n", date.Format(time.DateOnly))
			for project, duration := range projectTimes {
				fmt.Printf("%s: %s\n", project, duration.Truncate(time.Second))
			}
		} else {
			return printOutputJSONOrYAML(currentDay)
		}
		return nil
	},
}

func parseTime(date time.Time, timeString string) (time.Time, error) {
	parsedTime, err := time.Parse("15:04", timeString)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format: %w", err)
	}
	return time.Date(date.Year(), date.Month(), date.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, date.Location()), nil
}

func parseDuration(s string) (time.Duration, error) {
	var hours, minutes int
	_, err := fmt.Sscanf(s, "%d:%d", &hours, &minutes)
	if err == nil {
		return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute, nil
	}
	return time.ParseDuration(s)
}

func init() {
	attendanceCmd.AddCommand(attendanceAddCmd)

	attendanceAddCmd.Flags().StringVarP(&attendanceAddFlags.startTime, "start-time", "s", "", `Start time (as HH:MM)`)
}
