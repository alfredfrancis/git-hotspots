package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestRepo creates a temporary git repository for testing.
func setupTestRepo(t *testing.T) string {
	// Create a temporary directory
	tmpDir, err := ioutil.TempDir("", "git-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize a git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user (required for commits)
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user email: %v", err)
	}
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git user name: %v", err)
	}

	return tmpDir
}

// createCommit creates a commit with the given files and message.
func createCommit(t *testing.T, repoPath string, files []string, message string, commitTime time.Time) {
	for _, file := range files {
		filePath := filepath.Join(repoPath, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := ioutil.WriteFile(filePath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
		cmd := exec.Command("git", "add", file)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to add file %s: %v", file, err)
		}
	}

	// Set GIT_AUTHOR_DATE and GIT_COMMITTER_DATE for reproducible commit dates
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_AUTHOR_DATE=%s", commitTime.Format(time.RFC3339)),
		fmt.Sprintf("GIT_COMMITTER_DATE=%s", commitTime.Format(time.RFC3339)),
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
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
			Author:  "Test User",
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

	fileMap := make(map[string]int)
	for _, h := range fileHotspots {
		fileMap[h.Path] = h.Commits
	}

	if fileMap["fileA.txt"] != 3 {
		t.Errorf("Expected fileA.txt to have 3 commits, got %d", fileMap["fileA.txt"])
	}
	if fileMap["dir1/fileB.txt"] != 1 {
		t.Errorf("Expected dir1/fileB.txt to have 1 commit, got %d", fileMap["dir1/fileB.txt"])
	}

	// Check directory hotspots
	if len(dirHotspots) != 2 {
		t.Errorf("Expected 2 directory hotspots, got %d", len(dirHotspots))
	}

	dirMap := make(map[string]int)
	for _, h := range dirHotspots {
		dirMap[h.Path] = h.Commits
	}

	if dirMap["dir1"] != 2 {
		t.Errorf("Expected dir1 to have 2 commits, got %d", dirMap["dir1"])
	}
	if dirMap["dir2"] != 1 {
		t.Errorf("Expected dir2 to have 1 commit, got %d", dirMap["dir2"])
	}
}


