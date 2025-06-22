package main

import (
	"fmt"
	"os"
	"path/filepath"

	"git-hotspots/internal/git"
	"git-hotspots/pkg/ui"
)

func main() {
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

	// Display hotspots in UI
	ui.DisplayHotspots(fileHotspots, dirHotspots)
}


