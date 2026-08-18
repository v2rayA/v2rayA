package db

import (
	"encoding/json"
	"testing"
)

func TestMarshalSubscriptionListValueEscapesTextAndIncludesDatabaseID(t *testing.T) {
	encoded, err := marshalSubscriptionListValue(
		42,
		"quoted \"remark\"\nnext line",
		"https://example.invalid/subscription?name=\"test\"",
		"status\nline",
		"info\\value",
		nil,
		true,
		"auto_update_at_intervals",
		12,
		"2026-08-11T12:00:00Z",
	)
	if err != nil {
		t.Fatal(err)
	}

	var decoded subscriptionListValue
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if decoded.DatabaseID != 42 || decoded.Remarks != "quoted \"remark\"\nnext line" {
		t.Fatalf("decoded identity/text = (%d, %q)", decoded.DatabaseID, decoded.Remarks)
	}
	if decoded.AutoUpdateMode != "auto_update_at_intervals" || decoded.AutoUpdateIntervalHour != 12 {
		t.Fatalf("decoded policy = (%q, %d)", decoded.AutoUpdateMode, decoded.AutoUpdateIntervalHour)
	}
}

func TestListGetSubscriptionUsesDatabaseIDAfterDeletion(t *testing.T) {
	database := openSchemaTestDB(t)
	if err := InitSchema(database); err != nil {
		t.Fatal(err)
	}

	firstResult, err := database.Exec(
		`INSERT INTO subscriptions (address, remarks, sort) VALUES (?, ?, ?)`,
		"https://example.invalid/first", "first", 0,
	)
	if err != nil {
		t.Fatal(err)
	}
	firstID, err := firstResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	secondResult, err := database.Exec(
		`INSERT INTO subscriptions (address, remarks, sort) VALUES (?, ?, ?)`,
		"https://example.invalid/second", "second", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := secondResult.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	for _, server := range []struct {
		subID int64
		value string
	}{
		{firstID, `{"tag":"first-server"}`},
		{secondID, `{"tag":"second-server"}`},
	} {
		if _, err := database.Exec(
			`INSERT INTO servers (type, sub_id, config_json, sort) VALUES ('subscription_server', ?, ?, 0)`,
			server.subID, server.value,
		); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := database.Exec(`DELETE FROM servers WHERE sub_id = ?`, firstID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`DELETE FROM subscriptions WHERE id = ?`, firstID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE subscriptions SET sort = 0 WHERE id = ?`, secondID); err != nil {
		t.Fatal(err)
	}

	encoded, err := listGetSubscription(database, 0)
	if err != nil {
		t.Fatal(err)
	}
	var decoded subscriptionListValue
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("invalid subscription JSON: %v", err)
	}
	if decoded.DatabaseID != secondID {
		t.Fatalf("database ID = %d, want %d", decoded.DatabaseID, secondID)
	}
	if len(decoded.Servers) != 1 || string(decoded.Servers[0]) != `{"tag":"second-server"}` {
		t.Fatalf("servers = %s, want second subscription server", encoded)
	}
}
