package player

import (
	"math"

	"github.com/jesseward/impulse/pkg/module"
	"github.com/jesseward/impulse/pkg/xm"
)

var amigaPeriodTable = [12]uint16{
	// C-1 to B-1
	856, 808, 762, 720, 678, 640, 604, 570, 538, 508, 480, 453,
}

var xmSinTable = [32]uint8{
	0, 24, 49, 74, 97, 120, 141, 161, 180, 197, 212, 224, 235, 244, 250, 253,
	255, 253, 250, 244, 235, 224, 212, 197, 180, 161, 141, 120, 97, 74, 49, 24,
}

type XMTicker struct{}

func (t *XMTicker) ProcessTick(p *Player, playerState *playerState, channelState *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int) {
	xmModule, ok := p.module.(*xm.Module)
	if !ok {
		return
	}

	if tick == 0 {
		t.handleTickZero(p, xmModule, playerState, channelState, cell)
	}

	if cell.Effect == 0x0E && cell.EffectParam>>4 == 0x0D && tick == int(cell.EffectParam&0x0F) {
		// Note delay is a special case
		t.handleTickZero(p, xmModule, playerState, channelState, cell)
	}

	t.handleEffect(p, xmModule, playerState, channelState, cell, speed, bpm, nextRow, nextOrder, currentOrder, tick)

	t.updateEnvelopes(channelState, xmModule)
	t.applyVibrato(channelState)
	t.applyTremolo(channelState)
	t.applyAutovibrato(channelState, xmModule)

	if channelState.sample != nil && channelState.period > 0 {
		p.applyPorta(channelState)
	}
}

func (t *XMTicker) handleTickZero(p *Player, mod *xm.Module, playerState *playerState, state *channelState, cell *module.Cell) {
	if cell.Instrument > 0 && int(cell.Instrument) <= len(mod.Instruments) {
		instrument := mod.Instruments[cell.Instrument-1]
		if cell.Note > 0 && cell.Note < 97 {
			sampleIndex := instrument.SampleKeymap[cell.Note-1]
			if int(sampleIndex) < len(instrument.Samples) {
				state.sample = instrument.Samples[sampleIndex]
			}
		}
		state.sampleIndex = int(cell.Instrument)
		state.volume = float64(state.sample.Volume()) / 64.0
		state.panning = float64(state.sample.Panning()) / 255.0
	}

	if cell.Note > 0 && cell.Note < 97 { // Note is not KeyOff
		if state.sample != nil {
			state.period = t.getPeriod(cell.Note, int8(state.sample.Finetune()), mod.Header.Flags&1 == 0)
			state.samplePos = 0
			state.fadeout = 65535
			state.volumeEnvelopePos = 0
			state.panningEnvelopePos = 0
			state.autovibratoPos = 0
			state.sustained = true
		}
	} else if cell.Note == 97 { // KeyOff
		state.sustained = false
	}

	// Volume column effects
	if cell.Volume >= 0x10 && cell.Volume <= 0x50 {
		state.volume = float64(cell.Volume-0x10) / 64.0
	} else if cell.Volume >= 0x60 && cell.Volume <= 0x6F { // Volume slide down
		state.volume -= float64(cell.Volume&0x0F) / 64.0
	} else if cell.Volume >= 0x70 && cell.Volume <= 0x7F { // Volume slide up
		state.volume += float64(cell.Volume&0x0F) / 64.0
	} else if cell.Volume >= 0x80 && cell.Volume <= 0x8F { // Fine volume slide down
		state.volume -= float64(cell.Volume&0x0F) / 64.0
	} else if cell.Volume >= 0x90 && cell.Volume <= 0x9F { // Fine volume slide up
		state.volume += float64(cell.Volume&0x0F) / 64.0
	} else if cell.Volume >= 0xA0 && cell.Volume <= 0xAF { // Set vibrato speed
		state.vibratoSpeed = cell.Volume & 0x0F
	} else if cell.Volume >= 0xB0 && cell.Volume <= 0xBF { // Vibrato
		state.vibratoDepth = cell.Volume & 0x0F
	} else if cell.Volume >= 0xC0 && cell.Volume <= 0xCF { // Set panning
		state.panning = float64(cell.Volume&0x0F) / 15.0
	} else if cell.Volume >= 0xD0 && cell.Volume <= 0xDF { // Panning slide left
		state.panning -= float64(cell.Volume&0x0F) / 64.0
	} else if cell.Volume >= 0xE0 && cell.Volume <= 0xEF { // Panning slide right
		state.panning += float64(cell.Volume&0x0F) / 64.0
	} else if cell.Volume >= 0xF0 && cell.Volume <= 0xFF { // Tone portamento
		if cell.Volume&0x0F > 0 {
			state.portaSpeed = uint16(cell.Volume&0x0F) * 4
		}
	}
}

