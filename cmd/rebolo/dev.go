package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var devConfig = DefaultDevConfig()

// startDevServer starts the development server with hot reload
func startDevServer() {
	fmt.Println("Starting ReboloLang development server...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down development server...")
		cancel()
	}()

	// Check if frontend exists
	hasFrontend := false
	if _, err := os.Stat("frontend"); err == nil {
		hasFrontend = true
		fmt.Println("🎨 Frontend detected")
	}

	if hasFrontend {
		// 1. Install frontend dependencies if needed
		setupFrontendDependencies()
		
		// 2. Build frontend initially
		buildFrontend()
		
		// 3. Watch frontend for changes
		go watchAndCompileFrontend(ctx)
	} else {
		// Traditional mode: Setup Bun.js and compile assets initially
		setupBunAndAssets()
		
		// Start Bun watcher for assets (CSS/JS) in background
		go watchAndCompileAssets(ctx)
	}

	// Start Go server with hot reload for .go files
	startGoServerWithHotReload(ctx)
}

// setupBunAndAssets sets up Bun.js and compiles assets initially
func setupBunAndAssets() {
	// Check if Bun is installed
	if !isBunInstalled() {
		fmt.Println("🔧 Bun.js not found. Trying to use it from ~/.bun/bin...")

		// Try to use Bun from home directory
		homeDir, _ := os.UserHomeDir()
		bunPath := filepath.Join(homeDir, ".bun", "bin", "bun")
		if _, err := os.Stat(bunPath); err == nil {
			// Add to PATH temporarily
			os.Setenv("PATH", filepath.Dir(bunPath)+":"+os.Getenv("PATH"))
		} else {
			// Install Bun
			fmt.Println("📥 Installing Bun.js...")
			if err := installBun(); err != nil {
				log.Printf("⚠️  Bun.js installation failed: %v", err)
				log.Println("📝 Using fallback assets (direct copy of CSS/JS)")
				createFallbackAssets()
				return
			}
		}
	}

	// Build assets initially
	fmt.Println("⚡ Building initial assets with Bun...")
	if err := buildAssets(); err != nil {
		log.Printf("⚠️  Asset build failed: %v", err)
		createFallbackAssets()
	} else {
		fmt.Println("✅ Assets compiled successfully")
	}
}

// watchAndCompileAssets watches for CSS/JS changes and recompiles with Bun
func watchAndCompileAssets(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("❌ Failed to create asset watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch src directory
	srcDir := "src"
	if err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		log.Printf("❌ Failed to watch src directory: %v", err)
		return
	}

	fmt.Println("👀 Watching assets for changes (Bun.js)...")

	debounce := time.NewTimer(300 * time.Millisecond)
	debounce.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				ext := filepath.Ext(event.Name)
				if ext == ".css" || ext == ".js" || ext == ".ts" {
					debounce.Reset(300 * time.Millisecond)
				}
			}
		case <-debounce.C:
			fmt.Println("⚡ Recompiling assets...")
			if err := buildAssets(); err != nil {
				log.Printf("❌ Asset compilation failed: %v", err)
			} else {
				fmt.Println("✅ Assets recompiled")
			}
		case err := <-watcher.Errors:
			log.Printf("❌ Asset watcher error: %v", err)
		}
	}
}

// startGoServerWithHotReload starts the Go server and restarts it when .go files change
func startGoServerWithHotReload(ctx context.Context) {
	fmt.Println("🔥 Starting Go server with hot reload...")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Watch .go files recursively
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories, vendor, and node_modules
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})

	var cmd *exec.Cmd
	var serverStarted = make(chan bool, 1)

	// Function to start/restart the server
	startServer := func() {
		// Kill existing process
		if cmd != nil && cmd.Process != nil {
			fmt.Println("🔄 Restarting Go server...")
			cmd.Process.Kill()
			cmd.Wait()
		} else {
			fmt.Println("🚀 Starting Go server...")
		}

		// Start new process
		cmd = exec.Command("go", "run", "main.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()

		if err := cmd.Start(); err != nil {
			log.Printf("❌ Failed to start server: %v", err)
			return
		}

		// Signal that server started
		select {
		case serverStarted <- true:
		default:
		}
	}

	// Start server initially
	startServer()

	// Debounce timer for restarts
	debounce := time.NewTimer(500 * time.Millisecond)
	debounce.Stop()

	for {
		select {
		case <-ctx.Done():
			if cmd != nil && cmd.Process != nil {
				cmd.Process.Kill()
			}
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Only restart on .go file changes
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 && filepath.Ext(event.Name) == ".go" {
				fmt.Printf("🔄 Code changed: %s\n", filepath.Base(event.Name))
				debounce.Reset(500 * time.Millisecond)
			}
		case <-debounce.C:
			startServer()
		case err := <-watcher.Errors:
			log.Printf("❌ Watcher error: %v", err)
		}
	}
}

