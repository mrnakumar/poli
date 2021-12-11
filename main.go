package main

import (
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"mrnakumar.com/poli/fetch"
	"net/http"
	"os"
	"time"
)

const loggerId = "main"
const (
	FlagBearer     string = "bearer"
	FlagDbHost            = "dbHost"
	FlagDbName            = "dbName"
	FlagDbUser            = "dbUser"
	FlagDbPassword        = "dbPassword"
	FlagAction            = "action"
	FlagUserName          = "userName"
)
const actionDownloadUser = "downloadUser"
const actionDownloadTweets = "downloadTweetsForAllUsers"

type Flags struct {
	bearerToken string
	dbHost      string
	dbName      string
	dbUser      string
	dbPassword  string
	action      string
	userName    string
}

func main() {
	flags := parseFlags()
	twitterClient := fetch.HttpTwitterClient{Bearer: flags.bearerToken, Client: &http.Client{}}
	database := fetch.GetDb(flags.dbHost, flags.dbUser, flags.dbPassword, flags.dbName, false)
	defer closeDb(database)
	fetcher := fetch.Fetcher{
		TwitterClient: twitterClient,
		Database:      database,
	}
	if flags.action == actionDownloadUser {
		err := fetcher.AddUser(flags.userName)
		if err != nil {
			log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("failed to add username '%s'", flags.userName)
			return
		}
	}
	if flags.action == actionDownloadTweets {
		err := fetcher.GetAllUserTweets()
		if err != nil {
			log.Error().Str(constants.LoggerId, loggerId).Err(err).Msg("error in getting tweets")
			return
		}
	}
	log.Error().Str(constants.LoggerId, loggerId).Msgf("completed action '%s'", flags.action)
}

func parseFlags() Flags {
	var bearer string
	var dbHost string
	var dbName string
	var dbUser string
	var dbPassword string
	var action string
	var userName string

	flag.StringVar(&bearer, FlagBearer, "", "<Mandatory> Bearer Token")
	flag.StringVar(&dbHost, FlagDbHost, "", "<Mandatory> Database Host")
	flag.StringVar(&dbName, FlagDbName, "", "<Mandatory> Database Name")
	flag.StringVar(&dbUser, FlagDbUser, "", "<Mandatory> Database User")
	flag.StringVar(&dbPassword, FlagDbPassword, "", "<Mandatory> Database Password")
	flag.StringVar(&action, FlagAction, "", fmt.Sprintf("<Mandatory> action. Can be one of ['%s', '%s'", actionDownloadUser, actionDownloadTweets))
	flag.StringVar(&userName, FlagUserName, "", fmt.Sprintf("<Optional> The name of the user that is to be downloaded. Mandatory if %s = '%s'", FlagAction, actionDownloadUser))

	flag.Parse()
	flags := Flags{
		bearerToken: bearer,
		dbHost:      dbHost,
		dbName:      dbName,
		dbUser:      dbUser,
		dbPassword:  dbPassword,
		action:      action,
		userName:    userName,
	}
	validateOrExit(flags)
	return flags
}
func findUser(client fetch.HttpTwitterClient) {
	userName := "Profdilipmandal"
	response, err := client.FindUser(userName)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("Failed to find username '%s'", userName)
	} else {
		log.Info().Str(constants.LoggerId, loggerId).Msgf("%v",
			response.Data)
	}
}

func validateOrExit(flags Flags) {
	if flags.bearerToken == "" {
		printHelpAndExit("Bearer token is required")
	}
	if flags.action != actionDownloadTweets && flags.action != actionDownloadUser {
		printHelpAndExit(fmt.Sprintf("'%s' must be one of: ['%s', '%s']", FlagAction, actionDownloadUser, actionDownloadTweets))
	}
	if flags.action == actionDownloadUser && flags.userName == "" {
		printHelpAndExit(fmt.Sprintf("'%s' is required for '%s'", FlagUserName, actionDownloadUser))
	}
	if flags.dbHost == "" || flags.dbName == "" || flags.dbUser == "" || flags.dbPassword == "" {
		printHelpAndExit("Missing token related to database")
	}
}
func printHelpAndExit(msg string) {
	flag.PrintDefaults()
	if len(msg) > 0 {
		log.Info().Str(constants.LoggerId, loggerId).Msg(msg)
	}
	log.Error().Str(constants.LoggerId, loggerId).Msg("use -h to see help")
	os.Exit(constants.INVALID_FLAGS)
}
func getTweets(client fetch.HttpTwitterClient) {
	location, _ := time.LoadLocation("UTC")
	startTime := time.Now().In(location).AddDate(0, 0, -1).Format(time.RFC3339)
	if tweets, err := client.GetTweets("37365807", 100, "", startTime); err == nil {
		for _, tweet := range tweets.Tweets {
			log.Info().Msgf("id='%s', lang='%s', text='%s'\n", tweet.Id, tweet.Lang, tweet.Text)
		}
	} else {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msg("OOPs")
	}
}

func closeDb(ds *fetch.Database) {
	err := ds.Close()
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msg("failed to close datastore")
	}
}
