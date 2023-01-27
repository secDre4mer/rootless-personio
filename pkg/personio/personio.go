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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WorkingTimes []struct {
	ID         string      `json:"id"`
	EmployeeID int         `json:"employee_id"`
	Start      time.Time   `json:"start"`
	End        time.Time   `json:"end"`
	ActivityID interface{} `json:"activity_id"`
	Comment    string      `json:"comment"`
	ProjectID  interface{} `json:"project_id"`
}

type Client struct {
	BaseURL    string
	http       *http.Client
	EmployeeID int
}

func New(baseURL string) (*Client, error) {
	normalURL, err := NormalizeBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}
	return &Client{
		http:    &http.Client{Jar: jar},
		BaseURL: normalURL,
	}, nil
}

func NormalizeBaseURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimSuffix(u.Path, "/")
	return u.String(), nil
}

var employeeIDRegex = regexp.MustCompile(`window.EMPLOYEE\s*=\s*{\s*id:\s*(\d+),`)
var loginTokenRegex = regexp.MustCompile(`name="_token"[^>]*value="([^"]*)"`)
var loginTokenErrorRegex = regexp.MustCompile(`REDUX_INITIAL_STATE\.bladeState\.messages\s*=\s*{[^}]*error:\s*"((?:\\"|[^"])*)"`)

var (
	ErrEmployeeIDNotFound = errors.New("employee ID not found")
	UserAgent             = "Rootless-Personio-bot/0.1 (+https://github.com/jilleJr/rootless-personio)"
)

func (c *Client) LoginWithToken(email, pass, emailToken, securityToken string) error {
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

	return c.Login(email, pass)
}

type LoginTokenError struct {
	SecurityToken string
	Response      *http.Response
}