// isBunInstalled checks if Bun is available in PATH
func isBunInstalled() bool {
	_, err := exec.LookPath("bun")
	return err == nil
}

// installBun installs Bun.js using the official installer
func installBun() error {
	cmd := exec.Command("bash", "-c", "curl -fsSL https://bun.sh/install | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Add Bun to PATH for this session
	homeDir, _ := os.UserHomeDir()
	bunPath := filepath.Join(homeDir, ".bun", "bin")
	os.Setenv("PATH", bunPath+":"+os.Getenv("PATH"))

	return nil
}

// buildAssets builds the frontend assets with Bun
func buildAssets() error {
	if _, err := os.Stat("src/index.js"); os.IsNotExist(err) {
		return fmt.Errorf("src/index.js not found")
	}

	os.MkdirAll("public", 0755)

	// Build with Bun
	cmd := exec.Command("bun", "build", "src/index.js", "--outdir", "public", "--target", "browser")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("build failed: %w\n%s", err, string(output))
	}

	return nil
}

// createFallbackAssets creates basic CSS and JS files as fallback
func createFallbackAssets() {
	fmt.Println("📝 Creating fallback assets...")

	os.MkdirAll("public", 0755)

	// Copy CSS
	if cssData, err := os.ReadFile("src/styles.css"); err == nil {
		os.WriteFile("public/index.css", cssData, 0644)
		fmt.Println("   ✓ Copied styles.css → public/index.css")
	}

	// Copy JS (remove import statements)
	if jsData, err := os.ReadFile("src/index.js"); err == nil {
		jsContent := string(jsData)
		jsContent = strings.ReplaceAll(jsContent, "import './styles.css';", "")
		jsContent = strings.ReplaceAll(jsContent, `import "./styles.css";`, "")
		os.WriteFile("public/index.js", []byte(jsContent), 0644)
		fmt.Println("   ✓ Copied index.js → public/index.js")
	}

	fmt.Println("✅ Fallback assets created")
}

// setupFrontendDependencies installs frontend dependencies with Bun
func setupFrontendDependencies() {
	pkgPath := filepath.Join("frontend", "package.json")
	nodeModules := filepath.Join("frontend", "node_modules")
	
	// Check if dependencies are already installed
	if _, err := os.Stat(nodeModules); err == nil {
		return
	}
	
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return
	}
	
	fmt.Println("📦 Installing frontend dependencies...")
	cmd := exec.Command("bun", "install")
	cmd.Dir = "frontend"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Printf("⚠️  Failed to install dependencies: %v", err)
		log.Println("   Run manually: cd frontend && bun install")
	} else {
		fmt.Println("✅ Dependencies installed")
	}
}

// buildFrontend builds the frontend with Vite/Bun
func buildFrontend() {
	fmt.Println("⚡ Building frontend...")
	
	cmd := exec.Command("bun", "run", "build")
	cmd.Dir = "frontend"
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Printf("⚠️  Frontend build failed: %v", err)
		log.Printf("   Output: %s", string(output))
		return
	}
	
	fmt.Println("✅ Frontend built successfully")
}

// watchAndCompileFrontend watches frontend changes and rebuilds
func watchAndCompileFrontend(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("❌ Failed to create frontend watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Watch frontend/src directory
	frontendSrc := filepath.Join("frontend", "src")
	if err := filepath.Walk(frontendSrc, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		log.Printf("❌ Failed to watch frontend: %v", err)
		return
	}

	fmt.Println("👀 Watching frontend for changes...")

	debounce := time.NewTimer(500 * time.Millisecond)
	debounce.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				ext := filepath.Ext(event.Name)
				if ext == ".tsx" || ext == ".ts" || ext == ".jsx" || ext == ".js" || 
				   ext == ".vue" || ext == ".svelte" || ext == ".css" {
					fmt.Printf("🔄 Frontend changed: %s\n", filepath.Base(event.Name))
					debounce.Reset(500 * time.Millisecond)
				}
			}
		case <-debounce.C:
			buildFrontend()
		case err := <-watcher.Errors:
			log.Printf("❌ Frontend watcher error: %v", err)
		}
	}
}

