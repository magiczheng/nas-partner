package database

import (
		"database/sql"
		"log"
		"os"
		"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// SQLite 不支持并发写入，限制为单连接
	DB.SetMaxOpenConns(1)

	migrate()
	log.Println("database initialized")
}

func migrate() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ddns_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			enabled INTEGER DEFAULT 1,
			dns_provider TEXT NOT NULL,
			access_key_id TEXT DEFAULT '',
			access_key_secret TEXT DEFAULT '',
			extra_params TEXT DEFAULT '',
			ipv4_enabled INTEGER DEFAULT 1,
			ipv4_get_type TEXT DEFAULT 'auto',
			ipv4_url TEXT DEFAULT '',
			ipv4_net_interface TEXT DEFAULT '',
			ipv4_cmd TEXT DEFAULT '',
			ipv4_addr TEXT DEFAULT '',
			ipv6_enabled INTEGER DEFAULT 0,
			ipv6_get_type TEXT DEFAULT 'auto',
			ipv6_url TEXT DEFAULT '',
			ipv6_net_interface TEXT DEFAULT '',
			ipv6_cmd TEXT DEFAULT '',
			ipv6_addr TEXT DEFAULT '',
			domains TEXT NOT NULL DEFAULT '[]',
			ttl TEXT DEFAULT '600',
			interval INTEGER DEFAULT 300,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ddns_run_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT '',
			message TEXT NOT NULL DEFAULT '',
			ipv4_addr TEXT NOT NULL DEFAULT '',
			ipv6_addr TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
	}

	// 追加列（幂等，列已有时忽略错误）
	migrations := []string{
		"ALTER TABLE ddns_configs ADD COLUMN ipv4_addr TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE ddns_configs ADD COLUMN ipv6_addr TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE ddns_configs ADD COLUMN current_ipv4 TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE ddns_configs ADD COLUMN current_ipv6 TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE ddns_configs DROP COLUMN last_run_at",
		"ALTER TABLE ddns_configs DROP COLUMN last_status",
		}
	for _, m := range migrations {
		DB.Exec(m)
	}
}
