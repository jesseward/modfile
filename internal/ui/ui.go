package ui

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jesseward/impulse/internal/player"
	"github.com/jesseward/impulse/pkg/module"
	"golang.org/x/term"
)

const (
	minWidth  = 80
	minHeight = 24
)

var (
	titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("27")). // A nice purple
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
	borderColorStyle = lipgloss.NewStyle().BorderForeground(lipgloss.Color("15"))
)

func New(m module.Module) {
	// Get terminal dimensions
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v", err)
	}

	// Check if terminal is too small
	if width < minWidth || height < minHeight {
		fmt.Printf("Terminal too small. Minimum size is %d a %d, but yours is %d a %d.\n", minWidth, minHeight, width, height)
		return
	}

	stateUpdateChan := make(chan player.PlayerStateUpdate)
	opts := player.DefaultPlayerOptions()
	p := player.NewPlayer(m, func(format string, a ...interface{}) {}, stateUpdateChan, opts)
	audioPlayer, err := player.NewOtoPlayer(opts)
	if err != nil {
		panic(err)
	}

	mod := initialModel(m, p, audioPlayer)

	program := tea.NewProgram(mod, tea.WithAltScreen(), tea.WithMouseAllMotion())

	// Goroutine to listen for player state updates
	go func() {
		for update := range stateUpdateChan {
			program.Send(playerStateUpdateMsg(update))
		}
	}()

	// Goroutine for the timer
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			program.Send(playerTickMsg{})
		}
	}()

	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
	audioPlayer.Close()
}
