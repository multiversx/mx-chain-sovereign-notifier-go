package config

// Config holds notifier configuration
type Config struct {
	SubscribedAddresses []string        `toml:"subscribed_addresses"`
	WebSocketConfig     WebSocketConfig `toml:"web_socket"`
}

// WebSocketConfig holds web sockets config
type WebSocketConfig struct {
	Url                string `toml:"url"`
	MarshallerType     string `toml:"marshaller_type"`
	RetryDuration      uint32 `toml:"retry_duration"`
	BlockingAckOnError bool   `toml:"blocking-ack-on-error"`
}
