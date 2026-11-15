package db

import (
	"database/sql"
	"strings"
)

// SnapshotCatalog preserves database IDs and ordering when a deletion must be
// rolled back after a failed core reload. Callers must hold the mutation lock.
func SnapshotCatalog() (func() error, error) {
	type table struct {
		name, where, insert string
		values              [][]interface{}
	}
	tables := []table{
		{name: "subscriptions", where: "1"},
		{name: "servers", where: "1"},
		{name: "system_config", where: "key LIKE 'outbound.%:connectedServers'"},
	}
	err := ReadModifyWrite(func(tx *sql.Tx) error {
		for i := range tables {
			t := &tables[i]
			rows, err := tx.Query("SELECT * FROM " + t.name + " WHERE " + t.where)
			if err != nil {
				return err
			}
			defer rows.Close()
			columns, err := rows.Columns()
			if err != nil {
				return err
			}
			t.insert = "INSERT INTO " + t.name + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",") + ")"
			for rows.Next() {
				values := make([]interface{}, len(columns))
				pointers := make([]interface{}, len(columns))
				for j := range values {
					pointers[j] = &values[j]
				}
				if err := rows.Scan(pointers...); err != nil {
					return err
				}
				t.values = append(t.values, values)
			}
			if err := rows.Err(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return func() error {
		return ReadModifyWrite(func(tx *sql.Tx) error {
			for i := len(tables) - 1; i >= 0; i-- {
				t := tables[i]
				if _, err := tx.Exec("DELETE FROM " + t.name + " WHERE " + t.where); err != nil {
					return err
				}
			}
			for _, t := range tables {
				for _, values := range t.values {
					if _, err := tx.Exec(t.insert, values...); err != nil {
						return err
					}
				}
			}
			return nil
		})
	}, nil
}
