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

func ColorizeLevel(level string) string {
	switch level {
	case "Info":
		return color.New(color.FgHiCyan).SprintFunc()(level)
	case "Warning":
		return color.New(color.FgHiYellow).SprintFunc()(level)
	case "Alert":
		return color.New(color.FgHiRed).SprintFunc()(level)
	default:
		return level
	}
}
