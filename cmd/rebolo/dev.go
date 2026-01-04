package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var devConfig = DefaultDevConfig()

func startDevServer() {
	// Start Bun asset watcher
	go startBunAssetPipeline()

	// Start Go server with hot reload
	startGoServerWithReload()
}

func startBunAssetPipeline() {
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		fmt.Println("📦 No package.json found, creating default Bun setup...")
		createDefaultBunSetup()
	}

	// Install dependencies
	fmt.Println("📦 Installing Bun dependencies...")
	if err := runCommandArgs(devConfig.BunInstallCommand...); err != nil {
		log.Printf("Failed to install Bun dependencies: %v", err)
		return
	}

	// Start Bun in development mode with watch
	fmt.Println("⚡ Starting Bun asset pipeline...")
	go runBunDev()

	// Watch for frontend file changes
	go watchFrontendFiles()
}

func createDefaultBunSetup() {
	// Create directories
	os.MkdirAll(devConfig.FrontendSrcDir, 0755)
	os.MkdirAll(devConfig.FrontendOutDir, 0755)

	// Read and write package.json from template
	packageJSON, err := fs.ReadFile(templates, "templates/dev/package.json.tmpl")
	if err != nil {
		log.Printf("Failed to read package.json template: %v", err)
		return
	}
	os.WriteFile("package.json", packageJSON, 0644)

	// Read and write index.js from template
	indexJS, err := fs.ReadFile(templates, "templates/dev/index.js.tmpl")
	if err != nil {
		log.Printf("Failed to read index.js template: %v", err)
		return
	}
	os.WriteFile(filepath.Join(devConfig.FrontendSrcDir, "index.js"), indexJS, 0644)
}

func runBunDev() {
	cmd := exec.Command(devConfig.BunWatchCommand[0], devConfig.BunWatchCommand[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start Bun watcher: %v", err)
		return
	}

	// Keep the process running
	cmd.Wait()
}

func watchFrontendFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Failed to create frontend watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch configured src directory
	if err := watcher.Add(devConfig.FrontendSrcDir); err != nil {
		log.Printf("Failed to watch %s directory: %v", devConfig.FrontendSrcDir, err)
		return
	}

	fmt.Println("👀 Watching frontend files for changes...")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				for _, ext := range devConfig.FrontendWatchExtensions {
					if strings.HasSuffix(event.Name, ext) {
						fmt.Printf("🔄 Frontend file changed: %s\n", event.Name)
						break // Bun watcher will handle the rebuild automatically
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Frontend watcher error: %v", err)
		}
	}
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandArgs(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}
	return runCommand(args[0], args[1:]...)
}

func startGoServerWithReload() {
	fmt.Println("🔥 Starting Go server with hot reload...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Watch Go files
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && shouldSkipDir(path) {
			return filepath.SkipDir
		}

		if strings.HasSuffix(path, ".go") {
			watcher.Add(filepath.Dir(path))
		}

		return nil
	})

	var cmd *exec.Cmd
	restartServer := func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}

		fmt.Println("🔄 Restarting Go server...")
		cmd = exec.Command("go", "run", "main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
	}

	// Initial start
	restartServer()

	// Watch for changes
	debounce := time.NewTimer(devConfig.GoRestartDebounce)
	debounce.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				for _, ext := range devConfig.GoWatchExtensions {
					if strings.HasSuffix(event.Name, ext) {
						debounce.Reset(devConfig.GoRestartDebounce)
						break
					}
				}
			}

		case <-debounce.C:
			restartServer()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Go watcher error:", err)
		}
	}
}

func shouldSkipDir(path string) bool {
	for _, skip := range devConfig.GoSkipDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}
	return false
}
