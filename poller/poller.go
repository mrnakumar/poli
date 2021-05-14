package poller

import "fmt"

const pollFetchUrl = "https://api.twitter.com/2/tweets?ids=%s&expansions=attachments.poll_ids&poll.fields=duration_minutes,end_datetime,id,options,voting_status"

func Fetch(client TwitterClient, bearer string, tweetId string) {
	url := fmt.Sprintf(pollFetchUrl, tweetId)
	resp, err := client.GetPoll(url, bearer)
	if err == nil {
		fmt.Printf("%+v", resp)
	}
	fmt.Printf("%v", err)
}
