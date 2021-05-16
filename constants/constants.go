package constants

const (
	LoggerId = "logger"

	// exit codes
	INVALID_FLAGS                              = 1
	MENTION_STREAM_LISTEN                      = 2
	MENTIONS_INVALID_PAYLOAD_SKIP_LMIT_REACHED = 3

	// channel capacity
	MENTIONS_CHANNEL_SIZE = 100_000_0

	// others
	MENTIONS_INVALID_PAYLOAD_SKIP_LMIT = 3
	MENTION_STREAM_RECONNECT_DELAY     = 5
	IDS_PULL_LIMIT                     = 100
	IDS_PULL_TIMEOUT_SECONDS           = 10
)