func (t *XMTicker) handleEffect(p *Player, mod *xm.Module, playerState *playerState, state *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int) {
	effect := cell.Effect
	param := cell.EffectParam

	if effect == 0 && param == 0 {
		return
	}

	switch effect {
	case 0x00: // Arpeggio
		if tick > 0 {
			var noteOffset byte
			switch tick % 3 {
			case 1:
				noteOffset = param >> 4
			case 2:
				noteOffset = param & 0x0F
			}
			period := t.getPeriod(cell.Note+noteOffset, int8(state.sample.Finetune()), mod.Header.Flags&1 == 0)
			state.arpPeriod = period
		}
	case 0x01: // Porta Up
		if param > 0 {
			state.lastPortaUp = param
		}
		state.period -= uint16(state.lastPortaUp) * 4
	case 0x02: // Porta Down
		if param > 0 {
			state.lastPortaDown = param
		}
		state.period += uint16(state.lastPortaDown) * 4
	case 0x03: // Tone Porta
		if cell.Note > 0 && cell.Note < 97 {
			state.portaTarget = t.getPeriod(cell.Note, int8(state.sample.Finetune()), mod.Header.Flags&1 == 0)
		}
		if param > 0 {
			state.portaSpeed = uint16(param) * 4
		}
	case 0x04: // Vibrato
		if param&0xF0 > 0 {
			state.vibratoSpeed = param >> 4
		}
		if param&0x0F > 0 {
			state.vibratoDepth = param & 0x0F
		}
	case 0x05: // Tone Porta + Volume Slide
		t.handleEffect(p, mod, playerState, state, &module.Cell{Effect: 0x03}, speed, bpm, nextRow, nextOrder, currentOrder, tick)
		t.handleEffect(p, mod, playerState, state, &module.Cell{Effect: 0x0A, EffectParam: param}, speed, bpm, nextRow, nextOrder, currentOrder, tick)
	case 0x06: // Vibrato + Volume Slide
		t.handleEffect(p, mod, playerState, state, &module.Cell{Effect: 0x04}, speed, bpm, nextRow, nextOrder, currentOrder, tick)
		t.handleEffect(p, mod, playerState, state, &module.Cell{Effect: 0x0A, EffectParam: param}, speed, bpm, nextRow, nextOrder, currentOrder, tick)
	case 0x07: // Tremolo
		if param&0xF0 > 0 {
			state.tremoloSpeed = param >> 4
		}
		if param&0x0F > 0 {
			state.tremoloDepth = param & 0x0F
		}
	case 0x08: // Set Panning
		state.panning = float64(param) / 255.0
	case 0x09: // Sample Offset
		if tick == 0 {
			if param > 0 {
				state.lastSampleOffset = uint16(param)
			}
			state.samplePos = float64(state.lastSampleOffset * 256)
		}
	case 0x0A: // Volume Slide
		x := param >> 4
		y := param & 0x0F
		if x > 0 {
			state.volume += float64(x) / 64.0
		} else {
			state.volume -= float64(y) / 64.0
		}
	case 0x0B: // Position Jump
		if tick == 0 {
			*nextOrder = int(param)
			*nextRow = 0
		}
	case 0x0C: // Set Volume
		if tick == 0 {
			state.volume = float64(param) / 64.0
		}
	case 0x0D: // Pattern Break
		if tick == 0 {
			*nextOrder = *currentOrder + 1
			*nextRow = int(param>>4)*10 + int(param&0x0F)
		}
	case 0x0E: // Extended Effects
		t.handleExtendedEffect(state, param>>4, param&0x0F, nextRow, tick, playerState)
	case 0x0F: // Set Speed/BPM
		if tick == 0 {
			if param <= 0x1F {
				*speed = int(param)
			} else {
				*bpm = int(param)
			}
		}
	case 0x10: // G: Set global volume
		if tick == 0 && param <= 0x40 {
			playerState.globalVolume = float64(param) / 64.0
		}
	case 0x11: // H: Global volume slide
		if tick > 0 {
			x := param >> 4
			y := param & 0x0F
			if x > 0 {
				playerState.globalVolume += float64(x) / 64.0
			} else {
				playerState.globalVolume -= float64(y) / 64.0
			}
			if playerState.globalVolume < 0 {
				playerState.globalVolume = 0
			}
			if playerState.globalVolume > 1.0 {
				playerState.globalVolume = 1.0
			}
		}
	case 0x14: // K: Key off
		if tick == int(param) {
			state.fadeout = 0
		}
	case 0x15: // L: Set envelope position
		if tick == 0 {
			state.volumeEnvelopePos = uint16(param)
			state.panningEnvelopePos = uint16(param)
		}
	case 0x1A: // R: Multi retrig note
		if tick > 0 && param&0x0F > 0 && tick%int(param&0x0F) == 0 {
			state.samplePos = 0
			// also apply volume modification
			x := param >> 4
			switch x {
			case 1:
				state.volume -= 1.0 / 64.0
			case 2:
				state.volume -= 2.0 / 64.0
			case 3:
				state.volume -= 4.0 / 64.0
			case 4:
				state.volume -= 8.0 / 64.0
			case 5:
				state.volume -= 16.0 / 64.0
			case 6:
				state.volume = state.volume * 2 / 3
			case 7:
				state.volume = state.volume / 2
			case 9:
				state.volume += 1.0 / 64.0
			case 0xA:
				state.volume += 2.0 / 64.0
			case 0xB:
				state.volume += 4.0 / 64.0
			case 0xC:
				state.volume += 8.0 / 64.0
			case 0xD:
				state.volume += 16.0 / 64.0
			case 0xE:
				state.volume = state.volume * 3 / 2
			case 0xF:
				state.volume = state.volume * 2
			}
			if state.volume < 0 {
				state.volume = 0
			}
			if state.volume > 1.0 {
				state.volume = 1.0
			}
		}
	case 0x1C: // T: Tremor
		if param > 0 {
			state.tremorSpeed = param >> 4
			state.tremorDepth = param & 0x0F
		}
		onTicks := int(state.tremorSpeed)
		offTicks := int(state.tremorDepth)
		if tick%(onTicks+offTicks) >= onTicks {
			state.volume = 0
		}
	}
}

