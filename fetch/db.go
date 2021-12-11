package fetch

import (
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
	"sync"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

const dsLoggerId = "datastore"

var db *Database
var lock = new(sync.Mutex)

func GetDb(host string, userName string, password string, dbName string, forTest bool) *Database {
	if forTest {
		panic("datastore for testing is not yet implemented")
	}
	lock.Lock()
	defer lock.Unlock()
	if db == nil {
		connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", userName, password, host, dbName)
		handle, err := sql.Open("postgres", connStr)
		if err != nil {
			panic(err)
		}
		db = &Database{
			DB: handle,
		}
	}
	return db
}

func (ds *Database) GetUser(userName string) (*User, error) {
	rows, err := ds.DB.Query(fmt.Sprintf("SELECT * FROM users where name LIKE '%s'", userName))
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)
	hasNext := rows.Next()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if !hasNext {
		return nil, nil
	}
	var id string
	var name string
	var profilePictureUrl string
	err = rows.Scan(&id, &name, &profilePictureUrl)
	return &User{Id: id, Name: name, ProfilePictureUrl: profilePictureUrl}, nil
}

func (ds *Database) UpdateSinceId(userId TwitterUserId, kind string, since TweetId) error {
	query := fmt.Sprintf("INSERT INTO checkpoint VALUES ('%s', '%s', '%s') ON CONFLICT(user_id, type) DO UPDATE SET watermark = '%s' ", userId, kind, since, since)
	_, err := ds.DB.Exec(query)
	return err
}

func (ds *Database) GetSinceId(userId TwitterUserId) (TweetId, error) {
	rows, err := ds.DB.Query(fmt.Sprintf("SELECT watermark FROM checkpoint WHERE user_id='%s' AND type='twitter'", userId))
	if err != nil {
		return "", err
	}
	defer closeRows(rows)
	hasNext := rows.Next()
	if err = rows.Err(); err != nil {
		return "", err
	}
	if !hasNext {
		return "", nil
	}
	var waterMark string
	err = rows.Scan(&waterMark)
	return waterMark, err
}

func closeRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		log.Warn().Str(dsLoggerId, dsLoggerId).Err(err).Msg("failed not close result set")
	}
}

func (ds *Database) Close() error {
	return ds.DB.Close()
}

type User struct {
	Id                string
	Name              string
	ProfilePictureUrl string
}
