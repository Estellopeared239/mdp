package gui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	webview "github.com/webview/webview_go"

	"github.com/mxcoppell/md-preview-cli/internal/ipc"
	"github.com/mxcoppell/md-preview-cli/internal/renderer"
	"github.com/mxcoppell/md-preview-cli/internal/server"
	"github.com/mxcoppell/md-preview-cli/internal/watcher"
)

// activeHost is the package-level host reference needed by CGO callbacks.
var activeHost *Host

// Host is the single process that manages all preview windows.
type Host struct {
	mu        sync.Mutex
	windows   map[string]*WindowEntry
	nextID    int
	primaryWV webview.WebView // first webview, used for Run()/Dispatch()
	ipcSrv    *ipc.Server
	ctx       context.Context
	cancel    context.CancelFunc
}

// WindowEntry tracks a single preview window and its resources.
type WindowEntry struct {
	ID       string
	Filename string
	FilePath string
	Webview  webview.WebView
	Server   *server.Server
	Cancel   context.CancelFunc // cancels watchers
}

// RunHost is the host process entry point. It reads the initial config,
// sets up the dock icon, starts IPC, creates the first window, and runs
// the NSApp event loop.
func RunHost(cfgPath string) error {
	cfg, err := ReadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := &Host{
		windows: make(map[string]*WindowEntry),
		ctx:     ctx,
		cancel:  cancel,
	}
	activeHost = h

	// Initialize as regular app (dock icon visible)
	initHostMode()

	// Start IPC server
	ipcSrv, err := ipc.NewServer(h.handleIPC)
	if err != nil {
		return fmt.Errorf("starting IPC server: %w", err)
	}
	h.ipcSrv = ipcSrv
	go ipcSrv.Serve()

	// Create the first window (also sets primaryWV)
	if err := h.openWindowFromConfig(cfg); err != nil {
		ipcSrv.Close()
		return fmt.Errorf("creating first window: %w", err)
	}

	// Signal handler
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		h.primaryWV.Dispatch(func() {
			h.closeAllWindows()
		})
	}()

	// Run the NSApp event loop (blocks until Terminate)
	h.primaryWV.Run()

	// Cleanup
	fmt.Fprintf(os.Stderr, "md-preview-cli: shutting down\n")
	ipcSrv.Close()
	h.shutdownAllServers()
	return nil
}

// handleIPC processes an incoming IPC request from a CLI process.
func (h *Host) handleIPC(req ipc.OpenRequest) ipc.OpenResponse {
	cfg, err := ReadConfig(req.ConfigPath)
	if err != nil {
		return ipc.OpenResponse{Error: fmt.Sprintf("read config: %v", err)}
	}

	// Window creation must happen on the main thread
	done := make(chan string, 1)
	errCh := make(chan error, 1)
	h.primaryWV.Dispatch(func() {
		id, err := h.createWindow(cfg)
		if err != nil {
			errCh <- err
			return
		}
		done <- id
	})

	select {
	case id := <-done:
		return ipc.OpenResponse{OK: true, WindowID: id}
	case err := <-errCh:
		return ipc.OpenResponse{Error: err.Error()}
	}
}

// openWindowFromConfig creates a window from a Config. Called for the first window.
func (h *Host) openWindowFromConfig(cfg Config) error {
	id, err := h.createWindow(cfg)
	if err != nil {
		return err
	}
	_ = id
	return nil
}

// createWindow creates a webview + server + watchers for a config.
// Must be called on the main thread (or before Run() for the first window).
func (h *Host) createWindow(cfg Config) (string, error) {
	h.mu.Lock()
	h.nextID++
	id := fmt.Sprintf("w-%d", h.nextID)
	h.mu.Unlock()

	wCtx, wCancel := context.WithCancel(h.ctx)

	// Start HTTP server
	srv := server.New(server.Config{
		Port:       cfg.Port,
		Theme:      cfg.Theme,
		HTML:       cfg.HTML,
		TOC:        toServerTOC(cfg.TOC),
		Filename:   cfg.Filename,
		FilePath:   cfg.FilePath,
		ShowTOC:    cfg.ShowTOC,
		HasMath:    cfg.HasMath,
		HasMermaid: cfg.HasMermaid,
		WordCount:  cfg.WordCount,
		NoWatch:    cfg.NoWatch,
	})

	addr, err := srv.Start(wCtx)
	if err != nil {
		wCancel()
		return "", fmt.Errorf("starting server: %w", err)
	}
	url := fmt.Sprintf("http://%s", addr)
	fmt.Fprintf(os.Stderr, "md-preview-cli: listening on %s (%s)\n", url, cfg.Filename)

	// Start file watchers
	startFileWatchersWithCtx(wCtx, cfg, srv)

	// Create webview
	w := webview.New(false)
	hideWindowOffscreen(w.Window())

	title := cfg.Filename + " — md-preview-cli"
	w.SetTitle(title)
	w.SetSize(980, 1270, webview.HintNone)

	// Bind window management functions
	_ = w.Bind("moveWindowBy", func(dx, dy float64) {
		w.Dispatch(func() {
			moveWindowBy(w.Window(), int(dx), int(dy))
		})
	})

	_ = w.Bind("resizeWindowBy", func(dw, dh, shiftX float64) {
		w.Dispatch(func() {
			resizeWindowBy(w.Window(), int(dw), int(dh), int(shiftX))
		})
	})

	_ = w.Bind("showWindow", func(width, height int) {
		w.Dispatch(func() {
			showWindow(w.Window(), width, height)
		})
	})

	// closeThisWindow binding — routes through host
	windowID := id
	_ = w.Bind("closeThisWindow", func() {
		w.Dispatch(func() {
			h.CloseWindow(windowID)
		})
	})

	w.Navigate(url)

	// Wire server shutdown → close this window (not terminate app)
	srv.OnShutdown = func() {
		h.primaryWV.Dispatch(func() {
			h.CloseWindow(windowID)
		})
	}

	entry := &WindowEntry{
		ID:       id,
		Filename: cfg.Filename,
		FilePath: cfg.FilePath,
		Webview:  w,
		Server:   srv,
		Cancel:   wCancel,
	}

	h.mu.Lock()
	h.windows[id] = entry
	if h.primaryWV == nil {
		h.primaryWV = w
	}
	h.mu.Unlock()

	// Schedule frameless styling
	applyFramelessDirect(w.Window())

	return id, nil
}

