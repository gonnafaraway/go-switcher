package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/spf13/cobra"
)

const (
	goDownloadDir = "/usr/local/bin/go-switcher"
)

// Command represents a CLI command with its configuration and execution logic.
type Command struct {
	*cobra.Command
	config *Config
}

// Config holds the configuration for the Go version switcher.
type Config struct {
	GoDownloadDir string
}

// NewConfig creates a new configuration instance.
func NewConfig() *Config {
	return &Config{
		GoDownloadDir: goDownloadDir,
	}
}

// NewCommand creates a new command instance.
func NewCommand(config *Config) *Command {
	return &Command{
		Command: &cobra.Command{Use: "go-switcher"},
		config:  config,
	}
}

// NewListCommand creates a new list command.
func NewListCommand(config *Config) *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "list",
			Short: "List downloaded Go versions",
			Long:  "List all downloaded Go versions from the default directory",
		},
		config: config,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := listGoVersions(cmd, config); err != nil {
			log.Fatalf("Failed to list Go versions: %v", err)
		}
	}

	return cmd
}

// NewDownloadCommand creates a new download command.
func NewDownloadCommand(config *Config) *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "download [version]",
			Short: "Download a Go version",
			Long:  "Download a specific version of Go from the official source",
			Args:  cobra.MinimumNArgs(1),
		},
		config: config,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := downloadGoVersion(cmd, config, args[0]); err != nil {
			log.Fatalf("Failed to download Go version: %v", err)
		}
	}

	cmd.Flags().String("arch", "linux-amd64", "Architecture to download (e.g., linux-amd64, darwin-amd64, windows-amd64)")

	return cmd
}

// NewCleanCommand creates a new clean command.
func NewCleanCommand(config *Config) *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "clean",
			Short: "Remove all Go versions",
			Long:  "Remove all downloaded Go versions from the default directory",
		},
		config: config,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := cleanGoVersions(config); err != nil {
			log.Fatalf("Failed to clean Go versions: %v", err)
		}
	}

	return cmd
}

// NewSwitchCommand creates a new switch command.
func NewSwitchCommand(config *Config) *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "switch [version|number]",
			Short: "Switch Go version",
			Long:  "Switch to a specific Go version by updating environment variables",
			Args:  cobra.MinimumNArgs(1),
		},
		config: config,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := switchGoVersion(cmd, config, args[0]); err != nil {
			log.Fatalf("Failed to switch Go version: %v", err)
		}
	}

	cmd.Flags().String("arch", "linux-amd64", "Architecture to switch to (e.g., linux-amd64, darwin-amd64, windows-amd64)")

	return cmd
}

