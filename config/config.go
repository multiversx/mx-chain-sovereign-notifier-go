package config

// Config holds notifier configuration
type Config struct {
	SubscribedAddresses []string `toml:"subscribed_addresses"`
}
