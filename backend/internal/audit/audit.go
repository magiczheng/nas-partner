package audit

import "nas-partner/backend/internal/database"

func Log(username, action, detail, ip string) {
	database.DB.Exec(
		`INSERT INTO audit_logs (username, action, detail, ip) VALUES (?, ?, ?, ?)`,
		username, action, detail, ip,
	)
}
