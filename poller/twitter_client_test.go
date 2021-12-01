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
	InvalidJson   bool
	ReturnError   bool
	StatusCode    int
	requestNumber uint8
}

const tweetsPerResponse = 5
const sinceTweetId = "1301573587187331066"
const tweetsResponseBody1 = `{
    "data": [
        {
            "id": "1301573587187331075",
            "lang" : "en",
            "text": "Starting today, you can see your monthly Tweet usage for the v2 API in the"
        },
        {
            "id": "1301573587187331074",
            "lang" : "en",
            "text": "See how @PennMedCDH are using Twitter data to understand the COVID-19 health crisis"
        },
        {
            "id": "1301573587187331073",
            "lang" : "en",
            "text": "@blablauser001 you are nobody."
        },
        {
            "id": "1301573587187331072",
            "lang" : "en",
            "text": "Yes, I repeat you are no body."
        },
        {
            "id": "1301573587187331071",
            "lang" : "hi",
            "text": "ठीक स। चुप रह"
        }
    ],
    "meta": {
        "newest_id": "1301573587187331075",
        "next_token": "t3buvdr5pujq9g7bggsnf3ep2ha28",
        "result_count": 5
    }
}`

const tweetsResponseBody2 = `{
    "data": [
        {
            "id": "1301573587187331070",
            "lang" : "en",
            "text": "Starting today, you can see your monthly Tweet usage for the v2 API in the"
        },
        {
            "id": "1301573587187331069",
            "lang" : "en",
            "text": "See how @PennMedCDH are using Twitter data to understand the COVID-19 health crisis"
        },
        {
            "id": "1301573587187331068",
            "lang" : "en",
            "text": "@blablauser001 you are nobody."
        },
        {
            "id": "1301573587187331067",
            "lang" : "en",
            "text": "Yes, I repeat you are no body."
        },
        {
            "id": "1301573587187331066",
            "lang" : "hi",
            "text": "ठीक स। चुप रह"
        }
    ],
    "meta": {
        "newest_id": "1301573587187331070",
        "next_token": "",
        "result_count": 5
    }
}`

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	c.requestNumber++
	if c.ReturnError {
		return nil, errors.New("network error")
	}
	var bodyText string
	if c.InvalidJson {
		bodyText = "." + tweetsResponseBody1
	} else if c.requestNumber == 1 {
		bodyText = tweetsResponseBody1
	} else if c.requestNumber == 2 {
		bodyText = tweetsResponseBody2
	} else {
		return nil, errors.New("only two requests are supported by this test client")
	}
	body := []byte(bodyText)
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
		Client: &MockClient{},
	}
	response, err := twitterClient.GetTweets("37365807", tweetsPerResponse, sinceTweetId, "")
	if err != nil {
		t.Errorf("Error = %v; expected nil", err)
	}
	if len(response.Tweets) != 10 {
		t.Errorf("result count = %d; expected = 10", response.Meta.ResultCount)
	}
	nextToken := response.Meta.NextToken
	if len(nextToken) > 0 {
		t.Errorf("next token = %s; expected = ''", nextToken)
	}
	newestId := response.Meta.NewestId
	if strings.Compare(newestId, "1301573587187331075") != 0 {
		t.Errorf("newest id = %s; expected = '1301573587187331075'", newestId)
	}
}

func TestGetTweetsInvalidJson(t *testing.T) {
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{InvalidJson: true},
	}
	_, err := twitterClient.GetTweets("37365807", tweetsPerResponse, sinceTweetId, "")
	if err == nil {
		t.Error("expected error to be present")
	}
}

func TestGetTweetsClientError(t *testing.T) {
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{ReturnError: true},
	}
	_, err := twitterClient.GetTweets("37365807", tweetsPerResponse, sinceTweetId, "")
	if err == nil {
		t.Error("expected error to be present")
	}
}

func TestGetTweetsBothSinceIdAndSinceTimeMissing(t *testing.T) {
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{},
	}
	_, err := twitterClient.GetTweets("37365807", tweetsPerResponse, "", "")
	expectedError := "start tweet id or time must be provided"
	if strings.Compare(err.Error(), expectedError) != 0 {
		t.Errorf("expected error '%s'", expectedError)
	}
}

func TestGetTweetsInvalidTweetsPerResponse(t *testing.T) {
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{},
	}
	_, err := twitterClient.GetTweets("37365807", 4, "", "")
	expectedError := "tweetsPerRequest must be between 5 to 100, both inclusive"
	if strings.Compare(err.Error(), expectedError) != 0 {
		t.Errorf("expected error '%s'", expectedError)
	}
}

func TestGetTweetsHttpStatusNotOk(t *testing.T) {
	errorCode := 404
	twitterClient := HttpTwitterClient{
		Bearer: "",
		Client: &MockClient{StatusCode: errorCode},
	}
	_, err := twitterClient.GetTweets("37365807", tweetsPerResponse, sinceTweetId, "")
	if err == nil {
		t.Error("expected error to be present")
	}
	expected := fmt.Sprintf("unknown status code '%d'", errorCode)
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("error = '%s'; expected to contain: '%s'", err.Error(), expected)
	}
}
