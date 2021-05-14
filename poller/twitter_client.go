package poller

import (
	"encoding/json"
	"net/http"
	"time"
)

type TwitterClient interface {
	GetPoll(url string, bearer string) (PollFetchResponse, error)
}

type HttpTwitterClient struct {
	client *http.Client
}

func CreateHttpTwitterClient() HttpTwitterClient {
	return HttpTwitterClient{client: &http.Client{
	}}
}
func (c HttpTwitterClient) GetPoll(url string, bearer string) (data PollFetchResponse, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+bearer)
	res, err := c.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&data)
	return
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
