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

// Package personio is the main client library for accessing Personio
// as a non-admin user. Use the [New] function followed by [Client.Login]
// to get started.
package personio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	UserAgent = "Rootless-Personio-bot/0.1 (+https://github.com/jilleJr/rootless-personio)"
)

var (
	ErrEmployeeIDNotFound = errors.New("employee ID not found")
	ErrNotLoggedIn        = errors.New("not logged in")
)

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

func (c *Client) Raw(req *http.Request) (any, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	u.Fragment = req.URL.Fragment
	u.RawFragment = req.URL.RawFragment
	u.RawQuery = req.URL.RawQuery
	u.ForceQuery = req.URL.ForceQuery
	u.Path += req.URL.Path

	reqClone := *req
	reqClone.URL = u

	return DoRequest[any](c.http, &reqClone)
}

func (c *Client) assertLoggedIn() error {
	if c.EmployeeID == 0 {
		return ErrNotLoggedIn
	}
	return nil
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

func DoRequest[M any](client *http.Client, req *http.Request) (M, error) {
	var zero M // only returned on fail

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	if log.Logger.GetLevel() <= zerolog.TraceLevel {
		logRequest(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return zero, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if log.Logger.GetLevel() <= zerolog.TraceLevel {
		logRespone(resp)
	}

	contentType := resp.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return zero, fmt.Errorf("parse Content-Type header: %w", err)
	}
	if mediaType != "application/json" {
		return zero, fmt.Errorf("expected JSON response, but got %q", mediaType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("read body: %w", err)
	}

	var typedBody struct {
		Success bool `json:"success"`
		Error   struct {
			Code      int                 `json:"code"`
			Message   string              `json:"message"`
			ErrorData map[string][]string `json:"error_data"`
		} `json:"error"`
		Data M `json:"data"`
	}
	if err := json.Unmarshal(body, &typedBody); err != nil {
		return zero, fmt.Errorf("parse body: %w", err)
	}

	if !typedBody.Success {
		return zero, Error{
			Code:      typedBody.Error.Code,
			Message:   typedBody.Error.Message,
			ErrorData: typedBody.Error.ErrorData,
			Response:  resp,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, fmt.Errorf("non-2xx status code: %s", resp.Status)
	}

	return typedBody.Data, nil
}

func logRequest(req *http.Request) {
	var sb strings.Builder
	sb.WriteString("Request:")
	fmt.Fprintf(&sb, "\n\t> %s %s %s", req.Method, req.URL.RequestURI(), req.Proto)
	fmt.Fprintf(&sb, "\n\t> Host: %s", req.URL.Hostname())
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Fprintf(&sb, "\n\t> %s: %s", key, value)
		}
	}
	log.Trace().Msg(sb.String())
}

func logRespone(resp *http.Response) {
	var sb strings.Builder
	sb.WriteString("Response:")
	fmt.Fprintf(&sb, "\n\t< %s %s", resp.Proto, resp.Status)
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Fprintf(&sb, "\n\t< %s: %s", key, value)
		}
	}
	log.Trace().Msg(sb.String())
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
