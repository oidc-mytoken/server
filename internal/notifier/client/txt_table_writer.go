package notifier

import (
	"bytes"
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

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
	t.Bulk(data)
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
