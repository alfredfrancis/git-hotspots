package ui

import (
	"fmt"
	"sort"

	"git-hotspots/internal/git"

	"github.com/rivo/tview"
)

// DisplayHotspots displays the given file and directory hotspots in a terminal UI.
func DisplayHotspots(fileHotspots []git.Hotspot, dirHotspots []git.Hotspot) {
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
	fmt.Fprintln(fileTextView, "[yellow]Commits  File Path[-]")
	fmt.Fprintln(fileTextView, "[yellow]--------------------[-]")
	for i, hotspot := range fileHotspots {
		if i >= 10 { // Display top 10 files
			break
		}
		fmt.Fprintf(fileTextView, "%7d    %s\n", hotspot.Commits, hotspot.Path)
	}

	// Create a text view for directory hotspots
	dirTextView := tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(false)
	dirTextView.SetBorder(true).SetTitle("Top Hotspot Directories")

	// Populate directory hotspots
	fmt.Fprintln(dirTextView, "[yellow]Commits  Directory Path[-]")
	fmt.Fprintln(dirTextView, "[yellow]------------------------[-]")
	for i, hotspot := range dirHotspots {
		if i >= 10 { // Display top 10 directories
			break
		}
		fmt.Fprintf(dirTextView, "%7d    %s\n", hotspot.Commits, hotspot.Path)
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


