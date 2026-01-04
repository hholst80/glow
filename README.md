# Glow

Render markdown on the CLI, with _pizzazz_!

> **Note:** This is a fork of [charmbracelet/glow](https://github.com/charmbracelet/glow).
> See [Installation](#installation) for how to get binaries.

<p align="center">
    <img src="https://stuff.charm.sh/glow/glow-banner-github.gif" alt="Glow Logo">
    <a href="https://github.com/hholst80/glow/releases"><img src="https://img.shields.io/github/release/hholst80/glow.svg" alt="Latest Release"></a>
    <a href="https://github.com/hholst80/glow/actions"><img src="https://github.com/hholst80/glow/workflows/build/badge.svg" alt="Build Status"></a>
</p>

<p align="center">
    <img src="https://github.com/user-attachments/assets/c2246366-f84b-4847-b431-32a61ca07b74" width="800" alt="Glow UI Demo">
</p>

## What is it?

Glow is a terminal based markdown reader designed from the ground up to bring
out the beauty—and power—of the CLI.

Use it to discover markdown files, read documentation directly on the command
line. Glow will find local markdown files in subdirectories or a local
Git repository.

## Installation

### Download Binary

Download a pre-built static binary from the [GitHub Releases][releases] page.

**Available platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)
- FreeBSD (amd64, arm64)
- OpenBSD (amd64, arm64)
- NetBSD (amd64, arm64)

```bash
# Example: Download and install on Linux
curl -L https://github.com/hholst80/glow/releases/latest/download/glow_Linux_x86_64.tar.gz | tar xz
sudo mv glow /usr/local/bin/
```

### Build from Source (requires Go 1.21+)

```bash
git clone https://github.com/hholst80/glow.git
cd glow
go build
```

[releases]: https://github.com/hholst80/glow/releases

## The TUI

Simply run `glow` without arguments to start the textual user interface and
browse local. Glow will find local markdown files in the
current directory and below or, if you're in a Git repository, Glow will search
the repo.

Markdown files can be read with Glow's high-performance pager. Most of the
keystrokes you know from `less` are the same, but you can press `?` to list
the hotkeys.

### Outline Sidebar

Press `o` to toggle a right-aligned outline sidebar that shows a hierarchical
tree of markdown headings. The sidebar highlights the current section as you
scroll and supports jump-to navigation:

- `o` - Toggle outline visibility
- `Tab` - Switch focus between content and outline
- `j/k` - Navigate headings (when outline focused)
- `Enter` - Jump to selected heading
- `]/[` - Quick next/prev heading navigation

Enable outline on startup with `--outline` or `-o` flag, or set `showOutline: true`
in your config file.

## The CLI

In addition to a TUI, Glow has a CLI for working with Markdown. To format a
document use a markdown source as the primary argument:

```bash
# Read from file
glow README.md

# Read from stdin
echo "[Glow](https://github.com/charmbracelet/glow)" | glow -

# Fetch README from GitHub / GitLab
glow github.com/charmbracelet/glow

# Fetch markdown from HTTP
glow https://host.tld/file.md
```

### Word Wrapping

The `-w` flag lets you set a maximum width at which the output will be wrapped:

```bash
glow -w 60
```

### Paging

CLI output can be displayed in your preferred pager with the `-p` flag. This defaults
to the ANSI-aware `less -r` if `$PAGER` is not explicitly set.

### Styles

You can choose a style with the `-s` flag. When no flag is provided `glow` tries
to detect your terminal's current background color and automatically picks
either the `dark` or the `light` style for you.

```bash
glow -s [dark|light]
```

Alternatively you can also supply a custom JSON stylesheet:

```bash
glow -s mystyle.json
```

For additional usage details see:

```bash
glow --help
```

Check out the [Glamour Style Section](https://github.com/charmbracelet/glamour/blob/master/styles/gallery/README.md)
to find more styles. Or [make your own](https://github.com/charmbracelet/glamour/tree/master/styles)!

## The Config File

If you find yourself supplying the same flags to `glow` all the time, it's
probably a good idea to create a config file. Run `glow config`, which will open
it in your favorite $EDITOR. Alternatively you can manually put a file named
`glow.yml` in the default config path of you platform. If you're not sure where
that is, please refer to `glow --help`.

Here's an example config:

```yaml
# style name or JSON path (default "auto")
style: "light"
# mouse wheel support (TUI-mode only)
mouse: true
# use pager to display markdown
pager: true
# at which column should we word wrap?
width: 80
# show all files, including hidden and ignored.
all: false
# show line numbers (TUI-mode only)
showLineNumbers: false
# show outline sidebar (TUI-mode only)
showOutline: false
# preserve newlines in the output
preserveNewLines: false
```

## Contributing

See [AGENTS.md](AGENTS.md) for development guidelines.

## License

[MIT](https://github.com/hholst80/glow/raw/master/LICENSE)

---

This is a fork of [charmbracelet/glow](https://github.com/charmbracelet/glow).
Original project by [Charm](https://charm.sh).
