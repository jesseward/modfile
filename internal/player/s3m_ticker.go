package player

import (
	"github.com/jesseward/impulse/pkg/module"
)

var s3mPeriodTable = [12 * 9]uint16{
	// Octave 0
	27392, 25856, 24384, 23040, 21696, 20480, 19328, 18240, 17216, 16256, 15360, 14496,
	// Octave 1
	13696, 12928, 12192, 11520, 10848, 10240, 9664, 9120, 8608, 8128, 7680, 7248,
	// Octave 2
	6848, 6464, 6096, 5760, 5424, 5120, 4832, 4560, 4304, 4064, 3840, 3624,
	// Octave 3
	3424, 3232, 3048, 2880, 2712, 2560, 2416, 2280, 2152, 2032, 1920, 1812,
	// Octave 4
	1712, 1616, 1524, 1440, 1356, 1280, 1208, 1140, 1076, 1016, 960, 906,
	// Octave 5
	856, 808, 762, 720, 678, 640, 604, 570, 538, 508, 480, 453,
	// Octave 6
	428, 404, 381, 360, 339, 320, 302, 285, 269, 254, 240, 226,
	// Octave 7
	214, 202, 190, 180, 170, 160, 151, 143, 135, 127, 120, 113,
	// Octave 8
	107, 101, 95, 90, 85, 80, 75, 71, 67, 63, 60, 56,
}

type S3MTicker struct{}

func (t *S3MTicker) ProcessTick(p *Player, playerState *playerState, channelState *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int) {
	if tick == 0 {
		t.handleTickZero(p, cell, channelState, playerState)
	}
	t.handleEffect(p, channelState, cell, speed, bpm, nextRow, nextOrder, currentOrder, tick, playerState)
	if tick > 0 {
		applyVibrato(channelState)
		applyTremolo(channelState)
	}
}

func (t *S3MTicker) handleTickZero(p *Player, cell *module.Cell, state *channelState, playerState *playerState) {
	if cell.Instrument > 0 && int(cell.Instrument) <= len(p.module.Samples()) {
		state.sampleIndex = int(cell.Instrument)
		state.sample = p.module.Samples()[state.sampleIndex-1]
		if cell.Volume <= 64 {
			state.volume = float64(cell.Volume) / 64.0
		} else {
			state.volume = float64(state.sample.Volume()) / 64.0
		}
		if cell.Note > 0 {
			state.samplePos = 0
		}
	} else if cell.Volume <= 64 {
		state.volume = float64(cell.Volume) / 64.0
	}

	if cell.Note < 254 {
		octave := cell.Note >> 4
		note := cell.Note & 0x0F
		if int(octave)*12+int(note) < len(s3mPeriodTable) {
			period := s3mPeriodTable[octave*12+note]
			if state.sample != nil {
				c2spd := state.sample.Finetune()
				if c2spd == 0 {
					c2spd = 8363
				}
				state.period = uint16(float64(period) * 8363.0 / float64(c2spd))
				state.notePeriod = state.period
			} else {
				state.period = period
				state.notePeriod = period
			}
		}
	} else if cell.Note == 254 {
		state.volume = 0
	}
}

func (t *S3MTicker) handleEffect(p *Player, state *channelState, cell *module.Cell, speed, bpm, nextRow, nextOrder, currentOrder *int, tick int, playerState *playerState) {
	effect := cell.Effect
	param := cell.EffectParam

	switch effect {
	case 1: // Axx: Set Speed
		if tick == 0 {
			*speed = int(param)
		}
	case 2: // Bxx: Jump to order
		if tick == 0 {
			*nextOrder = int(param)
			*nextRow = 0
		}
	case 3: // Cxx: Break pattern to row
		if tick == 0 {
			*nextOrder = *currentOrder + 1
			*nextRow = int(param>>4)*10 + int(param&0x0F)
		}
	case 4: // Dxy: Volume slide
		applyVolumeSlide(state, param, tick, true)
	case 5: // Exx: Portamento Down
		applyPortamentoDown(state, param, tick, true)
	case 6: // Fxx: Portamento Up
		applyPortamentoUp(state, param, tick, true)
	case 7: // Gxx: Tone portamento
		t.tonePortamento(state, param)
	case 8: // Hxy: Vibrato
		t.vibrato(state, param)
	case 9: // Ixy: Tremor
		t.tremor(state, param, tick)
	case 10: // Jxy: Arpeggio
		applyArpeggio(state, param, tick, t)
	case 11: // Kxy: Vibrato + Volume slide
		t.vibrato(state, 0)
		applyVolumeSlide(state, param, tick, true)
	case 12: // Lxy: Porta + Volume slide
		t.tonePortamento(state, 0)
		applyVolumeSlide(state, param, tick, true)
	case 15: // Oxy: Set sample offset
		if tick == 0 {
			if param > 0 {
				state.lastSampleOffset = uint16(param)
			}
			state.samplePos = float64(state.lastSampleOffset * 256)
		}
	case 17: // Qxy: Retrig + Volume slide
		if tick > 0 && tick%int(param&0x0F) == 0 {
			state.samplePos = 0
			applyVolumeSlide(state, param>>4, 0, true)
		}
	case 18: // Rxy: Tremolo
		t.tremolo(state, param)
	case 19: // Sxx: Special
		t.specialEffect(state, param, nextRow, tick, playerState)
	case 20: // Txx: Set Tempo
		if tick == 0 && param > 32 {
			*bpm = int(param)
		}
	case 21: // Uxy: Fine Vibrato
		t.vibrato(state, param) // Fine vibrato is just vibrato with higher precision, handle in vibrato
	case 22: // Vxx: Set global volume
		if tick == 0 {
			if param <= 64 {
				playerState.globalVolume = float64(param) / 64.0
			}
		}
	}
}





