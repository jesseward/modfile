package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jesseward/impulse/internal/player"
	"github.com/jesseward/impulse/pkg/module"
)

type viewState int

const (
	showTracker viewState = iota
	showSamples
	showWaveform
	showQuitConfirmation
)

type playerStateUpdateMsg player.PlayerStateUpdate
type playerTickMsg struct{}
type clearFlashMessageMsg struct{}

type model struct {
	module      module.Module
	player      *player.Player
	audioPlayer *player.OtoPlayer
	stopChan    chan struct{}
	isPlaying   bool
	duration    int
	lastUpdate  player.PlayerStateUpdate

	width, height int
	activeView    viewState
	previousView  viewState
	flashMessage  string

	header   headerModel
	tracker  trackerModel
	sampler  samplerModel
	waveform waveformModel
	footer   footerModel
}

func initialModel(m module.Module, p *player.Player, ap *player.OtoPlayer) model {
	return model{
		module:      m,
		player:      p,
		audioPlayer: ap,
		isPlaying:   false,
		activeView:  showTracker,
		header:      newHeaderModel(m),
		tracker:     newTrackerModel(m),
		sampler:     newSamplerModel(m),
		footer:      newFooterModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := max(msg.Width, minWidth)
		height := max(msg.Height, minHeight)

		m.width = width
		m.height = height
		m.header.width = width
		m.footer.width = width - 2
		mainViewHeight := height - m.header.height() - m.footer.height()
		m.tracker.width = width
		m.tracker.height = mainViewHeight
		m.sampler.width = width
		m.sampler.height = mainViewHeight
		m.sampler.table.SetHeight(mainViewHeight - 4) // account for border and padding

		if m.waveform.sample != nil {
			m.waveform = newWaveformModel(m.waveform.sample, width, height)
		}

	case tea.KeyMsg:
		if m.activeView == showQuitConfirmation {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
			case "n", "N", "esc":
				m.activeView = m.previousView
			}
			return m, nil
		}

		if m.activeView == showWaveform {
			switch msg.String() {
			case "q", "ctrl+c", "enter", "esc":
				m.activeView = showSamples
			}
			return m, nil
		}

		switch key := msg.String(); key {
		case "q", "ctrl+c":
			m.previousView = m.activeView
			m.activeView = showQuitConfirmation
			return m, nil
		case " ":
			m.isPlaying = !m.isPlaying
			if m.isPlaying {
				m.flashMessage = "Playback started."
				m.stopChan = make(chan struct{})
				go m.player.WriteRaw(m.audioPlayer, m.stopChan)
			} else {
				m.flashMessage = "Playback paused."
				if m.stopChan != nil {
					close(m.stopChan)
				}
			}
			cmds = append(cmds, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return clearFlashMessageMsg{}
			}))
		case "tab":
			if m.activeView == showTracker {
				m.activeView = showSamples
				m.sampler.table.Focus()
			} else {
				m.activeView = showTracker
				m.sampler.table.Blur()
			}
		case "enter":
			if m.activeView == showSamples {
				selectedSampleIndex := m.sampler.table.Cursor()
				if selectedSampleIndex >= 0 && selectedSampleIndex < len(m.module.Samples()) {
					sample := m.module.Samples()[selectedSampleIndex]
					if sample.Length() == 0 {
						m.flashMessage = "Sample Is Empty"
						cmds = append(cmds, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
							return clearFlashMessageMsg{}
						}))
					} else {
						m.waveform = newWaveformModel(sample, m.width, m.height)
						m.activeView = showWaveform
					}
				}
			}

		}

	case playerStateUpdateMsg:
		m.lastUpdate = player.PlayerStateUpdate(msg)
		m.tracker.update(m.lastUpdate)
		m.header.update(m.lastUpdate, m.duration)
		return m, nil

	case playerTickMsg:
		if m.isPlaying {
			m.duration++
			m.header.update(m.lastUpdate, m.duration)
		}
		return m, nil

	case clearFlashMessageMsg:
		m.flashMessage = ""
		return m, nil
	}

	if m.activeView == showSamples {
		m.sampler, cmd = m.sampler.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var mainView string
	switch m.activeView {
	case showTracker:
		mainView = m.tracker.View()
	case showSamples:
		mainView = m.sampler.View()
	case showWaveform:
		mainView = m.waveform.View()
	case showQuitConfirmation:
		// Keep the background view
		switch m.previousView {
		case showTracker:
			mainView = m.tracker.View()
		case showSamples:
			mainView = m.sampler.View()
		}
	}

	var footerView string
	if m.flashMessage != "" {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("200")).
			Padding(0, 1)
		footerView = style.Render(m.flashMessage)
	} else {
		footerView = m.footer.View()
	}

	base := lipgloss.JoinVertical(lipgloss.Left,
		m.header.View(),
		mainView,
		footerView,
	)

	if m.activeView == showQuitConfirmation {
		dialogBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Inherit(borderColorStyle).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

		question := lipgloss.NewStyle().Width(50).Align(lipgloss.Center).Render("Are you sure you want to quit? (y/n)")
		ui := lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			dialogBox.Render(question),
		)
		return ui
	}

	return base
}
