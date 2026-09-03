package main

import (
	"strings"
	"testing"
	"time"

	"github.com/Fiend3d/catatui"
)

// TestRender draws the example at sizes from nothing to bigger than a screen,
// with the queue empty, part done and untouched. Rendering outside the area
// given panics in catatui, so this is what keeps the example honest when the
// library changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 8}, {80, 24}, {200, 60}}
	for i, d := range []*downloads{newDownloads(), partlyDone(), finished()} {
		for _, size := range sizes {
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("downloads %d, %dx%d: %v", i, size[0], size[1], err)
			}
			if err := terminal.Draw(func(f *catatui.Frame) { render(f, d) }); err != nil {
				t.Fatalf("downloads %d, %dx%d: %v", i, size[0], size[1], err)
			}
		}
	}
}

// partlyDone is every worker busy, halfway through, with some files still
// queued behind them.
func partlyDone() *downloads {
	d := newDownloads()
	for id := range numWorkers {
		d.next(id)
		d.inProgress[id].progress = 0.5
		d.inProgress[id].startedAt = time.Now().Add(-time.Second)
	}
	return d
}

// finished is the state the app quits in: nothing queued, nothing running.
func finished() *downloads { return &downloads{} }

// TestNextTakesEachDownloadOnce checks the queue hands every download out
// exactly once and then runs dry, which is what makes the app terminate.
func TestNextTakesEachDownloadOnce(t *testing.T) {
	d := newDownloads()
	seen := make(map[int]bool)
	for range numDownloads {
		next, ok := d.next(0)
		if !ok {
			t.Fatalf("queue ran dry after %d downloads, want %d", len(seen), numDownloads)
		}
		if seen[next.id] {
			t.Fatalf("download %d handed out twice", next.id)
		}
		seen[next.id] = true
	}
	if _, ok := d.next(0); ok {
		t.Errorf("queue handed out more than %d downloads", numDownloads)
	}
}

// TestRunToCompletion drives the whole download loop against a test terminal,
// finishing one file at a time, and checks two things: that the run ends, and
// that the viewport still holds the frame that was last drawn into it.
//
// The second is the one that matters. Every finished download calls
// InsertBefore, which scrolls the screen underneath the viewport; if the
// viewport is not repainted in full afterwards, the gauges freeze part-drawn
// and the run looks stuck even though it is not.
func TestRunToCompletion(t *testing.T) {
	backend := catatui.NewTestBackend(40, 20)
	if err := backend.SetCursorPosition(catatui.Position{Y: 12}); err != nil {
		t.Fatal(err)
	}
	terminal, err := catatui.NewTerminalWithViewport(backend, catatui.InlineViewport(8))
	if err != nil {
		t.Fatal(err)
	}

	// Buffered queues stand in for the worker goroutines: the test takes the
	// file out itself and reports it finished.
	d := newDownloads()
	workers := make([]worker, numWorkers)
	for i := range workers {
		workers[i] = worker{id: i, queue: make(chan download, 1)}
	}
	for _, w := range workers {
		if next, ok := d.next(w.id); ok {
			w.queue <- next
		}
	}

	finished := 0
	for step := 0; ; step++ {
		if step > numDownloads*2 {
			t.Fatalf("still running after %d steps, %d downloads finished", step, finished)
		}
		id := busyWorker(d)
		if id < 0 {
			t.Fatalf("no worker busy after %d downloads, and the run had not ended", finished)
		}
		<-workers[id].queue
		p := progress{workerID: id, downloadID: d.inProgress[id].id, done: true}

		done, err := apply(terminal, d, workers, p)
		if err != nil {
			t.Fatal(err)
		}
		finished++
		if err := terminal.Draw(func(f *catatui.Frame) { render(f, d) }); err != nil {
			t.Fatal(err)
		}
		if done {
			break
		}
	}

	if finished != numDownloads {
		t.Errorf("finished %d downloads, want %d", finished, numDownloads)
	}

	// "Done !" goes in immediately above the viewport, and the frame drawn
	// after it has to be all there.
	screen := backend.Buffer()
	if got := rowText(screen, 11); !strings.Contains(got, "Done !") {
		t.Errorf("row above the viewport is %q, want the closing line", got)
	}
	if got := rowText(screen, 12); !strings.Contains(got, "Progress") {
		t.Errorf("viewport's first row is %q, want the title of the frame just drawn", got)
	}

	// On the way out the viewport comes off the screen and the cursor lands on
	// its first row, so the shell prompt follows "Done !" with nothing in
	// between. Anything left behind here is a run of blank lines the user has
	// to scroll past.
	if err := finish(terminal); err != nil {
		t.Fatal(err)
	}
	cursor, err := backend.GetCursorPosition()
	if err != nil {
		t.Fatal(err)
	}
	if want := (catatui.Position{X: 0, Y: 12}); cursor != want {
		t.Errorf("cursor left at %+v, want %+v: the row right after \"Done !\"", cursor, want)
	}
	for y := cursor.Y; y < screen.Area.Bottom(); y++ {
		if got := rowText(screen, y); strings.TrimSpace(got) != "" {
			t.Errorf("row %d is %q, want the viewport blanked from the cursor down", y, got)
		}
	}
	if got := rowText(screen, 11); !strings.Contains(got, "Done !") {
		t.Errorf("row above the cursor is %q, want \"Done !\" still there", got)
	}
}

// busyWorker is the first worker with a download in flight, or -1.
func busyWorker(d *downloads) int {
	for i, p := range d.inProgress {
		if p != nil {
			return i
		}
	}
	return -1
}

// rowText reads one row of the screen back as a string.
func rowText(buf *catatui.Buffer, y uint16) string {
	var b strings.Builder
	for x := buf.Area.Left(); x < buf.Area.Right(); x++ {
		b.WriteString(buf.Get(x, y).GetSymbol())
	}
	return b.String()
}
