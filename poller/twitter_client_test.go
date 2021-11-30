package poller

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)
import "net/http"

type MockClient struct {
	Body        string
	ReturnError bool
	StatusCode  int
}

const tweetsResponseBody = `{
    "data": [
        {
            "id": "1301573587187331074",
            "text": "Starting today, you can see your monthly Tweet usage for the v2 API in the"
        },
        {
            "id": "1296887316556980230",
            "text": "See how @PennMedCDH are using Twitter data to understand the COVID-19 health crisis"
        }
    ],
    "meta": {
        "newest_id": "1301573587187331074",
        "next_token": "t3buvdr5pujq9g7bggsnf3ep2ha28",
        "oldest_id": "1296887316556980230",
        "previous_token": "t3equkmcd2zffvags2nkj0nhlrn78",
        "result_count": 2
    }
}`

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	if c.ReturnError {
		return nil, errors.New("network error")
	}
	body := []byte(c.Body)
	r := ioutil.NopCloser(bytes.NewReader(body))
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       r,
	}
	if c.StatusCode > 0 {
		response.StatusCode = c.StatusCode
	}
	return response, nil
}

func TestGetTweets(t *testing.T) {
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{
			Body: tweetsResponseBody,
		},
	}
	response, err := twitterClient.GetTweets("37365807")
	if err != nil {
		t.Errorf("Error = %v; expected nil", err)
	}
	if response.Meta.ResultCount != 2 {
		t.Errorf("ResultCount = %d; expected = 2", response.Meta.ResultCount)
	}

	var match int64 = 0
	for _, tweet := range response.Tweets {
		if tweet.Id == "1301573587187331074" || tweet.Id == "1296887316556980230" {
			match = match + 1
		}
	}

	if match != response.Meta.ResultCount {
		t.Errorf("tweetCount = %d; expected = %d", match, response.Meta.ResultCount)
	}

}

func TestGetTweetsInvalidJson(t *testing.T) {
	invalidJson := "." + tweetsResponseBody
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{Body: invalidJson},
	}
	_, err := twitterClient.GetTweets("37365807")
	if err == nil {
		t.Error("expected error to be present")
	}
}

func TestGetTweetsClientError(t *testing.T) {
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{ReturnError: true},
	}
	_, err := twitterClient.GetTweets("37365807")
	if err == nil {
		t.Error("expected error to be present")
	}
}

func TestGetTweetsHttpStatusNotOk(t *testing.T) {
	errorCode := 404
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{StatusCode: errorCode},
	}
	_, err := twitterClient.GetTweets("37365807")
	if err == nil {
		t.Error("expected error to be present")
	}
	expected := fmt.Sprintf("unknown status code '%d'", errorCode)
	if strings.Compare(err.Error(), expected) != 0 {
		t.Errorf("error = '%s'; expected: '%s'", err.Error(), expected)
	}
}
