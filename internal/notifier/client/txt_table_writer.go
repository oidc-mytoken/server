package notifier

import (
	"bytes"
	"io"

	"github.com/olekukonko/tablewriter"
)

func fPrintTable(out io.Writer, headers []string, data [][]string) {
	t := tablewriter.NewWriter(out)
	t.SetHeader(headers)
	// t.SetRowLine(true)
	t.SetColumnSeparator("I")
	t.AppendBulk(data)
	t.Render()
}

func fPrintSimpleTable(out io.Writer, headers []string, data map[string]string) {
	dataSlice := make([][]string, 0)
	for k, v := range data {
		dataSlice = append(
			dataSlice, []string{
				k,
				v,
			},
		)
	}
	fPrintTable(out, headers, dataSlice)
}

func generateSimpleTable(headers []string, data map[string]string) string {
	buf := bytes.NewBufferString("")
	fPrintSimpleTable(buf, headers, data)
	return buf.String()
}
