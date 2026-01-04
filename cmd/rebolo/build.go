package main

import (
	"fmt"
	"os"
	"os/exec"
)

func buildForProduction() {
	fmt.Println("🏗️  Building assets with Bun.js...")
	
	// Check if package.json exists
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		fmt.Println("❌ No package.json found. Run 'rebolo dev' first to set up assets.")
		return
	}
	
	// Install dependencies
	fmt.Println("📦 Installing dependencies...")
	if err := runBuildCommand("bun", "install"); err != nil {
		fmt.Printf("❌ Failed to install dependencies: %v\n", err)
		return
	}
	
	// Build assets for production
	fmt.Println("⚡ Building assets for production...")
	if err := runBuildCommand("bun", "run", "build"); err != nil {
		fmt.Printf("❌ Failed to build assets: %v\n", err)
		return
	}
	
	// Build Go binary
	fmt.Println("🔨 Building Go application...")
	if err := runBuildCommand("go", "build", "-o", "app", "main.go"); err != nil {
		fmt.Printf("❌ Failed to build Go application: %v\n", err)
		return
	}
	
	fmt.Println("✅ Build completed successfully!")
	fmt.Println("📦 Your application is ready:")
	fmt.Println("   - Binary: ./app")
	fmt.Println("   - Assets: ./public/")
	fmt.Println("")
	fmt.Println("🚀 To deploy:")
	fmt.Println("   1. Copy ./app binary to your server")
	fmt.Println("   2. Copy ./public/ directory to your server")
	fmt.Println("   3. Copy ./views/ directory to your server")
	fmt.Println("   4. Copy config.yml to your server")
	fmt.Println("   5. Run: ./app")
}

func runBuildCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
