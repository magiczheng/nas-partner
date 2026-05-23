package model

import (
	"encoding/json"
	"time"

	"nas-partner/backend/internal/database"
)

type DDNSConfig struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Enabled          bool      `json:"enabled"`
	DNSProvider      string    `json:"dns_provider"`
	AccessKeyID      string    `json:"access_key_id"`
	AccessKeySecret  string    `json:"access_key_secret"`
	ExtraParams      string    `json:"extra_params"`
	IPv4Enabled      bool      `json:"ipv4_enabled"`
	IPv4GetType      string    `json:"ipv4_get_type"`
	IPv4URL          string    `json:"ipv4_url"`
	IPv4NetInterface string    `json:"ipv4_net_interface"`
	IPv4Cmd          string    `json:"ipv4_cmd"`
	IPv6Enabled      bool      `json:"ipv6_enabled"`
	IPv6GetType      string    `json:"ipv6_get_type"`
	IPv6URL          string    `json:"ipv6_url"`
	IPv6NetInterface string    `json:"ipv6_net_interface"`
	IPv6Cmd          string    `json:"ipv6_cmd"`
	IPv4Addr         string    `json:"ipv4_addr"`
	IPv6Addr         string    `json:"ipv6_addr"`
	CurrentIPv4Addr  string    `json:"current_ipv4"`
	CurrentIPv6Addr  string    `json:"current_ipv6"`
	Domains          []string  `json:"domains"`
	TTL              string    `json:"ttl"`
	Interval         int       `json:"interval"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func List() ([]*DDNSConfig, error) {
	rows, err := database.DB.Query(
		`SELECT id, name, enabled, dns_provider, access_key_id, access_key_secret, extra_params,
			ipv4_enabled, ipv4_get_type, ipv4_url, ipv4_net_interface, ipv4_cmd,
			ipv6_enabled, ipv6_get_type, ipv6_url, ipv6_net_interface, ipv6_cmd,
			ipv4_addr, ipv6_addr, current_ipv4, current_ipv6,
			domains, ttl, interval, created_at, updated_at
		FROM ddns_configs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*DDNSConfig
	for rows.Next() {
		c := &DDNSConfig{}
		var domainsJSON string
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Enabled, &c.DNSProvider,
			&c.AccessKeyID, &c.AccessKeySecret, &c.ExtraParams,
			&c.IPv4Enabled, &c.IPv4GetType, &c.IPv4URL, &c.IPv4NetInterface, &c.IPv4Cmd,
			&c.IPv6Enabled, &c.IPv6GetType, &c.IPv6URL, &c.IPv6NetInterface, &c.IPv6Cmd,
			&c.IPv4Addr, &c.IPv6Addr, &c.CurrentIPv4Addr, &c.CurrentIPv6Addr, &domainsJSON, &c.TTL, &c.Interval,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(domainsJSON), &c.Domains)
		list = append(list, c)
	}
	return list, nil
}

func GetByID(id int64) (*DDNSConfig, error) {
	c := &DDNSConfig{}
	var domainsJSON string
	err := database.DB.QueryRow(
		`SELECT id, name, enabled, dns_provider, access_key_id, access_key_secret, extra_params,
			ipv4_enabled, ipv4_get_type, ipv4_url, ipv4_net_interface, ipv4_cmd,
			ipv6_enabled, ipv6_get_type, ipv6_url, ipv6_net_interface, ipv6_cmd,
			ipv4_addr, ipv6_addr, current_ipv4, current_ipv6,
			domains, ttl, interval, created_at, updated_at
		FROM ddns_configs WHERE id = ?`, id,
	).Scan(
		&c.ID, &c.Name, &c.Enabled, &c.DNSProvider,
		&c.AccessKeyID, &c.AccessKeySecret, &c.ExtraParams,
		&c.IPv4Enabled, &c.IPv4GetType, &c.IPv4URL, &c.IPv4NetInterface, &c.IPv4Cmd,
		&c.IPv6Enabled, &c.IPv6GetType, &c.IPv6URL, &c.IPv6NetInterface, &c.IPv6Cmd,
		&c.IPv4Addr, &c.IPv6Addr, &c.CurrentIPv4Addr, &c.CurrentIPv6Addr, &domainsJSON, &c.TTL, &c.Interval,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(domainsJSON), &c.Domains)
	return c, nil
}

