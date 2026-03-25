package notifier

import (
	"bytes"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

// TableRow represents a key-value pair for table generation with stable ordering
type TableRow struct {
	Key   string
	Value string
}

var tableWriteOptions []tablewriter.Option

func init() {
	tableWriteOptions = append(
		tableWriteOptions, tablewriter.WithRenderer(
			renderer.NewBlueprint(
				tw.Rendition{
					Symbols: tw.NewSymbolCustom("my-symbols").
						WithColumn("I").
						WithBottomRight("+").
						WithBottomLeft("+").
						WithMidRight("+").
						WithMidLeft("+").
						WithTopRight("+").
						WithTopLeft("+"),
				},
			),
		),
	)
}

func fPrintTable(out io.Writer, headers []string, data [][]string) {
	t := tablewriter.NewTable(out, tableWriteOptions...)
	t.Header(headers)
	// t.SetRowLine(true)
	_ = t.Bulk(data)
	_ = t.Render()
}

func fPrintSimpleTable(out io.Writer, headers []string, data []TableRow) {
	dataSlice := make([][]string, 0, len(data))
	for _, row := range data {
		dataSlice = append(
			dataSlice, []string{
				row.Key,
				row.Value,
			},
		)
	}
	fPrintTable(out, headers, dataSlice)
}

func generateSimpleTable(headers []string, data []TableRow) string {
	buf := bytes.NewBufferString("")
	fPrintSimpleTable(buf, headers, data)
	return buf.String()
}
