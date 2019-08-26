package controller

import (
	"fmt"

	"github.com/n3wscott/sources/cmd/demos/salmonrun/pkg/controller"
	moron "github.com/spencer-p/moroncloudevents"
)

func RegisterHandlers(svr *moron.Server, role, datapath string) error {
	switch role {
	case "salmon":
		svr.HandleCloudEvents(salmonEventReceiver)
		svr.HandleFunc("/websocket", salmonWebSocket)
	case "bear":
		svr.HandleCloudEvents(bearEventReceiver)
		svr.HandleFunc("/websocket", bearWebSocket)
	default:
		return fmt.Errorf("unknown role %q", role)
	}

	svr.Handle("/", http.FileServer(datapath))
}