func (c *DDNSConfig) Create() error {
	domainsJSON, _ := json.Marshal(c.Domains)
	if c.TTL == "" {
		c.TTL = "600"
	}
	_, err := database.DB.Exec(
		`INSERT INTO ddns_configs
		(name, enabled, dns_provider, access_key_id, access_key_secret, extra_params,
		 ipv4_enabled, ipv4_get_type, ipv4_url, ipv4_net_interface, ipv4_cmd,
		 ipv6_enabled, ipv6_get_type, ipv6_url, ipv6_net_interface, ipv6_cmd,
		 ipv4_addr, ipv6_addr, current_ipv4, current_ipv6, domains, ttl, interval)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Name, c.Enabled, c.DNSProvider, c.AccessKeyID, c.AccessKeySecret, c.ExtraParams,
		c.IPv4Enabled, c.IPv4GetType, c.IPv4URL, c.IPv4NetInterface, c.IPv4Cmd,
		c.IPv6Enabled, c.IPv6GetType, c.IPv6URL, c.IPv6NetInterface, c.IPv6Cmd,
		c.IPv4Addr, c.IPv6Addr, c.CurrentIPv4Addr, c.CurrentIPv6Addr, string(domainsJSON), c.TTL, c.Interval,
	)
	if err != nil {
		return err
	}
	return database.DB.QueryRow("SELECT last_insert_rowid()").Scan(&c.ID)
}

func (c *DDNSConfig) Update() error {
	domainsJSON, _ := json.Marshal(c.Domains)
	_, err := database.DB.Exec(
		`UPDATE ddns_configs SET
			name=?, enabled=?, dns_provider=?, access_key_id=?, access_key_secret=?, extra_params=?,
			ipv4_enabled=?, ipv4_get_type=?, ipv4_url=?, ipv4_net_interface=?, ipv4_cmd=?,
			ipv6_enabled=?, ipv6_get_type=?, ipv6_url=?, ipv6_net_interface=?, ipv6_cmd=?,
			ipv4_addr=?, ipv6_addr=?, current_ipv4=?, current_ipv6=?, domains=?, ttl=?, interval=?,
			updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		c.Name, c.Enabled, c.DNSProvider, c.AccessKeyID, c.AccessKeySecret, c.ExtraParams,
		c.IPv4Enabled, c.IPv4GetType, c.IPv4URL, c.IPv4NetInterface, c.IPv4Cmd,
		c.IPv6Enabled, c.IPv6GetType, c.IPv6URL, c.IPv6NetInterface, c.IPv6Cmd,
		c.IPv4Addr, c.IPv6Addr, c.CurrentIPv4Addr, c.CurrentIPv6Addr, string(domainsJSON), c.TTL, c.Interval,
		c.ID,
	)
	return err
}

func (c *DDNSConfig) UpdateCurrentIPs() error {
	_, err := database.DB.Exec(
		`UPDATE ddns_configs SET current_ipv4=?, current_ipv6=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		c.CurrentIPv4Addr, c.CurrentIPv6Addr, c.ID,
	)
	return err
}

func Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM ddns_configs WHERE id=?", id)
	return err
}

func Toggle(id int64) (*DDNSConfig, error) {
	c, err := GetByID(id)
	if err != nil {
		return nil, err
	}
	c.Enabled = !c.Enabled
	_, err = database.DB.Exec("UPDATE ddns_configs SET enabled=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", c.Enabled, id)
	if err != nil {
		return nil, err
	}
	return c, nil
}
