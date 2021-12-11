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

func (ds *Database) GetAllUsers() ([]*User, error) {
	rows, err := ds.DB.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)
	var users []*User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
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
	return scanUser(rows)
}

func (ds *Database) UpdateSinceId(userId TwitterUserId, kind WaterMarkType, since TweetId) error {
	query := fmt.Sprintf("INSERT INTO checkpoint VALUES ('%s', '%s', '%s') ON CONFLICT(user_id, type) DO UPDATE SET watermark = '%s' ", userId, kind, since, since)
	_, err := ds.DB.Exec(query)
	return err
}

func (ds *Database) GetSinceId(userId TwitterUserId) (TweetId, error) {
	rows, err := ds.DB.Query(fmt.Sprintf("SELECT watermark FROM checkpoint WHERE user_id='%s' AND type='%s'", userId, tweetWaterMark))
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

func scanUser(rows *sql.Rows) (*User, error) {
	var id string
	var name string
	var profilePictureUrl string
	err := rows.Scan(&id, &name, &profilePictureUrl)
	return &User{Id: id, Name: name, ProfilePictureUrl: profilePictureUrl}, err
}

func closeRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		log.Warn().Str(dsLoggerId, dsLoggerId).Err(err).Msg("failed to close result set")
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
