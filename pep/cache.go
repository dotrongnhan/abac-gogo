package pep

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"abac_go_example/models"
)

// DecisionCache provides caching for policy decisions
type DecisionCache struct {
	cache   map[string]*cacheEntry
	maxSize int
	ttl     time.Duration
	mu      sync.RWMutex
	stats   CacheStats
}

// cacheEntry represents a cached decision with metadata
type cacheEntry struct {
	result    *EnforcementResult
	timestamp time.Time
	hits      int64
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	HitRatio  float64 `json:"hit_ratio"`
}

// NewDecisionCache creates a new decision cache
func NewDecisionCache(maxSize int, ttl time.Duration) *DecisionCache {
	cache := &DecisionCache{
		cache:   make(map[string]*cacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a cached decision
func (c *DecisionCache) Get(request *models.EvaluationRequest) *EnforcementResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(request)
	entry, exists := c.cache[key]

	if !exists {
		c.stats.Misses++
		return nil
	}

	// Check if entry has expired
	if time.Since(entry.timestamp) > c.ttl {
		// Remove expired entry (will be cleaned up later)
		delete(c.cache, key)
		c.stats.Misses++
		return nil
	}

	// Update hit count and stats
	entry.hits++
	c.stats.Hits++

	// Create a copy to avoid race conditions
	result := &EnforcementResult{
		Decision:         entry.result.Decision,
		Allowed:          entry.result.Allowed,
		Reason:           entry.result.Reason,
		EvaluationTimeMs: entry.result.EvaluationTimeMs,
		CacheHit:         true, // Mark as cache hit
		Timestamp:        entry.result.Timestamp,
	}

	return result
}

// Set stores a decision in the cache
func (c *DecisionCache) Set(request *models.EvaluationRequest, result *EnforcementResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateKey(request)

	// Check if we need to evict entries
	if len(c.cache) >= c.maxSize {
		c.evictLRU()
	}

	// Store the entry
	c.cache[key] = &cacheEntry{
		result:    result,
		timestamp: time.Now(),
		hits:      0,
	}

	c.stats.Size = len(c.cache)
}

// generateKey creates a cache key from the evaluation request
func (c *DecisionCache) generateKey(request *models.EvaluationRequest) string {
	// Create a deterministic key from request components
	keyData := struct {
		SubjectID  string                 `json:"subject_id"`
		ResourceID string                 `json:"resource_id"`
		Action     string                 `json:"action"`
		Context    map[string]interface{} `json:"context"`
	}{
		SubjectID:  request.SubjectID,
		ResourceID: request.ResourceID,
		Action:     request.Action,
		Context:    request.Context,
	}

	jsonData, _ := json.Marshal(keyData)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

// evictLRU removes the least recently used entry
func (c *DecisionCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	var minHits int64 = -1

	// Find the entry with the oldest timestamp and lowest hits
	for key, entry := range c.cache {
		if minHits == -1 || entry.hits < minHits ||
			(entry.hits == minHits && entry.timestamp.Before(oldestTime)) {
			oldestKey = key
			oldestTime = entry.timestamp
			minHits = entry.hits
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		c.stats.Evictions++
	}
}

// cleanup periodically removes expired entries
func (c *DecisionCache) cleanup() {
	ticker := time.NewTicker(time.Minute * 5) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()

		for key, entry := range c.cache {
			if now.Sub(entry.timestamp) > c.ttl {
				delete(c.cache, key)
			}
		}

		c.stats.Size = len(c.cache)
		c.mu.Unlock()
	}
}

// GetStats returns current cache statistics
func (c *DecisionCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.cache)

	total := stats.Hits + stats.Misses
	if total > 0 {
		stats.HitRatio = float64(stats.Hits) / float64(total)
	}

	return stats
}

// Clear removes all entries from the cache
func (c *DecisionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.stats = CacheStats{}
}

// Size returns the current number of cached entries
func (c *DecisionCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}
