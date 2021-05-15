package main

import (
	"flag"
	"fmt"
	"mrnakumar.com/poli/poller"
)

func main() {
	var bearer string
	var tweetId string
	flag.StringVar(&bearer, "b", "", "<Mandatory> Bearer Token")
	flag.StringVar(&tweetId, "t", "", "<Mandatory> Tweet id")
	flag.Parse()
	if bearer == "" || tweetId == "" {
		flag.PrintDefaults()
		fmt.Println("use -h to see help")
		return
	}
	client := CreateHttpTwitterClient()
	poller.Fetch(client, bearer, tweetId)
}