func (t *XMTicker) handleExtendedEffect(state *channelState, command, value uint8, nextRow *int, tick int, playerState *playerState) {
	if tick == 0 {
		switch command {
		case 0x01: // E1x: Fine Porta Up
			state.period -= uint16(value) * 4
		case 0x02: // E2x: Fine Porta Down
			state.period += uint16(value) * 4
		case 0x03: // E3x: Set Gliss Control
			state.glissando = (value > 0)
		case 0x04: // E4x: Set Vibrato Control
			state.vibratoWave = value
		case 0x05: // E5x: Set Finetune
			state.finetune = int8(value)
		case 0x06: // E6x: Set Loop Begin/Loop
			if value == 0 {
				playerState.patternLoopRow = playerState.row
			} else {
				if playerState.patternLoopCount == 0 {
					playerState.patternLoopCount = int(value)
				}
				playerState.patternLoopCount--
				if playerState.patternLoopCount > 0 {
					*nextRow = playerState.patternLoopRow
				}
			}
		case 0x07: // E7x: Set Tremolo Control
			state.tremoloWave = value
		case 0x0A: // EAx: Fine Volume Slide Up
			state.volume += float64(value) / 64.0
		case 0x0B: // EBx: Fine Volume Slide Down
			state.volume -= float64(value) / 64.0
		case 0x0E: // EEx: Pattern Delay
			playerState.patternDelay = int(value)
		}
	}
	switch command {
	case 0x09: // E9x: Retrig Note
		if tick > 0 && value > 0 && tick%int(value) == 0 {
			state.samplePos = 0
		}
	case 0x0C: // ECx: Note Cut
		if tick == int(value) {
			state.volume = 0
		}
	}
}

