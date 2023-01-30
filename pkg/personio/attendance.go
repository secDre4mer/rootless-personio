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
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ----------------------

func (c *Client) GetWorkingTimes(from, to time.Time) (PersonioPeriods, error) {
	if err := c.assertLoggedIn(); err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", "/api/v1/attendances/periods", nil)
	//req.Header.Set("Accept", "application/json, text/plain, */*")

	//?filter[startDate]=2022-01-31&filter[endDate]=2022-03-06&filter[employee]=991824
	q := req.URL.Query()
	q.Add("filter[startDate]", from.Format("2006-01-02"))
	q.Add("filter[endDate]", to.Format("2006-01-02"))
	q.Add("filter[employee]", fmt.Sprintf("%d", c.EmployeeID))
	req.URL.RawQuery = q.Encode()

	resp, err := c.RawJSON(req)
	if err != nil {
		return nil, err
	}
	res, err := ParseResponseJSON[PersonioPeriods](resp)
	if err != nil {
		return nil, err
	}

	for k := range res {
		res[k].Attributes.Start = res[k].Attributes.Start.Local()
		res[k].Attributes.End = res[k].Attributes.End.Local()
		res[k].Attributes.CreatedAt = res[k].Attributes.CreatedAt.Local()
		res[k].Attributes.UpdatedAt = res[k].Attributes.UpdatedAt.Local()
	}
	return res, nil
}

type WorkingTimes []struct {
	ID         string      `json:"id"`
	EmployeeID int         `json:"employee_id"`
	Start      time.Time   `json:"start"`
	End        time.Time   `json:"end"`
	ActivityID interface{} `json:"activity_id"`
	Comment    string      `json:"comment"`
	ProjectID  interface{} `json:"project_id"`
}

func (c *Client) SetWorkingTimes(from, to time.Time) error {
	if err := c.assertLoggedIn(); err != nil {
		return err
	}

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

	req, err := http.NewRequest("POST", "/api/v1/attendances/periods", body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.RawJSON(req)
	if err != nil {
		return err
	}
	results, err := ParseResponseJSON[*PersonioPeriodsResult](resp)
	if err != nil {
		return err
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

type PersonioPeriods []struct {
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
}
