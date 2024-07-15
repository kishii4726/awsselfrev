package table

import (
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
)

func SetTable() *tablewriter.Table {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetHeader([]string{"SERVICE", "LEVEL", "RESOURCE", "REVIEW_ITEM"})

	return table
}

func Render(serviceName string, table *tablewriter.Table) {
	if table.NumLines() > 0 {
		table.Render()
	} else {
		log.Println(serviceName + ": No data to render.")
	}
}
