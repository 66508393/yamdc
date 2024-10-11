package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type sqliteStore struct {
	db *sql.DB
}

func (s *sqliteStore) init() error {
	createTable := `CREATE TABLE IF NOT EXISTS cache_tab (
        key TEXT PRIMARY KEY,
        value BLOB,
        expire_at INTEGER
    );`
	_, err := s.db.Exec(createTable)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) GetData(ctx context.Context, key string) ([]byte, error) {
	var val []byte
	now := time.Now().Unix()
	err := s.db.QueryRow("SELECT value FROM cache_tab WHERE key = ? and (expire_at > ? or expire_at = 0)", key, now).Scan(&val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (s *sqliteStore) PutData(ctx context.Context, key string, value []byte, expire time.Duration) error {
	var expireAt int64 = 0
	if expire > 0 {
		expireAt = time.Now().Add(expire).Unix()
	}
	_, err := s.db.Exec("INSERT OR REPLACE INTO cache_tab (key, value, expire_at) VALUES (?, ?, ?)", key, value, expireAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqliteStore) IsDataExist(ctx context.Context, key string) (bool, error) {
	var cnt int64
	now := time.Now().Unix()
	err := s.db.QueryRow("SELECT count(*) FROM cache_tab WHERE key = ? and (expire_at > ? or expire_at = 0)", key, now).Scan(&cnt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if cnt == 0 {
		return false, nil
	}
	return true, nil
}

func NewSqliteStorage(path string) (IStorage, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	s := &sqliteStore{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}
	return s, nil
}

func MustNewSqliteStorage(path string) IStorage {
	s, err := NewSqliteStorage(path)
	if err != nil {
		panic(err)
	}
	return s
}
