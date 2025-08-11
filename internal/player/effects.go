package player

// PeriodGetter is an interface for getting a period value for a note.
type PeriodGetter interface {
	GetPeriod(basePeriod uint16, semitoneOffset int) uint16
}

// sin_table is a 32-entry sine table for vibrato and tremolo effects.
var sin_table = [32]float64{
	0, 24, 49, 74, 97, 120, 141, 161, 180, 197, 212, 224, 235, 244, 250, 253,
	255, 253, 250, 244, 235, 224, 212, 197, 180, 161, 141, 120, 97, 74, 49, 24,
}

// applyArpeggio applies the arpeggio effect to a channel's state.
func applyArpeggio(state *channelState, param byte, tick int, getter PeriodGetter) {
	if tick == 0 {
		return
	}

	basePeriod := state.notePeriod
	x := int(param >> 4)
	y := int(param & 0x0F)

	switch tick % 3 {
	case 0:
		state.period = basePeriod
	case 1:
		state.period = getter.GetPeriod(basePeriod, x)
	case 2:
		state.period = getter.GetPeriod(basePeriod, y)
	}
}

// applyVibrato applies the vibrato effect to a channel's state.
func applyVibrato(state *channelState) {
	if state.vibratoDepth == 0 {
		return
	}
	var delta float64
	pos := state.vibratoPos
	wave := state.vibratoWave & 3
	switch wave {
	case 0: // Sine
		delta = sin_table[pos&31]
		if pos >= 32 {
			delta = -delta
		}
	case 1: // Ramp down (sawtooth)
		delta = float64(255 - (pos * 4))
	case 2: // Square
		if pos < 32 {
			delta = 255
		} else {
			delta = -255
		}
	}
	delta = delta * float64(state.vibratoDepth) / 128.0
	state.period += uint16(delta)
	state.vibratoPos = (state.vibratoPos + state.vibratoSpeed) & 63
}

// applyTremolo applies the tremolo effect to a channel's state.
func applyTremolo(state *channelState) {
	if state.tremoloDepth == 0 {
		return
	}
	var delta float64
	pos := state.tremoloPos
	wave := state.tremoloWave & 3
	switch wave {
	case 0: // Sine
		delta = sin_table[pos&31]
		if pos >= 32 {
			delta = -delta
		}
	case 1: // Ramp down (sawtooth)
		delta = float64(255 - (pos * 4))
	case 2: // Square
		if pos < 32 {
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

// applyVolumeSlide applies the volume slide effect to a channel's state.
func applyVolumeSlide(state *channelState, param byte, tick int, isS3M bool) {
	if param > 0 {
		state.lastVolSlide = param
	} else {
		param = state.lastVolSlide
	}

	x := param >> 4
	y := param & 0x0F

	if isS3M {
		if y == 0xF && x > 0 { // Fine slide up
			if tick == 0 {
				state.volume += float64(x) / 64.0
			}
			return
		} else if x == 0xF && y > 0 { // Fine slide down
			if tick == 0 {
				state.volume -= float64(y) / 64.0
			}
			return
		}
	}

	if tick > 0 {
		if x > 0 {
			state.volume += float64(x) / 64.0
		} else {
			state.volume -= float64(y) / 64.0
		}
	}

	if state.volume > 1.0 {
		state.volume = 1.0
	}
	if state.volume < 0 {
		state.volume = 0
	}
}

// applyPortamentoUp applies the portamento up effect to a channel's state.
func applyPortamentoUp(state *channelState, param byte, tick int, isS3M bool) {
	if tick == 0 {
		if isS3M {
			x := param >> 4
			y := param & 0x0F
			if x == 0xE { // Extra fine
				state.period -= uint16(y)
			} else if x == 0xF { // Fine
				state.period -= uint16(y) * 4
			}
		}
		return
	}

	var speed uint16
	if isS3M {
		if param > 0 {
			state.lastPorta = param
		}
		speed = uint16(state.lastPorta) * 4
	} else {
		if param > 0 {
			state.lastPortaUp = param
		}
		speed = uint16(state.lastPortaUp)
	}
	state.period -= speed
}

// applyPortamentoDown applies the portamento down effect to a channel's state.
func applyPortamentoDown(state *channelState, param byte, tick int, isS3M bool) {
	if tick == 0 {
		if isS3M {
			x := param >> 4
			y := param & 0x0F
			if x == 0xE { // Extra fine
				state.period += uint16(y)
			} else if x == 0xF { // Fine
				state.period += uint16(y) * 4
			}
		}
		return
	}

	var speed uint16
	if isS3M {
		if param > 0 {
			state.lastPorta = param
		}
		speed = uint16(state.lastPorta) * 4
	} else {
		if param > 0 {
			state.lastPortaDown = param
		}
		speed = uint16(state.lastPortaDown)
	}
	state.period += speed
}
