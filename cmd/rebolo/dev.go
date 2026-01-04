package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

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
	if err := runCommand("bun", "install"); err != nil {
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
	// Create package.json if it doesn't exist
	packageJSON := `{
  "name": "rebolo-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "bun --watch src/index.js --outdir=public",
    "build": "bun build src/index.js --outdir=public --minify",
    "watch": "bun build src/index.js --outdir=public --watch"
  },
  "devDependencies": {
    "bun": "latest"
  },
  "dependencies": {}
}`

	os.WriteFile("package.json", []byte(packageJSON), 0644)
	
	// Create src directory and index.js
	os.MkdirAll("src", 0755)
	os.MkdirAll("public", 0755)
	
	indexJS := `// ReboloLang Frontend Assets
console.log('🚀 ReboloLang app loaded!');

// Hot reload support
if (process.env.NODE_ENV === 'development') {
  const eventSource = new EventSource('/dev/reload');
  eventSource.onmessage = () => {
    console.log('🔄 Hot reloading...');
    location.reload();
  };
}

// Basic styling
document.addEventListener('DOMContentLoaded', function() {
  const style = document.createElement('style');
  style.textContent = ` + "`" + `
    body { 
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      margin: 0;
      padding: 2rem;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: white;
      min-height: 100vh;
    }
    .container { max-width: 800px; margin: 0 auto; text-align: center; }
    h1 { font-size: 3rem; margin-bottom: 1rem; }
    p { font-size: 1.2rem; opacity: 0.9; }
    .btn {
      background: #4CAF50;
      color: white;
      padding: 12px 24px;
      text-decoration: none;
      border-radius: 6px;
      display: inline-block;
      margin: 10px;
      transition: background 0.3s;
    }
    .btn:hover { background: #45a049; }
    .form-group {
      margin-bottom: 1rem;
      text-align: left;
    }
    .form-group label {
      display: block;
      margin-bottom: 5px;
      font-weight: bold;
    }
    .form-group input, .form-group textarea {
      width: 100%;
      padding: 8px;
      border: 1px solid #ddd;
      border-radius: 4px;
      font-size: 14px;
    }
  ` + "`" + `;
  document.head.appendChild(style);
});`

	os.WriteFile("src/index.js", []byte(indexJS), 0644)
}

func runBunDev() {
	cmd := exec.Command("bun", "run", "watch")
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
	
	// Watch src directory
	if err := watcher.Add("src"); err != nil {
		log.Printf("Failed to watch src directory: %v", err)
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
				if strings.HasSuffix(event.Name, ".js") || 
				   strings.HasSuffix(event.Name, ".css") ||
				   strings.HasSuffix(event.Name, ".ts") {
					fmt.Printf("🔄 Frontend file changed: %s\n", event.Name)
					// Bun watcher will handle the rebuild automatically
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
	debounce := time.NewTimer(100 * time.Millisecond)
	debounce.Stop()
	
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			
			if event.Op&fsnotify.Write == fsnotify.Write && strings.HasSuffix(event.Name, ".go") {
				debounce.Reset(100 * time.Millisecond)
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
	skipDirs := []string{"node_modules", ".git", "vendor", "public"}
	for _, skip := range skipDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}
	return false
}
