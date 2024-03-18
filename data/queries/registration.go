package queries

import (
	"database/sql"
	"fmt"
)

const (
	GetHostsQuery   = `SELECT hostname,key_id,proto FROM node_registration`
	GetHostQuery    = `SELECT hostname,key_id,proto FROM node_registration WHERE hostname=?`
	UpdateHostQuery = `UPDATE node_registration SET proto=? WHERE hostname=?`
	InsertHostQuery = `REPLACE INTO node_registration VALUES(?,?,?);`

	GetServerQuery    = `SELECT hostname, key FROM server_registration WHERE key=?`
	InsertServerQuery = `REPLACE INTO server_registration VALUES(?,?)`
)

type RegisteredHostsRow struct {
	Hostname string
	Key      string
	Data     []byte
}

func GetRegisteredHosts(db *sql.DB) chan *RegisteredHostsRow {
	rchan := make(chan *RegisteredHostsRow, 10)

	go func(db *sql.DB, rchan chan *RegisteredHostsRow) {
		defer close(rchan)
		rows, err := db.Query(GetHostsQuery)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to execute query: %w\n", err))
			return
		}
		defer rows.Close()
		for rows.Next() {
			r := RegisteredHostsRow{}
			err := rows.Scan(&r.Hostname, &r.Key, &r.Data )
			if err != nil {
				fmt.Println(fmt.Errorf("error loading row: %w\n", err))
			}
			rchan <- &r
		}

	}(db, rchan)

	return rchan
}

func GetRegisteredHost(db *sql.DB, hostname string) (*RegisteredHostsRow, error) {
	row := db.QueryRow(GetHostQuery, hostname)
	r := RegisteredHostsRow{}
	err := row.Scan(&r.Hostname, &r.Key, &r.Data, &r.Key)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func UpdateRegisteredHost(db *sql.DB, row *RegisteredHostsRow) error {
	_, err := db.Exec(UpdateHostQuery, row.Data, row.Hostname)
	return err
}

func InsertHostRegistration(db *sql.DB, row *RegisteredHostsRow) error {
	_, err := db.Exec(InsertHostQuery, row.Hostname, row.Key, row.Data)
	return err
}

type RegisteredServerRow struct {
	Hostname string
	Key      string
}

func GetRegisteredCoordinationServer(db *sql.DB, key string) (*RegisteredServerRow, error) {
	r := RegisteredServerRow{}
	row := db.QueryRow(GetServerQuery, key)
	err := row.Scan(r.Hostname, r.Key)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func SetRegisteredCoordinationServer(db *sql.DB, hostname, key string) error {
	_, err := db.Exec(InsertServerQuery, hostname, key)
	return err
}
