package mention

import (
	"fmt"
	"mrnakumar.com/poli/poller"
	"os"
)

type Listener struct {
	Client poller.TwitterClient
}

func (sl Listener) ListenSelf(bearer string) {
	filterBody := `{"add":[{"value": "@twitterdev is:reply"}]}`
	err := sl.Client.SetStreamFilter(filterBody, bearer)
	if err != nil {
		// TODO: replace with logs
		fmt.Printf("Failed to setup stream filter: %v", err)
		os.Exit(1)
	}
}
