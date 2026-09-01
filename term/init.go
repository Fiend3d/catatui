package term

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Fiend3d/catatui"
)

// Option configures Init.
type Option func(*config)

type config struct {
	in, out     *os.File
	altScreen   bool
	mouse       bool
	paste       bool
	focus       bool
	cursorShape CursorShape
	hasShape    bool
	viewport    catatui.Viewport
	hasVieport  bool
}

// WithMouse enables mouse reporting, so that EventMouse events arrive.
func WithMouse() Option { return func(c *config) { c.mouse = true } }

// WithBracketedPaste enables bracketed paste, so that pasted text arrives as a
// single EventPaste instead of a flood of key presses.
func WithBracketedPaste() Option { return func(c *config) { c.paste = true } }

// WithFocusReporting enables focus events.
func WithFocusReporting() Option { return func(c *config) { c.focus = true } }

// WithoutAlternateScreen keeps the application on the main screen, so that what
// it draws stays in the scrollback after it exits.
func WithoutAlternateScreen() Option { return func(c *config) { c.altScreen = false } }

// WithCursorShape sets how the terminal draws the cursor for the life of the
// program. The restore function puts it back.
//
// To change the shape while running — a text field switching to a bar while it
// has focus, say — use BackendOf and Backend.SetCursorShape instead.
func WithCursorShape(s CursorShape) Option {
	return func(c *config) { c.cursorShape, c.hasShape = s, true }
}

// WithViewport draws into the given viewport rather than the whole terminal.
func WithViewport(v catatui.Viewport) Option {
	return func(c *config) { c.viewport, c.hasVieport = v, true }
}

// WithIO uses the given files instead of os.Stdin and os.Stdout.
func WithIO(in, out *os.File) Option {
	return func(c *config) { c.in, c.out = in, out }
}

// Init puts the terminal into raw mode on the alternate screen and returns a
// ready catatui.Terminal along with the function that undoes it all.
//
// The restore function is safe to call more than once and is also installed to
// run if the program panics, because a panic that leaves the terminal in raw
// mode on the alternate screen leaves the user with an unusable shell and no
// visible stack trace.
//
//	terminal, restore, err := term.Init(term.WithMouse())
//	if err != nil {
//		return err
//	}
//	defer restore()
func Init(opts ...Option) (*catatui.Terminal, func(), error) {
	cfg := config{in: os.Stdin, out: os.Stdout, altScreen: true}
	for _, o := range opts {
		o(&cfg)
	}

	state, err := enterRawMode(cfg.in, cfg.out)
	if err != nil {
		return nil, nil, err
	}

	backend := NewBackend(cfg.out)
	if cfg.altScreen {
		backend.w.esc(seqAltScreenOn)
	}
	backend.w.esc(seqCursorHide)
	if cfg.mouse {
		backend.w.esc(seqMouseOn)
	}
	if cfg.paste {
		backend.w.esc(seqPasteOn)
	}
	if cfg.focus {
		backend.w.esc(seqFocusOn)
	}
	if cfg.hasShape {
		backend.w.setCursorShape(cfg.cursorShape)
	}
	backend.w.invalidateCursor()
	if err := backend.Flush(); err != nil {
		_ = exitRawMode(state)
		return nil, nil, err
	}

	var restoreOnce sync.Once
	restore := func() {
		restoreOnce.Do(func() {
			// The cursor shape is always reset, not just when this program set
			// it: a program that changed it while running through
			// Backend.SetCursorShape would otherwise leave the user's shell
			// with the wrong cursor, and there is no way to query what the
			// shape was before.
			backend.w.setCursorShape(CursorDefault)
			if cfg.focus {
				backend.w.esc(seqFocusOff)
			}
			if cfg.paste {
				backend.w.esc(seqPasteOff)
			}
			if cfg.mouse {
				backend.w.esc(seqMouseOff)
			}
			backend.w.esc(seqCursorShow)
			backend.w.csi("0m")
			if cfg.altScreen {
				backend.w.esc(seqAltScreenOff)
			}
			_ = backend.Flush()
			_ = exitRawMode(state)
		})
	}

	// A panic must not leave the terminal unusable. The handler restores and
	// re-panics so the stack trace still reaches the user, now legibly.
	installPanicRestore(restore)

	var terminal *catatui.Terminal
	if cfg.hasVieport {
		terminal, err = catatui.NewTerminalWithViewport(backend, cfg.viewport)
	} else {
		terminal, err = catatui.NewTerminal(backend)
	}
	if err != nil {
		restore()
		return nil, nil, err
	}
	return terminal, restore, nil
}

var (
	panicRestoreMu sync.Mutex
	panicRestores  []func()
)

// installPanicRestore registers a restore function to run if the process
// panics. Go has no panic hook, so RecoverAndRestore has to be deferred by the
// caller in the goroutine that might panic; this registry is what makes that
// one call enough regardless of how many terminals were opened.
func installPanicRestore(restore func()) {
	panicRestoreMu.Lock()
	defer panicRestoreMu.Unlock()
	panicRestores = append(panicRestores, restore)
}

