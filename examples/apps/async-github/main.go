// Command async-github fetches ratatui's open pull requests in the background
// and shows them in a table while the UI keeps running.
//
//	go run ./examples/apps/async-github
//
// j/k or the arrows scroll, q quits.
//
// The point is the shape of it: the fetch runs in a goroutine and writes into
// state the render loop reads under a lock, so a request that takes a second
// does not hold the frame up. The header says which state it is in — Idle,
// Loading, Loaded or the error — and the table fills in when the answer lands.
//
// There is no message passing here on purpose. Shared state is the simplest
// thing that works for a widget that fetches once, and a channel is what you
// reach for when the fetch has more to say than "here is the answer" — see the
// comment on pullRequestList.
//
// It hits the GitHub API unauthenticated, which is rate limited to sixty
// requests an hour per address. Set GITHUB_TOKEN to raise that; any token with
// no scopes will do, since the repository is public.
//
// Port of examples/apps/async-github @ ratatui-v0.30.2
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Fiend3d/catatui"
	"github.com/Fiend3d/catatui/term"
	"github.com/Fiend3d/catatui/widgets"
)

// pullsURL is the API call the example makes: the open pull requests of
// ratatui's own repository, most recently updated first.
const pullsURL = "https://api.github.com/repos/ratatui/ratatui/pulls?sort=updated&direction=desc"

// framesPerSecond is how often the screen is redrawn while the fetch is out.
// Nothing tells the loop that the answer has arrived, so it looks for itself.
const framesPerSecond = 60

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	defer term.RecoverAndRestore()

	terminal, restore, err := term.Init()
	if err != nil {
		return err
	}
	defer restore()

	events := term.NewEventReader(os.Stdin, os.Stdout)
	defer events.Close()

	ticker := time.NewTicker(time.Second / framesPerSecond)
	defer ticker.Stop()

	// The fetch outlives nothing: cancelling on the way out stops a request
	// still in flight from holding the program open.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := &app{pulls: newPullRequestList(pullsURL)}
	a.pulls.fetch(ctx)

	for !a.quit {
		if err := terminal.Draw(a.draw); err != nil {
			return err
		}
		select {
		case ev, ok := <-events.Events():
			if !ok {
				return events.Err()
			}
			a.handle(ev)
		case <-ticker.C:
		}
	}
	return nil
}

// app is a title bar over the list.
type app struct {
	quit  bool
	pulls *pullRequestList
}

func (a *app) handle(ev term.Event) {
	if ev.Kind != term.EventKey {
		return
	}
	switch {
	case ev.IsRune('q'), ev.IsKey(term.KeyEscape):
		a.quit = true
	case ev.IsRune('j'), ev.IsKey(term.KeyDown):
		a.pulls.scrollDown()
	case ev.IsRune('k'), ev.IsKey(term.KeyUp):
		a.pulls.scrollUp()
	}
}

func (a *app) draw(f *catatui.Frame) {
	rows := catatui.VerticalLayout(catatui.Length(1), catatui.Fill(1)).Split(f.Area())

	f.RenderWidget(
		catatui.LineFromStyledString("catatui async example",
			catatui.NewStyle().AddModifier(catatui.ModifierBold)).Centered(),
		rows[0])
	f.RenderWidget(a.pulls, rows[1])
}

// loadingState is where the fetch has got to.
type loadingState int

const (
	stateIdle loadingState = iota
	stateLoading
	stateLoaded
	stateError
)

// pullRequest is the three fields of a pull request the table shows.
type pullRequest struct {
	id    string
	title string
	url   string
}

// pullRequestList is a widget over state it shares with the goroutine filling
// it in. The widget is a pointer, so passing it to the goroutine and rendering
// it are the same object — this is what ratatui does with an Arc<RwLock<..>>
// and a clone.
//
// A goroutine writing into a mutex is the right shape for a widget that fetches
// once. Give it something to say beyond the answer — progress, a stream of
// updates, a retry — and a channel into the event loop is the better shape,
// because then the loop can redraw when there is news instead of looking sixty
// times a second.
type pullRequestList struct {
	url string

	mu    sync.Mutex
	state loadingState
	err   string
	pulls []pullRequest
	table widgets.TableState
}

