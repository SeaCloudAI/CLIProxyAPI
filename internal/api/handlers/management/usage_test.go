package management

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/usage"
)

type usagePersistenceRecorder struct {
	saved []usage.PersistentRecord
}

func (r *usagePersistenceRecorder) LoadRecords(context.Context) ([]usage.PersistentRecord, error) {
	return nil, nil
}

func (r *usagePersistenceRecorder) SaveRecords(_ context.Context, records []usage.PersistentRecord) error {
	r.saved = append(r.saved, records...)
	return nil
}

func TestImportUsageStatisticsPersistsSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stats := usage.NewRequestStatistics()
	recorder := &usagePersistenceRecorder{}
	prevStore := usage.SetPersistenceBackend(nil)
	t.Cleanup(func() {
		usage.SetPersistenceBackend(prevStore)
	})
	if err := usage.ConfigurePersistence(context.Background(), stats, recorder); err != nil {
		t.Fatalf("ConfigurePersistence error: %v", err)
	}

	payload := map[string]any{
		"version": 1,
		"usage": usage.StatisticsSnapshot{
			APIs: map[string]usage.APISnapshot{
				"persisted-key": {
					Models: map[string]usage.ModelSnapshot{
						"gpt-5.5": {
							Details: []usage.RequestDetail{{
								Timestamp: time.Date(2026, 5, 8, 13, 0, 0, 0, time.UTC),
								LatencyMs: 900,
								Source:    "codex@example.com",
								AuthIndex: "auth-2",
								Tokens: usage.TokenStats{
									InputTokens:  12,
									OutputTokens: 8,
									TotalTokens:  20,
								},
							}},
						},
					},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	h := &Handler{usageStats: stats}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v0/management/usage/import", bytes.NewReader(body))

	h.ImportUsageStatistics(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	if len(recorder.saved) != 1 {
		t.Fatalf("saved records = %d, want 1", len(recorder.saved))
	}
	if recorder.saved[0].APIName != "persisted-key" {
		t.Fatalf("api name = %q, want persisted-key", recorder.saved[0].APIName)
	}
}
