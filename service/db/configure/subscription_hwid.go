package configure

import (
	"database/sql"

	"github.com/v2rayA/v2rayA/db"
)

func SubscriptionHWIDGrants() (map[int64]string, error) {
	rows, err := db.GetDB().Query("SELECT id, address FROM subscriptions WHERE send_hwid")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	grants := map[int64]string{}
	for rows.Next() {
		var id int64
		var address string
		if err := rows.Scan(&id, &address); err != nil {
			return nil, err
		}
		grants[id] = address
	}
	return grants, rows.Err()
}

// A nil list leaves individual grants alone. An empty list revokes them all.
func SetSettingWithHWID(s *Setting, grants map[int64]string) error {
	if grants == nil {
		return SetSetting(s)
	}
	return db.ReadModifyWrite(func(tx *sql.Tx) error {
		if err := db.SetTx(tx, "system", "setting", s); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE subscriptions SET send_hwid = 0 WHERE send_hwid"); err != nil {
			return err
		}
		for id, address := range grants {
			if _, err := tx.Exec("UPDATE subscriptions SET send_hwid = 1 WHERE id = ? AND address = ?", id, address); err != nil {
				return err
			}
		}
		return nil
	})
}
