package table

import (
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
)

var FailOnly bool

func SetTable() *tablewriter.Table {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetHeader([]string{"SERVICE", "STATUS", "LEVEL", "RESOURCE", "SETTING", "ISSUE"})

	return table
}

func AddRow(t *tablewriter.Table, row []string) {
	if FailOnly && len(row) > 1 {
		status := row[1]
		if status == "Pass" || status == "-" {
			return
		}
	}
	t.Append(row)
}

func Render(serviceName string, table *tablewriter.Table) {
	if table.NumLines() > 0 {
		table.Render()
	} else {
		if !FailOnly {
			log.Println(serviceName + ": No data to render.")
		}
	}
}
