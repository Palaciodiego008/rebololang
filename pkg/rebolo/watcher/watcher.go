package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// AppInterface defines the minimal interface needed by FileWatcher
type AppInterface interface {
	UpdateLastChangeTime(t time.Time)
	ReloadTemplates()
}

// WatcherStats tracks statistics about file watching
type WatcherStats struct {
	TotalChanges    int
	TemplateChanges int
	AssetChanges    int
	CodeChanges     int
	LastChangeTime  time.Time
}

// FileWatcher monitors file changes and triggers reloads
type FileWatcher struct {
	watcher     *fsnotify.Watcher
	app         AppInterface
	subscribers []chan FileChangeEvent
	mu          sync.RWMutex
	debounce    map[string]time.Time
	debounceMu  sync.Mutex
	watchDirs   []string
	stats       WatcherStats
	statsMu     sync.RWMutex
}

// FileChangeEvent represents a file change notification
type FileChangeEvent struct {
	Path      string
	EventType string // "template", "asset", "code"
	Timestamp time.Time
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(app AppInterface, watchDirs []string) *FileWatcher {
	fw := &FileWatcher{
		app:         app,
		subscribers: make([]chan FileChangeEvent, 0),
		debounce:    make(map[string]time.Time),
		watchDirs:   watchDirs,
		stats:       WatcherStats{},
	}

	return fw
}

// Start starts the file watcher
func (fw *FileWatcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	fw.watcher = watcher

	// Add directories to watch
	for _, dir := range fw.watchDirs {
		if err := fw.addRecursive(dir); err != nil {
			log.Printf("⚠️  Failed to watch %s: %v", dir, err)
			continue
		}
		log.Printf("👁️  Watching: %s", dir)
	}

	// Start watching in goroutine
	go fw.processEvents()

	return nil
}

// Watch is an alias for Start for backward compatibility
func (fw *FileWatcher) Watch(ctx context.Context, dirs ...string) error {
	fw.watchDirs = dirs
	return fw.Start()
}

// addRecursive adds a directory and its subdirectories to the watcher
func (fw *FileWatcher) addRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories and node_modules
		if info != nil && info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return fw.watcher.Add(path)
		}
		return nil
	})
}

// processEvents handles file system events
func (fw *FileWatcher) processEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			fw.handleEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("❌ Watcher error: %v", err)
		}
	}
}

// handleEvent processes a single file system event
func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	// Debounce: ignore rapid successive events for the same file
	if !fw.shouldProcess(event.Name) {
		return
	}

	ext := filepath.Ext(event.Name)
	var eventType string

	switch ext {
	case ".html", ".tmpl":
		eventType = "template"
		fw.reloadTemplates()
	case ".css", ".js", ".ts", ".jsx", ".tsx":
		eventType = "asset"
		fw.recompileAssets()
	case ".go":
		eventType = "code"
		log.Printf("🔄 Code changed: %s (restart required)", event.Name)
	default:
		return // Ignore other file types
	}

	// Update statistics
	fw.statsMu.Lock()
	fw.stats.TotalChanges++
	fw.stats.LastChangeTime = time.Now()
	switch eventType {
	case "template":
		fw.stats.TemplateChanges++
	case "asset":
		fw.stats.AssetChanges++
	case "code":
		fw.stats.CodeChanges++
	}
	fw.statsMu.Unlock()

	// Notify all subscribers
	changeEvent := FileChangeEvent{
		Path:      event.Name,
		EventType: eventType,
		Timestamp: time.Now(),
	}

	log.Printf("🔥 Hot reload: %s (%s) at %s",
		filepath.Base(event.Name),
		eventType,
		changeEvent.Timestamp.Format("15:04:05.000"))
	fw.notifySubscribers(changeEvent)
}

// shouldProcess implements debouncing to avoid processing the same file too frequently
func (fw *FileWatcher) shouldProcess(path string) bool {
	fw.debounceMu.Lock()
	defer fw.debounceMu.Unlock()

	now := time.Now()
	if lastTime, exists := fw.debounce[path]; exists {
		if now.Sub(lastTime) < 500*time.Millisecond {
			return false
		}
	}
	fw.debounce[path] = now
	return true
}

// reloadTemplates reloads HTML templates without restarting the server
func (fw *FileWatcher) reloadTemplates() {
	start := time.Now()
	log.Printf("📝 Reloading templates...")

	// Delegate to the app
	fw.app.ReloadTemplates()
	fw.app.UpdateLastChangeTime(time.Now())

	duration := time.Since(start)
	log.Printf("✅ Templates reloaded in %v", duration)
}

// recompileAssets triggers asset recompilation with Bun
func (fw *FileWatcher) recompileAssets() {
	log.Printf("⚡ Recompiling assets with Bun...")
	// Update last change time for polling to detect changes
	fw.app.UpdateLastChangeTime(time.Now())
	// This will be triggered by the command watching process
	// Just notify subscribers - the dev command handles actual compilation
}

// Subscribe creates a new channel for receiving change events
func (fw *FileWatcher) Subscribe() chan FileChangeEvent {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	ch := make(chan FileChangeEvent, 10)
	fw.subscribers = append(fw.subscribers, ch)
	return ch
}

// Unsubscribe removes a channel from subscribers
func (fw *FileWatcher) Unsubscribe(ch chan FileChangeEvent) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	for i, sub := range fw.subscribers {
		if sub == ch {
			close(sub)
			fw.subscribers = append(fw.subscribers[:i], fw.subscribers[i+1:]...)
			return
		}
	}
}

// notifySubscribers sends an event to all subscribers
func (fw *FileWatcher) notifySubscribers(event FileChangeEvent) {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	for _, ch := range fw.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip this notification
		}
	}
}

// Stats returns current watcher statistics
func (fw *FileWatcher) Stats() WatcherStats {
	fw.statsMu.RLock()
	defer fw.statsMu.RUnlock()
	return fw.stats
}

// Close stops the file watcher
func (fw *FileWatcher) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Log statistics before closing
	fw.statsMu.RLock()
	stats := fw.stats
	fw.statsMu.RUnlock()
	if stats.TotalChanges > 0 {
		log.Printf("📊 Watcher stats: %d total changes (%d templates, %d assets, %d code)",
			stats.TotalChanges, stats.TemplateChanges, stats.AssetChanges, stats.CodeChanges)
	}

	// Close all subscriber channels
	for _, ch := range fw.subscribers {
		close(ch)
	}
	fw.subscribers = nil

	return fw.watcher.Close()
}
