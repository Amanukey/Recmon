package storage

import (
	"context"
	"database/sql"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) migrate() error {
	_, err := s.db.Exec(`PRAGMA journal_mode=WAL; CREATE TABLE IF NOT EXISTS projects (id TEXT PRIMARY KEY, name TEXT NOT NULL, created_at TEXT NOT NULL); CREATE TABLE IF NOT EXISTS assets (id INTEGER PRIMARY KEY AUTOINCREMENT, project_id TEXT NOT NULL, hostname TEXT NOT NULL, validation_status TEXT NOT NULL, first_seen TEXT NOT NULL, last_seen TEXT NOT NULL, UNIQUE(project_id, hostname)); CREATE TABLE IF NOT EXISTS dns_records (id INTEGER PRIMARY KEY AUTOINCREMENT, asset_id INTEGER NOT NULL, record_type TEXT NOT NULL, value TEXT NOT NULL, response_status TEXT NOT NULL, observed_at TEXT NOT NULL); CREATE TABLE IF NOT EXISTS http_services (id INTEGER PRIMARY KEY AUTOINCREMENT, asset_id INTEGER NOT NULL, scheme TEXT NOT NULL, port INTEGER NOT NULL, status_code INTEGER NOT NULL, title TEXT, response_bytes INTEGER NOT NULL, observed_at TEXT NOT NULL);`)
	return err
}
func (s *Store) DB() *sql.DB { return s.db }
func (s *Store) UpsertAsset(ctx context.Context, projectID, host, status, now string) (int64, error) {
	_, err := s.db.ExecContext(ctx, `INSERT INTO assets(project_id,hostname,validation_status,first_seen,last_seen) VALUES(?,?,?,?,?) ON CONFLICT(project_id,hostname) DO UPDATE SET validation_status=excluded.validation_status,last_seen=excluded.last_seen`, projectID, host, status, now, now)
	if err != nil {
		return 0, err
	}
	var id int64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM assets WHERE project_id=? AND hostname=?`, projectID, host).Scan(&id)
	return id, err
}
func (s *Store) AddDNS(ctx context.Context, assetID int64, recordType, value, status, now string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO dns_records(asset_id,record_type,value,response_status,observed_at) VALUES(?,?,?,?,?)`, assetID, recordType, value, status, now)
	return err
}
func (s *Store) AddHTTP(ctx context.Context, assetID int64, scheme string, port, code, bytes int, now string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO http_services(asset_id,scheme,port,status_code,response_bytes,observed_at) VALUES(?,?,?,?,?,?)`, assetID, scheme, port, code, bytes, now)
	return err
}
