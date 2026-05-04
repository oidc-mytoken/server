package notifier

import (
	"fmt"
	"testing"
)

func TestTable(*testing.T) {
	table := generateSimpleTable(
		nil, []TableRow{
			{
				Key:   "foo",
				Value: "bar",
			},
			{
				Key:   "FOO",
				Value: "BAR",
			},
		},
	)
	fmt.Println(table)
}
