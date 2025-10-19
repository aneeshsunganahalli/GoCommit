package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/dfanso/commit-msg/pkg/types"
	StoreUtils "github.com/dfanso/commit-msg/utils"
)

// CacheManager handles commit message caching operations.
type CacheManager struct {
	config   *types.CacheConfig
	entries  map[string]*types.CacheEntry
	stats    *types.CacheStats
	mutex    sync.RWMutex
	filePath string
	hasher   *DiffHasher
}

// NewCacheManager creates a new cache manager instance.
func NewCacheManager() (*CacheManager, error) {
	config := &types.CacheConfig{
		Enabled:         true,
		MaxEntries:      1000,
		MaxAgeDays:      30,
		CleanupInterval: 24, // 24 hours
		CacheFilePath:   "",
	}

	// Get cache file path
	cachePath, err := getCacheFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache file path: %w", err)
	}
	config.CacheFilePath = cachePath

	cm := &CacheManager{
		config:   config,
		entries:  make(map[string]*types.CacheEntry),
		stats:    &types.CacheStats{},
		filePath: cachePath,
		hasher:   NewDiffHasher(),
	}

	// Load existing cache
	if err := cm.loadCache(); err != nil {
		// If loading fails, start with empty cache
		fmt.Printf("Warning: Failed to load cache: %v\n", err)
	}

	return cm, nil
}

// Get retrieves a cached commit message if it exists.
func (cm *CacheManager) Get(provider types.LLMProvider, diff string, opts *types.GenerationOptions) (*types.CacheEntry, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	key := cm.hasher.GenerateCacheKey(provider, diff, opts)
	entry, exists := cm.entries[key]

	if !exists {
		cm.stats.TotalMisses++
		cm.updateHitRate()
		return nil, false
	}

	// Update access statistics
	entry.LastAccessedAt = time.Now().Format(time.RFC3339)
	entry.AccessCount++
	cm.stats.TotalHits++

	// Update hit rate
	cm.updateHitRate()

	return entry, true
}

// Set stores a commit message in the cache.
func (cm *CacheManager) Set(provider types.LLMProvider, diff string, opts *types.GenerationOptions, message string, cost float64, tokens *types.UsageInfo) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	key := cm.hasher.GenerateCacheKey(provider, diff, opts)
	now := time.Now().Format(time.RFC3339)

	entry := &types.CacheEntry{
		Message:          message,
		Provider:         provider,
		DiffHash:         cm.hasher.GenerateHash(diff, opts),
		StyleInstruction: getStyleInstruction(opts),
		Attempt:          getAttempt(opts),
		CreatedAt:        now,
		LastAccessedAt:   now,
		AccessCount:      1,
		Cost:             cost,
		Tokens:           tokens,
	}

	cm.entries[key] = entry
	cm.stats.TotalEntries = len(cm.entries)

	// Cleanup if we exceed max entries
	if len(cm.entries) > cm.config.MaxEntries {
		cm.cleanupOldEntries()
	}

	// Save to disk
	return cm.saveCache()
}

// Clear removes all entries from the cache.
func (cm *CacheManager) Clear() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.entries = make(map[string]*types.CacheEntry)
	cm.stats = &types.CacheStats{}

	// Remove cache file
	if err := os.Remove(cm.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	return nil
}

// GetStats returns cache statistics.
func (cm *CacheManager) GetStats() *types.CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// Calculate additional stats
	cm.calculateStats()

	// Return a copy to avoid race conditions
	statsCopy := *cm.stats
	return &statsCopy
}

// Cleanup removes old entries based on age and access count.
func (cm *CacheManager) Cleanup() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.cleanupOldEntries()
}

