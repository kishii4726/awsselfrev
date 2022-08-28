package color

import (
	"github.com/fatih/color"
)

func SetLevelColor() (string, string, string) {

	fgCyan := color.New(color.FgHiCyan).SprintFunc()
	fgYellow := color.New(color.FgHiYellow).SprintFunc()
	fgRed := color.New(color.FgHiRed).SprintFunc()

	return fgCyan("INFO"), fgYellow("Warning"), fgRed("Alert")
}
