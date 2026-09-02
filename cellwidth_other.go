//go:build !windows

package catatui

// clusterWidth is how far a terminal advances after printing one grapheme
// cluster: its unicode width, with a column added back for each halfwidth
// katakana sound mark.
//
// The sound marks are the one correction. Terminals render them in a column of
// their own even though Unicode scores them as zero-width combining marks.
//
// Everything else is plain unicode width, per cluster, which is what xterm,
// VTE, kitty and the rest agree on closely enough to draw with. It is also what
// ratatui measures, so a widget ported from it lays out the same here.
//
// Windows is the exception, and has a policy of its own; see
// cellwidth_windows.go for why a cluster there is not one advance but several.
func clusterWidth(cluster string, unisegWidth int) uint16 {
	w := uint16(min(max(unisegWidth, 0), maxU16))
	return SatAdd(w, countHalfwidthSoundMarks(cluster))
}
