
package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Command struct {
	ID			int
	Command 	string
	Frequency	int
	LastUsed	time.Time
	Tokens		string
}

type DB struct {
	conn *sql.DB
}

func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DB) migrate() error {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS commands (
			id			INTEGER PRIMARY KEY AUTOINCREMENT,
			command		TEXT UNIQUE NOT NULL,
			frequency	INTEGER DEFAULT 1,
			last_used	DATETIME NOT NULL,
			tokens		TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_last_used ON commands(last_used);
	`)

	return err
}


func (d *DB) Upsert (cmd, tokens string) error {
	_, err := d.conn.Exec(`
		INSERT INTO commands (command, frequency, last_used, tokens)
		values (?, 1, ?, ?)
		ON CONFLICT(command) DO UPDATE SET
			frequency = frequency + 1,
			last_used = excluded.last_used,
			tokens	  = excluded.tokens
	`, cmd, time.Now(), tokens)
	
	return err
}

func (d *DB) All () ([]Command, error) {
	rows, err := d.conn.Query(`
		SELECT id, command, frequency, last_used, tokens FROM commands
		ORDER BY last_used DESC
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var cmds []Command

	for rows.Next() {
		var c Command
		if err := rows.Scan(&c.ID, &c.Command, &c.Frequency, &c.LastUsed, &c.Tokens); err != nil {
			return nil, err
		}
		cmds = append(cmds, c)
	}
	return cmds, rows.Err()
}

func (d *DB) Close() error {
	return d.conn.Close()
}
