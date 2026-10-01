package configure

import "github.com/v2rayA/v2rayA/db"

type RoutingASource struct {
	URL           string `json:"url"`
	DirectUpdate  bool   `json:"directUpdate"`
	IntervalHours int    `json:"intervalHours"`
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