func (t *XMTicker) updateEnvelopes(state *channelState, mod *xm.Module) {
	if state.sampleIndex <= 0 || state.sampleIndex > len(mod.Instruments) {
		return
	}
	instrument := mod.Instruments[state.sampleIndex-1]

	// Volume Envelope
	if instrument.VolumeType&1 != 0 {
		state.volume = t.processEnvelope(state.volumeEnvelopePos, instrument.VolumeEnvelopePoints[:], instrument.NumVolumePoints, instrument.VolumeSustainPoint, instrument.VolumeLoopStartPoint, instrument.VolumeLoopEndPoint, instrument.VolumeType&2 != 0 && state.sustained, instrument.VolumeType&4 != 0) / 64.0
		state.volumeEnvelopePos++
	} else if !state.sustained {
		state.volume = 0
	}

	// Panning Envelope
	if instrument.PanningType&1 != 0 {
		state.panning = t.processEnvelope(state.panningEnvelopePos, instrument.PanningEnvelopePoints[:], instrument.NumPanningPoints, instrument.PanningSustainPoint, instrument.PanningLoopStartPoint, instrument.PanningLoopEndPoint, instrument.PanningType&2 != 0 && state.sustained, instrument.PanningType&4 != 0) / 255.0
		state.panningEnvelopePos++
	}

	// Fadeout
	if !state.sustained {
		state.fadeout += instrument.VolumeFadeout
		if state.fadeout > 65535 {
			state.fadeout = 65535
			state.volume = 0
		}
	}
}

func (t *XMTicker) processEnvelope(pos uint16, points []xm.EnvelopePoint, numPoints, sustainPoint, loopStart, loopEnd byte, sustain, loop bool) float64 {
	if numPoints == 0 {
		return 64.0
	}

	// Find the current segment
	var p1, p2 int
	for p1 = int(numPoints) - 1; p1 >= 0; p1-- {
		if points[p1].Frame <= pos {
			break
		}
	}
	if p1 < 0 {
		p1 = 0
	}

	if sustain && pos >= points[sustainPoint].Frame {
		return float64(points[sustainPoint].Value)
	}

	if loop && pos >= points[loopEnd].Frame {
		pos = points[loopStart].Frame + (pos-points[loopStart].Frame)%(points[loopEnd].Frame-points[loopStart].Frame)
		// Recalculate p1
		for p1 = int(numPoints) - 1; p1 >= 0; p1-- {
			if points[p1].Frame <= pos {
				break
			}
		}
	}

	p2 = p1 + 1
	if p2 >= int(numPoints) {
		return float64(points[numPoints-1].Value)
	}

	// Linear interpolation
	x1, y1 := float64(points[p1].Frame), float64(points[p1].Value)
	x2, y2 := float64(points[p2].Frame), float64(points[p2].Value)

	if x2 == x1 {
		return y1
	}

	return y1 + (y2-y1)*(float64(pos)-x1)/(x2-x1)
}

func (t *XMTicker) applyVibrato(state *channelState) {
	if state.vibratoDepth == 0 {
		return
	}
	pos := state.vibratoPos
	var delta float64
	switch state.vibratoWave & 3 {
	case 0: // Sine
		delta = float64(xmSinTable[pos&31])
		if pos >= 32 {
			delta = -delta
		}
	case 1: // Ramp down
		delta = float64(255 - pos*8)
	case 2: // Square
		if pos < 32 {
			delta = 255
		} else {
			delta = -255
		}
	case 3: // Ramp up
		delta = float64(pos*8 - 255)
	}
	delta = delta * float64(state.vibratoDepth) / 128.0
	state.period += uint16(delta)
	state.vibratoPos = (state.vibratoPos + state.vibratoSpeed) & 63
}

func (t *XMTicker) applyTremolo(state *channelState) {
	if state.tremoloDepth == 0 {
		return
	}
	pos := state.tremoloPos
	var volDelta float64
	switch state.tremoloWave & 3 {
	case 0: // Sine
		volDelta = float64(xmSinTable[pos&31])
	case 1: // Ramp down
		volDelta = float64(255 - pos*8)
	case 2: // Square
		volDelta = 255
	case 3: // Ramp up
		volDelta = float64(pos*8 - 255)
	}
	state.volume += volDelta * float64(state.tremoloDepth) / 32768.0
	state.tremoloPos = (state.tremoloPos + state.tremoloSpeed) & 63
}

func (t *XMTicker) applyAutovibrato(state *channelState, mod *xm.Module) {
	if state.sampleIndex <= 0 || state.sampleIndex > len(mod.Instruments) {
		return
	}
	instrument := mod.Instruments[state.sampleIndex-1]
	if instrument.VibratoDepth == 0 {
		return
	}

	sweep := float64(instrument.VibratoSweep)
	depth := float64(instrument.VibratoDepth)
	rate := float64(instrument.VibratoRate)

	var delta float64
	switch instrument.VibratoType {
	case 0: // Sine
		delta = math.Sin(float64(state.autovibratoPos)*math.Pi*2/256.0) * depth
	case 1: // Square
		if state.autovibratoPos < 128 {
			delta = depth
		} else {
			delta = -depth
		}
	case 2: // Ramp up
		delta = (float64(state.autovibratoPos)/128.0 - 1.0) * depth
	case 3: // Ramp down
		delta = (1.0 - float64(state.autovibratoPos)/128.0) * depth
	}

	if float64(state.autovibratoPos) < sweep {
		delta *= float64(state.autovibratoPos) / sweep
	}

	state.period += uint16(delta)
	state.autovibratoPos = (state.autovibratoPos + uint8(rate)) & 255
}

