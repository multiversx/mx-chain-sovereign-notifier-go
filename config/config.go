package config

// Config holds notifier configuration
type Config struct {
	SubscribedEvents []SubscribedEvent `toml:"subscribed_events"`
	WebSocketConfig  WebSocketConfig   `toml:"web_socket"`
}

type SubscribedEvent struct {
	Identifier string   `toml:"identifier"`
	Addresses  []string `toml:"addresses"`
}

// WebSocketConfig holds web sockets config
type WebSocketConfig struct {
	Url                string `toml:"url"`
	MarshallerType     string `toml:"marshaller_type"`
	RetryDuration      uint32 `toml:"retry_duration"`
	BlockingAckOnError bool   `toml:"blocking_ack_on_error"`
	HasherType         string `toml:"hasher_type"`
}
