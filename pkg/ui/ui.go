package ui

import (
	"fmt"
	"sort"

	"git-hotspots/internal/git"

	"github.com/rivo/tview"
)

// DisplayHotspots displays the given file and directory hotspots in a terminal UI.
// topCount specifies the number of top files and directories to display.
func DisplayHotspots(fileHotspots []git.Hotspot, dirHotspots []git.Hotspot, topCount int) {
	app := tview.NewApplication()

	// Sort hotspots for consistent display
	sort.Slice(fileHotspots, func(i, j int) bool {
		return fileHotspots[i].Commits > fileHotspots[j].Commits
	})
	sort.Slice(dirHotspots, func(i, j int) bool {
		return dirHotspots[i].Commits > dirHotspots[j].Commits
	})

	// Create a text view for file hotspots
	fileTextView := tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(false)
	fileTextView.SetBorder(true).SetTitle("Top Hotspot Files")

	// Populate file hotspots
	fmt.Fprintln(fileTextView, "[yellow]Commits  Top Contributor (Commits)  File Path[-]")
	fmt.Fprintln(fileTextView, "[yellow]-----------------------------------------------[-]")
	for i, hotspot := range fileHotspots {
		if i >= topCount { // Display top N files
			break
		}
		fmt.Fprintf(fileTextView, "%7d    %-20s (%d)    %s\n", 
			hotspot.Commits, 
			hotspot.TopContributor, 
			hotspot.AuthorCommits,
			hotspot.Path)
	}

	// Create a text view for directory hotspots
	dirTextView := tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(false)
	dirTextView.SetBorder(true).SetTitle("Top Hotspot Directories")

	// Populate directory hotspots
	fmt.Fprintln(dirTextView, "[yellow]Commits  Top Contributor (Commits)  Directory Path[-]")
	fmt.Fprintln(dirTextView, "[yellow]---------------------------------------------------[-]")
	for i, hotspot := range dirHotspots {
		if i >= topCount { // Display top N directories
			break
		}
		fmt.Fprintf(dirTextView, "%7d    %-20s (%d)    %s\n", 
			hotspot.Commits, 
			hotspot.TopContributor, 
			hotspot.AuthorCommits,
			hotspot.Path)
	}

	// Create a flex layout to arrange the text views
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(fileTextView, 0, 1, false).
		AddItem(dirTextView, 0, 1, false)

	// Set the root primitive and run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}


