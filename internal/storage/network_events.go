package storage

import "time"

// NetworkEvent represents a detected network change.
type NetworkEvent struct {
	ID            int64     `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	Reason        string    `json:"reason"`
	PrevSSID      string    `json:"prev_ssid"`
	PrevType      string    `json:"prev_type"`
	PrevInterface string    `json:"prev_interface"`
	PrevGateway   string    `json:"prev_gateway"`
	CurrSSID      string    `json:"curr_ssid"`
	CurrType      string    `json:"curr_type"`
	CurrInterface string    `json:"curr_interface"`
	CurrGateway   string    `json:"curr_gateway"`
}

// InsertNetworkEvent stores a network change event.
func (db *DB) InsertNetworkEvent(e *NetworkEvent) error {
	_, err := db.conn.Exec(`
		INSERT INTO network_events (timestamp, reason, prev_ssid, prev_type, prev_interface, prev_gateway, curr_ssid, curr_type, curr_interface, curr_gateway)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Timestamp, e.Reason, e.PrevSSID, e.PrevType, e.PrevInterface, e.PrevGateway,
		e.CurrSSID, e.CurrType, e.CurrInterface, e.CurrGateway,
	)
	return err
}

// GetNetworkEvents returns recent network change events.
func (db *DB) GetNetworkEvents(since time.Time) ([]NetworkEvent, error) {
	rows, err := db.conn.Query(`
		SELECT id, timestamp, reason, prev_ssid, prev_type, prev_interface, prev_gateway, curr_ssid, curr_type, curr_interface, curr_gateway
		FROM network_events
		WHERE timestamp >= ?
		ORDER BY timestamp DESC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []NetworkEvent
	for rows.Next() {
		var e NetworkEvent
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.Reason, &e.PrevSSID, &e.PrevType, &e.PrevInterface, &e.PrevGateway, &e.CurrSSID, &e.CurrType, &e.CurrInterface, &e.CurrGateway); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
