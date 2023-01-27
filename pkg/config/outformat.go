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

package config

import (
	"encoding"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/spf13/pflag"
)

type OutFormat string

const (
	OutFormatPretty OutFormat = "pretty"
	OutFormatJSON   OutFormat = "json"
	OutFormatYAML   OutFormat = "yaml"
)

func _() {
	// Ensure the type implements the interfaces
	f := OutFormatJSON
	var _ pflag.Value = &f
	var _ encoding.TextUnmarshaler = &f
	var _ jsonSchemaInterface = f
}

func (f OutFormat) String() string {
	return string(f)
}

func (f *OutFormat) Set(value string) error {
	switch OutFormat(value) {
	case OutFormatPretty:
		*f = OutFormatPretty
	case OutFormatJSON:
		*f = OutFormatJSON
	case OutFormatYAML:
		*f = OutFormatYAML
	default:
		return fmt.Errorf("unknown output format: %q, must be one of: pretty, json, yaml", value)
	}
	return nil
}

func (f *OutFormat) Type() string {
	return "out-format"
}

func (f *OutFormat) UnmarshalText(text []byte) error {
	return f.Set(string(text))
}

func (OutFormat) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:  "string",
		Title: "Output format",
		Enum: []any{
			OutFormatPretty,
			OutFormatJSON,
			OutFormatYAML,
		},
	}
}
