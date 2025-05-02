// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
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

// Package console contains code to pretty-print different types to the console.
package console

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/applejag/rootless-personio/pkg/personio"
	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"gopkg.in/typ.v4"
)

var (
	stdout = colorable.NewColorableStdout()
	stderr = colorable.NewColorableStderr()

	calendarMonthColor       = color.New(color.FgHiBlack, color.Italic)
	calendarWeekdayColor     = color.New(color.FgWhite, color.Underline)
	calendarEmptyColor       = color.New(color.Italic)
	calendarAttendedColor    = color.New(color.FgHiYellow)
	calendarAttendedDurColor = color.New(color.FgYellow)
	calendarWeekSumColor     = color.New(color.FgHiBlack, color.Italic)
	calendarAbsenceColor     = color.New(color.FgMagenta, color.Italic)

	usageHeaderColor = color.New(color.FgYellow, color.Underline, color.Italic)
	usageHelpColor   = color.New(color.FgHiBlack, color.Italic)
)

func PrintCalendarMonth(month time.Time, cal []personio.Timecard) {
	t := Table{}

	t.SetSpacing("  ")
	t.SetPrefix("  ")

	t.WriteColoredRow(calendarWeekdayColor,
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
		"Sunday")
	switch month.Weekday() {
	case time.Monday:
	case time.Tuesday:
		t.WriteCell("")
	case time.Wednesday:
		t.WriteCell("")
		t.WriteCell("")
	case time.Thursday:
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
	case time.Friday:
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
	case time.Saturday:
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
	case time.Sunday:
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
		t.WriteCell("")
	}

	day := month
	m := day.Month()
	var weekTime time.Duration
	for {
		dayStr := strconv.Itoa(day.Day())
		if calDay, ok := findCalendarDayAttendance(day, cal); ok && !calDay.IsOffDay {
			dur := time.Minute * time.Duration(calDay.TargetHours.ContractualWorkDurationMinutes)
			durStr := FormatDuration(dur)
			t.WriteCellWidth(fmt.Sprintf(
				"%s (%s)",
				calendarAttendedColor.Sprint(dayStr),
				calendarAttendedDurColor.Sprint(durStr),
			), len(dayStr)+len(durStr)+3)
			weekTime += dur
		} else {
			t.WriteCellColor(dayStr, calendarEmptyColor)
		}
		nextDay := day.AddDate(0, 0, 1)
		if nextDay.Month() != m || day.Weekday() == time.Sunday {
			for len(t.pendingRow) < 7 {
				t.WriteCell("")
			}
			t.WriteCellColor(fmt.Sprintf("∑ %s", FormatDuration(weekTime)), calendarWeekSumColor)
			t.CommitRow()
			weekTime = 0
		}
		if nextDay.Month() != m {
			break
		}
		day = nextDay
	}

	w := t.Width()
	monthStr := month.Month().String()
	calendarMonthColor.Printf("%s=== %s ===\n", strings.Repeat(" ", typ.Max(0, w/2-len(monthStr)/2-4)), monthStr)
	t.Fprintln(stdout)
}

// UsageTemplate returns a lightly colored usage template for Cobra.
func UsageTemplate() string {
	var sb strings.Builder
	usageHeaderColor.Fprint(&sb, "Usage:")
	sb.WriteString(`{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

`)
	usageHeaderColor.Fprint(&sb, "Aliases:")
	sb.WriteString(`
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

`)
	usageHeaderColor.Fprint(&sb, "Examples:")
	sb.WriteString(`
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

`)
	usageHeaderColor.Fprint(&sb, "Available Commands:")
	sb.WriteString(`{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

`)
	usageHeaderColor.Fprint(&sb, "Flags:")
	sb.WriteString(`
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

`)
	usageHeaderColor.Fprint(&sb, "Global Flags:")
	sb.WriteString(`
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

`)
	usageHeaderColor.Fprint(&sb, "Additional help topics:")
	sb.WriteString(`{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

`)
	usageHelpColor.Fprint(&sb, `Use "{{.CommandPath}} [command] --help" for more information about a command.`)
	sb.WriteString(`{{end}}`)
	sb.WriteByte('\n')
	return sb.String()
}

func findCalendarDayAttendance(day time.Time, days []personio.Timecard) (personio.Timecard, bool) {
	dayStr := day.Format(time.DateOnly)
	for _, calDay := range days {
		if calDay.Date == dayStr {
			return calDay, true
		}
	}
	return personio.Timecard{}, false
}
