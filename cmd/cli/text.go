package cli

import (
	"fmt"
	"strconv"
	"strings"
)

// All terminal tables use the same column sizing, spacing, and alignment.
type textTable struct {
	rows   [][]string
	widths []int
	right  []bool
}

func newTextTable(headings []string, rows [][]string, right []bool) textTable {
	table := textTable{rows: append([][]string{headings}, rows...), widths: make([]int, len(headings)), right: right}
	for i, row := range table.rows {
		table.rows[i] = make([]string, len(headings))
		for column, cell := range row {
			table.rows[i][column] = terminalCell(cell)
			table.widths[column] = max(table.widths[column], len(table.rows[i][column]))
		}
	}
	return table
}

func (table textTable) write(out *strings.Builder, indent string) {
	for index, row := range table.rows {
		if index == 1 {
			out.WriteString(indent)
			for column, width := range table.widths {
				if column > 0 {
					out.WriteString("  ")
				}
				out.WriteString(strings.Repeat("-", width))
			}
			out.WriteByte('\n')
		}
		out.WriteString(indent)
		for column, cell := range row {
			if column > 0 {
				out.WriteString("  ")
			}
			width := -table.widths[column]
			if column < len(table.right) && table.right[column] {
				width = -width
			}
			if column == len(row)-1 && width < 0 {
				out.WriteString(cell)
			} else {
				fmt.Fprintf(out, "%*s", width, cell)
			}
		}
		out.WriteByte('\n')
	}
}

func terminalCell(value string) string {
	// Escaping preserves data and prevents control characters or fallback glyph
	// widths from introducing extra rows or shifting terminal columns.
	quoted := strconv.QuoteToASCII(value)
	return quoted[1 : len(quoted)-1]
}
