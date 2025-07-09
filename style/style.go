package style

import "github.com/charmbracelet/lipgloss"

type colors struct {
	Green lipgloss.TerminalColor
	Red   lipgloss.TerminalColor
	Blue  lipgloss.TerminalColor
}

type palette struct {
	colors
	Dark colors
}

type border struct {
	Style lipgloss.Border
	Color lipgloss.TerminalColor
}

type borders struct {
	Multi border
	Error border
	Alt   border
	Dup   border
}

var (
	Palette = palette{
		colors: colors{
			Green: lipgloss.ANSIColor(148),
			Red:   lipgloss.ANSIColor(204),
			Blue:  lipgloss.ANSIColor(39),
		},
		Dark: colors{
			Green: lipgloss.ANSIColor(106),
			Red:   lipgloss.ANSIColor(204),
			Blue:  lipgloss.ANSIColor(67),
		},
	}

	App = lipgloss.NewStyle().Margin(1, 2)
	Top = lipgloss.NewStyle().MarginTop(2).MarginLeft(2).MarginRight(2)
	Mid = lipgloss.NewStyle().MarginTop(1).MarginLeft(2).MarginRight(2)
	Bot = lipgloss.NewStyle().MarginTop(1).MarginLeft(2).MarginRight(2)

	SepBot = lipgloss.NewStyle().MarginBottom(1)

	Highlight = lipgloss.NewStyle().Foreground(
		lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"},
	)
	Error = lipgloss.NewStyle().Foreground(Palette.Red)
	Faint = lipgloss.NewStyle().Faint(true)
	Alt   = lipgloss.NewStyle().Foreground(Palette.Blue)

	Border = borders{
		Multi: border{
			Style: newBorder("⋮"),
			Color: Palette.Green,
		},
		Error: border{
			Style: newBorder("•"),
			Color: Palette.Red,
		},
		Alt: border{
			Style: newBorder("∅"),
			Color: Alt.GetForeground(),
		},
		Dup: border{
			Style: newBorder("≡"),
			Color: Palette.Red,
		},
	}
)

func Center(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)
}

func (b border) Apply(style lipgloss.Style) lipgloss.Style {
	return style.
		Border(b.Style, false, false, false, true).
		BorderForeground(b.Color).
		PaddingLeft(1)
}

func newBorder(left string) lipgloss.Border {
	b := lipgloss.NormalBorder()
	b.Left = left
	return b
}
