package gc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/sitepod/sitepod/internal/storage"
)

// Config holds GC configuration
type Config struct {
	Enabled     bool          `json:"enabled"`
	Interval    time.Duration `json:"interval"`
	GracePeriod time.Duration `json:"grace_period"`
	MinVersions int           `json:"min_versions"`
	KeepDays    int           `json:"keep_days"`
}

// DefaultConfig returns default GC configuration
func DefaultConfig() Config {
	return Config{
		Enabled:     true,
		Interval:    24 * time.Hour,
		GracePeriod: 1 * time.Hour,
		MinVersions: 5,
		KeepDays:    30,
	}
}

// GC handles garbage collection of unused blobs and expired data
type GC struct {
	app     *pocketbase.PocketBase
	storage storage.Backend
	config  Config
}

// New creates a new GC instance
func New(app *pocketbase.PocketBase, storage storage.Backend, config Config) *GC {
	return &GC{
		app:     app,
		storage: storage,
		config:  config,
	}
}

// Start begins the GC background process
func (gc *GC) Start(ctx context.Context) {
	if !gc.config.Enabled {
		log.Println("GC is disabled")
		return
	}

	ticker := time.NewTicker(gc.config.Interval)
	defer ticker.Stop()

	// Delay initial run to allow migrations to complete
	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
		gc.Run(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("GC shutting down")
			return
		case <-ticker.C:
			gc.Run(ctx)
		}
	}
}

// Run performs a single GC cycle
func (gc *GC) Run(ctx context.Context) {
	log.Println("Starting GC cycle")
	start := time.Now()

	// 1. Clean up expired plans
	expiredPlans, err := gc.cleanExpiredPlans()
	if err != nil {
		log.Printf("Error cleaning expired plans: %v", err)
	}

	// 2. Clean up expired previews
	expiredPreviews, err := gc.cleanExpiredPreviews()
	if err != nil {
		log.Printf("Error cleaning expired previews: %v", err)
	}

	// 3. Clean up unreferenced blobs
	deletedBlobs, err := gc.cleanUnreferencedBlobs()
	if err != nil {
		log.Printf("Error cleaning unreferenced blobs: %v", err)
	}

	log.Printf("GC cycle completed in %v: plans=%d, previews=%d, blobs=%d",
		time.Since(start), expiredPlans, expiredPreviews, deletedBlobs)
}

func (gc *GC) cleanExpiredPlans() (int, error) {
	// Check if collection exists
	collection, err := gc.app.FindCollectionByNameOrId("plans")
	if err != nil || collection == nil {
		return 0, nil // Collection not ready yet, skip
	}

	// Format now as ISO8601 string for PocketBase filter
	now := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")

	// Find expired plans
	plans, err := gc.app.FindRecordsByFilter(
		"plans",
		"status = 'pending' && expires_at < {:now}",
		"",
		100,
		0,
		map[string]any{"now": now},
	)
	if err != nil {
		return 0, err
	}

	updated := 0
	for _, plan := range plans {
		plan.Set("status", "expired")
		if err := gc.app.Save(plan); err != nil {
			return updated, err
		}
		updated++
	}

	return updated, nil
}

func (gc *GC) cleanExpiredPreviews() (int, error) {
	// Check if collection exists
	collection, err := gc.app.FindCollectionByNameOrId("previews")
	if err != nil || collection == nil {
		return 0, nil // Collection not ready yet, skip
	}

	// Format now as ISO8601 string for PocketBase filter
	now := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")

	// Find expired previews
	previews, err := gc.app.FindRecordsByFilter(
		"previews",
		"expires_at < {:now}",
		"",
		100,
		0,
		map[string]any{"now": now},
	)
	if err != nil {
		return 0, err
	}

	deleted := 0
	var firstErr error
	for _, preview := range previews {
		// Delete preview file from storage
		project := preview.GetString("project")
		slug := preview.GetString("slug")
		if err := gc.storage.DeletePreview(project, slug); err != nil && firstErr == nil {
			firstErr = err
		}

		// Delete record
		if err := gc.app.Delete(preview); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		deleted++
	}

	return deleted, firstErr
}

func (gc *GC) cleanUnreferencedBlobs() (int, error) {
	// Check if collection exists
	collection, err := gc.app.FindCollectionByNameOrId("images")
	if err != nil || collection == nil {
		return 0, nil // Collection not ready yet, skip
	}

	// 1. Collect all referenced blob hashes from images
	referencedBlobs := make(map[string]bool)

	images, err := gc.app.FindRecordsByFilter("images", "1=1", "", 10000, 0, nil)
	if err != nil {
		return 0, err
	}

	for _, img := range images {
		var manifest map[string]storage.FileEntry
		if err := json.Unmarshal([]byte(img.GetString("manifest")), &manifest); err != nil {
			continue
		}
		for _, file := range manifest {
			referencedBlobs[file.Hash] = true
		}
	}

	// 2. List all blobs in storage
	allBlobs, err := gc.storage.ListBlobs()
	if err != nil {
		return 0, err
	}

	// 3. Find and delete unreferenced blobs (with grace period)
	deleted := 0
	for _, hash := range allBlobs {
		if referencedBlobs[hash] {
			continue
		}

		// Check grace period
		info, err := gc.storage.StatBlob(hash)
		if err != nil {
			continue
		}

		if time.Since(info.ModTime) < gc.config.GracePeriod {
			continue // Too new, might be in progress
		}

		// Delete
		if err := gc.storage.DeleteBlob(hash); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

// RunDryRun performs a dry run GC and returns what would be deleted
func (gc *GC) RunDryRun() (*DryRunResult, error) {
	result := &DryRunResult{}

	// Format now as ISO8601 string for PocketBase filter
	now := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")

	// Count expired plans
	plans, err := gc.app.FindRecordsByFilter(
		"plans",
		"status = 'pending' && expires_at < {:now}",
		"",
		1000,
		0,
		map[string]any{"now": now},
	)
	if err == nil {
		result.ExpiredPlans = len(plans)
	}

	// Count expired previews
	previews, err := gc.app.FindRecordsByFilter(
		"previews",
		"expires_at < {:now}",
		"",
		1000,
		0,
		map[string]any{"now": now},
	)
	if err == nil {
		result.ExpiredPreviews = len(previews)
	}

	// Count unreferenced blobs
	referencedBlobs := make(map[string]bool)
	images, _ := gc.app.FindRecordsByFilter("images", "1=1", "", 10000, 0, nil)
	for _, img := range images {
		var manifest map[string]storage.FileEntry
		if err := json.Unmarshal([]byte(img.GetString("manifest")), &manifest); err != nil {
			continue
		}
		for _, file := range manifest {
			referencedBlobs[file.Hash] = true
		}
	}

	allBlobs, _ := gc.storage.ListBlobs()
	for _, hash := range allBlobs {
		if !referencedBlobs[hash] {
			info, err := gc.storage.StatBlob(hash)
			if err != nil {
				continue
			}
			if time.Since(info.ModTime) >= gc.config.GracePeriod {
				result.UnreferencedBlobs++
				result.ReclaimableBytes += info.Size
			}
		}
	}

	return result, nil
}

// DryRunResult contains the results of a dry run
type DryRunResult struct {
	ExpiredPlans      int   `json:"expired_plans"`
	ExpiredPreviews   int   `json:"expired_previews"`
	UnreferencedBlobs int   `json:"unreferenced_blobs"`
	ReclaimableBytes  int64 `json:"reclaimable_bytes"`
}
