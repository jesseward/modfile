package player

import "github.com/jesseward/impulse/pkg/module"

// Ticker defines the interface for processing a single tick of a row for a specific module format.
type Ticker interface {
	// ProcessTick handles the logic for a single tick, including effects and period calculations.
	ProcessTick(p *Player, playerState *playerState, channelState *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int)
	RenderChannelTick(p *Player, state *channelState, tickBuffer []int, samplesPerTick int)
}
