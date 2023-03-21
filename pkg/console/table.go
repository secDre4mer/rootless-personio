// Dinkur the task time tracking utility.
// <https://github.com/dinkur/dinkur>
//
// SPDX-FileCopyrightText: 2021 Kalle Fagerberg
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

package console

import (
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

type tableCell struct {
	s string
	w int
}

type Table struct {
	colWidth   []int
	rows       [][]tableCell
	pendingRow []tableCell
	prefix     string
	spacing    string
}

func (t *Table) SetPrefix(prefix string) {
	t.prefix = prefix
}

func (t *Table) SetSpacing(spacing string) {
	t.spacing = spacing
}

func (t *Table) WriteColoredRow(c *color.Color, headers ...string) {
	for _, cell := range headers {
		t.WriteCellWidth(c.Sprint(cell), utf8.RuneCountInString(cell))
	}
	t.CommitRow()
}

func (t *Table) WriteCell(s string) {
	t.pendingRow = append(t.pendingRow, tableCell{s, utf8.RuneCountInString(s)})
}

func (t *Table) WriteCellColor(s string, c *color.Color) {
	t.WriteCellWidth(c.Sprint(s), utf8.RuneCountInString(s))
}

func (t *Table) WriteCellWidth(s string, width int) {
	t.pendingRow = append(t.pendingRow, tableCell{s, width})
}

func (t *Table) CommitRow() {
	t.rows = append(t.rows, t.pendingRow)
	t.expandColWidths(t.pendingRow)
	t.pendingRow = nil
}

func (t *Table) Rows() int {
	return len(t.rows)
}

func (t *Table) Println() {
	t.Fprintln(os.Stdout)
}

func (t *Table) Fprintln(w io.Writer) {
	var sb strings.Builder
	rowsWithSpaces := len(t.colWidth) - 1
	spaces := strings.Repeat(" ", t.WidestCellWidth())
	colSpaces := make([]string, len(t.colWidth))
	for i, w := range t.colWidth {
		colSpaces[i] = spaces[:w]
	}
	for _, row := range t.rows {
		sb.WriteString(t.prefix)
		for i, cell := range row {
			if i > 0 {
				sb.WriteString(t.spacing)
			}
			sb.WriteString(cell.s)
			if i < rowsWithSpaces {
				sb.WriteString(colSpaces[i][cell.w:])
			}
		}
		sb.WriteByte('\n')
	}
	w.Write([]byte(sb.String()))
}

func (t *Table) WidestCellWidth() int {
	var width int
	for _, w := range t.colWidth {
		if w > width {
			width = w
		}
	}
	return width
}

func (t *Table) Width() int {
	var width int
	for _, w := range t.colWidth {
		width += w
	}
	width += len(t.prefix)
	width += len(t.spacing) * (len(t.colWidth) - 1)
	return width
}

func (t *Table) expandColWidths(cells []tableCell) {
	for len(t.colWidth) < len(cells) {
		t.colWidth = append(t.colWidth, 0)
	}
	for i, cell := range cells {
		if cell.w > t.colWidth[i] {
			t.colWidth[i] = cell.w
		}
	}
}
