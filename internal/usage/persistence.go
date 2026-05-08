package usage

import (
	"context"
	"strings"
	"sync"
	"time"

	coreusage "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/usage"
)

// PersistentRecord stores the normalized request detail used for database persistence.
type PersistentRecord struct {
	ID        string
	APIName   string
	ModelName string
	Detail    RequestDetail
}

// PersistenceBackend persists normalized usage records and restores them on startup.
type PersistenceBackend interface {
	LoadRecords(ctx context.Context) ([]PersistentRecord, error)
	SaveRecords(ctx context.Context, records []PersistentRecord) error
}

var (
	persistenceMu      sync.RWMutex
	persistenceBackend PersistenceBackend
)

// SetPersistenceBackend swaps the global persistence backend and returns the previous one.
func SetPersistenceBackend(store PersistenceBackend) PersistenceBackend {
	persistenceMu.Lock()
	defer persistenceMu.Unlock()

	prev := persistenceBackend
	persistenceBackend = store
	return prev
}

func getPersistenceBackend() PersistenceBackend {
	persistenceMu.RLock()
	defer persistenceMu.RUnlock()
	return persistenceBackend
}

// ConfigurePersistence restores persisted usage records into the provided statistics store
// and enables subsequent persistence for new records.
func ConfigurePersistence(ctx context.Context, stats *RequestStatistics, store PersistenceBackend) error {
	if store == nil {
		SetPersistenceBackend(nil)
		return nil
	}
	if stats == nil {
		stats = GetRequestStatistics()
	}
	records, err := store.LoadRecords(ctx)
	if err != nil {
		return err
	}
	stats.ReplaceWithRecords(records)
	SetPersistenceBackend(store)
	return nil
}

// PersistSnapshot writes a management snapshot into the configured persistence backend.
func PersistSnapshot(ctx context.Context, snapshot StatisticsSnapshot) error {
	return StorePersistentRecords(ctx, RecordsFromSnapshot(snapshot))
}

// StorePersistentRecords writes normalized records to the configured persistence backend.
func StorePersistentRecords(ctx context.Context, records []PersistentRecord) error {
	store := getPersistenceBackend()
	if store == nil || len(records) == 0 {
		return nil
	}
	return store.SaveRecords(ctx, records)
}

// RecordsFromSnapshot converts a statistics snapshot into normalized persistent records.
func RecordsFromSnapshot(snapshot StatisticsSnapshot) []PersistentRecord {
	if len(snapshot.APIs) == 0 {
		return nil
	}
	records := make([]PersistentRecord, 0)
	for apiName, apiSnapshot := range snapshot.APIs {
		for modelName, modelSnapshot := range apiSnapshot.Models {
			for _, detail := range modelSnapshot.Details {
				records = append(records, BuildPersistentRecordFromDetail(apiName, modelName, detail))
			}
		}
	}
	return records
}

// BuildPersistentRecord normalizes a runtime usage record into a persistable detail.
func BuildPersistentRecord(ctx context.Context, record coreusage.Record) PersistentRecord {
	timestamp := record.RequestedAt
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	apiName := strings.TrimSpace(record.APIKey)
	if apiName == "" {
		apiName = resolveAPIIdentifier(ctx, record)
	}
	failed := record.Failed
	if !failed {
		failed = !resolveSuccess(ctx)
	}
	return normalisePersistentRecord(PersistentRecord{
		APIName:   apiName,
		ModelName: record.Model,
		Detail: RequestDetail{
			Timestamp: timestamp,
			LatencyMs: normaliseLatency(record.Latency),
			Source:    record.Source,
			AuthIndex: record.AuthIndex,
			Tokens:    normaliseDetail(record.Detail),
			Failed:    failed,
		},
	})
}

// BuildPersistentRecordFromDetail normalizes a snapshot detail into a persistable record.
func BuildPersistentRecordFromDetail(apiName, modelName string, detail RequestDetail) PersistentRecord {
	return normalisePersistentRecord(PersistentRecord{
		APIName:   apiName,
		ModelName: modelName,
		Detail:    detail,
	})
}

func normalisePersistentRecord(record PersistentRecord) PersistentRecord {
	record.APIName = strings.TrimSpace(record.APIName)
	if record.APIName == "" {
		record.APIName = "unknown"
	}
	record.ModelName = strings.TrimSpace(record.ModelName)
	if record.ModelName == "" {
		record.ModelName = "unknown"
	}
	if record.Detail.Timestamp.IsZero() {
		record.Detail.Timestamp = time.Now()
	}
	if record.Detail.LatencyMs < 0 {
		record.Detail.LatencyMs = 0
	}
	record.Detail.Tokens = normaliseTokenStats(record.Detail.Tokens)
	record.ID = strings.TrimSpace(record.ID)
	if record.ID == "" {
		record.ID = dedupKey(record.APIName, record.ModelName, record.Detail)
	}
	return record
}
