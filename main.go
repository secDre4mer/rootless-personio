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

package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/jilleJr/rootless-personio/pkg/personio"
)

func main() {
	baseURL, ok := os.LookupEnv("PERSONIO_URL")
	if !ok {
		log.Fatalln("Environment variable PERSONIO_URL must be set.")
	}
	p, err := personio.New(baseURL)
	if err != nil {
		log.Fatalf("Create client: %s", err)
	}

	email, ok := os.LookupEnv("PERSONIO_EMAIL")
	if !ok {
		log.Fatalln("Environment variable PERSONIO_EMAIL must be set.")
	}
	pass, ok := os.LookupEnv("PERSONIO_PASS")
	if !ok {
		log.Fatalln("Environment variable PERSONIO_PASS must be set.")
	}

	secToken, ok := os.LookupEnv("PERSONIO_SEC_TOKEN")
	if ok {
		emailToken, ok := os.LookupEnv("PERSONIO_EMAIL_TOKEN")
		if !ok {
			log.Fatalln("Environment variable PERSONIO_EMAIL_TOKEN must be set when PERSONIO_SEC_TOKEN is set.")
		}
		if err := p.UnlockAndLogin(email, pass, emailToken, secToken); err != nil {
			log.Fatalf("Login with tokens: %s", err)
		}

	} else {
		if err := p.Login(email, pass); err != nil {
			log.Fatalf("Login: %s", err)
		}
	}

	log.Printf("Logged in as employee %d", p.EmployeeID)

	me, err := p.GetCurrentEmployeeData()
	if err != nil {
		log.Fatalf("Get current employee: %s", err)
	}

	b, err := json.MarshalIndent(me, "    ", "  ")
	if err != nil {
		log.Fatalf("Marshal as JSON: %s", err)
	}
	log.Printf("Current employee:\n    %s", b)
}