func (e LoginTokenError) Error() string {
	return fmt.Sprintf("token sent to email inbox, use with security token: %s", e.SecurityToken)
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
				return LoginTokenError{
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

func (c *Client) SetWorkingTimes(from, to time.Time) error {
	path := c.BaseURL + "/api/v1/attendances/periods"

	payload := []struct {
		ID         string      `json:"id"`
		EmployeeID int         `json:"employee_id"`
		Start      string      `json:"start"`
		End        string      `json:"end"`
		ActivityID interface{} `json:"activity_id"`
		Comment    string      `json:"comment"`
		ProjectID  interface{} `json:"project_id"`
	}{
		{
			ID:         uuid.New().String(),
			EmployeeID: c.EmployeeID,
			Start:      from.Format("2006-01-02T15:04:05Z"),
			End:        to.Format("2006-01-02T15:04:05Z"),
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode body: %w", err)
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	results, err := DoRequest[PersonioPeriodsResult](c.http, req)
	if err != nil {
		return fmt.Errorf("HTTP request: %w", err)
	}

	log.Printf("Got %d results", len(*results))
	return nil
}

type PersonioPeriodsResult []struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		LegacyID       int         `json:"legacy_id"`
		LegacyStatus   string      `json:"legacy_status"`
		Start          time.Time   `json:"start"`
		End            time.Time   `json:"end"`
		Comment        string      `json:"comment"`
		LegacyBreakMin int         `json:"legacy_break_min"`
		Origin         string      `json:"origin"`
		CreatedAt      time.Time   `json:"created_at"`
		UpdatedAt      time.Time   `json:"updated_at"`
		DeletedAt      interface{} `json:"deleted_at"`
	} `json:"attributes"`
	Relationships struct {
		Project struct {
			Data struct {
				ID interface{} `json:"id"`
			} `json:"data"`
		} `json:"project"`
		Employee struct {
			Data struct {
				ID int `json:"id"`
			} `json:"data"`
		} `json:"employee"`
		Company struct {
			Data struct {
				ID int `json:"id"`
			} `json:"data"`
		} `json:"company"`
		AttendanceDay struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"attendance_day"`
		CreatedBy struct {
			Data struct {
				ID int `json:"id"`
			} `json:"data"`
		} `json:"created_by"`
		UpdatedBy struct {
			Data struct {
				ID int `json:"id"`
			} `json:"data"`
		} `json:"updated_by"`
	} `json:"relationships"`
}

type PersonioPeriods struct {
	Success bool `json:"success"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`

	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Start           time.Time `json:"start"`
			End             time.Time `json:"end"`
			LegacyBreakMin  int       `json:"legacy_break_min"`
			Comment         string    `json:"comment"`
			PeriodType      string    `json:"period_type"`
			CreatedAt       time.Time `json:"created_at"`
			UpdatedAt       time.Time `json:"updated_at"`
			EmployeeID      int       `json:"employee_id"`
			CreatedBy       int       `json:"created_by"`
			AttendanceDayID string    `json:"attendance_day_id"`
		} `json:"attributes"`
	} `json:"data"`
}

func (c *Client) GetWorkingTimes(from, to time.Time) (*PersonioPeriods, error) {
	path := c.BaseURL + "api/v1/attendances/periods"

	req, _ := http.NewRequest("GET", path, nil)
	req.Header.Set("accept", "application/json")
	//req.Header.Set("Accept", "application/json, text/plain, */*")

	//?filter[startDate]=2022-01-31&filter[endDate]=2022-03-06&filter[employee]=991824
	q := req.URL.Query()
	q.Add("filter[startDate]", from.Format("2006-01-02"))
	q.Add("filter[endDate]", to.Format("2006-01-02"))
	q.Add("filter[employee]", fmt.Sprintf("%d", c.EmployeeID))
	req.URL.RawQuery = q.Encode()

	response, err := c.http.Do(req)
	if err != nil {
		log.Printf("cannot get workingtimes %v\n", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("Received %d response code %s", response.StatusCode, path)
	}

	dataRes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("cannot read body %v\n", err)
	}
	//log.Println(string(dataRes))
	var res PersonioPeriods
	json.Unmarshal(dataRes, &res)
	for k := range res.Data {
		res.Data[k].Attributes.Start = res.Data[k].Attributes.Start.Local()
		res.Data[k].Attributes.End = res.Data[k].Attributes.End.Local()
		res.Data[k].Attributes.CreatedAt = res.Data[k].Attributes.CreatedAt.Local()
		res.Data[k].Attributes.UpdatedAt = res.Data[k].Attributes.UpdatedAt.Local()
	}
	//pretty.Println(res)
	if !res.Success {
		return nil, Error{
			Code:     res.Error.Code,
			Message:  res.Error.Message,
			Response: response,
		}
	}
	return &res, nil
}

func DoRequest[M any](client *http.Client, req *http.Request) (*M, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("cannot get workingtimes %v\n", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("parse Content-Type header: %w", err)
	}
	if mediaType != "application/json" {
		return nil, fmt.Errorf("expected JSON response, but got %q", mediaType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var typedBody struct {
		Success bool `json:"success"`
		Error   struct {
			Code      int                 `json:"code"`
			Message   string              `json:"message"`
			ErrorData map[string][]string `json:"error_data"`
		} `json:"error"`
		Data *M `json:"data"`
	}
	if err := json.Unmarshal(body, &typedBody); err != nil {
		return nil, fmt.Errorf("parse body: %w", err)
	}

	if !typedBody.Success {
		return nil, Error{
			Code:      typedBody.Error.Code,
			Message:   typedBody.Error.Message,
			ErrorData: typedBody.Error.ErrorData,
			Response:  resp,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non-2xx status code: %s", resp.Status)
	}

	return typedBody.Data, nil
}

type Error struct {
	Code      int
	Message   string
	ErrorData map[string][]string
	Response  *http.Response
}

func (e Error) Error() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s (code %d)", e.Message, e.Code)
	for _, errs := range e.ErrorData {
		for _, err := range errs {
			sb.WriteByte(' ')
			sb.WriteString(err)
		}
	}
	return sb.String()
}
