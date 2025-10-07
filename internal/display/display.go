package display

import (
	"fmt"

	"github.com/pterm/pterm"
)

// FileStatistics holds statistics about changed files
type FileStatistics struct {
	StagedFiles    []string
	UnstagedFiles  []string
	UntrackedFiles []string
	TotalFiles     int
	LinesAdded     int
	LinesDeleted   int
}

// ShowFileStatistics displays file statistics with colored output
func ShowFileStatistics(stats *FileStatistics) {
	pterm.DefaultSection.Println("Changes Summary")

	// Create bullet list items
	bulletItems := []pterm.BulletListItem{}

	if len(stats.StagedFiles) > 0 {
		bulletItems = append(bulletItems, pterm.BulletListItem{
			Level:       0,
			Text:        pterm.Green(fmt.Sprintf("Staged files: %d", len(stats.StagedFiles))),
			TextStyle:   pterm.NewStyle(pterm.FgGreen),
			BulletStyle: pterm.NewStyle(pterm.FgGreen),
		})
		for i, file := range stats.StagedFiles {
			if i < 5 { // Show first 5 files
				bulletItems = append(bulletItems, pterm.BulletListItem{
					Level: 1,
					Text:  file,
				})
			}
		}
		if len(stats.StagedFiles) > 5 {
			bulletItems = append(bulletItems, pterm.BulletListItem{
				Level: 1,
				Text:  pterm.Gray(fmt.Sprintf("... and %d more", len(stats.StagedFiles)-5)),
			})
		}
	}

	if len(stats.UnstagedFiles) > 0 {
		bulletItems = append(bulletItems, pterm.BulletListItem{
			Level:       0,
			Text:        pterm.Yellow(fmt.Sprintf("Unstaged files: %d", len(stats.UnstagedFiles))),
			TextStyle:   pterm.NewStyle(pterm.FgYellow),
			BulletStyle: pterm.NewStyle(pterm.FgYellow),
		})
		for i, file := range stats.UnstagedFiles {
			if i < 3 {
				bulletItems = append(bulletItems, pterm.BulletListItem{
					Level: 1,
					Text:  file,
				})
			}
		}
		if len(stats.UnstagedFiles) > 3 {
			bulletItems = append(bulletItems, pterm.BulletListItem{
				Level: 1,
				Text:  pterm.Gray(fmt.Sprintf("... and %d more", len(stats.UnstagedFiles)-3)),
			})
		}
	}

	if len(stats.UntrackedFiles) > 0 {
		bulletItems = append(bulletItems, pterm.BulletListItem{
			Level:       0,
			Text:        pterm.Cyan(fmt.Sprintf("Untracked files: %d", len(stats.UntrackedFiles))),
			TextStyle:   pterm.NewStyle(pterm.FgCyan),
			BulletStyle: pterm.NewStyle(pterm.FgCyan),
		})
		for i, file := range stats.UntrackedFiles {
			if i < 3 {
				bulletItems = append(bulletItems, pterm.BulletListItem{
					Level: 1,
					Text:  file,
				})
			}
		}
		if len(stats.UntrackedFiles) > 3 {
			bulletItems = append(bulletItems, pterm.BulletListItem{
				Level: 1,
				Text:  pterm.Gray(fmt.Sprintf("... and %d more", len(stats.UntrackedFiles)-3)),
			})
		}
	}

	pterm.DefaultBulletList.WithItems(bulletItems).Render()
}

// ShowCommitMessage displays the commit message in a styled panel
func ShowCommitMessage(message string) {
	pterm.DefaultSection.Println("Generated Commit Message")

	// Create a panel with the commit message
	panel := pterm.DefaultBox.
		WithTitle("Commit Message").
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgLightGreen)).
		WithHorizontalString("─").
		WithVerticalString("│").
		WithTopLeftCornerString("┌").
		WithTopRightCornerString("┐").
		WithBottomLeftCornerString("└").
		WithBottomRightCornerString("┘")

	panel.Println(pterm.LightGreen(message))
}

// ShowChangesPreview displays a preview of changes with line statistics
func ShowChangesPreview(stats *FileStatistics) {
	pterm.DefaultSection.Println("Changes Preview")

	// Create info boxes
	if stats.LinesAdded > 0 || stats.LinesDeleted > 0 {
		infoData := [][]string{
			{"Lines Added", pterm.Green(fmt.Sprintf("+%d", stats.LinesAdded))},
			{"Lines Deleted", pterm.Red(fmt.Sprintf("-%d", stats.LinesDeleted))},
			{"Total Files", pterm.Cyan(fmt.Sprintf("%d", stats.TotalFiles))},
		}

		pterm.DefaultTable.WithHasHeader(false).WithData(infoData).Render()
	} else {
		pterm.Info.Println("No line statistics available for unstaged changes")
	}
}
