package main

import (
	"bytes"
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

func TestCLIIntegration(t *testing.T) {
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Create some commits
	now := time.Now()
	createCommit(t, tmpDir, []string{"file1.txt"}, "Initial commit", now.Add(-24*time.Hour))
	createCommit(t, tmpDir, []string{"file1.txt", "file2.txt"}, "Add file2", now.Add(-12*time.Hour))
	createCommit(t, tmpDir, []string{"dir1/file3.txt"}, "Add file3 in dir1", now.Add(-6*time.Hour))

	// Build the CLI tool
	buildCmd := exec.Command("go", "build", "-o", "git-hotspots", ".")
	buildCmd.Dir = "/home/ubuntu/git-hotspots/cmd/git-hotspots"
	var buildErr bytes.Buffer
	buildCmd.Stderr = &buildErr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build git-hotspots executable: %v\nStderr: %s", err, buildErr.String())
	}

	// Run the CLI tool against the test repository
	cliCmd := exec.Command("./git-hotspots", tmpDir)
	cliCmd.Dir = "/home/ubuntu/git-hotspots/cmd/git-hotspots"
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

	cliCmd = exec.Command("./git-hotspots", nonGitDir)
	cliCmd.Dir = "/home/ubuntu/git-hotspots/cmd/git-hotspots"
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


