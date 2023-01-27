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
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/jilleJr/rootless-personio/pkg/util"
)

type Config struct {
	BaseURL string `yaml:"baseUrl" jsonschema:"oneof_type=string;null" jsonschema_extras:"format=uri"`
	Auth    Auth
	Output  OutFormat
	Log     Log
}

type Auth struct {
	Email    string `jsonschema:"oneof_type=string;null" jsonschema_extras:"format=email"`
	Password string `jsonschema:"oneof_type=string;null"`
}

type Log struct {
	Format LogFormat
	Level  LogLevel
}

type jsonSchemaInterface interface {
	JSONSchema() *jsonschema.Schema
}

func Schema() *jsonschema.Schema {
	r := new(jsonschema.Reflector)
	r.KeyNamer = util.ToCamelCase
	r.Namer = func(t reflect.Type) string {
		return util.ToCamelCase(t.Name())
	}
	r.RequiredFromJSONSchemaTags = true
	s := r.Reflect(&Config{})
	s.ID = "https://github.com/jilleJr/rootless-personio/raw/main/personio.schema.json"
	return s
}
