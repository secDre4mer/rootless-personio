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

import "testing"

func TestNormalizeBaseURL(t *testing.T) {
	var tests = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "keeps as-is",
			url:  "https://example.personio.de",
			want: "https://example.personio.de",
		},
		{
			name: "trim trailing slash",
			url:  "https://example.personio.de/",
			want: "https://example.personio.de",
		},
		{
			name: "different path",
			url:  "https://example.personio.de/some/reverse/proxy/",
			want: "https://example.personio.de/some/reverse/proxy",
		},
		{
			name: "removes extra junk",
			url:  "https://example.personio.de?foo=bar#heello",
			want: "https://example.personio.de",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeBaseURL(tc.url)
			if err != nil {
				t.Fatalf("want %q, got error: %s", tc.url, err)
			}
			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
