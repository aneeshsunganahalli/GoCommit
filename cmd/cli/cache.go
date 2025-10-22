package cmd

import (
	"fmt"
	"time"

	"github.com/dfanso/commit-msg/cmd/cli/store"
	"github.com/pterm/pterm"
)

// ShowCacheStats displays cache statistics.
func ShowCacheStats(Store *store.StoreMethods) error {
	stats := Store.GetCacheStats()

	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).
		WithTextStyle(pterm.NewStyle(pterm.FgWhite, pterm.Bold)).
		Println("Commit Message Cache Statistics")

	pterm.Println()

	// Create statistics table
	statsData := [][]string{
		{"Total Entries", fmt.Sprintf("%d", stats.TotalEntries)},
		{"Cache Hits", fmt.Sprintf("%d", stats.TotalHits)},
		{"Cache Misses", fmt.Sprintf("%d", stats.TotalMisses)},
		{"Hit Rate", fmt.Sprintf("%.2f%%", stats.HitRate*100)},
		{"Total Cost Saved", fmt.Sprintf("$%.4f", stats.TotalCostSaved)},
		{"Cache Size", formatBytes(stats.CacheSizeBytes)},
	}

	if stats.OldestEntry != "" {
		statsData = append(statsData, []string{"Oldest Entry", formatTime(stats.OldestEntry)})
	}

	if stats.NewestEntry != "" {
		statsData = append(statsData, []string{"Newest Entry", formatTime(stats.NewestEntry)})
	}

	pterm.DefaultTable.WithHasHeader(false).WithData(statsData).Render()

	pterm.Println()

	// Show cache status
	if stats.TotalEntries == 0 {
		pterm.Info.Println("Cache is empty. Generate some commit messages to start building the cache.")
	} else {
		pterm.Success.Printf("Cache is active with %d entries\n", stats.TotalEntries)

		if stats.HitRate > 0 {
			pterm.Info.Printf("Cache hit rate: %.2f%% (%.0f%% of requests served from cache)\n",
				stats.HitRate*100, stats.HitRate*100)
		}

		if stats.TotalCostSaved > 0 {
			pterm.Success.Printf("Total cost saved: $%.4f\n", stats.TotalCostSaved)
		}
	}

	return nil
}

// ClearCache removes all cached messages.
func ClearCache(Store *store.StoreMethods) error {
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).
		WithTextStyle(pterm.NewStyle(pterm.FgWhite, pterm.Bold)).
		Println("Clear Cache")

	pterm.Println()

	// Get current stats before clearing
	stats := Store.GetCacheStats()

	if stats.TotalEntries == 0 {
		pterm.Info.Println("Cache is already empty.")
		return nil
	}

	// Confirm before clearing
	confirm, err := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(false).
		Show(fmt.Sprintf("Are you sure you want to clear %d cached entries? This action cannot be undone.", stats.TotalEntries))

	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirm {
		pterm.Info.Println("Cache clear cancelled.")
		return nil
	}

	// Clear the cache
	if err := Store.ClearCache(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	pterm.Success.Println("Cache cleared successfully!")
	pterm.Info.Printf("Removed %d entries and saved $%.4f in future API costs\n",
		stats.TotalEntries, stats.TotalCostSaved)

	return nil
}

// CleanupCache removes old cached messages.
func CleanupCache(Store *store.StoreMethods) error {
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgYellow)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack, pterm.Bold)).
		Println("Cleanup Cache")

	pterm.Println()

	// Get stats before cleanup
	statsBefore := Store.GetCacheStats()

	if statsBefore.TotalEntries == 0 {
		pterm.Info.Println("Cache is empty. Nothing to cleanup.")
		return nil
	}

	pterm.Info.Println("Removing old and unused cached entries...")

	// Perform cleanup
	if err := Store.CleanupCache(); err != nil {
		return fmt.Errorf("failed to cleanup cache: %w", err)
	}

	// Get stats after cleanup
	statsAfter := Store.GetCacheStats()
	removed := statsBefore.TotalEntries - statsAfter.TotalEntries

	if removed > 0 {
		pterm.Success.Printf("Cleanup completed! Removed %d old entries.\n", removed)
		pterm.Info.Printf("Cache now contains %d entries\n", statsAfter.TotalEntries)
	} else {
		pterm.Info.Println("No old entries found to remove.")
	}

	return nil
}

// Helper functions

// formatBytes formats bytes into human-readable format.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatTime formats a timestamp string for display.
func formatTime(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("2006-01-02 15:04:05")
}
