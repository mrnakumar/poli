package poller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"mrnakumar.com/poli/constants"
	"net/http"
	"os"
	"strings"
	"time"
)

const setStreamFilterUrl = "https://api.twitter.com/2/tweets/search/stream/rules"
const listenMentionUrl = "https://api.twitter.com/2/tweets/search/stream?tweet.fields=conversation_id,created_at"
const loggerId = "twitter_client"

type TwitterClient interface {
	GetPoll(url string) (PollFetchResponse, error)
	SetStreamFilter(filterBody string) error
	MentionStream() <-chan MentionElement
}

type HttpTwitterClient struct {
	bearer string
	client *http.Client
}

func CreateHttpTwitterClient(bearer string) HttpTwitterClient {
	return HttpTwitterClient{bearer: bearer, client: &http.Client{}}
}
func (c HttpTwitterClient) GetPoll(url string) (data PollFetchResponse, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	addBearer(req, c.bearer)
	res, err := c.client.Do(req)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("could not get poll for url '%s'", url)
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Warn().Str(constants.LoggerId, loggerId).Err(err).Msgf("could not get poll for url '%s'. status '%d'",
			url, res.StatusCode)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&data)
	return
}

func (c HttpTwitterClient) SetStreamFilter(filterBody string) error {
	req, _ := http.NewRequest("POST", setStreamFilterUrl, strings.NewReader(filterBody))
	addBearer(req, c.bearer)
	req.Header.Add("Content-type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("status '%s'", res.StatusCode)
	}
	return nil
}

func (c HttpTwitterClient) MentionStream() <-chan MentionElement {
	reader := c.connectToMentionStream(false)
	mentions := make(chan MentionElement, constants.MENTIONS_CHANNEL_SIZE)
	go func() {
		invalid := 0
		for {
			var mention MentionElement
			line, err := reader.ReadBytes('\n')
			if err != nil {
				log.Error().Str(constants.LoggerId, loggerId).Err(err).Msg("could not reach mention response")
				reader = c.connectToMentionStream(true)
			} else {
				err = json.NewDecoder(bytes.NewReader(line)).Decode(&mention)
				if err != nil {
					invalid = invalid + 1
					log.Warn().Str(constants.LoggerId, loggerId).Err(err).Msgf("invalid mention payload '%s'", line)
					if invalid > constants.MENTIONS_INVALID_PAYLOAD_SKIP_LMIT {
						log.Panic().Str(constants.LoggerId, loggerId).Msg("invalid mention payload limit reached")
						os.Exit(constants.MENTIONS_INVALID_PAYLOAD_SKIP_LMIT_REACHED)
					}
				} else {
					mentions <- mention
					invalid = 0
				}
			}
		}
	}()
	return mentions
}

func (c HttpTwitterClient) connectToMentionStream(delay bool) *bufio.Reader {
	if delay {
		log.Info().Str(constants.LoggerId, loggerId).Msgf("going to sleep for '%d' minutes before reconnecting "+
			"to mention stream",
			constants.MENTION_STREAM_RECONNECT_DELAY)
		time.Sleep(constants.MENTION_STREAM_RECONNECT_DELAY * time.Minute)
	}
	req, _ := http.NewRequest("GET", listenMentionUrl, nil)
	addBearer(req, c.bearer)
	res, err := c.client.Do(req)
	if err != nil {
		log.Panic().Str(constants.LoggerId, loggerId).Err(err).Msg("could not listen to mention stream")
		os.Exit(constants.MENTION_STREAM_LISTEN)
	}
	if res.StatusCode != http.StatusOK {
		log.Panic().Str(constants.LoggerId, loggerId).Msgf("could not listen to mention stream. status '%s'",
			res.StatusCode)
		os.Exit(constants.MENTION_STREAM_LISTEN)
	}
	return bufio.NewReader(res.Body)
}
func addBearer(req *http.Request, bearer string) {
	req.Header.Add("Authorization", "Bearer "+bearer)
}

type PollFetchResponse struct {
	Data     []Data   `json:"data"`
	Includes Includes `json:"includes"`
}

type Data struct {
	ID          string       `json:"id"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}
type Attachment struct {
	PollIds []string `json:"poll_ids"`
}
type Options struct {
	Position int    `json:"position"`
	Label    string `json:"label"`
	Votes    int    `json:"votes"`
}
type Polls struct {
	EndDatetime     time.Time `json:"end_datetime"`
	DurationMinutes int       `json:"duration_minutes"`
	ID              string    `json:"id"`
	VotingStatus    string    `json:"voting_status"`
	Options         []Options `json:"options"`
}
type Includes struct {
	Polls []Polls `json:"polls"`
}

type MentionElement struct {
	Data struct {
		Text           string    `json:"text"`
		ID             string    `json:"id"`
		ConversationID string    `json:"conversation_id"`
		CreatedAt      time.Time `json:"created_at"`
	} `json:"data"`
}
