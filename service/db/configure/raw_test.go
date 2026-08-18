package configure

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

func TestSubscriptionRawDatabaseIDIsInternal(t *testing.T) {
	raw, err := Bytes2SubscriptionRaw([]byte(`{
		"_databaseId": 42,
		"address": "https://example.invalid/subscription",
		"status": "",
		"servers": [],
		"info": "",
		"autoSelect": false,
		"autoUpdateMode": "auto_update_at_intervals",
		"autoUpdateIntervalHour": 12
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if raw.DatabaseID != 42 {
		t.Fatalf("database ID = %d, want 42", raw.DatabaseID)
	}

	encoded, err := jsoniter.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if gjson.GetBytes(encoded, "_databaseId").Exists() {
		t.Fatalf("internal database ID leaked into exported JSON: %s", encoded)
	}
}
