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
	"strings"
)

var (
	csrfTokenErrorRegex = regexp.MustCompile(`REDUX_INITIAL_STATE\.bladeState\.messages\s*=\s*{[^}]*error:\s*"((?:\\"|[^"])*)"`)
)

func (c *Client) UnlockAndLogin(email, pass, emailToken, csrfToken string) error {
	if err := c.UnlockWithToken(emailToken, csrfToken); err != nil {
		return err
	}
	return c.Login(email, pass)
}

func (c *Client) UnlockWithToken(emailToken, csrfToken string) error {
	params := url.Values{}
	params.Set("_token", strings.TrimSpace(csrfToken))
	params.Set("token", strings.TrimSpace(emailToken))

	req, err := http.NewRequest(http.MethodPost, "/login/token-auth", strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}

	c.csrfToken = csrfToken

	resp, err := c.RawForm(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
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

	if _, err := c.RawForm(req); err != nil {
		return err
	}

	userActivity, err := c.getUserActivity()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEmployeeIDNotFound, err)
	}
	c.EmployeeID = userActivity.Visitor.ID
	return nil
}

// getUserActivity seems to get info about the currently logged in user.
//
// Don't know for certain what this endpoint is, so keeping the function as
// private in the meantime.
func (c *Client) getUserActivity() (*userActivity, error) {
	req, err := http.NewRequest(http.MethodGet, "/user-activity/api/v1/pendo", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := c.RawJSON(req)
	if err != nil {
		return nil, err
	}
	return ParseResponseJSON[*userActivity](resp)
}

type userActivity struct {
	Account userActivityAccount `json:"account"`
	Enabled bool                `json:"enabled"`
	Visitor userActivityVisitor `json:"visitor"`
}

type userActivityVisitor struct {
	EmploymentType      string `json:"employment_type"`        // ex: "internal"
	HasReports          bool   `json:"has_reports"`            // ex: false
	HasSupervisor       bool   `json:"has_supervisor"`         // ex: true
	ID                  int    `json:"id"`                     // ex: 8234095
	Role                string `json:"role"`                   // ex: "Employee"
	Status              string `json:"status"`                 // ex: "active"
	UserCreatedAt       string `json:"user_created_at"`        // ex: "2006-01-02 15:04:05"
	UserIsAccountOwner  bool   `json:"user_is_account_owner"`  // ex: false
	UserIsAdmin         bool   `json:"user_is_admin"`          // ex: false
	UserIsContractOwner bool   `json:"user_is_contract_owner"` // ex: false
	UserLang            string `json:"user_lang"`              // ex: "en"
}

type userActivityAccount struct {
	AccountActiveEmployees     int    `json:"account_active_employees"`     // ex: 123
	AccountCreatedAt           string `json:"account_created_at"`           // ex: "2006-01-02 15:04:05"
	AccountExpiresOn           string `json:"account_expires_on"`           // ex: "2006-01-02"
	AccountHostname            string `json:"account_hostname"`             // ex: "mycompany"
	AccountIsTest              bool   `json:"account_is_test"`              // ex: false
	AccountName                string `json:"account_name"`                 // ex: "My Company GmbH"
	AccountPlanID              int    `json:"account_plan_id"`              // ex: 123
	AccountSubcompaniesEnabled bool   `json:"account_subcompanies_enabled"` // ex: false
	Addons                     string `json:"addons"`                       // ex: "productivity_plus,customization_plus,automation_plus"
	BillingCountry             string `json:"billing_country"`              // ex: "DE"
	EnabledPayrollIntegration  any    `json:"enabled_payroll_integration"`  // ex: null
	ID                         int    `json:"id"`                           // ex: 1234
	IsTrialAccount             bool   `json:"is_trial_account"`             // ex: false
	PlanStatus                 string `json:"plan_status"`                  // ex: "active"
	PlanType                   string `json:"plan_type"`                    // ex: "professional"
	PlanVersion                int    `json:"plan_version"`                 // ex: 5
	TrialConversionDate        any    `json:"trial_conversion_date"`        // ex: null
}

type LockedAccountError struct {
	CSRFToken string
	Response  *http.Response
}

func (e LockedAccountError) Error() string {
	return fmt.Sprintf("account locked; token sent to email inbox, use with CSRF token: %s", e.CSRFToken)
}