// listGoVersions lists all downloaded Go versions.
func listGoVersions(cmd *cobra.Command, config *Config) error {
	fmt.Printf("= Go versions from %s =\n", config.GoDownloadDir)

	if _, err := os.Stat(config.GoDownloadDir); os.IsNotExist(err) {
		fmt.Println("No Go versions found. Directory does not exist.")
		return nil
	}

	dir, err := os.ReadDir(config.GoDownloadDir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	var (
		versions      []string
		architectures []string
		paths         []string
	)

	for _, fi := range dir {
		if !fi.IsDir() {
			continue
		}

		versionDir := fmt.Sprintf("%s/%s", config.GoDownloadDir, fi.Name())
		archDirs, err := os.ReadDir(versionDir)
		if err != nil {
			fmt.Printf("Error reading version directory %s: %v\n", fi.Name(), err)
			continue
		}

		for _, archFi := range archDirs {
			if archFi.IsDir() {
				versions = append(versions, fi.Name())
				architectures = append(architectures, archFi.Name())
				paths = append(paths, fmt.Sprintf("%s/%s/%s", config.GoDownloadDir, fi.Name(), archFi.Name()))
			}
		}
	}

	if len(versions) == 0 {
		fmt.Println("No Go versions found")
		return nil
	}

	for i := range versions {
		fmt.Printf("%d. %s (%s)\n", i+1, versions[i], architectures[i])
	}

	cmd.Root().Annotations = map[string]string{
		"versions":      strings.Join(versions, ","),
		"architectures": strings.Join(architectures, ","),
		"paths":         strings.Join(paths, ","),
	}

	return nil
}

// downloadGoVersion downloads a specific Go version.
func downloadGoVersion(cmd *cobra.Command, config *Config, version string) error {
	fmt.Println("= Downloading version from official resource =")

	arch, _ := cmd.Flags().GetString("arch")
	archiveName := fmt.Sprintf("go%s.%s.tar.gz", version, arch)
	tmpArchivePath := fmt.Sprintf("/tmp/%s", archiveName)
	targetDir := fmt.Sprintf("%s/%s/%s", config.GoDownloadDir, version, arch)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	if err := sh.Run("wget", "-O", tmpArchivePath, fmt.Sprintf("https://go.dev/dl/%s", archiveName)); err != nil {
		return fmt.Errorf("download archive: %w", err)
	}

	if err := sh.Run("tar", "-xzf", tmpArchivePath, "-C", targetDir); err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	if err := os.Remove(tmpArchivePath); err != nil {
		log.Printf("Warning: failed to remove archive: %v", err)
	}

	fmt.Println("= Download finished successfully =")
	return nil
}

// cleanGoVersions removes all downloaded Go versions.
func cleanGoVersions(config *Config) error {
	fmt.Printf("= Removing all Go versions from %s =\n", config.GoDownloadDir)

	if _, err := os.Stat(config.GoDownloadDir); os.IsNotExist(err) {
		fmt.Println("No Go versions found. Directory does not exist.")
		return nil
	}

	dir, err := os.ReadDir(config.GoDownloadDir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	if len(dir) == 0 {
		fmt.Println("No Go versions found to clean")
		return nil
	}

	for _, fi := range dir {
		if !fi.IsDir() {
			continue
		}

		versionDir := fmt.Sprintf("%s/%s", config.GoDownloadDir, fi.Name())
		fmt.Printf("Removing version: %s\n", fi.Name())
		if err := os.RemoveAll(versionDir); err != nil {
			log.Printf("Warning: failed to remove %s: %v", versionDir, err)
		}
	}

	fmt.Println("= Cleanup completed =")
	return nil
}

// switchGoVersion switches to a specific Go version.
func switchGoVersion(cmd *cobra.Command, config *Config, versionOrNumber string) error {
	var (
		version    string
		arch       string
		versionDir string
	)

	dir, err := os.ReadDir(config.GoDownloadDir)
	if err != nil {
		return fmt.Errorf("read directory: %w", err)
	}

	var (
		versions      []string
		architectures []string
		paths         []string
	)

	for _, fi := range dir {
		if !fi.IsDir() {
			continue
		}

		versionDir := fmt.Sprintf("%s/%s", config.GoDownloadDir, fi.Name())
		archDirs, err := os.ReadDir(versionDir)
		if err != nil {
			log.Printf("Error reading version directory %s: %v", fi.Name(), err)
			continue
		}

		for _, archFi := range archDirs {
			if archFi.IsDir() {
				versions = append(versions, fi.Name())
				architectures = append(architectures, archFi.Name())
				paths = append(paths, fmt.Sprintf("%s/%s/%s", config.GoDownloadDir, fi.Name(), archFi.Name()))
			}
		}
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Go versions found")
	}

	if num, err := strconv.Atoi(versionOrNumber); err == nil {
		if num < 1 || num > len(versions) {
			return fmt.Errorf("invalid number. Please choose a number between 1 and %d", len(versions))
		}

		version = versions[num-1]
		arch = architectures[num-1]
		versionDir = paths[num-1]
	} else {
		arch, _ = cmd.Flags().GetString("arch")
		version = versionOrNumber
		versionDir = fmt.Sprintf("%s/%s/%s", config.GoDownloadDir, version, arch)
	}

	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("Go version %s for architecture %s is not installed", version, arch)
	}

	fmt.Printf("= Switching to Go version %s for architecture %s =\n", version, arch)

	goBinPath := fmt.Sprintf("%s/go/bin", versionDir)
	goPath := fmt.Sprintf("%s/workspace", versionDir)
	goRoot := fmt.Sprintf("%s/go", versionDir)

	if err := os.MkdirAll(goPath, 0755); err != nil {
		return fmt.Errorf("create workspace directory: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	profilePath := fmt.Sprintf("%s/.profile", homeDir)
	profileContent, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("read profile: %w", err)
	}

	profileLines := strings.Split(string(profileContent), "\n")
	var newProfileLines []string

	for _, line := range profileLines {
		if !strings.Contains(line, "export PATH=$PATH:") &&
			!strings.Contains(line, "export GOPATH=") &&
			!strings.Contains(line, "export GOROOT=") {
			newProfileLines = append(newProfileLines, line)
		}
	}

	newProfileLines = append(newProfileLines, "")
	newProfileLines = append(newProfileLines, "# Go environment variables")
	newProfileLines = append(newProfileLines, fmt.Sprintf("export PATH=$PATH:%s", goBinPath))
	newProfileLines = append(newProfileLines, fmt.Sprintf("export GOPATH=%s", goPath))
	newProfileLines = append(newProfileLines, fmt.Sprintf("export GOROOT=%s", goRoot))

	if err := os.WriteFile(profilePath, []byte(strings.Join(newProfileLines, "\n")), 0644); err != nil {
		return fmt.Errorf("write profile: %w", err)
	}

	fmt.Println("= Successfully switched to Go version", version, "=")
	fmt.Println("PATH now includes:", goBinPath)
	fmt.Println("GOPATH is set to:", goPath)
	fmt.Println("GOROOT is set to:", goRoot)
	fmt.Println("Changes have been written to", profilePath)
	fmt.Println("Please log out and log back in for changes to take effect")

	return nil
}

func main() {
	config := NewConfig()
	rootCmd := NewCommand(config)

	rootCmd.AddCommand(
		NewListCommand(config).Command,
		NewDownloadCommand(config).Command,
		NewCleanCommand(config).Command,
		NewSwitchCommand(config).Command,
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
}
