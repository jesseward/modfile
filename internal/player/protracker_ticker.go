package player

import "github.com/jesseward/impulse/pkg/module"


var periodTable = [16 * 36]uint16{
	856, 808, 762, 720, 678, 640, 604, 570, 538, 508, 480, 453, 428, 404, 381, 360, 339, 320, 302, 285, 269, 254, 240, 226, 214, 202, 190, 180, 170, 160, 151, 143, 135, 127, 120, 113,
	850, 802, 757, 715, 674, 637, 601, 567, 535, 505, 477, 450, 425, 401, 379, 357, 337, 318, 300, 284, 268, 253, 239, 225, 213, 201, 189, 179, 169, 159, 150, 142, 134, 126, 119, 113,
	844, 796, 752, 709, 670, 632, 597, 563, 532, 502, 474, 447, 422, 398, 376, 355, 335, 316, 298, 282, 266, 251, 237, 224, 211, 199, 188, 177, 167, 158, 149, 141, 133, 125, 118, 112,
	838, 791, 746, 704, 665, 628, 592, 559, 528, 498, 470, 444, 419, 395, 373, 352, 332, 314, 296, 280, 264, 249, 235, 222, 209, 198, 187, 176, 166, 157, 148, 140, 132, 125, 118, 111,
	832, 785, 741, 699, 660, 623, 588, 555, 524, 495, 467, 441, 416, 392, 370, 350, 330, 312, 294, 278, 262, 247, 233, 220, 208, 196, 185, 175, 165, 156, 147, 139, 131, 124, 117, 110,
	826, 779, 736, 694, 655, 619, 584, 551, 520, 491, 463, 437, 413, 390, 368, 347, 328, 309, 292, 276, 260, 245, 232, 219, 206, 195, 184, 174, 164, 155, 146, 138, 130, 123, 116, 109,
	820, 774, 730, 689, 651, 614, 580, 547, 516, 487, 460, 434, 410, 387, 365, 345, 325, 307, 290, 274, 258, 244, 230, 217, 205, 193, 183, 172, 163, 154, 145, 137, 129, 122, 115, 109,
	814, 768, 725, 684, 646, 610, 575, 543, 513, 484, 457, 431, 407, 384, 363, 342, 323, 305, 288, 272, 256, 242, 228, 216, 204, 192, 181, 171, 161, 152, 144, 136, 128, 121, 114, 108,
	907, 856, 808, 762, 720, 678, 640, 604, 570, 538, 508, 480, 453, 428, 404, 381, 360, 339, 320, 302, 285, 269, 254, 240, 226, 214, 202, 190, 180, 170, 160, 151, 143, 135, 127, 120,
	900, 850, 802, 757, 715, 675, 636, 601, 567, 535, 505, 477, 450, 425, 401, 379, 357, 337, 318, 300, 284, 268, 253, 238, 225, 212, 200, 189, 179, 169, 159, 150, 142, 134, 126, 119,
	894, 844, 796, 752, 709, 670, 632, 597, 563, 532, 502, 474, 447, 422, 398, 376, 355, 335, 316, 298, 282, 266, 251, 237, 223, 211, 199, 188, 177, 167, 158, 149, 141, 133, 125, 118,
	887, 838, 791, 746, 704, 665, 628, 592, 559, 528, 498, 470, 444, 419, 395, 373, 352, 332, 314, 296, 280, 264, 249, 235, 222, 209, 198, 187, 176, 166, 157, 148, 140, 132, 125, 118,
	881, 832, 785, 741, 699, 660, 623, 588, 555, 524, 494, 467, 441, 416, 392, 370, 350, 330, 312, 294, 278, 262, 247, 233, 220, 208, 196, 185, 175, 165, 156, 147, 139, 131, 123, 117,
	875, 826, 779, 736, 694, 655, 619, 584, 551, 520, 491, 463, 437, 413, 390, 368, 347, 328, 309, 292, 276, 260, 245, 232, 219, 206, 195, 184, 174, 164, 155, 146, 138, 130, 123, 116,
	868, 820, 774, 730, 689, 651, 614, 580, 547, 516, 487, 460, 434, 410, 387, 365, 345, 325, 307, 290, 274, 258, 244, 230, 217, 205, 193, 183, 172, 163, 154, 145, 137, 129, 122, 115,
	862, 814, 768, 725, 684, 646, 610, 575, 543, 513, 484, 457, 431, 407, 384, 363, 342, 323, 305, 288, 272, 256, 242, 228, 216, 203, 192, 181, 171, 161, 152, 144, 136, 128, 121, 114,
}

