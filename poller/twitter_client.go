package poller

import (
	"encoding/json"
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
	GetPoll(url string, bearer string) (PollFetchResponse, error)
	SetStreamFilter(filterBody string, bearer string) error
	MentionStream(bearer string) (*MentionElement, error)
}

type HttpTwitterClient struct {
	client *http.Client
}

func CreateHttpTwitterClient() HttpTwitterClient {
	return HttpTwitterClient{client: &http.Client{}}
}
func (c HttpTwitterClient) GetPoll(url string, bearer string) (data PollFetchResponse, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	addBearer(req, bearer)
	res, err := c.client.Do(req)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("could not get poll for url '%s'", url)
		return
	}
	if res.StatusCode != http.StatusOK {
		log.Warn().Str(constants.LoggerId, loggerId).Err(err).Msgf("could not get poll for url '%s'. status '%d'", url, res.StatusCode)
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&data)
	return
}

func (c HttpTwitterClient) SetStreamFilter(filterBody string, bearer string) error {
	req, _ := http.NewRequest("POST", setStreamFilterUrl, strings.NewReader(filterBody))
	addBearer(req, bearer)
	req.Header.Add("Content-type", "application/json")
	res, err := c.client.Do(req)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msg("could not set stream filter")
		return err
	}
	if res.StatusCode != http.StatusCreated {
		log.Warn().Str(constants.LoggerId, loggerId).Msgf("could not set stream filter. status '%d'", res.StatusCode)
	} else {
		log.Info().Str(constants.LoggerId, loggerId).Msg("set stream filter is successful")
	}
	return nil
}

func (c HttpTwitterClient) MentionStream(bearer string) (mention *MentionElement, err error) {
	req, _ := http.NewRequest("GET", listenMentionUrl, nil)
	addBearer(req, bearer)
	res, err := c.client.Do(req)
	if err != nil {
		log.Panic().Str(constants.LoggerId, loggerId).Err(err).Msgf("could not listen to mention stream. status '%s'", res.StatusCode)
		os.Exit(constants.MENTION_STREAM_LISTEN)
	}
	if res.StatusCode != http.StatusOK {
		log.Panic().Str(constants.LoggerId, loggerId).Msgf("could not listen to mention stream. status '%s'", res.StatusCode)
		os.Exit(constants.MENTION_STREAM_LISTEN)
	}
	err = json.NewDecoder(res.Body).Decode(&mention)
	return
}
func addBearer(req *http.Request, bearer string) {
	req.Header.Add("Authorization", "Bearer "+bearer)
}

type PollFetchResponse struct {
	Data     []Data   `json:"data"`
	Includes Includes `json:"includes"`
}

type Data struct {
	ID   string `json:"id"`
	Text string `json:"text"`
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
