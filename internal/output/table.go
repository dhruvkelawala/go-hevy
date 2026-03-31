package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func PrintTable(w io.Writer, headers []string, rows [][]string) {
	table := tablewriter.NewWriter(w)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetRowLine(false)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows)
	table.Render()
}

func PrintKeyValueTable(w io.Writer, rows [][2]string) {
	tableRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		tableRows = append(tableRows, []string{row[0], row[1]})
	}
	PrintTable(w, []string{"Field", "Value"}, tableRows)
}

func PrintCompact(w io.Writer, lines []string) error {
	_, err := fmt.Fprintln(w, strings.Join(lines, "\n"))
	return err
}

func ValueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
