package fetch

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"time"
)

const fetcherLoggerId = "fetcher"
const tweetFetchSize = 100 // maximum allowed
const tweetWaterMark WaterMarkType = "twitter"

type WaterMarkType = string

type Fetcher struct {
	TwitterClient HttpTwitterClient
	Database      *Database
}

func (f *Fetcher) AddUser(userName string) error {
	user, err := f.Database.GetUser(userName)
	if err != nil {
		return err
	}
	if user != nil {
		// user already exists
		log.Info().Str(constants.LoggerId, fetcherLoggerId).Msgf("user %s already exists", userName)
		return nil
	}
	response, err := f.TwitterClient.FindUser(userName)
	if err != nil {
		return err
	}
	data := response.Data
	_, err = f.Database.DB.Exec(fmt.Sprintf("INSERT INTO users values ('%s', '%s', '%s')", data.Id, data.UserName, data.ProfileImageUrl))
	return err
}

func (f *Fetcher) GetAllUserTweets() error {
	users, err := f.Database.GetAllUsers()
	if err != nil {
		log.Error().Str(constants.LoggerId, fetcherLoggerId).Err(err).Msg("failed to get all users from DB")
		return errors.New("could not get users from DB")
	} else if users == nil || len(users) == 0 {
		log.Info().Str(constants.LoggerId, fetcherLoggerId).Err(err).Msg("no users in DB")
		return errors.New("no users to get tweets for")
	} else {
		var successCount = 0
		for _, user := range users {
			err = f.GetUserTweets(user.Name)
			if err == nil {
				successCount++
			} else {
				log.Error().Str(constants.LoggerId, fetcherLoggerId).Err(err).Msgf("error in getting tweets for user '%s'", user.Name)
			}
		}
		log.Info().Str(constants.LoggerId, fetcherLoggerId).Err(err).Msgf("success in fetching tweets for '%d' users out of a total of '%d'", successCount, len(users))
		if successCount == len(users) {
			return nil
		} else {
			return fmt.Errorf("failed to get tweets for '%d' users", len(users)-successCount)
		}
	}

}
func (f *Fetcher) GetUserTweets(userName string) error {
	logFailure := func(failure string, err error) {
		failureMsg := fmt.Sprintf("failed in '%s' for userName '%s'", failure, userName)
		log.Error().Str(constants.LoggerId, fetcherLoggerId).Err(err).Msg(failureMsg)
	}
	rollbackOrLogOnError := func(txn *sql.Tx) {
		err2 := txn.Rollback()
		if err2 != nil {
			log.Error().Str(constants.LoggerId, fetcherLoggerId).Err(err2)
		}
	}
	user, err := f.Database.GetUser(userName)
	if err != nil {
		logFailure("getting user", err)
		return err
	}
	if user == nil {
		return fmt.Errorf("not found user '%s'", userName)
	}
	var startTime string
	sinceId, err := f.Database.GetSinceId(user.Id)
	if err != nil {
		return err
	}
	if len(sinceId) == 0 {
		location, _ := time.LoadLocation("UTC")
		// Get tweets for past two days
		startTime = time.Now().In(location).AddDate(0, 0, -2).Format(time.RFC3339)
	}
	tweetsResponse, err := f.TwitterClient.GetTweets(user.Id, tweetFetchSize, sinceId, startTime)
	if err != nil {
		logFailure("getting tweets", err)
		return err
	}

	if len(tweetsResponse.Tweets) > 0 {
		txn, err := f.Database.DB.Begin()
		if err != nil {
			logFailure("saving tweets to datastore", err)
			return err
		}
		stmt, err := txn.Prepare(pq.CopyIn("tweets", "id", "text", "lang", "user_id"))
		if err != nil {
			logFailure("saving tweets to datastore", err)
			return err
		}

		for _, tweetsResponse := range tweetsResponse.Tweets {
			_, err = stmt.Exec(tweetsResponse.Id, tweetsResponse.Text, tweetsResponse.Lang, user.Id)
			if err != nil {
				logFailure("loop: saving tweets to datastore", err)
				rollbackOrLogOnError(txn)
				return err
			}
		}
		_, err = stmt.Exec()
		if err != nil {
			logFailure("statement execute: saving tweets to datastore", err)
			return err
		}

		err = f.Database.UpdateSinceId(user.Id, tweetWaterMark, tweetsResponse.Meta.NewestId)
		if err != nil {
			logFailure("checkpointing last read tweet id to datastore", err)
			rollbackOrLogOnError(txn)
			return err
		}
		err = stmt.Close()
		if err != nil {
			logFailure("statement close: saving tweets to datastore", err)
			rollbackOrLogOnError(txn)
			return err
		}
		return txn.Commit()
	}
	return nil
}
