package interactive

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// simple prepend
// msg will be prepended with a newline to rest
// use it like append(currentLines, line, line2)
//
// prependLine("hello", "a", "b") output will look like
// "a
// b
// hello"
func prependLine(currentLines string, msg ...string) string {
	var builder strings.Builder
	for _, line := range msg {
		builder.WriteString(line + "\n")
	}
	builder.WriteString(currentLines)
	return builder.String()
}

var lightDark = lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))

var helpStyleDescription = lipgloss.NewStyle().Foreground(
	lightDark(lipgloss.Color("#B2B2B2"), lipgloss.Color("#cacaca")),
)
var helpStyleKey = lipgloss.NewStyle().Foreground(
	lightDark(lipgloss.Color("#909090"), lipgloss.Color("#a2a2a2")),
)
var helpStyleSeparator = lipgloss.NewStyle().Foreground(
	lightDark(lipgloss.Color("#DDDADA"), lipgloss.Color("#3C3C3C")),
)

var baseText = lipgloss.NewStyle().TabWidth(2).Foreground(lipgloss.Color("252"))
var baseTextDim = baseText.Foreground(lipgloss.Color("248"))
var boldStyle = baseText.Bold(true)

var headerStyle = boldStyle.Foreground(lipgloss.Color("252"))
var tableContent = baseText.Foreground(lipgloss.Color("252"))
var tableContentEven = baseText.Foreground(lipgloss.Color("248"))

var greenStyle = baseText.Foreground(lipgloss.Color("#75FBAB"))
var greenStyleDim = greenStyle.Foreground(lipgloss.Color("#59B980"))
var redStyle = baseText.Foreground(lipgloss.Color("#FF7698"))
var redStyleDim = redStyle.Foreground(lipgloss.Color("#BA5F75"))

func textDimmer(isDim bool, text string) string {
	if isDim {
		return baseTextDim.Render(text)
	}
	return baseText.Render(text)
}

func greenRedBoolText(isGreen bool, isDim bool, text string) string {
	if isGreen {
		if isDim {
			return greenStyleDim.Render(text)
		}
		return greenStyle.Render(text)
	} else {
		if isDim {
			return redStyleDim.Render(text)
		}
		return redStyle.Render(text)
	}
}

func tableStyleFunc(row, col int) lipgloss.Style {
	if row == 0 {
		return headerStyle
	}

	if row%2 == 0 {
		return tableContentEven
	}

	return tableContent
}

var viewportStyleBlue = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("248")).
	PaddingRight(2)
