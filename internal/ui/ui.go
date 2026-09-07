package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Color is a minimal ANSI color helper. Colors are disabled when noColor is
// true or NO_COLOR is set.
type Color struct {
	enabled bool
}

func NewColor(noColor bool) *Color {
	enabled := !noColor && os.Getenv("NO_COLOR") == ""
	return &Color{enabled: enabled}
}

func (c *Color) wrap(code, s string) string {
	if !c.enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (c *Color) Green(s string) string  { return c.wrap("32", s) }
func (c *Color) Red(s string) string    { return c.wrap("31", s) }
func (c *Color) Yellow(s string) string { return c.wrap("33", s) }

// Table renders aligned columns using text/tabwriter.
type Table struct {
	headers []string
	rows    [][]string
}

func NewTable(headers ...string) *Table {
	return &Table{headers: headers}
}

func (t *Table) AddRow(cells ...string) {
	t.rows = append(t.rows, cells)
}

func (t *Table) Render(w io.Writer) {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	if len(t.headers) > 0 {
		fmt.Fprintln(tw, strings.Join(t.headers, "\t"))
	}
	for _, r := range t.rows {
		fmt.Fprintln(tw, strings.Join(r, "\t"))
	}
	tw.Flush()
}

// PrintJSON writes v as indented JSON to stdout.
func PrintJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
