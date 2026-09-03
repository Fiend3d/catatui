package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
)

// The tests never reach the network. The fetch takes the URL it is given, so
// they point it at a server in this process instead, which is also the only way
// to test the failure paths.

// samplePayload is the shape of the API's answer, cut down to the fields the
// table uses plus one it ignores.
const samplePayload = `[
	{"number": 1234, "title": "Add a widget", "html_url": "https://github.com/ratatui/ratatui/pull/1234", "state": "open"},
	{"number": 1235, "title": "Fix the layout", "html_url": "https://github.com/ratatui/ratatui/pull/1235", "state": "open"}
]`

// TestRender draws the app at sizes from nothing to bigger than a screen, in
// each of the states the fetch can be in. Rendering outside the area given
// panics in catatui, so this is what keeps the example honest when the library
// changes.
func TestRender(t *testing.T) {
	sizes := [][2]uint16{{0, 0}, {1, 1}, {3, 2}, {10, 4}, {40, 12}, {80, 24}, {200, 60}}

	states := map[string]func(*pullRequestList){
		"idle":    func(*pullRequestList) {},
		"loading": func(l *pullRequestList) { l.setState(stateLoading) },
		"loaded":  func(l *pullRequestList) { l.onLoad(samplePulls(t)) },
		"a long error": func(l *pullRequestList) {
			l.onError(&testError{strings.Repeat("rate limit exceeded. ", 20)})
		},
		"scrolled past the end": func(l *pullRequestList) {
			l.onLoad(samplePulls(t))
			for range 10 {
				l.scrollDown()
			}
		},
		"a very long title": func(l *pullRequestList) {
			l.onLoad([]pullRequest{{"1", strings.Repeat("long ", 100), strings.Repeat("u", 300)}})
		},
	}

	for name, setup := range states {
		for _, size := range sizes {
			a := &app{pulls: newPullRequestList("")}
			setup(a.pulls)
			terminal, err := catatui.NewTerminal(catatui.NewTestBackend(size[0], size[1]))
			if err != nil {
				t.Fatalf("%s at %dx%d: %v", name, size[0], size[1], err)
			}
			if err := terminal.Draw(a.draw); err != nil {
				t.Fatalf("%s at %dx%d: %v", name, size[0], size[1], err)
			}
		}
	}
}

// TestTheTableShowsWhatWasFetched checks the three columns end up on screen in
// the order the widths describe.
func TestTheTableShowsWhatWasFetched(t *testing.T) {
	l := newPullRequestList("")
	l.onLoad(samplePulls(t))

	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 80, 6))
	l.Render(buf.Area, buf)

	text := buf.String()
	for _, want := range []string{"Pull Requests", "Loaded", "1234", "Add a widget",
		"github.com/ratatui/ratatui/pull/1234", "j/k to scroll, q to quit"} {
		if !strings.Contains(text, want) {
			t.Errorf("the table reads\n%s\nwant it to contain %q", text, want)
		}
	}
}

// TestTheStatusIsShown checks the header names each state, which is the only
// sign the fetch is doing anything.
func TestTheStatusIsShown(t *testing.T) {
	l := newPullRequestList("")
	if got := l.status(); got != "Idle" {
		t.Errorf("a list that has not started says %q", got)
	}

	l.setState(stateLoading)
	if got := l.status(); got != "Loading" {
		t.Errorf("a list waiting on the API says %q", got)
	}

	l.onError(&testError{"403 Forbidden: rate limit exceeded"})
	if got := l.status(); !strings.Contains(got, "rate limit exceeded") {
		t.Errorf("a failed fetch says %q, want the reason in it", got)
	}

	l.onLoad(samplePulls(t))
	if got := l.status(); got != "Loaded" {
		t.Errorf("a finished fetch says %q", got)
	}
}

