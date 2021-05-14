package main

import "mrnakumar.com/poli/poller"

func main() {
	client := poller.CreateHttpTwitterClient()
	poller.Fetch(client, bearer, tweetId)
}