var sin_table = [32]float64{
	0, 24, 49, 74, 97, 120, 141, 161, 180, 197, 212, 224, 235, 244, 250, 253,
	255, 253, 250, 244, 235, 224, 212, 197, 180, 161, 141, 120, 97, 74, 49, 24,
}

type ProtrackerTicker struct{}

func (t *ProtrackerTicker) ProcessTick(p *Player, playerState *playerState, channelState *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int) {
	if tick == 0 {
		t.handleTickZero(p, cell, channelState)
	}
	t.handleEffect(p, channelState, cell, speed, bpm, nextRow, nextOrder, currentOrder, tick, playerState)
}

func (t *ProtrackerTicker) handleTickZero(p *Player, cell *module.Cell, state *channelState) {
	if cell.SampleNumber > 0 && int(cell.SampleNumber) <= len(p.module.Samples()) {
		state.sampleIndex = int(cell.SampleNumber)
		state.sample = p.module.Samples()[state.sampleIndex-1]
		state.volume = float64(state.sample.Volume()) / 64.0
		if cell.Period > 0 {
			state.samplePos = 0
		}
	}
	if cell.Period > 0 {
		if cell.Effect == 0x03 || cell.Effect == 0x05 {
			state.portaTarget = cell.Period
		} else {
			state.period = cell.Period
		}
	}
}

