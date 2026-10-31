package cli

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// Markdown cells retain JSON literals, including exact integers and escaped
// newlines, so data cannot introduce extra table rows or Markdown instructions.
func markdownCode(value string) string {
	longest, run := 0, 0
	for _, character := range value {
		if character == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	fence := strings.Repeat("`", longest+1)
	value = strings.NewReplacer("|", "\\|", "\r", "\\r", "\n", "\\n").Replace(value)
	return fence + " " + value + " " + fence
}

func markdownCell(value jsontext.Value) string {
	if len(value) == 0 {
		return "—"
	}
	return markdownCode(string(value))
}

func writeMarkdown(w io.Writer, title string, value any) error {
	raw, err := resultValue(value)
	if err != nil {
		return err
	}
	var body bytes.Buffer
	fmt.Fprintf(&body, "# Hot Aisle %s\n\n", title)
	switch raw[0] {
	case '[':
		var rows []jsontext.Value
		if err := json.Unmarshal(raw, &rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			body.WriteString("No items.\n")
			break
		}
		objects := make([]map[string]jsontext.Value, 0, len(rows))
		columns := map[string]bool{}
		for _, row := range rows {
			if row.Kind() != '{' {
				objects = nil
				break
			}
			var object map[string]jsontext.Value
			if err := json.Unmarshal(row, &object); err != nil {
				return err
			}
			objects = append(objects, object)
			for key := range object {
				columns[key] = true
			}
		}
		if len(objects) == len(rows) && len(columns) > 0 {
			keys := slices.Sorted(maps.Keys(columns))
			for _, key := range keys {
				fmt.Fprintf(&body, "| %s ", markdownCode(key))
			}
			body.WriteString("|\n")
			body.WriteString(strings.Repeat("| --- ", len(keys)))
			body.WriteString("|\n")
			for _, object := range objects {
				for _, key := range keys {
					fmt.Fprintf(&body, "| %s ", markdownCell(object[key]))
				}
				body.WriteString("|\n")
			}
		} else {
			for _, row := range rows {
				fmt.Fprintf(&body, "- %s\n", markdownCell(row))
			}
		}
	case '{':
		var object map[string]jsontext.Value
		if err := json.Unmarshal(raw, &object); err != nil {
			return err
		}
		body.WriteString("| Field | Value |\n| --- | --- |\n")
		for _, key := range slices.Sorted(maps.Keys(object)) {
			fmt.Fprintf(&body, "| %s | %s |\n", markdownCode(key), markdownCell(object[key]))
		}
	default:
		fmt.Fprintln(&body, markdownCell(raw))
	}
	_, err = w.Write(body.Bytes())
	return err
}
