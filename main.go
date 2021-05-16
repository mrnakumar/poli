package main

import (
	"flag"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"mrnakumar.com/poli/mention"
	"mrnakumar.com/poli/poller"
	"os"
)

const loggerId = "main"

func main() {
	var bearer string
	var tweetId string
	var mentionId string
	flag.StringVar(&bearer, "b", "", "<Mandatory> Bearer Token")
	flag.StringVar(&tweetId, "t", "", "<Mandatory> Tweet id")
	flag.StringVar(&mentionId, "m", "", "<Mandatory> Id to look for in mentions")
	flag.Parse()
	if bearer == "" || tweetId == "" {
		flag.PrintDefaults()
		log.Panic().Str(constants.LoggerId, loggerId).Msg("use -h to see help")
		os.Exit(constants.INVALID_FLAGS)
	}
	client := poller.CreateHttpTwitterClient()
	mentionListener := mention.Listener{Client: client, MentionId: mentionId}
	mentionListener.ListenSelf(bearer)
}
