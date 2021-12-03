package main

import (
	"flag"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"mrnakumar.com/poli/poller"
	"net/http"
	"os"
	"time"
)

const loggerId = "main"

func main() {
	var bearer string
	flag.StringVar(&bearer, "b", "", "<Mandatory> Bearer Token")
	flag.Parse()
	if bearer == "" {
		flag.PrintDefaults()
		log.Panic().Str(constants.LoggerId, loggerId).Msg("use -h to see help")
		os.Exit(constants.INVALID_FLAGS)
	}
	client := poller.HttpTwitterClient{Bearer: bearer, Client: &http.Client{}}
	//getTweets(client)
	findUser(client)

}

func findUser(client poller.HttpTwitterClient) {
	userName := "Profdilipmandal"
	response, err := client.FindUser(userName)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("Failed to find username '%s'", userName)
	} else {
		log.Info().Str(constants.LoggerId, loggerId).Msgf("%v",
			response.Data)
	}
}

func getTweets(client poller.HttpTwitterClient) {
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
