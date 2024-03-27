package interactive

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

type commandSent string
type configurationChangeSent string

// TODO: different modes to show different keymaps
type helpKeyMap struct {
	SweepEnable  key.Binding
	SweepDisable key.Binding

	CarrierEnable  key.Binding
	CarrierDisable key.Binding

	ModulationEnable  key.Binding
	ModulationDisable key.Binding

	IdlePatternEnable  key.Binding
	IdlePatternDisable key.Binding

	Quit key.Binding
}

func (h helpKeyMap) Update(m model) helpKeyMap {
	// update will update the model with disabled/enabled key status

	// currently not expecting an update on screen keypress debouncing
	// so this is more tied to the stream state than if they just pressed a key

	var enableCommands bool
	if !m.streamClosed && m.withinOperationTime {
		enableCommands = true
	}

	h.SweepEnable.SetEnabled(enableCommands)
	h.SweepDisable.SetEnabled(enableCommands)
	h.CarrierEnable.SetEnabled(enableCommands)
	h.CarrierDisable.SetEnabled(enableCommands)
	h.ModulationEnable.SetEnabled(enableCommands)
	h.ModulationDisable.SetEnabled(enableCommands)
	h.IdlePatternEnable.SetEnabled(enableCommands)
	h.IdlePatternDisable.SetEnabled(enableCommands)

	return h
}

func defaultKeyMap() helpKeyMap {
	return helpKeyMap{
		SweepEnable: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "enable sweep"),
		),
		SweepDisable: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "disable sweep"),
		),
		ModulationEnable: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "enable modulation"),
		),
		ModulationDisable: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "disable modulation"),
		),
		CarrierEnable: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "enable carrier"),
		),
		CarrierDisable: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "enable carrier"),
		),
		IdlePatternEnable: key.NewBinding(
			key.WithKeys("7"),
			key.WithHelp("7", "enable idle pattern"),
		),
		IdlePatternDisable: key.NewBinding(
			key.WithKeys("8"),
			key.WithHelp("8", "disable idle pattern"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (h helpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		h.SweepEnable,
		h.SweepDisable,
		h.CarrierEnable,
		h.CarrierDisable,
		h.ModulationEnable,
		h.ModulationDisable,
		h.IdlePatternEnable,
		h.IdlePatternDisable,
		h.Quit,
	}
}

func (h helpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			h.SweepEnable,
			h.SweepDisable,
			h.CarrierEnable,
			h.CarrierDisable,
			h.ModulationEnable,
			h.ModulationDisable,
			h.IdlePatternEnable,
			h.IdlePatternDisable,
		},
		{h.Quit},
	}
}

type (
	screenIntervalTick struct {
		duration time.Duration
	}
)

func startScreenInterval(duration time.Duration) tea.Cmd {
	return tea.Tick(duration, func(time.Time) tea.Msg {
		return screenIntervalTick{duration: duration}
	})
}
