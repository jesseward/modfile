package player

// PlayerStateUpdate contains information about the player's current state,
// such as the current order, pattern, and row.
type PlayerStateUpdate struct {
	Order   int
	Pattern int
	Row     int
	Speed   int
	BPM     int
}
