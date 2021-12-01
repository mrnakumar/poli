package poller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"mrnakumar.com/poli/constants"
	"net/http"
	"strconv"
	"strings"
)

const loggerId = "twitter_client"
const userTweetsUrl = "https://api.twitter.com/2/users/:id/tweets"

type TweetId = string
type TwitterUserId = string
type StartTimeISO8601ZoneUTC = string

type HttpTwitterClient struct {
	Bearer string
	Client HttpClient
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GetTweets
// sinceId takes precedence over startTime. If both of these are missing then error is returned.
// tweetsPerRequest must be between 5 and 100
func (c HttpTwitterClient) GetTweets(userId TwitterUserId, tweetsPerRequest uint8, sinceId TweetId,
	startTime StartTimeISO8601ZoneUTC) (*TweetsResponse, error) {
	url, err := tweetsUrl(userId, tweetsPerRequest, "", sinceId, startTime)
	if err != nil {
		return nil, err
	}

	result := &TweetsResponse{}
	for {
		log.Info().Str(constants.LoggerId, loggerId).Msg("Entering loop")

		var tweets TweetsResponse
		err = getRequest(&c, url, &tweets)
		if err == nil {
			// process the response
			existingNewestId := result.Meta.NewestId
			result.Meta = tweets.Meta
			if len(existingNewestId) > 0 {
				// only update if from first fetch
				result.Meta.NewestId = existingNewestId
			}
			if tweets.Tweets != nil {
				if result.Tweets == nil {
					result.Tweets = tweets.Tweets
				} else {
					for _, tweet := range tweets.Tweets {
						result.Tweets = append(result.Tweets, tweet)
					}
				}
			} else {
				// strange no tweets are returned. meta has already been taken so just break out of loop
				break
			}
			if tweets.Meta.NextToken == "" {
				// not sufficient tweets. means end reached
				log.Info().Str(constants.LoggerId, loggerId).Msgf("received '%d' tweets for userId '%s'.",
					len(tweets.Tweets), userId)
				break
			}
		} else {
			// error encountered in getting the response from twitter
			return nil, err
		}
		url, _ = tweetsUrl(userId, tweetsPerRequest, result.Meta.NextToken, "", "")
	}
	log.Info().Str(constants.LoggerId, loggerId).Msgf("returning a total of '%d' tweets for user id '%s'.",
		len(result.Tweets), userId)
	return result, err
}

func getRequest(c *HttpTwitterClient, url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	addBearer(req, c.Bearer)
	res, err := c.Client.Do(req)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("failed to get for url '%s'", url)
		return fmt.Errorf("request failed for url '%s'", url)
	}
	if res.StatusCode != http.StatusOK {
		msg, err := ioutil.ReadAll(res.Body)
		if err == nil {
			log.Warn().Str(constants.LoggerId, loggerId).Err(err).
				Msgf("failed to parse error response. received status code '%d', message '%s' for url '%s'",
					res.StatusCode, msg, url)
		}
		return fmt.Errorf("unknown status code '%d' for url '%s'", res.StatusCode, url)
	}
	defer closeOrLogWarningIfFailed(res.Body)
	err = json.NewDecoder(res.Body).Decode(&v)
	if err != nil {
		log.Error().Str(constants.LoggerId, loggerId).Err(err).Msgf("failed to decode response for url '%s", url)
		body, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return fmt.Errorf("failed to decode response body '%s' for url '%s'", string(body), url)
		}
		return fmt.Errorf("failed to decode response body for url '%s'", url)
	}
	return err
}

func tweetsUrl(userId TwitterUserId, tweetsPerRequest uint8, paginationToken string, sinceId TweetId,
	startTime StartTimeISO8601ZoneUTC) (string, error) {
	if tweetsPerRequest < 5 || tweetsPerRequest > 100 {
		return "", errors.New("tweetsPerRequest must be between 5 to 100, both inclusive")
	}
	var tweetsUrl = strings.ReplaceAll(userTweetsUrl, ":id", userId)
	var queryPart = "?max_results=" + strconv.FormatUint(uint64(tweetsPerRequest), 10)
	if len(paginationToken) > 0 {
		queryPart = queryPart + "&pagination_token=" + paginationToken
	} else if len(strings.TrimSpace(sinceId)) > 0 {
		queryPart = queryPart + "&since_id=" + sinceId
	} else {
		if len(strings.TrimSpace(startTime)) == 0 {
			return "", errors.New("start tweet id or time must be provided")
		}
		if len(strings.TrimSpace(startTime)) > 0 {
			queryPart = queryPart + "&start_time=" + startTime
		}
	}
	queryPart = queryPart + "&tweet.fields=id,text,lang"
	return tweetsUrl + queryPart, nil
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
	Lang string `json:"lang"`
}

type Meta struct {
	NewestId    string `json:"newest_id"`
	NextToken   string `json:"next_token"`
	ResultCount uint8  `json:"result_count"`
}
