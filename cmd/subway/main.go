package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fatih/color"
)

const TOR_DIR = ".subway"

func main() {
	config := DefaultConfig()

	var genConfig bool
	var loadConfigPath string
	var cleanup bool

	flag.BoolVar(&genConfig, "gen", false, "Generate default config file")
	flag.StringVar(&loadConfigPath, "load", "", "Load config from specified file")
	flag.StringVar(&config.Root, "root", config.Root, "Root directory to serve")
	flag.IntVar(&config.Port, "port", config.Port, "Port to serve on")
	flag.StringVar(&config.TorPath, "tor", config.TorPath, "Path to Tor executable")
	flag.StringVar(&config.PortForwarding, "pf", config.PortForwarding, "Port forwarding (e.g., localhost:8080)")
	flag.BoolVar(&cleanup, "clean", false, "Clean up tor/ directory and subway.json")
	flag.Parse()

	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if genConfig {
		err := GenerateDefaultConfig(DEFAULT_CONFIG_FILE)
		if err != nil {
			fmt.Printf("Failed to generate default config: %v\n", err)
		} else {
			fmt.Printf("Default config generated: %s\n", DEFAULT_CONFIG_FILE)
		}
		return
	}

	if cleanup {
		if err := cleanUp(); err != nil {
			fmt.Println(red(fmt.Sprintf("Cleanup failed: %v", err)))
		} else {
			fmt.Println(green("Cleanup successful"))
		}
		return
	}

	if loadConfigPath != "" {
		loadedConfig, err := LoadConfig(loadConfigPath)
		if err != nil {
			fmt.Printf("Failed to load config from %s: %v\n", loadConfigPath, err)
		} else {
			config = loadedConfig
		}
	} else {
		loadedConfig, err := LoadConfig(DEFAULT_CONFIG_FILE)
		if err == nil {
			config = loadedConfig
		}
	}

	// Override config with command-line flags
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "root":
			config.Root = f.Value.String()
		case "port":
			config.Port, _ = f.Value.(flag.Getter).Get().(int)
		case "tor":
			config.TorPath = f.Value.String()
		case "pf":
			config.PortForwarding = f.Value.String()
		}
	})

	if !checkTorInstallation(config.TorPath) {
		fmt.Println(red("Tor is not installed or the provided path is incorrect. Please install Tor or provide the correct path."))
		return
	}

	if config.PortForwarding != "" {
		if err := checkPortAvailability(config.PortForwarding); err != nil {
			fmt.Println(red(fmt.Sprintf("Port forwarding error: %v", err)))
			return
		}
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(red(fmt.Sprintf("Failed to get current directory: %v", err)))
		return
	}

	torDir := filepath.Join(currentDir, TOR_DIR)
	if err := os.MkdirAll(torDir, 0700); err != nil {
		fmt.Println(red(fmt.Sprintf("Failed to create Tor directory: %v", err)))
		return
	}

	torDataDir := filepath.Join(torDir, "data")
	if err := os.MkdirAll(torDataDir, 0700); err != nil {
		fmt.Println(red(fmt.Sprintf("Failed to create Tor data directory: %v", err)))
		return
	}

	torrcPath := filepath.Join(torDir, "torrc")
	if err := generateTorrc(torrcPath, torDataDir, config.Port, config.PortForwarding); err != nil {
		fmt.Println(red(fmt.Sprintf("Failed to generate torrc: %v", err)))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	torCmd := startTor(ctx, config.TorPath, torrcPath)
	defer func(Process *os.Process) {
		err := Process.Kill()
		if err != nil {
			fmt.Println(red(fmt.Sprintf("Failed to kill Tor process: %v", err)))
		}
	}(torCmd.Process)

	onionAddress, err := waitForOnionAddress(filepath.Join(torDataDir, "hostname"))
	if err != nil {
		fmt.Println(red(fmt.Sprintf("Failed to get .onion address: %v", err)))
	} else {
		fmt.Printf("Onion address: %s\n", green(onionAddress))
	}

	var server *http.Server
	if config.PortForwarding == "" {
		server = setupServer(config)
		go func() {
			fmt.Printf("Starting server on http://localhost:%d\n", config.Port)
			fmt.Printf("Serving directory: %s\n", config.Root)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Println(red(fmt.Sprintf("HTTP server error: %v", err)))
			}
		}()
	} else {
		fmt.Printf("Forwarding traffic to %s\n", config.PortForwarding)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down...")

	if server != nil {
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			fmt.Println(red(fmt.Sprintf("Server forced to shutdown: %v", err)))
		}
	}

	fmt.Println("Stopped")
}

func checkPortAvailability(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("the specified port is not available: %v", err)
	}
	err1 := conn.Close()
	if err1 != nil {
		return err1
	}
	return nil
}

func cleanUp() error {
	if err := os.RemoveAll(TOR_DIR); err != nil {
		return fmt.Errorf("failed to remove %s/ directory: %w", TOR_DIR, err)
	}

	if err := os.Remove(DEFAULT_CONFIG_FILE); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove subway.json: %w", err)
		}
	}

	return nil
}
