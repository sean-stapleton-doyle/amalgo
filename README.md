# amalgo
[![Build](https://github.com/sean-stapleton-doyle/amalgo/actions/workflows/go.yml/badge.svg)](https://github.com/sean-stapleton-doyle/amalgo/actions/workflows/go.yml) 

`amalgo` is a command-line tool that recursively scans a directory for files with specified extensions and concatenates them into a single, amalgamated output file. It allows for filters files using `.gitignore` rules, custom patterns, and directory exclusions.

I use it for creating a single context file of a project's source code to use with AI coding assistants or for generating simple documentation bundles.

## Installation

You can install `amalgo` by building from the source.

### From Source

1.  Clone the repository:

    ```bash
    git clone https://github.com/sean-stapleton-doyle/amalgo.git
    cd amalgo
    ```

2.  Build the binary:

    ```bash
    make build
    ```

    This will create the `amalgo` executable in the project's root directory.
    To install to your path:
    ```
    make install
    ```

## Usage

The basic command requires you to specify the file extension(s) to include.

```bash
amalgo --ext <extension> [flags]
```

### Examples

**1. Concatenate all `.go` files in the current directory**
This will create an output file named `concat.md`.

```bash
amalgo -e .go
```

**2. Combine `.rs` and `.toml` files into a single output file**
You can specify extensions as a comma-separated list or by repeating the flag.

```bash
amalgo --ext .rs,.toml --out rust_project.md
```

or

```bash
amalgo -e .rs -e .toml -o rust_project.md
```

**3. Scan a specific directory and pipe the output to the clipboard**
This example scans the `src` directory for Python files and uses `pbcopy` (macOS) to copy the result.

```bash
amalgo -d ./src -e .py -o - | pbcopy
```

**4. Include hidden files and add a custom ignore pattern**
This command includes hidden files (like `.config.js`) but specifically ignores any `*.test.js` files.

```bash
amalgo -e .js --include-hidden --ignore-pattern "*.test.js"
```

-----

## Command-line Flags

Here is a complete list of available flags:

| Flag | Shorthand | Description | Default |
| :--- | :---: | :--- | :--- |
| `--dir` | `-d` | Root directory to scan. | `.` |
| `--ext` | `-e` | File extension(s) to include. Can be repeated or comma-separated. | **(Required)** |
| `--out` | `-o` | Output file path. Use `-` for standard output. | `concat.<format>` |
| `--ignore-dirs` | `-i` | Directory names to ignore. | `.git`, `node_modules`, `vendor` |
| `--ignore-pattern`| `-p` | Custom gitignore-style patterns to exclude. Can be repeated. | `     ` |
| `--heading-level` | `-l` | Markdown heading level for file headers (1-6). | `1` |
| `--format` | `-f` | Output format. | `markdown` |
| `--include-hidden`| | Include hidden files and directories (those starting with `.`). | `false` |
| `--use-gitignore` | | Automatically use `.gitignore` in the base directory if present. | `true` |
| `--gitignore` | `-g` | Path to a specific `.gitignore` file to use. | Auto-detected |

-----

## Development

This project uses a `Makefile` for common dev tasks.

### Prerequisites

  * Go (version 1.23+)
  * `golangci-lint` (for the `lint` target)

### Key Makefile Commands

  * `make build`: Build the `amalgo` binary.
  * `make test`: Run all tests.
  * `make fmt`: Format the Go source code.
  * `make lint`: Run the linter to check for code style and errors.
  * `make tidy`: Tidy and verify Go modules.
  * `make run`: Build and run the tool with a default example (`-e .go`).
  * `make clean`: Remove build artifacts.

To see all available commands, run:

```bash
make help
```

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
