package player

import (
	"encoding/binary"

	"github.com/jesseward/impulse/pkg/module"
)

// PlayerOptions defines the configuration for the audio player.
type PlayerOptions struct {
	SampleRate  int
	NumChannels int
	BitDepth    int // in bytes
}

// DefaultPlayerOptions returns a default set of player options (CD quality).
func DefaultPlayerOptions() PlayerOptions {
	return PlayerOptions{
		SampleRate:  44100,
		NumChannels: 2,
		BitDepth:    2, // 16-bit
	}
}

type Player struct {
	module          module.Module
	ticker          Ticker
	log             func(format string, a ...interface{})
	StateUpdateChan chan<- PlayerStateUpdate
	opts            PlayerOptions
}

func NewPlayer(module module.Module, log func(format string, a ...interface{}), stateUpdateChan chan<- PlayerStateUpdate, opts PlayerOptions) *Player {
	var ticker Ticker
	switch module.Type() {
	case "Protracker":
		ticker = &ProtrackerTicker{}
	case "S3M":
		ticker = &S3MTicker{}
	case "FastTracker II Extended Module":
		ticker = &XMTicker{}
	}

	return &Player{
		module:          module,
		ticker:          ticker,
		log:             log,
		StateUpdateChan: stateUpdateChan,
		opts:            opts,
	}
}

func (p *Player) WriteRaw(player AudioPlayer, stopChan <-chan struct{}) error {
	if otoPlayer, ok := player.(*OtoPlayer); ok {
		otoPlayer.player.Play()
	}

	audioChan, errChan := p.renderSongByRow(stopChan)

	for {
		select {
		case audioBuf, ok := <-audioChan:
			if !ok {
				return nil
			}
			buf := make([]byte, len(audioBuf)*p.opts.BitDepth)
			for i, sample := range audioBuf {
				clipped := int16(max(min(sample, 32767), -32768))
				binary.LittleEndian.PutUint16(buf[i*2:], uint16(clipped))
			}
			if _, err := player.Write(buf); err != nil {
				return err
			}
		case err := <-errChan:
			return err
		case <-stopChan:
			return nil
		}
	}
}

func (p *Player) renderSongByRow(stopChan <-chan struct{}) (<-chan []int, <-chan error) {
	audioChan := make(chan []int)
	errChan := make(chan error, 1)

	go func() {
		defer close(audioChan)
		defer close(errChan)

		playerState := playerState{
			speed:    p.module.DefaultSpeed(),
			bpm:      p.module.DefaultBPM(),
			channels: make([]channelState, p.module.NumChannels()),
		}
		for i := range playerState.channels {
			playerState.channels[i] = defaultChannelState()
		}

		orderIndex := 0
		rowIndex := 0

		for orderIndex < p.module.SongLength() {
			select {
			case <-stopChan:
				return
			default:
			}

			patternIndex := p.module.PatternOrder()[orderIndex]
			if int(patternIndex) >= p.module.NumPatterns() {
				orderIndex++
				continue
			}

			numRows := p.module.NumRows(int(patternIndex))
			if rowIndex >= numRows {
				rowIndex = 0
				orderIndex++
				continue
			}

			playerState.row = rowIndex
			playerState.pattern = int(patternIndex)
			playerState.order = orderIndex

			if p.StateUpdateChan != nil {
				p.StateUpdateChan <- PlayerStateUpdate{
					Order:   orderIndex,
					Pattern: int(patternIndex),
					Row:     rowIndex,
					Speed:   playerState.speed,
					BPM:     playerState.bpm,
				}
			}

			rowBuffer, newRow, newOrder := p.processRow(&playerState, patternIndex)
			audioChan <- rowBuffer

			if newOrder != -1 {
				orderIndex = newOrder
				rowIndex = newRow
			} else if newRow != -1 {
				rowIndex = newRow
			} else {
				rowIndex++
			}
		}
	}()

	return audioChan, errChan
}

func (p *Player) processRow(state *playerState, pattern int) ([]int, int, int) {
	var rowBuffer []int

	nextOrder := -1
	nextRow := -1

	if state.patternDelay > 0 {
		state.patternDelay--
	} else {
		for tick := 0; tick < state.speed; tick++ {
			samplesPerTick := int(float64(p.opts.SampleRate) * 2.5 / float64(state.bpm))
			tickBuffer := make([]int, samplesPerTick*p.opts.NumChannels)

			for ch := 0; ch < p.module.NumChannels(); ch++ {
				cell := p.module.PatternCell(pattern, state.row, ch)
				channel := &state.channels[ch]
				p.ticker.ProcessTick(p, state, channel, &cell, &state.speed, &state.bpm, &nextRow, &nextOrder, &state.order, tick)
				if channel.sample != nil && channel.period > 0 {
					p.applyPorta(channel)
					p.ticker.RenderChannelTick(p, channel, tickBuffer, samplesPerTick)
				}
			}
			rowBuffer = append(rowBuffer, tickBuffer...)
		}
	}
	return rowBuffer, nextRow, nextOrder
}

func (p *Player) applyPorta(state *channelState) {
	if state.portaTarget > 0 {
		if state.period < state.portaTarget {
			state.period += state.portaSpeed
			if state.period > state.portaTarget {
				state.period = state.portaTarget
			}
		} else if state.period > state.portaTarget {
			state.period -= state.portaSpeed
			if state.period < state.portaTarget {
				state.period = state.portaTarget
			}
		}
		if xmTicker, ok := p.ticker.(*XMTicker); ok {
			if state.glissando {
				xmTicker.roundPeriodToSemitone(state)
			}
		}
	}
}

func (p *Player) pan(numChannels int, panning float64, sampleValue float64) (int, int) {
	if numChannels == 1 {
		return int(sampleValue), 0
	}
	return int(sampleValue * (1.0 - panning)), int(sampleValue * panning)
}

type playerState struct {
	order            int
	row              int
	pattern          int
	speed            int
	bpm              int
	globalVolume     float64
	channels         []channelState
	patternDelay     int
	patternLoopRow   int
	patternLoopCount int
}

type channelState struct {
	sample             module.Sample
	sampleIndex        int
	period             uint16
	samplePos          float64
	volume             float64
	portaTarget        uint16
	portaSpeed         uint16
	vibratoSpeed       uint8
	vibratoDepth       uint8
	vibratoWave        uint8
	vibratoPos         uint8
	tremoloSpeed       uint8
	tremoloDepth       uint8
	tremoloWave        uint8
	tremoloPos         uint8
	tremorSpeed        uint8
	tremorDepth        uint8
	arpPeriod          uint16
	glissando          bool
	finetune           int8
	panning            float64
	lastPortaUp        byte
	lastPortaDown      byte
	lastSampleOffset   uint16
	volumeEnvelopePos  uint16
	panningEnvelopePos uint16
	fadeout            uint16
	autovibratoPos     uint8
	sustained          bool
	lastVolSlide       byte
	lastPorta          byte
	stereo             float64
}

func defaultChannelState() channelState {
	return channelState{
		volume:      1.0,
		sampleIndex: -1,
		panning:     0.5,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
