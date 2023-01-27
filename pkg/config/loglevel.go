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

package config

import (
	"encoding"

	"github.com/invopop/jsonschema"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

// LogLevel is an enum of different log levels / severities.
//
// This is just a wrapper around the [zerolog.Level] type,
// where all the different enum values are defined.
type LogLevel zerolog.Level

// LogLevelDefault is the default log format.
// Used in the [LogLevel.JSONSchema] method.
var LogLevelDefault = LogLevel(zerolog.WarnLevel)

func _() {
	// Ensure the type implements the interfaces
	l := LogLevel(zerolog.DebugLevel)
	var _ pflag.Value = &l
	var _ encoding.TextUnmarshaler = &l
	var _ jsonSchemaInterface = l
}

// String implements [fmt.Stringer] and [pflag.Value].
//
// Used by cobra when showing the default value of a flag.
func (l LogLevel) String() string {
	return zerolog.Level(l).String()
}

// Set implements [pflag.Value].
//
// Used by cobra when setting the new value for a flag.
func (l *LogLevel) Set(value string) error {
	lvl, err := zerolog.ParseLevel(value)
	if err != nil {
		return err
	}
	*l = LogLevel(lvl)
	return nil
}

// Type implements [pflag.Value].
//
// Used by cobra when rendering the list of flags and their types.
func (l *LogLevel) Type() string {
	return "log-level"
}

// MarshalText implements [encoding.TextMarshaler].
//
// Used when printing YAML config files.
func (l LogLevel) MarshalText() ([]byte, error) {
	return zerolog.Level(l).MarshalText()
}

// UnmarshalText implements [encoding.TextUnmarshaler].
//
// Used when parsing YAML config files.
func (l *LogLevel) UnmarshalText(text []byte) error {
	return l.Set(string(text))
}

// JSONSchema returns the custom JSON schema definition for this type.
func (LogLevel) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Logging level",
		Enum: []any{
			LogLevel(zerolog.TraceLevel),
			LogLevel(zerolog.DebugLevel),
			LogLevel(zerolog.InfoLevel),
			LogLevel(zerolog.WarnLevel),
			LogLevel(zerolog.ErrorLevel),
			LogLevel(zerolog.FatalLevel),
			LogLevel(zerolog.PanicLevel),
			LogLevel(zerolog.Disabled),
		},
		Default: LogLevelDefault,
	}
}
