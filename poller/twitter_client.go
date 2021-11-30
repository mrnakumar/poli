package poller

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"mrnakumar.com/poli/constants"
	"net/http"
	"strings"
)

const loggerId = "twitter_client"
const userTweetsUrl = "https://api.twitter.com/2/users/:id/tweets?max_results=5"

type TweetId = string
type TwitterUserId = string

type HttpTwitterClient struct {
	Bearer string
	Client HttpClient
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (c HttpTwitterClient) GetTweets(userId TwitterUserId) (*TweetsResponse, error) {
	url := strings.ReplaceAll(userTweetsUrl, ":id", userId)
	var tweets TweetsResponse
	err := getRequest(&c, url, &tweets)
	if err == nil {
		return &tweets, nil
	}
	return nil, err
}

func getRequest(c *HttpTwitterClient, url string, v interface{}) error {
	req, _ := http.NewRequest("GET", url, nil)
	addBearer(req, c.Bearer)
	res, err := c.Client.Do(req)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("failed to get for url '%s'", url)
		return err
	}
	if res.StatusCode != http.StatusOK {
		log.Warn().Str(constants.LoggerId, loggerId).Err(err).Msgf("unknown status code '%d' for url '%s'",
			res.StatusCode, url)
		return fmt.Errorf("unknown status code '%d'", res.StatusCode)
	}
	defer closeOrLogWarningIfFailed(res.Body)
	err = json.NewDecoder(res.Body).Decode(&v)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err)
		return err
	}
	return err
}

func addBearer(req *http.Request, bearer string) {
	req.Header.Add("Authorization", "Bearer "+bearer)
}

func closeOrLogWarningIfFailed(body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		log.Warn().Str(constants.LoggerId, loggerId).Err(err).Msg("Failed to close response body")
	}
}

type UserInfo struct {
}

type TweetsResponse struct {
	Tweets []Tweet `json:"data"`
	Meta   Meta    `json:"meta"`
}

type Tweet struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

type Meta struct {
	NewestId    string `json:"newest_id"`
	NextToken   string `json:"next_token"`
	ResultCount int64  `json:"result_count"`
}