// TestAFetchFillsTheList runs the whole path — goroutine, HTTP, decode — against
// a server in this process.
func TestAFetchFillsTheList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("the request carries no User-Agent, which GitHub rejects")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(samplePayload))
	}))
	defer server.Close()

	l := newPullRequestList(server.URL)
	l.fetch(context.Background())

	waitFor(t, l, stateLoaded)
	if len(l.pulls) != 2 {
		t.Fatalf("the fetch left %d pull requests, want 2", len(l.pulls))
	}
	if got := l.pulls[0]; got.id != "1234" || got.title != "Add a widget" {
		t.Errorf("the first row is %+v", got)
	}
	if _, ok := l.table.Selected(); !ok {
		t.Errorf("nothing is selected after a fetch that found something")
	}
}

// TestAFailedFetchSaysWhy checks a refusal reaches the header rather than
// leaving the app on Loading forever, and that the API's own explanation
// survives — a rate limit is the one everybody meets, and the status alone does
// not say so.
func TestAFailedFetchSaysWhy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
	}))
	defer server.Close()

	l := newPullRequestList(server.URL)
	l.fetch(context.Background())

	waitFor(t, l, stateError)
	if !strings.Contains(l.err, "API rate limit exceeded") {
		t.Errorf("the error reads %q, want the API's explanation in it", l.err)
	}
	if !strings.Contains(l.err, "403") {
		t.Errorf("the error reads %q, want the status in it", l.err)
	}
}

// TestGarbageIsAnErrorNotAPanic checks an answer that is not the JSON expected
// leaves the app in a state it can draw.
func TestGarbageIsAnErrorNotAPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html>not json</html>"))
	}))
	defer server.Close()

	l := newPullRequestList(server.URL)
	l.fetch(context.Background())

	waitFor(t, l, stateError)
	buf := catatui.NewBuffer(catatui.NewRect(0, 0, 40, 5))
	l.Render(buf.Area, buf) // would panic if the error left the widget half built
}

// TestIncompleteRowsAreDropped checks a pull request missing one of the three
// fields is left out rather than drawn as a blank row.
func TestIncompleteRowsAreDropped(t *testing.T) {
	payload := `[
		{"number": 1, "title": "Complete", "html_url": "https://example.com/1"},
		{"number": 2, "title": "No URL"},
		{"title": "No number", "html_url": "https://example.com/3"},
		{"number": 4, "html_url": "https://example.com/4"}
	]`
	pulls, err := decodePullRequests(strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if len(pulls) != 1 || pulls[0].id != "1" {
		t.Errorf("decoding gave %+v, want only the complete row", pulls)
	}
}

// TestScrollingMovesTheSelection checks j and k reach the table state through
// the lock.
func TestScrollingMovesTheSelection(t *testing.T) {
	a := &app{pulls: newPullRequestList("")}
	a.pulls.onLoad(samplePulls(t))

	key := func(r rune) term.Event {
		return term.Event{Kind: term.EventKey, Key: term.KeyRune, Rune: r}
	}

	a.handle(key('j'))
	if got, ok := a.pulls.table.Selected(); !ok || got != 1 {
		t.Errorf("j selected %v (selected: %v), want the second row", got, ok)
	}
	a.handle(key('k'))
	if got, ok := a.pulls.table.Selected(); !ok || got != 0 {
		t.Errorf("k selected %v (selected: %v), want back at the first row", got, ok)
	}

	a.handle(key('q'))
	if !a.quit {
		t.Errorf("q did not quit")
	}
}

// samplePulls is the sample payload decoded, for the tests that do not need a
// server.
func samplePulls(t *testing.T) []pullRequest {
	t.Helper()
	pulls, err := decodePullRequests(strings.NewReader(samplePayload))
	if err != nil {
		t.Fatal(err)
	}
	return pulls
}

// waitFor blocks until the fetch reaches a state, or gives up. The fetch is in
// a goroutine, so there is nothing else to synchronise on — the app itself just
// redraws until the answer appears.
func waitFor(t *testing.T, l *pullRequestList, want loadingState) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		l.mu.Lock()
		state := l.state
		l.mu.Unlock()
		if state == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("the fetch never reached %v", want)
}

// testError is an error with a message chosen by the test.
type testError struct{ message string }

func (e *testError) Error() string { return e.message }
