package notifier

import (
	"fmt"
	"testing"
)

func TestTable(*testing.T) {
	table := generateSimpleTable(
		nil, map[string]string{
			"foo": "bar",
			"FOO": "BAR",
		},
	)
	fmt.Println(table)
}
