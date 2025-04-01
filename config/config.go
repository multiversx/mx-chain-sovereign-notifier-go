package config

// Config holds notifier configuration
type Config struct {
	SubscribedEvents    []SubscribedEvent `toml:"subscribed_events"`
	HasherType          string            `toml:"hasher_type"`
	WebSocketConfig     WebSocketConfig   `toml:"web_socket"`
	AddressPubKeyConfig PubkeyConfig      `toml:"address_pubkey_converter"`
}

// SubscribedEvent holds subscribed events config
type SubscribedEvent struct {
	Identifier string   `toml:"identifier"`
	Addresses  []string `toml:"addresses"`
}

// WebSocketConfig holds web sockets config
type WebSocketConfig struct {
	Url                string `toml:"url"`
	MarshallerType     string `toml:"marshaller_type"`
	Mode               string `toml:"mode"`
	RetryDuration      uint32 `toml:"retry_duration"`
	WithAcknowledge    bool   `toml:"with_acknowledge"`
	BlockingAckOnError bool   `toml:"blocking_ack_on_error"`
	AcknowledgeTimeout int    `toml:"acknowledge_timeout"`
	Version            uint32 `toml:"version"`
}

// PubkeyConfig will map the public key configuration
type PubkeyConfig struct {
	Length int
	Hrp    string
}
