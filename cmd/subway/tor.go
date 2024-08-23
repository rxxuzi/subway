package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const torrcTemplate = `SocksPort 9050
HiddenServiceDir {{.TorDataDir}}
{{if .PortForwarding}}
HiddenServicePort 80 {{.PortForwarding}}
{{else}}
HiddenServicePort 80 127.0.0.1:{{.Port}}
{{end}}
`

func generateTorrc(path, torDataDir string, port int, portForwarding string) error {
	tmpl, err := template.New("torrc").Parse(torrcTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, struct {
		TorDataDir     string
		Port           int
		PortForwarding string
	}{
		TorDataDir:     filepath.ToSlash(torDataDir),
		Port:           port,
		PortForwarding: portForwarding,
	})
}

func startTor(ctx context.Context, torPath, torrcPath string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, torPath, "-f", torrcPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Println(fmt.Sprintf("Failed to start Tor: %v", err))
		os.Exit(1)
	}
	return cmd
}

func checkTorInstallation(torPath string) bool {
	cmd := exec.Command(torPath, "--version")
	err := cmd.Run()
	return err == nil
}

func waitForOnionAddress(hostnameFile string) (string, error) {
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(hostnameFile); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	content, err := os.ReadFile(hostnameFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}
