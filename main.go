package main

import (
	"flag"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"mrnakumar.com/poli/poller"
	"net/http"
	"os"
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
	if tweets, err := client.GetTweets("37365807"); err == nil {
		for _, tweet := range tweets.Tweets {
			log.Info().Msg(tweet.Text)
			log.Info().Msg("\n")
		}
	} else {
		log.Panic().Str(constants.LoggerId, loggerId).Err(err)
	}

}
