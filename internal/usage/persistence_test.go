package usage

import (
	"context"
	"testing"
	"time"
)

type memoryPersistence struct {
	saved []PersistentRecord
}

func (m *memoryPersistence) LoadRecords(context.Context) ([]PersistentRecord, error) {
	return []PersistentRecord{
		{
			ID:        "persisted-1",
			APIName:   "persisted-key",
			ModelName: "gpt-5.5",
			Detail: RequestDetail{
				Timestamp: time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
				LatencyMs: 1200,
				Source:    "codex@example.com",
				AuthIndex: "auth-1",
				Tokens: TokenStats{
					InputTokens:     10,
					OutputTokens:    20,
					ReasoningTokens: 5,
					CachedTokens:    3,
					TotalTokens:     35,
				},
			},
		},
	}, nil
}

func (m *memoryPersistence) SaveRecords(_ context.Context, records []PersistentRecord) error {
	m.saved = append(m.saved, records...)
	return nil
}

func TestConfigurePersistenceRestoresSnapshot(t *testing.T) {
	prevStore := SetPersistenceBackend(nil)
	t.Cleanup(func() {
		SetPersistenceBackend(prevStore)
	})

	stats := NewRequestStatistics()
	stats.RecordPersistent(PersistentRecord{
		ID:        "stale",
		APIName:   "stale-key",
		ModelName: "gpt-5.4",
		Detail: RequestDetail{
			Timestamp: time.Date(2026, 5, 8, 10, 0, 0, 0, time.UTC),
			Tokens: TokenStats{
				TotalTokens: 1,
			},
		},
	})

	store := &memoryPersistence{}
	if err := ConfigurePersistence(context.Background(), stats, store); err != nil {
		t.Fatalf("ConfigurePersistence error: %v", err)
	}

	snapshot := stats.Snapshot()
	if snapshot.TotalRequests != 1 {
		t.Fatalf("total requests = %d, want 1", snapshot.TotalRequests)
	}
	if _, ok := snapshot.APIs["stale-key"]; ok {
		t.Fatal("stale in-memory statistics were not replaced")
	}
	modelStats := snapshot.APIs["persisted-key"].Models["gpt-5.5"]
	if modelStats.TotalTokens != 35 {
		t.Fatalf("total tokens = %d, want 35", modelStats.TotalTokens)
	}
	if len(modelStats.Details) != 1 {
		t.Fatalf("details len = %d, want 1", len(modelStats.Details))
	}
}