func (t *ProtrackerTicker) handleEffect(p *Player, state *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int, playerState *playerState) {
	effect := cell.Effect
	val := cell.EffectParam
	switch effect {
	// Arpeggio alternates between the base note and two other notes, creating a chord-like effect.
	case 0x00: // Arpeggio
		if val > 0 && tick > 0 {
			arp_note := state.period
			switch tick % 3 {
			case 1:
				arp_note = t.getPeriod(state.period, int(val>>4))
			case 2:
				arp_note = t.getPeriod(state.period, int(val&0x0F))
			}
			state.period = arp_note
		}
	// Porta Up slides the pitch of the note up.
	case 0x01: // Porta Up
		if tick > 0 {
			if val > 0 {
				state.lastPortaUp = val
			}
			state.period -= uint16(state.lastPortaUp)
			if state.period < 113 {
				state.period = 113
			}
		}
	// Porta Down slides the pitch of the note down.
	case 0x02: // Porta Down
		if tick > 0 {
			if val > 0 {
				state.lastPortaDown = val
			}
			state.period += uint16(state.lastPortaDown)
			if state.period > 856 {
				state.period = 856
			}
		}
	// Tone Portamento slides the pitch from the previous note to the new note.
	case 0x03: // Tone Portamento
		if cell.Period > 0 {
			state.portaTarget = cell.Period
		}
		if val > 0 {
			state.portaSpeed = uint16(val)
		}
	// Vibrato oscillates the pitch of the note.
	case 0x04: // Vibrato
		if val>>4 > 0 {
			state.vibratoSpeed = val >> 4
		}
		if val&0x0F > 0 {
			state.vibratoDepth = val & 0x0F
		}
		t.applyVibrato(state)
	// Tone Portamento + Volume Slide combines a tone portamento with a volume slide.
	case 0x05: // Tone Portamento + Volume Slide
		t.handleEffect(p, state, &module.Cell{Effect: 0x03}, speed, bpm, nextRow, nextOrder, currentOrder, tick, playerState)
		t.handleEffect(p, state, &module.Cell{Effect: 0x0A, EffectParam: val}, speed, bpm, nextRow, nextOrder, currentOrder, tick, playerState)
	// Vibrato + Volume Slide combines a vibrato with a volume slide.
	case 0x06: // Vibrato + Volume Slide
		t.handleEffect(p, state, &module.Cell{Effect: 0x04}, speed, bpm, nextRow, nextOrder, currentOrder, tick, playerState)
		t.handleEffect(p, state, &module.Cell{Effect: 0x0A, EffectParam: val}, speed, bpm, nextRow, nextOrder, currentOrder, tick, playerState)
	// Tremolo oscillates the volume of the note.
	case 0x07: // Tremolo
		if val>>4 > 0 {
			state.tremoloSpeed = val >> 4
		}
		if val&0x0F > 0 {
			state.tremoloDepth = val & 0x0F
		}
		t.applyTremolo(state)
	// Set Panning sets the stereo panning of the channel.
	case 0x08: // Set Panning
		state.panning = float64(val) / 255.0
	// Set Sample Offset starts the sample from a specific offset.
	case 0x09: // Set Sample Offset
		if val > 0 {
			state.lastSampleOffset = uint16(val)
		}
		state.samplePos = float64(state.lastSampleOffset * 256)
	// Volume Slide slides the volume up or down.
	case 0x0A: // Volume Slide
		if tick > 0 {
			if val>>4 > 0 {
				state.lastVolSlide = val >> 4
				state.volume += float64(state.lastVolSlide) / 64.0
			} else {
				state.lastVolSlide = val & 0x0F
				state.volume -= float64(state.lastVolSlide) / 64.0
			}
			if state.volume < 0 {
				state.volume = 0
			}
			if state.volume > 1 {
				state.volume = 1
			}
		}
	// Position Jump jumps to a specific pattern in the order list.
	case 0x0B: // Position Jump
		*nextOrder = int(val)
		*nextRow = 0
	// Set Volume sets the volume of the channel.
	case 0x0C: // Set Volume
		state.volume = float64(val) / 64.0
	// Pattern Break jumps to a specific row in the next pattern.
	case 0x0D: // Pattern Break
		*nextOrder = *currentOrder + 1
		*nextRow = int(val>>4)*10 + int(val&0x0F)
	// Extended Effects handles extended effects.
	case 0x0E: // Extended Effects
		t.handleExtendedEffect(state, val>>4, val&0x0F, nextRow, tick, playerState)
	// Set Speed/BPM sets the speed (ticks per row) or BPM (beats per minute) of the song.
	case 0x0F: // Set Speed/BPM
		if val > 0 {
			if val <= 32 {
				*speed = int(val)
			} else {
				*bpm = int(val)
			}
		}
	}
}

func (t *ProtrackerTicker) handleExtendedEffect(state *channelState, command, value uint8, nextRow *int, tick int, playerState *playerState) {
	switch command {
	// Fine Porta Up slides the pitch of the note up by a small amount.
	case 0x01: // Fine Porta Up
		if tick == 0 {
			state.period -= uint16(value)
		}
	// Fine Porta Down slides the pitch of the note down by a small amount.
	case 0x02: // Fine Porta Down
		if tick == 0 {
			state.period += uint16(value)
		}
	// Glissando Control enables or disables glissando, which causes tone portamento to slide in discrete steps.
	case 0x03: // Glissando Control
		state.glissando = (value > 0)
	// Set Vibrato Waveform sets the waveform used for vibrato.
	case 0x04: // Set Vibrato Waveform
		state.vibratoWave = value
	// Set Finetune sets the finetune value for the instrument.
	case 0x05: // Set Finetune
		// Protracker does not support changing finetune with effects.
	// Pattern Loop loops a section of a pattern.
	case 0x06: // Pattern Loop
		if value == 0 {
			playerState.patternLoopRow = *nextRow
		} else {
			if playerState.patternLoopCount == 0 {
				playerState.patternLoopCount = int(value)
			} else {
				playerState.patternLoopCount--
			}
			if playerState.patternLoopCount > 0 {
				*nextRow = playerState.patternLoopRow
			}
		}
	// Set Tremolo Waveform sets the waveform used for tremolo.
	case 0x07: // Set Tremolo Waveform
		state.tremoloWave = value
	// Retrig Note re-triggers the note after a specified number of ticks.
	case 0x09: // Retrig Note
		if tick > 0 && tick%int(value) == 0 {
			state.samplePos = 0
		}
	// Fine Volume Slide Up slides the volume up by a small amount.
	case 0x0A: // Fine Volume Slide Up
		if tick == 0 {
			state.volume += float64(value) / 64.0
			if state.volume > 1.0 {
				state.volume = 1.0
			}
		}
	// Fine Volume Slide Down slides the volume down by a small amount.
	case 0x0B: // Fine Volume Slide Down
		if tick == 0 {
			state.volume -= float64(value) / 64.0
			if state.volume < 0 {
				state.volume = 0
			}
		}
	// Note Cut cuts the note after a specified number of ticks.
	case 0x0C: // Note Cut
		if tick == int(value) {
			state.volume = 0
		}
	// Note Delay delays the start of the note by a specified number of ticks.
	case 0x0D: // Note Delay
		if tick < int(value) {
			// This requires more complex handling of note triggering
		}
	// Pattern Delay delays the start of the next pattern by a specified number of rows.
	case 0x0E: // Pattern Delay
		playerState.patternDelay = int(value)
	}
}

