package audit

import (
	"time"

	"nas-partner/backend/internal/database"
)

type Entry struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

func Log(username, action, detail, ip string) {
	database.DB.Exec(
		`INSERT INTO audit_logs (username, action, detail, ip) VALUES (?, ?, ?, ?)`,
		username, action, detail, ip,
	)
}

func List(limit int) ([]Entry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := database.DB.Query(
		`SELECT id, username, action, detail, ip, created_at
		 FROM audit_logs ORDER BY id DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Username, &e.Action, &e.Detail, &e.IP, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	if list == nil {
		list = []Entry{}
	}
	return list, nil
}
