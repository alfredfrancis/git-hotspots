package git

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// setupTestRepo creates a temporary git repository for testing.
func setupTestRepo(t *testing.T) string {
	// Create a temporary directory
	tmpDir, err := ioutil.TempDir("", "git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize a git repository
	_, err = git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// We'll set the user config in the createCommit function instead
	// as we don't need global config for our tests

	return tmpDir
}

// createCommit creates a commit with the given files and message.
func createCommit(t *testing.T, repoPath string, files []string, message string, commitTime time.Time) {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		t.Fatalf("Failed to open repository: %v", err)
	}

	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("Failed to get worktree: %v", err)
	}

	// Create and add files
	for _, file := range files {
		filePath := filepath.Join(repoPath, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := ioutil.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
		
		// Add the file to the staging area
		_, err = wt.Add(file)
		if err != nil {
			t.Fatalf("Failed to add file %s: %v", file, err)
		}
	}

	// Create commit with the specified time
	commit, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  commitTime,
		},
		Committer: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  commitTime,
		},
	})
	if err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify the commit was created
	_, err = repo.CommitObject(commit)
	if err != nil {
		t.Fatalf("Failed to get commit object: %v", err)
	}
}

func TestIsGitRepository(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	if !IsGitRepository(tmpDir) {
		t.Errorf("Expected %s to be a git repository, but it's not", tmpDir)
	}

	nonGitDir, err := ioutil.TempDir("", "non-git-test-")
	if err != nil {
		t.Fatalf("Failed to create non-git temp dir: %v", err)
	}
	defer os.RemoveAll(nonGitDir)

	if IsGitRepository(nonGitDir) {
		t.Errorf("Expected %s not to be a git repository, but it is", nonGitDir)
	}
}

func TestAnalyzeCommits(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	now := time.Now()
	createCommit(t, tmpDir, []string{"file1.txt"}, "Initial commit", now.Add(-24*time.Hour))
	createCommit(t, tmpDir, []string{"file1.txt", "file2.txt"}, "Add file2", now.Add(-12*time.Hour))
	createCommit(t, tmpDir, []string{"dir1/file3.txt"}, "Add file3 in dir1", now.Add(-6*time.Hour))

	commits, err := AnalyzeCommits(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCommits failed: %v", err)
	}

	if len(commits) != 3 {
		t.Errorf("Expected 3 commits, got %d", len(commits))
	}

	// Check the latest commit
	latestCommit := commits[0] // git log returns in reverse chronological order
	if !strings.Contains(latestCommit.Message, "Add file3") {
		t.Errorf("Expected latest commit message to contain 'Add file3', got %s", latestCommit.Message)
	}
	if len(latestCommit.Files) != 1 || latestCommit.Files[0] != "dir1/file3.txt" {
		t.Errorf("Expected latest commit to affect dir1/file3.txt, got %v", latestCommit.Files)
	}

	// Test --since=1 year filter
	oldCommitTime := now.Add(-366 * 24 * time.Hour) // More than 1 year ago
	createCommit(t, tmpDir, []string{"old_file.txt"}, "Old commit", oldCommitTime)

	commitsAfterOld, err := AnalyzeCommits(tmpDir)
	if err != nil {
		t.Fatalf("AnalyzeCommits failed after adding old commit: %v", err)
	}

	if len(commitsAfterOld) != 3 {
		t.Errorf("Expected 3 commits (filtered by 1 year), got %d", len(commitsAfterOld))
	}
}

func TestIdentifyHotspots(t *testing.T) {
	commits := []CommitInfo{
		{
			Hash:    "hash1",
			Author:  "Test User",
			Date:    time.Now(),
			Message: "Commit 1",
			Files:   []string{"fileA.txt", "dir1/fileB.txt"},
		},
		{
			Hash:    "hash2",
			Author:  "Another User",
			Date:    time.Now(),
			Message: "Commit 2",
			Files:   []string{"fileA.txt", "dir2/fileC.txt"},
		},
		{
			Hash:    "hash3",
			Author:  "Test User",
			Date:    time.Now(),
			Message: "Commit 3",
			Files:   []string{"fileA.txt", "dir1/fileD.txt"},
		},
	}

	fileHotspots, dirHotspots := IdentifyHotspots(commits)

	// Check file hotspots
	if len(fileHotspots) != 4 {
		t.Errorf("Expected 4 file hotspots, got %d", len(fileHotspots))
	}

	// Create maps for easier testing
	fileMap := make(map[string]Hotspot)
	for _, h := range fileHotspots {
		fileMap[h.Path] = h
	}

	// Check fileA.txt (should have 3 commits, with Test User as top contributor with 2 commits)
	fileA := fileMap["fileA.txt"]
	if fileA.Commits != 3 {
		t.Errorf("Expected fileA.txt to have 3 commits, got %d", fileA.Commits)
	}
	if fileA.TopContributor != "Test User" {
		t.Errorf("Expected fileA.txt top contributor to be 'Test User', got '%s'", fileA.TopContributor)
	}
	if fileA.AuthorCommits != 2 {
		t.Errorf("Expected fileA.txt top contributor to have 2 commits, got %d", fileA.AuthorCommits)
	}

	// Check dir1/fileB.txt (should have 1 commit from Test User)
	fileB := fileMap["dir1/fileB.txt"]
	if fileB.Commits != 1 {
		t.Errorf("Expected dir1/fileB.txt to have 1 commit, got %d", fileB.Commits)
	}
	if fileB.TopContributor != "Test User" {
		t.Errorf("Expected dir1/fileB.txt top contributor to be 'Test User', got '%s'", fileB.TopContributor)
	}

	// Check directory hotspots
	if len(dirHotspots) != 2 {
		t.Errorf("Expected 2 directory hotspots, got %d", len(dirHotspots))
	}

	dirMap := make(map[string]Hotspot)
	for _, h := range dirHotspots {
		dirMap[h.Path] = h
	}

	// Check dir1 (should have 2 commits, with Test User as top contributor with 2 commits)
	dir1 := dirMap["dir1"]
	if dir1.Commits != 2 {
		t.Errorf("Expected dir1 to have 2 commits, got %d", dir1.Commits)
	}
	if dir1.TopContributor != "Test User" {
		t.Errorf("Expected dir1 top contributor to be 'Test User', got '%s'", dir1.TopContributor)
	}
	if dir1.AuthorCommits != 2 {
		t.Errorf("Expected dir1 top contributor to have 2 commits, got %d", dir1.AuthorCommits)
	}

	// Check dir2 (should have 1 commit from Another User)
	dir2 := dirMap["dir2"]
	if dir2.Commits != 1 {
		t.Errorf("Expected dir2 to have 1 commit, got %d", dir2.Commits)
	}
	if dir2.TopContributor != "Another User" {
		t.Errorf("Expected dir2 top contributor to be 'Another User', got '%s'", dir2.TopContributor)
	}
}


