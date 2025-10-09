package display

import (
	"fmt"
	"testing"
)

func TestShowFileStatistics(t *testing.T) {
	t.Parallel()

	t.Run("displays staged files", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			StagedFiles: []string{"file1.go", "file2.go", "file3.go"},
			TotalFiles:  3,
		}

		// Just test that the function doesn't panic
		ShowFileStatistics(stats)
	})

	t.Run("displays unstaged files", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			UnstagedFiles: []string{"file1.js", "file2.js"},
			TotalFiles:    2,
		}

		// Just test that the function doesn't panic
		ShowFileStatistics(stats)
	})

	t.Run("displays untracked files", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			UntrackedFiles: []string{"newfile.txt"},
			TotalFiles:     1,
		}

		// Just test that the function doesn't panic
		ShowFileStatistics(stats)
	})

	t.Run("limits displayed files", func(t *testing.T) {
		t.Parallel()

		// Create more files than the display limits
		stagedFiles := make([]string, MaxStagedFiles+2)
		for i := range stagedFiles {
			stagedFiles[i] = fmt.Sprintf("file%d.go", i)
		}

		stats := &FileStatistics{
			StagedFiles: stagedFiles,
			TotalFiles:  len(stagedFiles),
		}

		// Just test that the function doesn't panic
		ShowFileStatistics(stats)
	})

	t.Run("handles empty statistics", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			StagedFiles:    []string{},
			UnstagedFiles:  []string{},
			UntrackedFiles: []string{},
			TotalFiles:     0,
		}

		// Just test that the function doesn't panic
		ShowFileStatistics(stats)
	})
}

func TestShowCommitMessage(t *testing.T) {
	t.Parallel()

	t.Run("displays commit message", func(t *testing.T) {
		t.Parallel()

		message := "feat: add new feature"
		// Just test that the function doesn't panic
		ShowCommitMessage(message)
	})

	t.Run("handles empty message", func(t *testing.T) {
		t.Parallel()

		message := ""
		// Just test that the function doesn't panic
		ShowCommitMessage(message)
	})

	t.Run("handles multiline message", func(t *testing.T) {
		t.Parallel()

		message := "feat: add new feature\n\nThis is a detailed description\nwith multiple lines"
		// Just test that the function doesn't panic
		ShowCommitMessage(message)
	})
}

func TestShowChangesPreview(t *testing.T) {
	t.Parallel()

	t.Run("displays line statistics", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			LinesAdded:   10,
			LinesDeleted: 5,
			TotalFiles:   3,
		}

		// Just test that the function doesn't panic
		ShowChangesPreview(stats)
	})

	t.Run("handles zero statistics", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			LinesAdded:   0,
			LinesDeleted: 0,
			TotalFiles:   0,
		}

		// Just test that the function doesn't panic
		ShowChangesPreview(stats)
	})

	t.Run("handles only added lines", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			LinesAdded:   15,
			LinesDeleted: 0,
			TotalFiles:   2,
		}

		// Just test that the function doesn't panic
		ShowChangesPreview(stats)
	})

	t.Run("handles only deleted lines", func(t *testing.T) {
		t.Parallel()

		stats := &FileStatistics{
			LinesAdded:   0,
			LinesDeleted: 8,
			TotalFiles:   1,
		}

		// Just test that the function doesn't panic
		ShowChangesPreview(stats)
	})
}

func TestFileStatisticsConstants(t *testing.T) {
	t.Parallel()

	if MaxStagedFiles <= 0 {
		t.Fatal("MaxStagedFiles should be positive")
	}

	if MaxUnstagedFiles <= 0 {
		t.Fatal("MaxUnstagedFiles should be positive")
	}

	if MaxUntrackedFiles <= 0 {
		t.Fatal("MaxUntrackedFiles should be positive")
	}
}

func TestFileStatisticsStruct(t *testing.T) {
	t.Parallel()

	stats := &FileStatistics{
		StagedFiles:    []string{"file1.go", "file2.go"},
		UnstagedFiles:  []string{"file3.js"},
		UntrackedFiles: []string{"file4.txt"},
		TotalFiles:     3,
		LinesAdded:     10,
		LinesDeleted:   5,
	}

	if len(stats.StagedFiles) != 2 {
		t.Fatalf("expected 2 staged files, got %d", len(stats.StagedFiles))
	}

	if len(stats.UnstagedFiles) != 1 {
		t.Fatalf("expected 1 unstaged file, got %d", len(stats.UnstagedFiles))
	}

	if len(stats.UntrackedFiles) != 1 {
		t.Fatalf("expected 1 untracked file, got %d", len(stats.UntrackedFiles))
	}

	if stats.TotalFiles != 3 {
		t.Fatalf("expected 3 total files, got %d", stats.TotalFiles)
	}

	if stats.LinesAdded != 10 {
		t.Fatalf("expected 10 lines added, got %d", stats.LinesAdded)
	}

	if stats.LinesDeleted != 5 {
		t.Fatalf("expected 5 lines deleted, got %d", stats.LinesDeleted)
	}
}
