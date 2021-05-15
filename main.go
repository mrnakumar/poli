package main

import (
	"flag"
	"fmt"
	"mrnakumar.com/poli/mention"
	"mrnakumar.com/poli/poller"
)

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
		fmt.Println("use -h to see help")
		return
	}
	client := poller.CreateHttpTwitterClient()
	mentionListener := mention.Listener{Client: client, MentionId: mentionId}
	mentionListener.ListenSelf(bearer)
	//poller.Fetch(client, bearer, tweetId)
}
