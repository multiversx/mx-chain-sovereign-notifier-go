package process

type SovereignNotifier interface {
}

// WSClient defines what a websocket client should do
type WSClient interface {
	Start()
	Close()
}
