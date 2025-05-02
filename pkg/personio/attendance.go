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
	"net/http"
	"net/url"
	"time"

	"github.com/applejag/rootless-personio/pkg/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (c *Client) GetMyAttendanceCalendar(startDate, endDate time.Time) ([]Timecard, error) {
	return c.GetAttendanceCalendar(c.EmployeeID, startDate, endDate)
}

func (c *Client) GetAttendanceCalendar(employeeID int, startDate, endDate time.Time) ([]Timecard, error) {
	if err := c.assertLoggedIn(); err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Set("start_date", startDate.Format(time.DateOnly))
	queryParams.Set("end_date", endDate.Format(time.DateOnly))

	req, err := http.NewRequest("GET", fmt.Sprintf(
		"/svc/attendance-bff/v1/timesheet/%d?%s",
		employeeID, queryParams.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.RawJSON(req)
	if err != nil {
		return nil, err
	}

	var timesheet TimecardResponse
	if err := json.NewDecoder(resp.Body).Decode(&timesheet); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return timesheet.Timecards, nil
}

func (c *Client) SetAttendance(date time.Time, periods []Period) error {
	if err := c.assertLoggedIn(); err != nil {
		return err
	}

	var requestPeriods []RequestPeriod

	for i := range periods {
		if periods[i].ID == uuid.Nil {
			periods[i].ID = uuid.New()
		}
		periods[i].Start = PersonioTime{periods[i].Start.Truncate(time.Second).UTC()}
		periods[i].End = PersonioTime{periods[i].End.Truncate(time.Second).UTC()}
		if periods[i].Type == "" {
			periods[i].Type = PeriodTypeWork
		}
		requestPeriods = append(requestPeriods, RequestPeriod(periods[i]))
	}

	body, err := json.Marshal(SetAttendanceDayRequest{
		EmployeeID: c.EmployeeID,
		Periods:    requestPeriods,
	})
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(body)

	dayID, err := c.GetOrNewDayUUID(date)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, "/svc/attendance-api/v1/days/"+dayID.String(), bodyReader)
	if err != nil {
		return err
	}

	resp, err := c.RawJSON(req)
	if err != nil {
		return err
	}

	// Currently don't care about the response
	_, err = ParseResponseJSON[any](resp)
	return err
}

// DeleteAttendance will delete a day's attendance.
// Note: this seems to be broken in the Personio API, it returns a 403 error (despite being used by the web UI).
func (c *Client) DeleteAttendance(date time.Time) error {
	if err := c.assertLoggedIn(); err != nil {
		return err
	}

	dayID, err := c.GetOrNewDayUUID(date)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, "/svc/attendance-api/v1/days/"+dayID.String(), nil)
	if err != nil {
		return err
	}

	_, err = c.RawJSON(req)
	return err
}

// GetOrNewDayUUID will either lookup a day's ID (from cache or by querying
// the API), or generate a new ID and store this new ID in cache.
//
// After the remote lookup to the API, the client caches which days in the same
// month that has undefined IDs.
func (c *Client) GetOrNewDayUUID(date time.Time) (uuid.UUID, error) {
	id, err := c.GetDayUUID(date)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get day UUID: %w", err)
	}
	if id != nil {
		return *id, nil
	}
	newID := uuid.New()
	dateString := date.Format(time.DateOnly)
	c.dayIDCache[dateString] = &newID
	log.Debug().Str("day", dateString).Stringer("uuid", newID).
		Msg("Randomized new UUID for day.")
	return newID, nil
}

// GetDayUUID will lookup a day's ID (from cache or by querying the API),
// or nil if it is undefined.
//
// The Personio API want the client to generate the IDs, so an undefined day ID
// means you are free to generate your own ID.
//
// After the remote lookup to the API, the client caches which days in the same
// month that has undefined IDs.
func (c *Client) GetDayUUID(date time.Time) (*uuid.UUID, error) {
	dateString := date.Format(time.DateOnly)
	// Cache contains nil values on "known to be undefined day IDs"
	if id, ok := c.dayIDCache[dateString]; ok {
		return id, nil
	}
	startDate, endDate := util.TimeFullMonth(date)
	cal, err := c.GetMyAttendanceCalendar(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get days for range %s-%s: %w",
			startDate.Format(time.DateOnly),
			endDate.Format(time.DateOnly),
			err)
	}

	c.cacheDayIDs(cal)
	return c.dayIDCache[dateString], nil
}

func (c *Client) cacheDayIDs(days []Timecard) {

	// Cache known days
	for _, day := range days {
		id := day.DayID
		if id == nil {
			continue
		}
		c.dayIDCache[day.Date] = id
		log.Debug().Str("day", day.Date).Stringer("uuid", id).
			Msg("Cached existing UUID for day.")
	}
}
