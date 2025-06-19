# Git Hotspots CLI Tool

This is a command-line interface (CLI) tool written in Go that analyzes Git commit history to identify code hotspots. Hotspots are defined as files and directories with the most commits within the last year, indicating areas of frequent change.

## Features

- Checks if the current directory is a Git repository.
- Analyzes Git commits from the last 1 year.
- Identifies top hotspot files and directories based on commit count.
- Presents the hotspots in a clear, terminal-based user interface.

## Installation

To install the `git-hotspots` CLI tool, follow these steps:

1.  **Ensure Go is installed:** If you don't have Go installed, you can download it from the official Go website: [https://golang.org/doc/install](https://golang.org/doc/install)

2.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-username/git-hotspots.git # Replace with actual repo URL if available
    cd git-hotspots
    ```

3.  **Build the executable:**
    ```bash
    go build -o git-hotspots ./cmd/git-hotspots
    ```

    This will create an executable named `git-hotspots` in the current directory.

4.  **(Optional) Add to your PATH:** To run the tool from any directory, move the executable to a directory included in your system's PATH (e.g., `/usr/local/bin`):
    ```bash
    sudo mv git-hotspots /usr/local/bin/
    ```

## Usage

Navigate to the root directory of a Git repository you want to analyze and run the tool:

```bash
git-hotspots
```

Alternatively, you can specify the path to a Git repository:

```bash
git-hotspots /path/to/your/repo
```

The tool will display a terminal UI showing the top 10 hotspot files and top 10 hotspot directories.

## Example Output

```
┌───────────────────────────────Top Hotspot Files──────────────────────────────┐
│Commits  File Path                                                            │
│--------------------                                                          │
│      2    file1.txt                                                          │
│      1    src/util.go                                                        │
│      1    src/main.go                                                        │
│      1    file2.txt                                                          │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
┌────────────────────────────Top Hotspot Directories───────────────────────────┐
│Commits  Directory Path                                                       │
│------------------------                                                      │
│      2    src                                                                │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Development

### Project Structure

-   `cmd/git-hotspots/main.go`: The main entry point for the CLI application.
-   `internal/git/`: Contains the core logic for Git repository analysis.
-   `pkg/ui/`: Contains the logic for the terminal user interface.

### Running Tests

To run the unit and integration tests, navigate to the project root and execute:

```bash
go test ./...
```

## Contributing

Feel free to open issues or submit pull requests if you have suggestions or improvements.