// loadCache loads the cache from disk.
func (cm *CacheManager) loadCache() error {
	if _, err := os.Stat(cm.filePath); os.IsNotExist(err) {
		return nil // No cache file exists yet
	}

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var cacheData struct {
		Entries map[string]*types.CacheEntry `json:"entries"`
		Stats   *types.CacheStats            `json:"stats"`
	}

	if err := json.Unmarshal(data, &cacheData); err != nil {
		return fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	cm.entries = cacheData.Entries
	if cacheData.Stats != nil {
		cm.stats = cacheData.Stats
	}

	return nil
}

// saveCache saves the cache to disk.
func (cm *CacheManager) saveCache() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(cm.filePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cacheData := struct {
		Entries map[string]*types.CacheEntry `json:"entries"`
		Stats   *types.CacheStats            `json:"stats"`
		Config  *types.CacheConfig           `json:"config"`
	}{
		Entries: cm.entries,
		Stats:   cm.stats,
		Config:  cm.config,
	}

	data, err := json.MarshalIndent(cacheData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(cm.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// cleanupOldEntries removes old entries based on age and access count.
func (cm *CacheManager) cleanupOldEntries() error {
	now := time.Now()
	maxAge := time.Duration(cm.config.MaxAgeDays) * 24 * time.Hour

	var keysToDelete []string

	for key, entry := range cm.entries {
		createdAt, err := time.Parse(time.RFC3339, entry.CreatedAt)
		if err != nil {
			// If we can't parse the date, consider it old
			keysToDelete = append(keysToDelete, key)
			continue
		}

		// Remove entries older than max age
		if now.Sub(createdAt) > maxAge {
			keysToDelete = append(keysToDelete, key)
			continue
		}
	}

	// If we still have too many entries, remove least recently accessed
	if len(cm.entries)-len(keysToDelete) > cm.config.MaxEntries {
		cm.removeLeastAccessed(keysToDelete)
	}

	// Delete the selected entries
	for _, key := range keysToDelete {
		delete(cm.entries, key)
	}

	cm.stats.TotalEntries = len(cm.entries)

	return nil
}

// removeLeastAccessed removes the least recently accessed entries.
func (cm *CacheManager) removeLeastAccessed(existingKeysToDelete []string) {
	type entryWithKey struct {
		key   string
		entry *types.CacheEntry
	}

	var entries []entryWithKey
	for key, entry := range cm.entries {
		// Skip entries already marked for deletion
		skip := false
		for _, existingKey := range existingKeysToDelete {
			if key == existingKey {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		entries = append(entries, entryWithKey{key: key, entry: entry})
	}

	// Sort by last accessed time (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, entries[i].entry.LastAccessedAt)
		timeJ, _ := time.Parse(time.RFC3339, entries[j].entry.LastAccessedAt)
		return timeI.Before(timeJ)
	})

	// Add the oldest entries to deletion list
	entriesToRemove := len(entries) - (cm.config.MaxEntries - len(existingKeysToDelete))
	for i := 0; i < entriesToRemove && i < len(entries); i++ {
		existingKeysToDelete = append(existingKeysToDelete, entries[i].key)
	}
}

// updateHitRate calculates and updates the hit rate.
func (cm *CacheManager) updateHitRate() {
	total := cm.stats.TotalHits + cm.stats.TotalMisses
	if total > 0 {
		cm.stats.HitRate = float64(cm.stats.TotalHits) / float64(total)
	}
}

// calculateStats calculates additional statistics.
func (cm *CacheManager) calculateStats() {
	if len(cm.entries) == 0 {
		cm.stats.OldestEntry = ""
		cm.stats.NewestEntry = ""
		cm.stats.CacheSizeBytes = 0
		return
	}

	var oldest, newest time.Time
	var totalCost float64

	for _, entry := range cm.entries {
		createdAt, err := time.Parse(time.RFC3339, entry.CreatedAt)
		if err != nil {
			continue
		}

		if oldest.IsZero() || createdAt.Before(oldest) {
			oldest = createdAt
			cm.stats.OldestEntry = entry.CreatedAt
		}

		if newest.IsZero() || createdAt.After(newest) {
			newest = createdAt
			cm.stats.NewestEntry = entry.CreatedAt
		}

		totalCost += entry.Cost
	}

	cm.stats.TotalCostSaved = totalCost

	// Calculate cache file size
	if stat, err := os.Stat(cm.filePath); err == nil {
		cm.stats.CacheSizeBytes = stat.Size()
	}
}

// getCacheFilePath returns the path to the cache file.
func getCacheFilePath() (string, error) {
	configPath, err := StoreUtils.GetConfigPath()
	if err != nil {
		return "", err
	}

	// Use the same directory as config but with cache.json filename
	configDir := filepath.Dir(configPath)
	return filepath.Join(configDir, "cache.json"), nil
}

// Helper functions
func getStyleInstruction(opts *types.GenerationOptions) string {
	if opts == nil {
		return ""
	}
	return opts.StyleInstruction
}

func getAttempt(opts *types.GenerationOptions) int {
	if opts == nil {
		return 1
	}
	return opts.Attempt
}
