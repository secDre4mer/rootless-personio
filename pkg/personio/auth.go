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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	employeeIDRegex      = regexp.MustCompile(`window.EMPLOYEE\s*=\s*{\s*id:\s*(\d+),`)
	loginTokenRegex      = regexp.MustCompile(`name="_token"[^>]*value="([^"]*)"`)
	loginTokenErrorRegex = regexp.MustCompile(`REDUX_INITIAL_STATE\.bladeState\.messages\s*=\s*{[^}]*error:\s*"((?:\\"|[^"])*)"`)
)

func (c *Client) UnlockAndLogin(email, pass, emailToken, securityToken string) error {
	if err := c.UnlockWithToken(emailToken, securityToken); err != nil {
		return err
	}
	return c.Login(email, pass)
}

func (c *Client) UnlockWithToken(emailToken, securityToken string) error {
	params := url.Values{}
	params.Set("_token", strings.TrimSpace(securityToken))
	params.Set("token", strings.TrimSpace(emailToken))

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/login/token-auth", strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	if strings.HasSuffix(resp.Request.URL.Path, "/login/token-auth") {
		errorMatch := loginTokenErrorRegex.FindSubmatch(body)
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

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/login/index", strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	idMatch := employeeIDRegex.FindSubmatch(body)
	if idMatch == nil {
		if strings.HasSuffix(resp.Request.URL.Path, "/login/token-auth") {
			tokenMatch := loginTokenRegex.FindSubmatch(body)
			if tokenMatch != nil {
				return LockedAccountError{
					SecurityToken: string(tokenMatch[1]),
					Response:      resp,
				}
			}
		}

		return ErrEmployeeIDNotFound
	}

	id, err := strconv.Atoi(string(idMatch[1]))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEmployeeIDNotFound, err)
	}

	c.EmployeeID = id
	return nil
}

type LockedAccountError struct {
	SecurityToken string
	Response      *http.Response
}

func (e LockedAccountError) Error() string {
	return fmt.Sprintf("account locked; token sent to email inbox, use with security token: %s", e.SecurityToken)
}
