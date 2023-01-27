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
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

// LogFormat is an enum of different log formats.
type LogFormat string

// LogFormatDefault is the default log format.
// Used in the [LogFormat.JSONSchema] method.
var LogFormatDefault = LogFormatPretty

// Available [LogFormat] values.
const (
	LogFormatPretty LogFormat = "pretty"
	LogFormatJSON   LogFormat = "json"
)

func _() {
	// Ensure the type implements the interfaces
	f := LogFormatJSON
	var _ pflag.Value = &f
	var _ encoding.TextUnmarshaler = &f
	var _ jsonSchemaInterface = f
}

// String implements [fmt.Stringer] and [pflag.Value].
//
// Used by cobra when showing the default value of a flag.
func (f LogFormat) String() string {
	return string(f)
}

// Set implements [pflag.Value].
//
// Used by cobra when setting the new value for a flag.
func (f *LogFormat) Set(value string) error {
	switch LogFormat(value) {
	case LogFormatPretty:
		*f = LogFormatPretty
	case LogFormatJSON:
		*f = LogFormatJSON
	default:
		return fmt.Errorf("unknown log format: %q, must be one of: pretty, json", value)
	}
	return nil
}

// Type implements [pflag.Value].
//
// Used by cobra when rendering the list of flags and their types.
func (f *LogFormat) Type() string {
	return "log-format"
}

// UnmarshalText implements [encoding.TextUnmarshaler].
//
// Used when parsing YAML config files.
func (f *LogFormat) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

// JSONSchema returns the custom JSON schema definition for this type.
func (LogFormat) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Logging format",
		Enum: []any{
			LogFormatPretty,
			LogFormatJSON,
		},
		Default: LogFormatDefault,
	}
}
