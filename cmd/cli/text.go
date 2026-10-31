package cli

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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

func (table textTable) width() int {
	width := max(0, len(table.widths)-1) * 2
	for _, column := range table.widths {
		width += column
	}
	return width
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

type textNode struct {
	kind   jsontext.Kind
	value  string
	fields map[string]textNode
	items  []textNode
}

func parseText(raw jsontext.Value) (textNode, error) {
	node := textNode{kind: raw.Kind()}
	switch node.kind {
	case '{':
		var fields map[string]jsontext.Value
		if err := json.Unmarshal(raw, &fields); err != nil {
			return node, err
		}
		node.fields = make(map[string]textNode, len(fields))
		for key, value := range fields {
			child, err := parseText(value)
			if err != nil {
				return node, err
			}
			node.fields[key] = child
		}
	case '[':
		var items []jsontext.Value
		if err := json.Unmarshal(raw, &items); err != nil {
			return node, err
		}
		for _, value := range items {
			child, err := parseText(value)
			if err != nil {
				return node, err
			}
			node.items = append(node.items, child)
		}
	case '"':
		if err := json.Unmarshal(raw, &node.value); err != nil {
			return node, err
		}
	default:
		node.value = string(raw)
	}
	return node, nil
}

func writeText(w io.Writer, value any) error {
	raw, err := resultValue(value)
	if err != nil {
		return err
	}
	node, err := parseText(raw)
	if err != nil {
		return err
	}
	var out strings.Builder
	node.write(&out, "")
	_, err = io.WriteString(w, out.String())
	return err
}

func (node textNode) inline() bool {
	if node.kind == '{' {
		return false
	}
	if node.kind == '[' {
		return !slices.ContainsFunc(node.items, func(item textNode) bool { return item.kind == '{' || item.kind == '[' })
	}
	return true
}

func (node textNode) text(key string) string {
	switch node.kind {
	case 'n':
		return "Not available"
	case 't':
		return "Yes"
	case 'f':
		return "No"
	case '[':
		if len(node.items) == 0 {
			return "None"
		}
		values := make([]string, 0, len(node.items))
		for _, item := range node.items {
			values = append(values, item.text(key))
		}
		return strings.Join(values, ", ")
	case '"':
		if node.value == "" || node.value == "0001-01-01T00:00:00Z" {
			return "Not set"
		}
	case '0':
		switch textFieldKey(key) {
		case "available_balance", "minimum_balance":
			return textMoney(node.value) + " USD"
		case "hourly_rate", "hourly_price_cents", "on_demand_price":
			return textMoney(node.value) + " USD/hour"
		case "ram_capacity", "disk_capacity", "ram_bytes", "disk_bytes", "capacity":
			if bytes, err := strconv.ParseUint(node.value, 10, 64); err == nil {
				return gibibytes(new(bytes)) + " GiB"
			}
		case "minimum_reservation_minutes":
			return node.value + " min"
		}
	}
	return node.value
}

func textMoney(cents string) string {
	amount := strings.TrimPrefix(cents, "-")
	if len(amount) < 3 {
		amount = strings.Repeat("0", 3-len(amount)) + amount
	}
	result := "$" + amount[:len(amount)-2] + "." + amount[len(amount)-2:]
	if strings.HasPrefix(cents, "-") {
		result = "-" + result
	}
	return result
}

func textFieldKey(key string) string {
	switch key {
	case "MinimumReservationMinutes":
		return "minimum_reservation_minutes"
	case "OnDemandPrice":
		return "on_demand_price"
	default:
		return strings.ToLower(key)
	}
}

func textLabel(key string) string {
	key = textFieldKey(key)
	switch key {
	case "ram_capacity", "ram_bytes":
		return "RAM"
	case "disk_capacity", "disk_bytes":
		return "Disk"
	case "minimum_reservation_minutes":
		return "Minimum reservation"
	case "on_demand_price", "hourly_price_cents":
		return "Hourly price"
	case "cpus":
		return "CPUs"
	case "gpus":
		return "GPUs"
	}
	words := strings.Split(key, "_")
	for i, word := range words {
		switch word {
		case "api", "cpu", "dns", "gpu", "ip", "os", "ram", "ssh", "url", "usd", "vm":
			words[i] = strings.ToUpper(word)
		}
	}
	label := strings.Join(words, " ")
	if label != "" {
		first, width := utf8.DecodeRuneInString(label)
		label = string(unicode.ToUpper(first)) + label[width:]
	}
	return label
}

func textKeys(fields map[string]textNode) []string {
	priority := []string{"user", "name", "email", "handle", "prefix", "label", "type", "fingerprint", "state", "host", "ip_address", "dns_name", "port", "description", "roles", "effective_roles", "cpu_cores", "ram_capacity", "disk_capacity"}
	order := func(key string) int {
		index := slices.Index(priority, textFieldKey(key))
		if index == -1 {
			return len(priority)
		}
		return index
	}
	return slices.SortedFunc(maps.Keys(fields), func(a, b string) int {
		return cmp.Or(cmp.Compare(order(a), order(b)), strings.Compare(textLabel(a), textLabel(b)))
	})
}

func (node textNode) write(out *strings.Builder, indent string) {
	switch node.kind {
	case '{':
		if len(node.fields) == 0 {
			fmt.Fprintln(out, indent+"None")
			return
		}
		keys := textKeys(node.fields)
		width := 0
		for _, key := range keys {
			if node.fields[key].inline() {
				width = max(width, len(terminalCell(textLabel(key)))+1)
			}
		}
		section := false
		for _, key := range keys {
			child := node.fields[key]
			if child.inline() {
				if section {
					out.WriteByte('\n')
				}
				prefix := fmt.Sprintf("%s%-*s  ", indent, width, terminalCell(textLabel(key))+":")
				writeTextField(out, prefix, terminalCell(child.text(key)))
				section = false
			} else {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				fmt.Fprintln(out, indent+terminalCell(textLabel(key)))
				child.write(out, indent+"  ")
				section = true
			}
		}
	case '[':
		if len(node.items) == 0 {
			fmt.Fprintln(out, indent+"None")
			return
		}
		if table, ok := node.table(); ok && table.width()+len(indent) <= 120 {
			table.write(out, indent)
			return
		}
		for i, item := range node.items {
			if i > 0 {
				out.WriteByte('\n')
			}
			if item.inline() {
				fmt.Fprintln(out, indent+"- "+terminalCell(item.text("")))
			} else {
				fmt.Fprintf(out, "%s%d.\n", indent, i+1)
				item.write(out, indent+"  ")
			}
		}
	default:
		fmt.Fprintln(out, indent+terminalCell(node.text("")))
	}
}

func writeTextField(out *strings.Builder, prefix, value string) {
	out.WriteString(prefix)
	width := max(30, 100-len(prefix))
	if len(value) <= width {
		fmt.Fprintln(out, value)
		return
	}
	line := 0
	for word := range strings.FieldsSeq(value) {
		if line > 0 && line+1+len(word) > width {
			out.WriteByte('\n')
			out.WriteString(strings.Repeat(" ", len(prefix)))
			line = 0
		}
		if line > 0 {
			out.WriteByte(' ')
			line++
		}
		out.WriteString(word)
		line += len(word)
	}
	out.WriteByte('\n')
}

func (node textNode) table() (textTable, bool) {
	fields := map[string]textNode{}
	for _, item := range node.items {
		if item.kind != '{' {
			return textTable{}, false
		}
		for key, value := range item.fields {
			if !value.inline() {
				return textTable{}, false
			}
			fields[key] = value
		}
	}
	if len(fields) == 0 {
		return textTable{}, false
	}
	keys := textKeys(fields)
	headings, right := make([]string, len(keys)), make([]bool, len(keys))
	for i, key := range keys {
		headings[i] = textLabel(key)
		right[i] = fields[key].kind == '0'
	}
	rows := make([][]string, 0, len(node.items))
	for _, item := range node.items {
		row := make([]string, len(keys))
		for i, key := range keys {
			if value, exists := item.fields[key]; exists {
				row[i] = value.text(key)
			} else {
				row[i] = "Not set"
			}
		}
		rows = append(rows, row)
	}
	return newTextTable(headings, rows, right), true
}

func textAPIMessage(message string) string {
	var raw jsontext.Value
	if json.Unmarshal([]byte(message), &raw) != nil {
		return terminalCell(message)
	}
	node, err := parseText(raw)
	if err != nil {
		return message
	}
	var out strings.Builder
	for _, key := range []string{"message", "detail", "error"} {
		if primary, exists := node.fields[key]; exists && primary.kind == '"' && primary.value != "" {
			fmt.Fprintln(&out, terminalCell(primary.value))
			delete(node.fields, key)
			if len(node.fields) == 0 {
				return strings.TrimSpace(out.String())
			}
			break
		}
	}
	node.write(&out, "")
	return strings.TrimSpace(out.String())
}
