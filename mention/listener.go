package mention

import (
	"fmt"
	"mrnakumar.com/poli/poller"
	"os"
)

type Listener struct {
	Client    poller.TwitterClient
	MentionId string
}

func (sl Listener) ListenSelf(bearer string) {
	filterBody := fmt.Sprintf(`{"add":[{"value": "@%s is:reply"}]}`, sl.MentionId)
	err := sl.Client.SetStreamFilter(filterBody, bearer)
	if err != nil {
		// TODO: replace with logs
		fmt.Printf("Failed to setup stream filter: %v", err)
		os.Exit(1)
	}
	res, err := sl.Client.MentionStream(bearer)
	if err != nil {
		fmt.Printf("Couldn't get mention: %v\n", err)
	} else if res == nil {
		fmt.Println("Empty response")
	} else {
		fmt.Printf("%+v", res)
	}
}
