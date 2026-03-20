package gui

import (
	"log"
	"sync"
)

// ExitMode controls when the application exits.
type ExitMode int

const (
	// ExitOnLastClose exits when the last window is closed.
	ExitOnLastClose ExitMode = iota
	// ExitOnMainClose exits when the main (first) window is closed.
	ExitOnMainClose
)

// App manages multiple windows in a single application.
type App struct {
	mu       sync.Mutex
	windows  map[uint32]*Window
	order    []uint32
	mainID   uint32
	ExitMode ExitMode
	pending  chan WindowCfg
}

// NewApp creates an App with an empty window registry.
func NewApp() *App {
	return &App{
		windows: make(map[uint32]*Window),
		pending: make(chan WindowCfg, 16),
	}
}

// Register associates a platform window ID with a Window.
// Duplicate IDs are ignored with a warning.
func (a *App) Register(id uint32, w *Window) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.windows[id]; exists {
		log.Printf("gui: App.Register: duplicate window ID %d ignored", id)
		return
	}
	a.windows[id] = w
	a.order = append(a.order, id)
	w.app = a
	w.platformID = id
	if len(a.order) == 1 {
		a.mainID = id
	}
}

// Unregister removes a window. Returns true if the app should
// exit based on ExitMode.
//
// Clears w.app and w.platformID without holding Window.mu. This is
// safe because Unregister runs on the main thread, which is also the
// only thread that reads these fields.
func (a *App) Unregister(id uint32) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	w := a.windows[id]
	if w != nil {
		w.app = nil
		w.platformID = 0
	}
	delete(a.windows, id)
	for i, oid := range a.order {
		if oid == id {
			a.order = append(a.order[:i], a.order[i+1:]...)
			break
		}
	}
	switch a.ExitMode {
	case ExitOnMainClose:
		return id == a.mainID
	default:
		return len(a.windows) == 0
	}
}

// Window returns the Window for the given platform ID, or nil.
func (a *App) Window(id uint32) *Window {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.windows[id]
}

// Windows returns all registered windows in creation order.
func (a *App) Windows() []*Window {
	a.mu.Lock()
	defer a.mu.Unlock()
	ws := make([]*Window, 0, len(a.order))
	for _, id := range a.order {
		if w, ok := a.windows[id]; ok {
			ws = append(ws, w)
		}
	}
	return ws
}

// OpenWindow queues a new window for creation on the next frame.
// The internal buffer holds up to 16 pending requests. Requests
// beyond that are dropped with a log warning.
func (a *App) OpenWindow(cfg WindowCfg) {
	select {
	case a.pending <- cfg:
	default:
		log.Println("gui: App.OpenWindow: pending buffer full, " +
			"window request dropped")
	}
}

// PendingOpen returns the channel of window configs to create.
func (a *App) PendingOpen() <-chan WindowCfg {
	return a.pending
}

// Broadcast calls fn for every registered window. Iterates under
// the lock to avoid allocating a snapshot slice.
func (a *App) Broadcast(fn func(*Window)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, id := range a.order {
		if w, ok := a.windows[id]; ok {
			fn(w)
		}
	}
}
