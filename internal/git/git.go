package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// IsGitRepository checks if the given path is a Git repository.
func IsGitRepository(path string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// CommitInfo holds information about a commit.
type CommitInfo struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
	Files   []string
}

// AnalyzeCommits analyzes git commits in the last year and returns commit information.
func AnalyzeCommits(repoPath string) ([]CommitInfo, error) {
	var commits []CommitInfo

	// Get commits from the last year
	cmd := exec.Command("git", "log", "--pretty=format:%H|%an|%ad|%s", "--date=iso", "--name-only", "--since=1 year")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git log: %w", err)
	}

	rawCommits := strings.Split(strings.TrimSpace(string(output)), "\n\n")

	for _, rawCommit := range rawCommits {
		if rawCommit == "" {
			continue
		}
		parts := strings.SplitN(rawCommit, "\n", 2)
		if len(parts) < 2 {
			continue
		}

		header := parts[0]
		fileList := strings.Split(strings.TrimSpace(parts[1]), "\n")

		headerParts := strings.SplitN(header, "|", 4)
		if len(headerParts) < 4 {
			continue
		}

		commitHash := headerParts[0]
		author := headerParts[1]
		dateStr := headerParts[2]
		message := headerParts[3]

		commitDate, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %s: %w", dateStr, err)
		}

		var files []string
		for _, file := range fileList {
			if file != "" {
				files = append(files, file)
			}
		}

		commits = append(commits, CommitInfo{
			Hash:    commitHash,
			Author:  author,
			Date:    commitDate,
			Message: message,
			Files:   files,
		})
	}

	return commits, nil
}

// Hotspot represents a file or directory with its commit count.
type Hotspot struct {
	Path  string
	Commits int
}

// IdentifyHotspots identifies hotspot files and directories.
func IdentifyHotspots(commits []CommitInfo) ([]Hotspot, []Hotspot) {
	fileCommits := make(map[string]int)
	dirCommits := make(map[string]int)

	for _, commit := range commits {
		for _, file := range commit.Files {
			fileCommits[file]++
			dir := filepath.Dir(file)
			if dir != "." {
				dirCommits[dir]++
			}
		}
	}

	var fileHotspots []Hotspot
	for path, count := range fileCommits {
		fileHotspots = append(fileHotspots, Hotspot{Path: path, Commits: count})
	}

	var dirHotspots []Hotspot
	for path, count := range dirCommits {
		dirHotspots = append(dirHotspots, Hotspot{Path: path, Commits: count})
	}

	// Sort hotspots by commit count in descending order
	// (Sorting will be done in a separate utility function or later in UI)

	return fileHotspots, dirHotspots
}