func (t *ProtrackerTicker) applyVibrato(state *channelState) {
	var delta float64
	switch state.vibratoWave & 3 {
	case 0: // Sine
		delta = sin_table[state.vibratoPos&31]
		if state.vibratoPos >= 32 {
			delta = -delta
		}
	case 1: // Ramp down
		delta = float64(255 - (state.vibratoPos * 8))
	case 2: // Square
		if state.vibratoPos < 32 {
			delta = 255
		} else {
			delta = -255
		}
	}
	delta = delta * float64(state.vibratoDepth) / 128.0
	state.period += uint16(delta)
	state.vibratoPos = (state.vibratoPos + state.vibratoSpeed) & 63
}

func (t *ProtrackerTicker) applyTremolo(state *channelState) {
	var delta float64
	switch state.tremoloWave & 3 {
	case 0: // Sine
		delta = sin_table[state.tremoloPos&31]
		if state.tremoloPos >= 32 {
			delta = -delta
		}
	case 1: // Ramp down
		delta = float64(255 - (state.tremoloPos * 8))
	case 2: // Square
		if state.vibratoPos < 32 {
			delta = 255
		} else {
			delta = -255
		}
	}
	delta = delta * float64(state.tremoloDepth) / 64.0
	state.volume += delta / 64.0
	if state.volume < 0 {
		state.volume = 0
	}
	if state.volume > 1.0 {
		state.volume = 1.0
	}
	state.tremoloPos = (state.tremoloPos + state.tremoloSpeed) & 63
}

func (t *ProtrackerTicker) getPeriod(period uint16, offset int) uint16 {
	finetune := 0
	note := 0
	for i := range periodTable {
		if periodTable[i] == period {
			finetune = i / 36
			note = i % 36
			break
		}
	}
	note += offset
	if note < 0 {
		note = 0
	}
	if note > 35 {
		note = 35
	}
	return periodTable[finetune*36+note]
}

func (t *ProtrackerTicker) RenderChannelTick(p *Player, state *channelState, tickBuffer []int, samplesPerTick int) {
	if state.sample == nil || state.period == 0 || state.sample.Length() == 0 || state.sampleIndex == -1 {
		return
	}

	freq := 7093789.2 / (float64(state.period) * 2.0)
	step := freq / float64(p.opts.SampleRate)

	for i := 0; i < samplesPerTick; i++ {
		sampleLength := float64(state.sample.Length() * 2)
		loopStart := float64(state.sample.LoopStart() * 2)
		loopLength := float64(state.sample.LoopLength() * 2)

		if state.samplePos >= sampleLength {
			if loopLength > 1 {
				state.samplePos = loopStart
			} else {
				continue
			}
		}

		if int(state.samplePos) < len(state.sample.Data()) {
			sampleValue := float64(state.sample.Data()[int(state.samplePos)]) * state.volume
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