package main

import (
	"fmt"
	"os"
	"path/filepath"

	"git-hotspots/internal/git"
	"git-hotspots/pkg/ui"
)

// testMode is used to disable UI in tests
var testMode bool = false

func main() {
	// Check for test mode flag
	for _, arg := range os.Args {
		if arg == "--test-mode" {
			testMode = true
			// Remove the flag from args
			newArgs := make([]string, 0, len(os.Args)-1)
			for _, a := range os.Args {
				if a != "--test-mode" {
					newArgs = append(newArgs, a)
				}
			}
			os.Args = newArgs
			break
		}
	}

	// Determine the repository path
	repoPath := "."
	if len(os.Args) > 1 {
		repoPath = os.Args[1]
	}

	// Resolve the absolute path
	absoluteRepoPath, err := filepath.Abs(repoPath)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Check if it's a Git repository
	if !git.IsGitRepository(absoluteRepoPath) {
		fmt.Printf("Error: %s is not a Git repository.\n", absoluteRepoPath)
		os.Exit(1)
	}

	// Analyze commits
	commits, err := git.AnalyzeCommits(absoluteRepoPath)
	if err != nil {
		fmt.Printf("Error analyzing commits: %v\n", err)
		os.Exit(1)
	}

	// Identify hotspots
	fileHotspots, dirHotspots := git.IdentifyHotspots(commits)

	// In test mode, just print a summary instead of launching the UI
	if testMode {
		fmt.Println("Git Hotspots Analysis Summary:")
		fmt.Println("\nTop File Hotspots:")
		for i, h := range fileHotspots {
			if i >= 5 {
				break
			}
			fmt.Printf("- %s: %d commits (Top contributor: %s with %d commits)\n", 
				h.Path, h.Commits, h.TopContributor, h.AuthorCommits)
		}
		
		fmt.Println("\nTop Directory Hotspots:")
		for i, h := range dirHotspots {
			if i >= 5 {
				break
			}
			fmt.Printf("- %s: %d commits (Top contributor: %s with %d commits)\n", 
				h.Path, h.Commits, h.TopContributor, h.AuthorCommits)
		}
	} else {
		// Display hotspots in UI
		ui.DisplayHotspots(fileHotspots, dirHotspots)
	}
}