// RecoverAndRestore restores every terminal opened by Init and then re-panics.
//
// Defer it in main and in any goroutine that draws:
//
//	defer term.RecoverAndRestore()
func RecoverAndRestore() {
	if r := recover(); r != nil {
		RestoreAll()
		panic(r)
	}
}

// RestoreAll restores every terminal opened by Init. Init's own restore
// function is idempotent, so calling both is harmless.
func RestoreAll() {
	panicRestoreMu.Lock()
	fns := append([]func(){}, panicRestores...)
	panicRestoreMu.Unlock()
	for _, f := range fns {
		f()
	}
}

// EventReader turns a terminal's byte stream into Events.
//
// It runs a goroutine that reads input and, on platforms without a resize
// signal, another that watches the terminal size. Events arrive on the channel
// returned by Events.
type EventReader struct {
	events chan Event
	stop   chan struct{}
	closer sync.Once
	err    error
	errMu  sync.Mutex
}

// escapeTimeout is how long a trailing ESC waits for the rest of a sequence
// before being reported as the escape key.
//
// Terminal input is genuinely ambiguous here: ESC alone is the escape key, and
// ESC is also the first byte of every arrow and function key. The only way to
// tell them apart is to wait, and this is the usual compromise: long enough that
// a real sequence always arrives whole, short enough that pressing escape does
// not feel laggy.
const escapeTimeout = 50 * time.Millisecond

// NewEventReader starts reading events from in, reporting resizes measured on
// sizeSrc. Pass the same file for both when reading from a terminal.
func NewEventReader(in *os.File, sizeSrc *os.File) *EventReader {
	r := &EventReader{
		events: make(chan Event, 128),
		stop:   make(chan struct{}),
	}
	go r.readLoop(in)
	if sizeSrc != nil {
		go watchResize(sizeSrc, func(s catatui.Size) {
			r.send(Event{Kind: EventResize, Size: s})
		}, r.stop)
	}
	return r
}

// Events returns the channel events arrive on. It is closed when the input
// stream ends or Close is called.
func (r *EventReader) Events() <-chan Event { return r.events }

// Err returns the error that ended reading, if any. io.EOF is reported as nil.
func (r *EventReader) Err() error {
	r.errMu.Lock()
	defer r.errMu.Unlock()
	return r.err
}

// Close stops reading.
func (r *EventReader) Close() {
	r.closer.Do(func() { close(r.stop) })
}

func (r *EventReader) send(ev Event) {
	select {
	case r.events <- ev:
	case <-r.stop:
	}
}

// readLoop reads bytes and parses them into events.
//
// The pending buffer holds bytes that did not yet form a complete sequence. A
// lone trailing ESC is resolved by the timeout described at escapeTimeout.
func (r *EventReader) readLoop(in *os.File) {
	defer close(r.events)

	type readResult struct {
		data []byte
		err  error
	}
	reads := make(chan readResult, 1)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := in.Read(buf)
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			select {
			case reads <- readResult{chunk, err}:
			case <-r.stop:
				return
			}
			if err != nil {
				return
			}
		}
	}()

	var pending []byte
	var timer *time.Timer
	var timeout <-chan time.Time

	for {
		select {
		case <-r.stop:
			return

		case res := <-reads:
			pending = append(pending, res.data...)
			pending = r.drain(pending)
			if res.err != nil {
				// Flush a trailing escape before giving up on the stream.
				if len(pending) == 1 && pending[0] == 0x1b {
					r.send(Event{Kind: EventKey, Key: KeyEscape})
				}
				if !errors.Is(res.err, io.EOF) {
					r.errMu.Lock()
					r.err = res.err
					r.errMu.Unlock()
				}
				return
			}
			// Arm the escape timeout only while an incomplete sequence is held.
			if len(pending) > 0 {
				if timer == nil {
					timer = time.NewTimer(escapeTimeout)
				} else {
					timer.Reset(escapeTimeout)
				}
				timeout = timer.C
			} else {
				timeout = nil
			}

		case <-timeout:
			timeout = nil
			if len(pending) == 0 {
				continue
			}
			if pending[0] == 0x1b {
				// Nothing more came: this really was the escape key. If other
				// bytes follow it, they are an unrecognised sequence and get
				// dropped so the stream can resynchronise.
				r.send(Event{Kind: EventKey, Key: KeyEscape})
				pending = pending[1:]
			}
			pending = r.drain(pending)
		}
	}
}

// drain parses as many complete events as it can, returning the leftover bytes.
func (r *EventReader) drain(pending []byte) []byte {
	for len(pending) > 0 {
		ev, n, ok := parse(pending)
		if n == 0 && !ok {
			// Incomplete; wait for more bytes.
			return pending
		}
		pending = pending[n:]
		if ok {
			r.send(ev)
		}
	}
	return nil
}
