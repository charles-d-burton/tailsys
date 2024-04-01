package queries

import (
	"database/sql"
	"fmt"
	"regexp"
)

const (
	GetHostsQuery   = `SELECT hostname,key_id,proto FROM node_registration`
	GetHostQuery    = `SELECT hostname,key_id,proto FROM node_registration WHERE hostname=?`
	UpdateHostQuery = `UPDATE node_registration SET proto=? WHERE hostname=?`
	InsertHostQuery = `REPLACE INTO node_registration VALUES(?,?,?);`

	GetServerQuery    = `SELECT hostname, key FROM server_registration WHERE key=?`
	InsertServerQuery = `REPLACE INTO server_registration VALUES(?,?)`
)

type RegisteredHostsData struct {
	Hostname string
	Key      string
	Data     []byte
}

func GetMatchRegisteredHosts(db *sql.DB, pattern string) (chan *RegisteredHostsData, error) {
	rchan := make(chan *RegisteredHostsData, 1000)

	re, err := regexp.CompilePOSIX(pattern)
	if err != nil {
		return nil, err
	}

	go func(db *sql.DB, re *regexp.Regexp, rchan chan *RegisteredHostsData) {
		defer close(rchan)
		rows, err := db.Query(GetHostsQuery)
		if err != nil {
			fmt.Println(fmt.Errorf("problem getting hosts %w", err))
		}
		defer rows.Close()
		for rows.Next() {
			r := RegisteredHostsData{}
			err := rows.Scan(&r.Hostname, &r.Key, &r.Data)
			if err != nil {
				fmt.Println(fmt.Errorf("problem getting hosts %w", err))
			}
      fmt.Println("found a row: ", r.Hostname)
			if re.MatchString(r.Hostname) {
				rchan <- &r
			}
		}
	}(db, re, rchan)
	return rchan, nil
}

func GetRegisteredHosts(db *sql.DB) chan *RegisteredHostsData {
	rchan := make(chan *RegisteredHostsData, 10)

	go func(db *sql.DB, rchan chan *RegisteredHostsData) {
		defer close(rchan)
		rows, err := db.Query(GetHostsQuery)
		defer rows.Close()
		if err != nil {
			fmt.Println(fmt.Errorf("unable to execute query: %w\n", err))
			return
		}
		for rows.Next() {
			r := RegisteredHostsData{}
			err := rows.Scan(&r.Hostname, &r.Key, &r.Data)
			if err != nil {
				fmt.Println(fmt.Errorf("error loading row: %w\n", err))
			}
			rchan <- &r
		}

	}(db, rchan)

	return rchan
}

func GetRegisteredHost(db *sql.DB, hostname string) (*RegisteredHostsData, error) {
	row := db.QueryRow(GetHostQuery, hostname)
	r := RegisteredHostsData{}
	err := row.Scan(&r.Hostname, &r.Key, &r.Data, &r.Key)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func UpdateRegisteredHost(db *sql.DB, row *RegisteredHostsData) error {
	_, err := db.Exec(UpdateHostQuery, row.Data, row.Hostname)
	return err
}

func InsertHostRegistration(db *sql.DB, row *RegisteredHostsData) error {
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
