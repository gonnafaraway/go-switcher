# Go Version Switcher

A simple CLI tool to manage multiple Go versions on your system.

## Features

- Download specific Go versions
- List installed versions
- Switch between versions
- Clean up old versions
- Support for different architectures

## Installation

```bash
# Install the latest version
go install github.com/gonnafaraway/go-switcher@latest

# Install a specific version
go install github.com/gonnafaraway/go-switcher@v1.0.0
```

## Usage

### List installed versions

```bash
go-switcher list
```

Output:
```
= Go versions from /usr/local/bin/go-switcher =
1. 1.23.0 (linux-amd64)
2. 1.22.0 (linux-amd64)
3. 1.23.0 (darwin-amd64)
```

### Download a Go version

```bash
# Download specific version
go-switcher download 1.23.0

# Download with specific architecture
go-switcher download 1.23.0 --arch darwin-amd64
```

### Switch Go version

```bash
# Switch using version number
go-switcher switch 1.23.0

# Switch using list number
go-switcher switch 1

# Switch with specific architecture
go-switcher switch 1.23.0 --arch darwin-amd64
```

### Clean up versions

```bash
go-switcher clean
```

## Directory Structure

```
/usr/local/bin/go-switcher/
├── 1.23.0/
│   ├── linux-amd64/
│   │   ├── go/          # GOROOT
│   │   └── workspace/   # GOPATH
│   └── darwin-amd64/
│       ├── go/
│       └── workspace/
└── 1.22.0/
    └── linux-amd64/
        ├── go/
        └── workspace/
```

## Environment Variables

The tool manages these environment variables in your `~/.profile`:

- `PATH`: Updated to include the selected Go version's bin directory
- `GOPATH`: Set to the version-specific workspace directory
- `GOROOT`: Set to the Go installation directory

## Requirements

- Go 1.16 or later
- Linux/macOS (Windows support planned)
- `wget` and `tar` commands

## License

MIT 