func (t *S3MTicker) tonePortamento(state *channelState, param byte) {
	if param > 0 {
		state.portaSpeed = uint16(param) * 4
	}
	// Note is set in handleTickZero
}

func (t *S3MTicker) vibrato(state *channelState, param byte) {
	if param > 0 {
		state.vibratoSpeed = param >> 4
		state.vibratoDepth = param & 0x0F
	}
}

func (t *S3MTicker) tremolo(state *channelState, param byte) {
	if param > 0 {
		state.tremoloSpeed = param >> 4
		state.tremoloDepth = param & 0x0F
	}
}

func (t *S3MTicker) tremor(state *channelState, param byte, tick int) {
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

func (t *S3MTicker) specialEffect(state *channelState, param byte, nextRow *int, tick int, playerState *playerState) {
	cmd := param >> 4
	val := param & 0x0F
	switch cmd {
	case 0x0: // S0x: Set Filter
		// Not implemented
	case 0x1: // S1x: Set Glissando Control
		state.glissando = (val > 0)
	case 0x2: // S2x: Set Finetune
		// Not implemented
	case 0x3: // S3x: Set Vibrato Waveform
		state.vibratoWave = val
	case 0x4: // S4x: Set Tremolo Waveform
		state.tremoloWave = val
	case 0x8: // S8x: Set Panning
		if tick == 0 {
			state.panning = float64(val) / 15.0
		}
	case 0xA: // SAx: Stereo Control
		if tick == 0 {
			state.stereo = float64(val) / 15.0
		}
	case 0xB: // SBx: Pattern loop
		if tick == 0 {
			if val == 0 {
				playerState.patternLoopRow = playerState.row
			} else {
				if playerState.patternLoopCount == 0 {
					playerState.patternLoopCount = int(val)
					*nextRow = playerState.patternLoopRow
				} else {
					playerState.patternLoopCount--
					if playerState.patternLoopCount > 0 {
						*nextRow = playerState.patternLoopRow
					}
				}
			}
		}
	case 0xC: // SCx: Note Cut
		if tick == int(val) {
			state.volume = 0
		}
	case 0xD: // SDx: Note Delay
		if tick < int(val) {
			// requires restructuring note handling
		}
	case 0xE: // SEx: Pattern Delay
		if tick == 0 {
			playerState.patternDelay = int(val) * playerState.speed
		}
	case 0xF: // SFx: Funkrepeat
		// Not implemented
	}
}



func (t *S3MTicker) GetPeriod(period uint16, offset int) uint16 {
	// this is not correct.
	// find the note from the period
	var note, octave int
	for i, p := range s3mPeriodTable {
		if p == period {
			octave = i / 12
			note = i % 12
			break
		}
	}
	note += offset
	if note < 0 {
		note = 0
	}
	if note > 11 {
		note = 11
	}
	return s3mPeriodTable[octave*12+note]
}

func (t *S3MTicker) RenderChannelTick(p *Player, state *channelState, tickBuffer []int, samplesPerTick int) {
	if state.sample == nil || state.period == 0 || state.sample.Length() == 0 || state.sampleIndex == -1 {
		return
	}

	freq := 14317456.0 / float64(state.period)
	step := freq / float64(p.opts.SampleRate)

	is16Bit := state.sample.Flags()&4 != 0
	isStereo := state.sample.Flags()&2 != 0
	bytesPerSample := 1
	if is16Bit {
		bytesPerSample = 2
	}
	numChannels := 1
	if isStereo {
		numChannels = 2
	}

	for i := 0; i < samplesPerTick; i++ {
		sampleLength := float64(state.sample.Length())
		loopBegin := float64(state.sample.LoopStart())
		loopEnd := float64(state.sample.LoopEnd())
		loopLength := loopEnd - loopBegin
		hasLoop := state.sample.Flags()&1 != 0

		if hasLoop && loopLength > 1 {
			if state.samplePos >= loopEnd {
				state.samplePos -= loopLength
			}
		} else {
			if state.samplePos >= sampleLength {
				state.sample = nil
				return
			}
		}

		byteOffset := int(state.samplePos) * bytesPerSample * numChannels
		sampleData := state.sample.Data()
		if byteOffset >= len(sampleData) {
			continue
		}

		offset := i * p.opts.NumChannels
		if isStereo {
			var left, right float64
			if is16Bit {
				if byteOffset/2+1 < len(sampleData) {
					lSample := sampleData[byteOffset/2]
					rSample := sampleData[byteOffset/2+1]
					left = float64(lSample) * state.volume
					right = float64(rSample) * state.volume
				}
			} else {
				lSample := sampleData[byteOffset]
				rSample := sampleData[byteOffset+1]
				left = float64(lSample) * state.volume
				right = float64(rSample) * state.volume
			}
			tickBuffer[offset] += int(left)
			if p.opts.NumChannels > 1 {
				tickBuffer[offset+1] += int(right)
			}
		} else { // Mono
			var sampleValue float64
			if is16Bit {
				if byteOffset/2 < len(sampleData) {
					sample := sampleData[byteOffset/2]
					sampleValue = float64(sample) * state.volume
				}
			} else {
				if byteOffset < len(sampleData) {
					sample := sampleData[byteOffset]
					sampleValue = float64(sample) * state.volume
				}
			}
			left, right := p.pan(p.opts.NumChannels, state.panning, sampleValue)
			tickBuffer[offset] += left
			if p.opts.NumChannels > 1 {
				tickBuffer[offset+1] += right
			}
		}
		state.samplePos += step
	}
}
