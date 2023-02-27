package process

// SovereignNotifier defines what a sovereign notifier should do
type SovereignNotifier interface {
}

// WSClient defines what a websocket client should do
type WSClient interface {
	Start()
	Close()
}
