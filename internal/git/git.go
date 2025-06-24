package git

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// IsGitRepository checks if the given path is a Git repository.
func IsGitRepository(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
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

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Create a new log options
	since := time.Now().AddDate(-1, 0, 0) // Last year
	logOptions := &git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
		Since: &since,
	}

	// Get the commit iterator
	commitIter, err := repo.Log(logOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit iterator: %w", err)
	}

	// Iterate through the commits
	err = commitIter.ForEach(func(c *object.Commit) error {
		// Get the files changed in this commit
		fileStats, err := getFilesInCommit(c)
		if err != nil {
			return fmt.Errorf("failed to get files in commit %s: %w", c.Hash.String(), err)
		}

		var files []string
		for _, fs := range fileStats {
			files = append(files, fs)
		}

		// Create a CommitInfo object
		commitInfo := CommitInfo{
			Hash:    c.Hash.String(),
			Author:  c.Author.Name,
			Date:    c.Author.When,
			Message: c.Message,
			Files:   files,
		}

		commits = append(commits, commitInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate through commits: %w", err)
	}

	return commits, nil
}

// Hotspot represents a file or directory with its commit count and top contributor.
type Hotspot struct {
	Path           string
	Commits        int
	TopContributor string
	AuthorCommits  int
}

// getFilesInCommit returns a list of files changed in a commit
func getFilesInCommit(commit *object.Commit) ([]string, error) {
	var files []string

	// Get the commit tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	// Check if this commit has parents
	parents := commit.Parents()
	parentsCount := commit.NumParents()

	if parentsCount == 0 {
		// If this is the first commit (no parents), list all files in the tree
		err = tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// For each parent, get the changes
		seenFiles := make(map[string]bool)
		
		// Close the parents iterator when done
		defer parents.Close()
		
		// Iterate through all parents
		for {
			parent, err := parents.Next()
			if err == plumbing.ErrObjectNotFound {
				// Skip this parent if not found
				continue
			} else if err != nil {
				// End of parents or other error
				break
			}
			
			// Get parent tree
			parentTree, err := parent.Tree()
			if err != nil {
				continue // Skip this parent if we can't get its tree
			}
			
			// Get changes between parent and this commit
			changes, err := tree.Diff(parentTree)
			if err != nil {
				continue // Skip this parent if we can't get changes
			}
			
			// Extract file paths from changes
			for _, change := range changes {
				action, err := change.Action()
				if err != nil {
					continue
				}
				
				// Only include files that were added, modified, or deleted
				if action == merkletrie.Insert || action == merkletrie.Modify || action == merkletrie.Delete {
					if change.From.Name != "" && !seenFiles[change.From.Name] {
						files = append(files, change.From.Name)
						seenFiles[change.From.Name] = true
					} else if change.To.Name != "" && !seenFiles[change.To.Name] {
						files = append(files, change.To.Name)
						seenFiles[change.To.Name] = true
					}
				}
			}
		}
		
		// If we couldn't get any files from parents, try to list all files in the tree
		if len(files) == 0 {
			err = tree.Files().ForEach(func(f *object.File) error {
				files = append(files, f.Name)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}

// IdentifyHotspots identifies hotspot files and directories.
func IdentifyHotspots(commits []CommitInfo) ([]Hotspot, []Hotspot) {
	fileCommits := make(map[string]int)
	dirCommits := make(map[string]int)
	fileAuthors := make(map[string]map[string]int) // file -> author -> commit count
	dirAuthors := make(map[string]map[string]int)  // dir -> author -> commit count

	// Initialize maps
	for _, commit := range commits {
		author := commit.Author
		for _, file := range commit.Files {
			// Track file commits
			fileCommits[file]++
			
			// Track file authors
			if _, ok := fileAuthors[file]; !ok {
				fileAuthors[file] = make(map[string]int)
			}
			fileAuthors[file][author]++
			
			// Track directory commits
			dir := filepath.Dir(file)
			if dir != "." {
				dirCommits[dir]++
				
				// Track directory authors
				if _, ok := dirAuthors[dir]; !ok {
					dirAuthors[dir] = make(map[string]int)
				}
				dirAuthors[dir][author]++
			}
		}
	}

	// Create file hotspots with top contributor information
	var fileHotspots []Hotspot
	for path, count := range fileCommits {
		topContributor := ""
		topContributions := 0
		
		// Find top contributor for this file
		for author, authorCommits := range fileAuthors[path] {
			if authorCommits > topContributions {
				topContributor = author
				topContributions = authorCommits
			}
		}
		
		fileHotspots = append(fileHotspots, Hotspot{
			Path:           path,
			Commits:        count,
			TopContributor: topContributor,
			AuthorCommits:  topContributions,
		})
	}

	// Create directory hotspots with top contributor information
	var dirHotspots []Hotspot
	for path, count := range dirCommits {
		topContributor := ""
		topContributions := 0
		
		// Find top contributor for this directory
		for author, authorCommits := range dirAuthors[path] {
			if authorCommits > topContributions {
				topContributor = author
				topContributions = authorCommits
			}
		}
		
		dirHotspots = append(dirHotspots, Hotspot{
			Path:           path,
			Commits:        count,
			TopContributor: topContributor,
			AuthorCommits:  topContributions,
		})
	}

	// Sort hotspots by commit count in descending order
	// (Sorting will be done in a separate utility function or later in UI)

	return fileHotspots, dirHotspots
}


