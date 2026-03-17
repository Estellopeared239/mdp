<p align="center">
  <img src="assets/dock-icon.png" width="128" height="128" alt="mdp icon">
</p>

# mdp

A CLI tool that renders markdown files in a native frameless macOS window. Designed for terminal-only agents like Claude Code.

## Features

- Native frameless window with rounded corners and shadow
- CommonMark + GFM rendering with syntax highlighting (Chroma)
- KaTeX math and Mermaid diagram support (lazy-loaded)
- Live reload on file changes (fsnotify + polling fallback)
- Multi-window support — each file opens in its own window (re-opening brings existing window to front)
- Table of contents sidebar
- Dark/light/system theme with CSS custom properties
- Vim-inspired keyboard shortcuts
- Pipe from stdin for fire-and-forget previews

## Install

### Homebrew (macOS)

```bash
brew tap mxcoppell/tap
brew install mdp

# Upgrade to latest version
brew update && brew upgrade mdp
```

### Build from source

Requires Go 1.24+ and macOS (CGO + Cocoa).

```bash
git clone https://github.com/mxcoppell/mdp.git
cd mdp
make deps    # download vendored JS dependencies (first time only)
make build   # debug build → ./bin/mdp
```

## Usage

```bash
# Preview a file with live reload
mdp README.md

# Preview multiple files (each in its own window)
mdp doc1.md doc2.md

# Pipe from stdin
echo "# Hello World" | mdp
cat README.md | mdp

# Options
mdp --theme dark README.md
mdp --toc README.md
mdp --browser README.md    # open in system browser instead
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down / up |
| `n` / `p` | Next / previous heading |
| `]` | Toggle table of contents |
| `Cmd+F` | Search |
| `T` | Toggle theme |
| `+` / `-` | Zoom in / out |
| `0` | Reset zoom |
| `h` | Show shortcuts |
| `Esc` | Close window |

## Build

```bash
make build     # debug build with symbols (~21 MB)
make release   # stripped release build (~16 MB)
make test      # run all tests
make clean     # remove build artifacts
```

## License

MIT
