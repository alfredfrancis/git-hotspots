package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"git-hotspots/internal/git"
	"git-hotspots/pkg/ui"
)

// testMode is used to disable UI in tests
var testMode bool = false

func main() {
	// Define flags
	topCount := flag.Int("top", 10, "Number of top files and directories to display")
	flag.Bool("test-mode", false, "Run in test mode (no UI)")
	
	// Parse flags
	flag.Parse()
	
	// Check for test mode flag
	if flag.Lookup("test-mode").Value.String() == "true" {
		testMode = true
	}

	// Determine the repository path
	repoPath := "."
	if flag.NArg() > 0 {
		repoPath = flag.Arg(0)
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
		displayCount := 5 // Default for test mode
		if *topCount < displayCount {
			displayCount = *topCount
		}
		
		for i, h := range fileHotspots {
			if i >= displayCount {
				break
			}
			fmt.Printf("- %s: %d commits (Top contributor: %s with %d commits)\n", 
				h.Path, h.Commits, h.TopContributor, h.AuthorCommits)
		}
		
		fmt.Println("\nTop Directory Hotspots:")
		for i, h := range dirHotspots {
			if i >= displayCount {
				break
			}
			fmt.Printf("- %s: %d commits (Top contributor: %s with %d commits)\n", 
				h.Path, h.Commits, h.TopContributor, h.AuthorCommits)
		}
	} else {
		// Display hotspots in UI
		ui.DisplayHotspots(fileHotspots, dirHotspots, *topCount)
	}
}