// CloseWindow closes a single window and its resources.
func (h *Host) CloseWindow(id string) {
	h.mu.Lock()
	entry, ok := h.windows[id]
	if !ok {
		h.mu.Unlock()
		return
	}
	delete(h.windows, id)
	remaining := len(h.windows)
	isPrimary := entry.Webview == h.primaryWV
	h.mu.Unlock()

	// Stop watchers and server
	entry.Cancel()
	entry.Server.Shutdown()

	if remaining == 0 {
		// Last window: stop the run loop
		if isPrimary {
			h.primaryWV.Terminate()
		} else {
			entry.Webview.Destroy()
			h.primaryWV.Terminate()
		}
	} else {
		// Not the last window: just close this one
		if isPrimary {
			// Primary is closing but others remain — pick a new primary
			h.mu.Lock()
			for _, e := range h.windows {
				h.primaryWV = e.Webview
				break
			}
			h.mu.Unlock()
			entry.Webview.Destroy()
		} else {
			entry.Webview.Destroy()
		}
	}
}

// closeAllWindows closes every window. Called from signal handler or dock menu Quit.
func (h *Host) closeAllWindows() {
	h.mu.Lock()
	ids := make([]string, 0, len(h.windows))
	for id := range h.windows {
		ids = append(ids, id)
	}
	h.mu.Unlock()

	for _, id := range ids {
		h.CloseWindow(id)
	}
}

// shutdownAllServers ensures all servers are fully stopped after the run loop exits.
func (h *Host) shutdownAllServers() {
	h.mu.Lock()
	entries := make([]*WindowEntry, 0, len(h.windows))
	for _, e := range h.windows {
		entries = append(entries, e)
	}
	h.mu.Unlock()

	for _, e := range entries {
		e.Cancel()
		e.Server.Shutdown()
		e.Server.Wait()
	}
}

// ActivateWindow brings a window to the front.
func (h *Host) ActivateWindow(id string) {
	h.mu.Lock()
	entry, ok := h.windows[id]
	h.mu.Unlock()
	if !ok {
		return
	}
	activateWindow(entry.Webview.Window())
}

// WindowCount returns the number of open windows.
func (h *Host) WindowCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.windows)
}

// WindowList returns ordered window IDs and filenames.
func (h *Host) WindowList() []WindowEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := make([]WindowEntry, 0, len(h.windows))
	for _, e := range h.windows {
		list = append(list, *e)
	}
	return list
}

// OpenFile reads a markdown file, renders it, and opens a new window.
func (h *Host) OpenFile(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "md-preview-cli: resolve path: %v\n", err)
		return
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "md-preview-cli: read file: %v\n", err)
		return
	}

	result := renderer.Render(string(data))
	cfg := Config{
		Theme:      "system",
		HTML:       result.HTML,
		TOC:        rendererTOCToGUI(result.TOC),
		RawContent: string(data),
		Filename:   filepath.Base(absPath),
		FilePath:   absPath,
		WatchFiles: []string{absPath},
		HasMath:    result.HasMath,
		HasMermaid: result.HasMermaid,
		WordCount:  result.WordCount,
	}

	h.primaryWV.Dispatch(func() {
		if _, err := h.createWindow(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "md-preview-cli: open window: %v\n", err)
		}
	})
}

// startFileWatchersWithCtx starts watchers that can be cancelled per-window.
func startFileWatchersWithCtx(ctx context.Context, cfg Config, srv *server.Server) {
	if cfg.NoWatch || len(cfg.WatchFiles) == 0 {
		return
	}

	for _, file := range cfg.WatchFiles {
		absPath, err := filepath.Abs(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "md-preview-cli: resolve path error (%s): %v\n", file, err)
			continue
		}

		var w watcher.Watcher
		if cfg.Poll > 0 {
			w = watcher.NewPollWatcher(absPath, cfg.Poll)
		} else {
			w, err = watcher.NewFileWatcher(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "md-preview-cli: watcher error (%s): %v\n", file, err)
				continue
			}
		}

		go func() {
			if err := w.Start(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "md-preview-cli: watcher error: %v\n", err)
			}
		}()

		go func() {
			for newContent := range w.Content() {
				result := renderer.Render(newContent)
				srv.UpdateContent(result.HTML, rendererTOCToServer(result.TOC), result.HasMath, result.HasMermaid, result.WordCount)
			}
		}()
	}
}

func rendererTOCToGUI(entries []renderer.TOCEntry) []TOCEntry {
	result := make([]TOCEntry, len(entries))
	for i, e := range entries {
		result[i] = TOCEntry{
			ID:       e.ID,
			Text:     e.Text,
			Level:    e.Level,
			Children: rendererTOCToGUI(e.Children),
		}
	}
	return result
}

func rendererTOCToServer(entries []renderer.TOCEntry) []server.TOCEntry {
	result := make([]server.TOCEntry, len(entries))
	for i, e := range entries {
		result[i] = server.TOCEntry{
			ID:       e.ID,
			Text:     e.Text,
			Level:    e.Level,
			Children: rendererTOCToServer(e.Children),
		}
	}
	return result
}