func (t *XMTicker) roundPeriodToSemitone(state *channelState) {
	// With linear frequencies, 1 semitone is 64 period units.
	// We want to find the closest period that corresponds to a semitone.
	// This is equivalent to finding the closest multiple of 64.
	// The formula in libxm is a bit more complex, let's analyze it:
	// new_period = (((ch->period + ch->finetune * 4 + 32) & 0xFFC0) - ch->finetune * 4);
	// This is for linear periods. Let's stick to a simpler rounding for now.
	// A simpler approach is to find the note for the period and then get the period for that note.

	// For now, let's do a simple rounding.
	// This is not entirely correct, but it's a start.
	if state.period > 0 {
		// find the note for the period
		note := 1.0 + float64(7680-int(state.period)-int(state.finetune)/2)/64.0
		note = math.Round(note)
		state.period = uint16(7680 - (int(note)-1)*64 - int(state.finetune)/2)
	}
}

func (t *XMTicker) getPeriod(note byte, finetune int8, isAmiga bool) uint16 {
	if isAmiga {
		noteIndex := (note - 1) % 12
		octave := (note - 1) / 12
		period := amigaPeriodTable[noteIndex]
		return period >> octave
	} else {
		// Linear
		return uint16(7680 - (int(note)-1)*64 - int(finetune)/2)
	}
}

func (t *XMTicker) RenderChannelTick(p *Player, state *channelState, tickBuffer []int, samplesPerTick int) {
	if state.sample == nil || state.period == 0 || state.sample.Length() == 0 || state.sampleIndex == -1 {
		return
	}

	xmModule, ok := p.module.(*xm.Module)
	if !ok {
		return
	}

	var freq float64
	if xmModule.Header.Flags&1 == 0 { // Amiga
		freq = 3546895.0 / float64(state.period)
	} else { // Linear
		freq = 8363 * math.Pow(2, (4608.0-float64(state.period))/768.0)
	}

	step := freq / float64(p.opts.SampleRate)
	sampleData := state.sample.Data()
	sampleLength := float64(state.sample.Length())
	loopStart := float64(state.sample.LoopStart())
	loopEnd := float64(state.sample.LoopEnd())
	hasLoop := state.sample.LoopLength() > 2
	isPingPong := state.sample.IsPingPong()

	for i := 0; i < samplesPerTick; i++ {
		if state.samplePos >= sampleLength {
			if hasLoop {
				if isPingPong {
					// Invert direction
					step = -step
					// Move sample position back within bounds
					state.samplePos = sampleLength - 1
				} else {
					state.samplePos -= (loopEnd - loopStart)
				}
			} else {
				continue
			}
		} else if isPingPong && state.samplePos < loopStart && step < 0 {
			// Invert direction
			step = -step
			state.samplePos = loopStart
		}

		pos := int(state.samplePos)
		if pos < 0 {
			pos = 0
		}

		if pos < len(sampleData) {
			// Linear interpolation
			posFloor := int(math.Floor(state.samplePos))
			posCeil := posFloor + 1
			t := state.samplePos - float64(posFloor)

			sample1 := float64(sampleData[posFloor])
			var sample2 float64
			if posCeil < len(sampleData) {
				sample2 = float64(sampleData[posCeil])
			} else if hasLoop {
				if isPingPong {
					sample2 = float64(sampleData[posFloor-1])
				} else {
					sample2 = float64(sampleData[int(loopStart)])
				}
			} else {
				sample2 = sample1
			}

			sampleValue := (sample1*(1.0-t) + sample2*t) * state.volume
			left, right := p.pan(p.opts.NumChannels, state.panning, sampleValue)

			offset := i * p.opts.NumChannels
			tickBuffer[offset] += left
			if p.opts.NumChannels > 1 {
				tickBuffer[offset+1] += right
			}
			state.samplePos += step
		}
	}
}
