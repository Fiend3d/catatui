// Command inline shows the inline viewport: a fixed block of the terminal that
// redraws in place while finished lines scroll away above it.
//
//	go run ./examples/apps/inline
//
// Four workers download ten files between them. Each finished download is
// written into the scrollback with Terminal.InsertBefore, so it survives
// everything the viewport draws afterwards, and stays on screen when the
// program exits.
//
// Press q to quit.
//
// Port of examples/apps/inline @ ratatui-v0.30.2
package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// numDownloads is how many files are downloaded in all, and numWorkers how many
// are in flight at once.
const (
	numDownloads = 10
	numWorkers   = 4
)

// download is a file waiting to be fetched. The size is made up, and is only
// there to decide how long the worker sleeps for.
type download struct {
	id   int
	size int
}

// downloadInProgress is a download a worker has started but not finished.
type downloadInProgress struct {
	id        int
	startedAt time.Time
	progress  float64
}

// downloads is the queue and what is being worked on right now.
//
// ratatui's example keeps the running downloads in a map from worker to
// download; here they are a slice indexed by worker, which is already in worker
// order and so needs no sorting to render.
type downloads struct {
	pending    []download
	inProgress [numWorkers]*downloadInProgress
}

// next hands the given worker the next pending download, if there is one.
func (d *downloads) next(workerID int) (download, bool) {
	if len(d.pending) == 0 {
		return download{}, false
	}
	next := d.pending[0]
	d.pending = d.pending[1:]
	d.inProgress[workerID] = &downloadInProgress{id: next.id, startedAt: time.Now()}
	return next, true
}

// running counts the downloads currently in flight.
func (d *downloads) running() int {
	var n int
	for _, p := range d.inProgress {
		if p != nil {
			n++
		}
	}
	return n
}

// newDownloads queues up the whole list, each with a random size.
func newDownloads() *downloads {
	pending := make([]download, numDownloads)
	for i := range pending {
		pending[i] = download{id: i, size: rand.IntN(1000)}
	}
	return &downloads{pending: pending}
}

// progress is a worker reporting how far along its download is.
type progress struct {
	workerID   int
	downloadID int
	// ratio is how much of the file is done, from 0 to 1. It is only
	// meaningful while done is false.
	ratio float64
	done  bool
}

// worker fetches whatever is sent to it, reporting as it goes.
type worker struct {
	id    int
	queue chan download
}

// startWorkers starts one goroutine per worker, all reporting to the same
// channel. Each one runs until its queue is closed, which is what stops them
// when the program is done.
func startWorkers(updates chan<- progress) []worker {
	workers := make([]worker, numWorkers)
	for id := range workers {
		queue := make(chan download)
		workers[id] = worker{id: id, queue: queue}
		go func() {
			for d := range queue {
				remaining := d.size
				for remaining > 0 {
					wait := min(remaining, 10)
					time.Sleep(time.Duration(wait) * 10 * time.Millisecond)
					remaining = max(remaining-10, 0)
					ratio := float64(d.size-remaining) / float64(d.size)
					updates <- progress{workerID: id, downloadID: d.id, ratio: ratio}
				}
				updates <- progress{workerID: id, downloadID: d.id, done: true}
			}
		}()
	}
	return workers
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run drives the downloads until they are all finished, or until q is pressed.
func run() error {
	defer term.RecoverAndRestore()

	// An inline viewport owns eight rows at the cursor and nothing else, so
	// whatever was on screen before stays visible above them. It implies no
	// alternate screen: that is the whole point of drawing inline, and it is
	// what lets InsertBefore push finished lines into the scrollback.
	terminal, restore, err := term.Init(term.WithViewport(catatui.InlineViewport(8)))
	if err != nil {
		return err
	}
	defer func() {
		_ = finish(terminal)
		restore()
	}()

	// The reader has to start after Init: placing an inline viewport means
	// asking the terminal where its cursor is, and that reply comes back on
	// the same input this reader would be draining.
	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	updates := make(chan progress)
	workers := startWorkers(updates)
	defer func() {
		// Closing the queues is what ends the workers. One of them may be
		// part-way through a file and blocked reporting progress to a loop
		// that has stopped listening; that is fine here, because the program
		// is exiting. A longer-lived app would give them a way to give up.
		for _, w := range workers {
			close(w.queue)
		}
	}()

	d := newDownloads()
	for _, w := range workers {
		if next, ok := d.next(w.id); ok {
			w.queue <- next
		}
	}

	// The clock is what keeps the "(123ms)" counters moving while nothing else
	// is happening.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	redraw := true
	for {
		if redraw {
			if err := terminal.Draw(func(f *catatui.Frame) { render(f, d) }); err != nil {
				return err
			}
		}
		redraw = true

		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			// A resize needs nothing but a redraw: Draw resizes the terminal
			// to match the backend before it renders.
			if ev.Kind == term.EventKey && (ev.IsRune('q') || ev.IsCtrl('c')) {
				return nil
			}
		case <-ticker.C:
		case p := <-updates:
			done, err := apply(terminal, d, workers, p)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
			// A progress report on its own does not earn a frame; the tick
			// draws them, which keeps the redraws to five a second.
			redraw = p.done
		}
	}
}

// finish takes the viewport back off the screen on the way out.
//
// The viewport is the program's own scratch space, not scrollback: what is
// worth keeping has already been written above it by InsertBefore. So it is
// blanked, and the cursor left on its first row, which puts the shell prompt
// directly under the last line inserted rather than below eight blank ones.
// restore flushes both.
func finish(terminal *catatui.Terminal) error {
	if err := terminal.Clear(); err != nil {
		return err
	}
	return terminal.SetCursorPosition(terminal.ViewportArea().AsPosition())
}

// apply folds one worker report into the downloads, printing a line above the
// viewport for each one that finishes and handing the worker its next file. It
// reports whether every download is now done.
func apply(terminal *catatui.Terminal, d *downloads, workers []worker, p progress) (bool, error) {
	current := d.inProgress[p.workerID]
	if current == nil {
		return false, nil
	}
	if !p.done {
		current.progress = p.ratio
		return false, nil
	}

	d.inProgress[p.workerID] = nil
	if err := insertFinishedLine(terminal, current); err != nil {
		return false, err
	}

	if next, ok := d.next(p.workerID); ok {
		workers[p.workerID].queue <- next
		return false, nil
	}
	if d.running() > 0 {
		return false, nil
	}
	return true, terminal.InsertBefore(1, func(buf *catatui.Buffer) {
		buf.SetString(buf.Area.X, buf.Area.Y, "Done !", catatui.NewStyle())
	})
}

// insertFinishedLine writes one line into the scrollback above the viewport.
func insertFinishedLine(terminal *catatui.Terminal, d *downloadInProgress) error {
	line := catatui.NewLine(
		catatui.NewSpan("Finished "),
		catatui.NewStyledSpan(fmt.Sprintf("download %d", d.id),
			catatui.NewStyle().AddModifier(catatui.ModifierBold)),
		catatui.NewSpan(fmt.Sprintf(" in %dms", time.Since(d.startedAt).Milliseconds())),
	)
	return terminal.InsertBefore(1, func(buf *catatui.Buffer) {
		buf.SetLine(buf.Area.X, buf.Area.Y, line, buf.Area.Width)
	})
}
