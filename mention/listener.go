package mention

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"mrnakumar.com/poli/poller"
	"os"
	"strings"
	"time"
)

const loggerId = "listener"
const pollFetchUrl = "https://api.twitter.com/2/tweets?ids=%s&expansions=attachments.poll_ids&poll.fields=duration_minutes,end_datetime,id,options,voting_status"

type Listener struct {
	Client    poller.TwitterClient
	MentionId string
}

func (sl Listener) ListenSelf() {
	filterBody := fmt.Sprintf(`{"add":[{"value": "@%s is:reply"}]}`, sl.MentionId)
	err := sl.Client.SetStreamFilter(filterBody)
	if err != nil {
		log.Panic().Str(constants.LoggerId, loggerId).Err(err).Msg("could not set stream filter")
		os.Exit(1)
	} else {
		log.Info().Str(constants.LoggerId, loggerId).Msg("set stream filter is successful")
	}
	mentions := sl.Client.MentionStream()
	var ids []poller.MentionElement
	for {
		select {
		case mention := <-mentions:
			{
				ids = append(ids, mention)
				if len(ids) == constants.IDS_PULL_LIMIT {
					sl.drainIfNotEmpty(ids)
					ids = make([]poller.MentionElement, 0)
				}
			}
		case <-time.After(constants.IDS_PULL_TIMEOUT_SECONDS * time.Second):
			{
				sl.drainIfNotEmpty(ids)
				ids = make([]poller.MentionElement, 0)
			}
		}
	}
}

func (sl Listener) drainIfNotEmpty(mentions []poller.MentionElement) {
	if len(mentions) > 0 {
		var tweetIds []string
		for _, m := range mentions {
			tweetIds = append(tweetIds, m.Data.ConversationID)
		}
		sl.fetch(tweetIds)
	}
}
func (sl Listener) fetch(tweetIds []string) {
	url := fmt.Sprintf(pollFetchUrl, strings.Join(tweetIds, ","))
	resp, err := sl.Client.GetPoll(url)
	if err == nil {
		keepOnlyClosed(&resp)

		fmt.Printf("%+v", resp)
	} else {
		log.Error().Str(constants.LoggerId, loggerId).Strs("tweetIds", tweetIds).Err(err).Msg("failed to fetch tweets")
	}
}

func keepOnlyClosed(resp *poller.PollFetchResponse) {
	if resp.Includes.Polls != nil {
		temp := resp.Includes.Polls[:0]
		for _, poll := range resp.Includes.Polls {
			if poll.VotingStatus == "closed" {
				temp = append(temp, poll)
			}
		}
		resp.Includes.Polls = temp
	}
}
