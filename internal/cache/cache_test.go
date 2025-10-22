package cache

import (
	"path/filepath"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestCacheManager_GetSet(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create cache manager with custom path
	cm := &CacheManager{
		config: &types.CacheConfig{
			Enabled:         true,
			MaxEntries:      1000,
			MaxAgeDays:      30,
			CleanupInterval: 24,
			CacheFilePath:   filepath.Join(tempDir, "test-cache.json"),
		},
		entries:  make(map[string]*types.CacheEntry),
		stats:    &types.CacheStats{},
		filePath: filepath.Join(tempDir, "test-cache.json"),
		hasher:   NewDiffHasher(),
	}

	// Test data
	provider := types.ProviderOpenAI
	diff := "test diff content"
	opts := &types.GenerationOptions{
		StyleInstruction: "test style",
		Attempt:          1,
	}
	message := "test commit message"
	cost := 0.001

	// Test setting a cache entry
	err := cm.Set(provider, diff, opts, message, cost, nil)
	if err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	// Test getting the cache entry
	entry, found := cm.Get(provider, diff, opts)
	if !found {
		t.Fatalf("Cache entry not found after setting")
	}

	if entry.Message != message {
		t.Errorf("Expected message %s, got %s", message, entry.Message)
	}

	if entry.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, entry.Provider)
	}

	if entry.Cost != cost {
		t.Errorf("Expected cost %f, got %f", cost, entry.Cost)
	}

	// Test getting non-existent entry
	_, found = cm.Get(types.ProviderClaude, diff, opts)
	if found {
		t.Errorf("Expected cache miss for different provider")
	}
}

func TestCacheManager_Stats(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create cache manager with custom path
	cm := &CacheManager{
		config: &types.CacheConfig{
			Enabled:         true,
			MaxEntries:      1000,
			MaxAgeDays:      30,
			CleanupInterval: 24,
			CacheFilePath:   filepath.Join(tempDir, "test-cache.json"),
		},
		entries:  make(map[string]*types.CacheEntry),
		stats:    &types.CacheStats{},
		filePath: filepath.Join(tempDir, "test-cache.json"),
		hasher:   NewDiffHasher(),
	}

	// Initial stats should be empty
	stats := cm.GetStats()
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries, got %d", stats.TotalEntries)
	}

	// Add some entries
	provider := types.ProviderOpenAI
	diff1 := "test diff 1"
	diff2 := "test diff 2"
	opts := &types.GenerationOptions{Attempt: 1}

	cm.Set(provider, diff1, opts, "message 1", 0.001, nil)
	cm.Set(provider, diff2, opts, "message 2", 0.002, nil)

	// Test cache hit
	_, found := cm.Get(provider, diff1, opts)
	if !found {
		t.Errorf("Expected cache hit")
	}

	// Test cache miss
	_, found = cm.Get(types.ProviderClaude, diff1, opts)
	if found {
		t.Errorf("Expected cache miss for different provider")
	}

	// Test another cache miss
	_, found = cm.Get(provider, "different diff", opts)
	if found {
		t.Errorf("Expected cache miss for different diff")
	}

	// Check stats
	stats = cm.GetStats()
	if stats.TotalEntries != 2 {
		t.Errorf("Expected 2 entries, got %d", stats.TotalEntries)
	}

	if stats.TotalHits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.TotalHits)
	}

	if stats.TotalMisses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.TotalMisses)
	}

	expectedHitRate := float64(stats.TotalHits) / float64(stats.TotalHits+stats.TotalMisses)
	if stats.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %f, got %f", expectedHitRate, stats.HitRate)
	}
}

func TestCacheManager_Clear(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create cache manager with custom path
	cm := &CacheManager{
		config: &types.CacheConfig{
			Enabled:         true,
			MaxEntries:      1000,
			MaxAgeDays:      30,
			CleanupInterval: 24,
			CacheFilePath:   filepath.Join(tempDir, "test-cache.json"),
		},
		entries:  make(map[string]*types.CacheEntry),
		stats:    &types.CacheStats{},
		filePath: filepath.Join(tempDir, "test-cache.json"),
		hasher:   NewDiffHasher(),
	}

	// Add some entries
	provider := types.ProviderOpenAI
	diff := "test diff"
	opts := &types.GenerationOptions{Attempt: 1}

	cm.Set(provider, diff, opts, "message", 0.001, nil)

	// Verify entry exists
	_, found := cm.Get(provider, diff, opts)
	if !found {
		t.Fatalf("Cache entry not found after setting")
	}

	// Clear cache
	err := cm.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify entry is gone
	_, found = cm.Get(provider, diff, opts)
	if found {
		t.Errorf("Cache entry found after clearing")
	}

	// Check stats
	stats := cm.GetStats()
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.TotalEntries)
	}
}

func TestCacheManager_Persistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, "test-cache.json")

	// Create first cache manager and add entry
	cm1 := &CacheManager{
		config: &types.CacheConfig{
			Enabled:         true,
			MaxEntries:      1000,
			MaxAgeDays:      30,
			CleanupInterval: 24,
			CacheFilePath:   cacheFile,
		},
		entries:  make(map[string]*types.CacheEntry),
		stats:    &types.CacheStats{},
		filePath: cacheFile,
		hasher:   NewDiffHasher(),
	}

	provider := types.ProviderOpenAI
	diff := "test diff"
	opts := &types.GenerationOptions{Attempt: 1}
	message := "test message"
	cost := 0.001

	err := cm1.Set(provider, diff, opts, message, cost, nil)
	if err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	// Create second cache manager (should load from file)
	cm2 := &CacheManager{
		config: &types.CacheConfig{
			Enabled:         true,
			MaxEntries:      1000,
			MaxAgeDays:      30,
			CleanupInterval: 24,
			CacheFilePath:   cacheFile,
		},
		entries:  make(map[string]*types.CacheEntry),
		stats:    &types.CacheStats{},
		filePath: cacheFile,
		hasher:   NewDiffHasher(),
	}

	// Load from file
	if err := cm2.loadCache(); err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	// Verify entry exists in second manager
	entry, found := cm2.Get(provider, diff, opts)
	if !found {
		t.Fatalf("Cache entry not found in second manager")
	}

	if entry.Message != message {
		t.Errorf("Expected message %s, got %s", message, entry.Message)
	}
}

func TestDiffHasher_NormalizeDiff(t *testing.T) {
	hasher := NewDiffHasher()

	diff := `diff --git a/file.txt b/file.txt
index 1234567..abcdefg 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2 updated
 line3`

	normalized := hasher.normalizeDiff(diff)

	// Should not contain git headers
	if containsString(normalized, "diff --git") {
		t.Errorf("Normalized diff should not contain git headers")
	}

	if containsString(normalized, "index ") {
		t.Errorf("Normalized diff should not contain index line")
	}

	if containsString(normalized, "+++") {
		t.Errorf("Normalized diff should not contain +++ line")
	}

	if containsString(normalized, "---") {
		t.Errorf("Normalized diff should not contain --- line")
	}

	// Should contain the actual changes
	if !containsString(normalized, "-line2") {
		t.Errorf("Normalized diff should contain removed line")
	}

	if !containsString(normalized, "+line2 updated") {
		t.Errorf("Normalized diff should contain added line")
	}
}
