package controller

const (
	EVENT_SOURCE      = "https://github.com/n3wscott/sources/cmd/demos/salmonrun"
	SALMON_EVENT_TYPE = "com.github.n3wscott.sources.demos.salmonrun.salmon"
	BEAR_EVENT_TYPE   = "com.github.n3wscott.sources.demos.salmonrun.bear"
)

var (
	// conns is a collection of send methods for connections
	conns = make(map[string]ConnectionSender)

	// A collection of timeout message channels for the bear
	timeoutchans = make(map[string]chan Message)
)
