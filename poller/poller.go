package poller

import "fmt"

const pollFetchUrl = "https://api.twitter.com/2/tweets?ids=%s&expansions=attachments.poll_ids&poll.fields=duration_minutes,end_datetime,id,options,voting_status"

func Fetch(client TwitterClient, bearer string, tweetId string) {
	url := fmt.Sprintf(pollFetchUrl, tweetId)
	resp, err := client.GetPoll(url, bearer)
	if err == nil {
		fmt.Printf("%+v\n", resp)
		fmt.Printf("After.....\n")
		keepOnlyClosed(&resp)
		fmt.Printf("%+v", resp)
	} else {
		fmt.Printf("%v", err)
	}
}

func keepOnlyClosed(resp *PollFetchResponse) {
	if resp.Includes.Polls != nil {
		copy := resp.Includes.Polls[:0]
		for _, poll := range resp.Includes.Polls {
			if poll.VotingStatus == "closed" {
				copy = append(copy, poll)
			}
		}
		resp.Includes.Polls = copy
	}
}
