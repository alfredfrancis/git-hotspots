package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec" // Still needed for CLI commands
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

func TestCLIIntegration(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Create some commits
	now := time.Now()
	createCommit(t, tmpDir, []string{"file1.txt"}, "Initial commit", now.Add(-24*time.Hour))
	createCommit(t, tmpDir, []string{"file1.txt", "file2.txt"}, "Add file2", now.Add(-12*time.Hour))
	createCommit(t, tmpDir, []string{"dir1/file3.txt"}, "Add file3 in dir1", now.Add(-6*time.Hour))

	// Build the CLI tool
	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	buildCmd := exec.Command("go", "build", "-o", "git-hotspots", ".")
	buildCmd.Dir = currentDir
	var buildErr bytes.Buffer
	buildCmd.Stderr = &buildErr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build git-hotspots executable: %v\nStderr: %s", err, buildErr.String())
	}

	// Run the CLI tool against the test repository with test mode flag
	cliCmd := exec.Command("./git-hotspots", "--test-mode", tmpDir)
	cliCmd.Dir = currentDir
	var out bytes.Buffer
	cliCmd.Stdout = &out
	cliCmd.Stderr = &out // Capture stderr as well

	// tview requires a terminal, so running it directly in a test will fail.
	// For integration tests, we can only check if the command exits successfully
	// and if there are no unexpected errors printed to stdout/stderr.
	// A more robust integration test would involve mocking the tview library
	// or using a pseudo-terminal, which is out of scope for a basic CLI test.

	// We need to prevent the tview UI from launching during tests.
	// One way is to pass an environment variable or a flag to the main function
	// to indicate that it's running in test mode and should skip UI display.
	// For simplicity, we'll just check for the expected error output for now.

	if err := cliCmd.Run(); err != nil {
		t.Errorf("CLI tool failed with error: %v\nOutput: %s", err, out.String())
	}

	// Basic check: ensure no panic/fatal errors are printed
	outputStr := out.String()
	if strings.Contains(outputStr, "Error:") || strings.Contains(outputStr, "panic:") {
		t.Errorf("CLI tool output contains errors or panics: %s", outputStr)
	}

	// Test case for non-git directory
	nonGitDir, err := ioutil.TempDir("", "non-git-test-")
	if err != nil {
		t.Fatalf("Failed to create non-git temp dir: %v", err)
	}
	defer os.RemoveAll(nonGitDir)

	cliCmd = exec.Command("./git-hotspots", "--test-mode", nonGitDir)
	cliCmd.Dir = currentDir
	out.Reset()
	cliCmd.Stdout = &out
	cliCmd.Stderr = &out

	if err := cliCmd.Run(); err == nil {
		t.Errorf("Expected CLI tool to fail for non-git directory, but it succeeded")
	}
	outputStr = out.String()
	if !strings.Contains(outputStr, "is not a Git repository") {
		t.Errorf("Expected error message for non-git repository, got: %s", outputStr)
	}
}


