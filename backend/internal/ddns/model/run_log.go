package model

import (
	"fmt"

	"nas-partner/backend/internal/database"
	"time"
)

type DDNSRunLog struct {
	ID        int64     `json:"id"`
	ConfigID  int64     `json:"config_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	IPv4Addr  string    `json:"ipv4_addr"`
	IPv6Addr  string    `json:"ipv6_addr"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateRunLog(configID int64, status, message, ipv4Addr, ipv6Addr string) (*DDNSRunLog, error) {
	result, err := database.DB.Exec(
		`INSERT INTO ddns_run_logs (config_id, status, message, ipv4_addr, ipv6_addr)
		 VALUES (?, ?, ?, ?, ?)`,
		configID, status, message, ipv4Addr, ipv6Addr,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &DDNSRunLog{
		ID:       id,
		ConfigID: configID,
		Status:   status,
		Message:  message,
		IPv4Addr: ipv4Addr,
		IPv6Addr: ipv6Addr,
	}, nil
}

func ListRunLogsByConfigID(configID int64, limit int) ([]*DDNSRunLog, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := database.DB.Query(
		`SELECT id, config_id, status, message, ipv4_addr, ipv6_addr, created_at
		 FROM ddns_run_logs WHERE config_id = ? ORDER BY id DESC LIMIT ?`,
		configID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*DDNSRunLog
	for rows.Next() {
		l := &DDNSRunLog{}
		if err := rows.Scan(&l.ID, &l.ConfigID, &l.Status, &l.Message, &l.IPv4Addr, &l.IPv6Addr, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func GetLatestRunLogByConfigID(configID int64) (*DDNSRunLog, error) {
	l := &DDNSRunLog{}
	err := database.DB.QueryRow(
		`SELECT id, config_id, status, message, ipv4_addr, ipv6_addr, created_at
		 FROM ddns_run_logs WHERE config_id = ? ORDER BY id DESC LIMIT 1`,
		configID,
	).Scan(&l.ID, &l.ConfigID, &l.Status, &l.Message, &l.IPv4Addr, &l.IPv6Addr, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func DeleteAllByConfigID(configID int64) error {
	_, err := database.DB.Exec(`DELETE FROM ddns_run_logs WHERE config_id = ?`, configID)
	return err
}

func DeleteOlderThan(days int) (int64, error) {
	result, err := database.DB.Exec(
		`DELETE FROM ddns_run_logs WHERE created_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
