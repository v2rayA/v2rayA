package configure

import (
	"database/sql"
	"github.com/v2rayA/v2rayA/db"
)

type RoutingASource struct {
	SubscriptionID int64  `json:"subscriptionId,omitempty"`
	URL            string `json:"url"`
	DirectUpdate   bool   `json:"directUpdate"`
	IntervalHours  int    `json:"intervalHours"`
}

func GetRoutingASource() (source RoutingASource) {
	_ = db.Get("system", "routingASource", &source)
	if source.IntervalHours <= 0 {
		source.IntervalHours = 24
	}
	return
}

func SetRoutingASource(source RoutingASource) error {
	return db.Set("system", "routingASource", source)
}

// Store the source, rules and (when requested) split mode as one change.
func SetRoutingAConfig(text string, source RoutingASource, setting *Setting) error {
	return db.ReadModifyWrite(func(tx *sql.Tx) error {
		if err := db.SetTx(tx, "system", "routingA", text); err != nil {
			return err
		}
		if err := db.SetTx(tx, "system", "routingASource", source); err != nil {
			return err
		}
		if setting != nil {
			return db.SetTx(tx, "system", "setting", setting)
		}
		return nil
	})
}
