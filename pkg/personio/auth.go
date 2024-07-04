// SPDX-FileCopyrightText: 2022 Jonas Riedel
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

package personio

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	csrfTokenErrorRegex = regexp.MustCompile(`REDUX_INITIAL_STATE\.bladeState\.messages\s*=\s*{[^}]*error:\s*"((?:\\"|[^"])*)"`)
)

func (c *Client) UnlockAndLogin(email, pass, emailToken string) error {
	if err := c.UnlockWithToken(emailToken); err != nil {
		return fmt.Errorf("unlock account: %w", err)
	}
	return c.Login(email, pass)
}

func (c *Client) UnlockWithToken(emailToken string) error {
	params := url.Values{}
	params.Set("token", strings.TrimSpace(emailToken))

	req, err := http.NewRequest(http.MethodPost, "/login/token-auth", strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	resp, err := c.RawForm(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	if strings.HasSuffix(resp.Request.URL.Path, "/login/token-auth") {
		errorMatch := csrfTokenErrorRegex.FindSubmatch(body)
		if errorMatch != nil {
			return fmt.Errorf("error from page: %s", errorMatch[1])
		}
		return errors.New("did not unlock account, and found no error on page")
	}
	return nil
}

func (c *Client) Login(email, pass string) error {
	params := url.Values{}
	params.Set("email", email)
	params.Set("password", pass)

	req, err := http.NewRequest(http.MethodPost, "/login/index", strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	resp, err := c.RawForm(req)
	if err != nil {
		return err
	}

	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return fmt.Errorf("parse base URL: %w", err)
	}

	if resp.Request.URL.Host != baseURL.Host {
		return fmt.Errorf("%w: want host %q, got %q", ErrUnexpectedRedirect, baseURL.Host, resp.Request.URL.Host)
	}

	if strings.HasSuffix(resp.Request.URL.Path, "/login/token-auth") {
		return ErrUnlockRequired
	}

	if strings.TrimPrefix(resp.Request.URL.Path, "/") != "" {
		return fmt.Errorf("%w: want path \"/\", got %q", ErrUnexpectedRedirect, resp.Request.URL.Path)
	}

	userActivity, err := c.getUserActivity()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEmployeeIDNotFound, err)
	}
	c.EmployeeID = userActivity.User.ID
	return nil
}

// getUserActivity seems to get info about the currently logged in user.
//
// Don't know for certain what this endpoint is, so keeping the function as
// private in the meantime.
func (c *Client) getUserActivity() (*navigationContext, error) {
	req, err := http.NewRequest(http.MethodGet, "/api/v1/navigation/context", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.RawJSON(req)
	if err != nil {
		return nil, err
	}
	return ParseResponseJSON[*navigationContext](resp)
}

type navigationContext struct {
	User navigationContextUser `json:"user"`
}

type navigationContextUser struct {
	ID                int    // ex: 123
	Type              string // ex: "employee"
	IsAdmin           bool   // ex: false
	FullName          string // ex: "My Name"
	Position          string // ex: "DevOps Engineer"
	ProfilePictureURL string // ex: "/image-service/v1/images/1234/medium/186027875a90b993ea726ee9e7fbe79f7219d9b9.png"
	Impersonated      bool   // ex: false
	ImpersonatorId    any    // ex: null
	Context           string // ex: "client"
	PreferredName     string // ex: "My Name"
	FirstName         string // ex: "My"
	LastName          string // ex: "Name"
	Email             string // ex: "my.name@example.com"
}