func newPullRequestList(url string) *pullRequestList {
	return &pullRequestList{url: url}
}

// fetch starts the request in the background and returns straight away.
func (l *pullRequestList) fetch(ctx context.Context) {
	l.setState(stateLoading)
	go func() {
		pulls, err := fetchPullRequests(ctx, l.url)
		if err != nil {
			l.onError(err)
			return
		}
		l.onLoad(pulls)
	}()
}

func (l *pullRequestList) setState(state loadingState) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = state
}

func (l *pullRequestList) onLoad(pulls []pullRequest) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = stateLoaded
	l.pulls = append(l.pulls, pulls...)
	if len(l.pulls) > 0 {
		l.table.Select(0)
	}
}

func (l *pullRequestList) onError(err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.state = stateError
	l.err = err.Error()
}

func (l *pullRequestList) scrollDown() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.table.ScrollDownBy(1)
}

func (l *pullRequestList) scrollUp() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.table.ScrollUpBy(1)
}

// status is the line drawn in the top right, which is the only sign the fetch
// is doing anything at all.
func (l *pullRequestList) status() string {
	switch l.state {
	case stateIdle:
		return "Idle"
	case stateLoading:
		return "Loading"
	case stateLoaded:
		return "Loaded"
	default:
		return "Error: " + l.err
	}
}

// Render draws the table under the lock, since drawing it moves the scroll
// offset the goroutine may be writing beside.
func (l *pullRequestList) Render(area catatui.Rect, buf *catatui.Buffer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	block := widgets.Bordered().
		Title("Pull Requests").
		TitleTop(catatui.LineFromString(l.status()).Right()).
		TitleBottom(catatui.LineFromString("j/k to scroll, q to quit"))

	rows := make([]widgets.Row, len(l.pulls))
	for i, pull := range l.pulls {
		rows[i] = widgets.NewRowFromStrings(pull.id, pull.title, pull.url)
	}

	table := widgets.NewTable(rows,
		catatui.Length(5), // the number
		catatui.Fill(1),   // the title, which takes what is left
		catatui.Max(49),   // and the URL, which has a length it does not exceed
	).
		Block(block).
		HighlightSpacing(widgets.HighlightSpacingAlways).
		HighlightSymbol(">>").
		RowHighlightStyle(catatui.NewStyle().Bg(catatui.ColorBlue))

	catatui.RenderStatefulWidget(table, area, buf, &l.table)
}

// fetchPullRequests calls the GitHub API and returns what it says.
func fetchPullRequests(ctx context.Context, url string) ([]pullRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	// GitHub rejects a request without one of these.
	req.Header.Set("User-Agent", "catatui-async-github-example")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, statusError(resp)
	}
	return decodePullRequests(resp.Body)
}

// decodePullRequests reads the fields the table shows out of the API's answer,
// dropping anything missing one of them, as ratatui's filter_map does.
func decodePullRequests(r io.Reader) ([]pullRequest, error) {
	var payload []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		URL    string `json:"html_url"`
	}
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decoding the pull requests: %w", err)
	}

	pulls := make([]pullRequest, 0, len(payload))
	for _, item := range payload {
		if item.Number == 0 || item.Title == "" || item.URL == "" {
			continue
		}
		pulls = append(pulls, pullRequest{
			id:    fmt.Sprint(item.Number),
			title: item.Title,
			url:   item.URL,
		})
	}
	return pulls, nil
}

// statusError turns a refusal into the line the header shows. The API explains
// itself in the body, and a rate limit — the one everybody meets — says so
// there rather than in the status alone.
func statusError(resp *http.Response) error {
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil && body.Message != "" {
		return fmt.Errorf("%s: %s", resp.Status, body.Message)
	}
	return fmt.Errorf("%s", resp.Status)
}
