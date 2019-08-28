/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"log"
	"net/http"
	"path"

	moron "github.com/spencer-p/moroncloudevents"
)

func withLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}

func RegisterHandlers(svr *moron.Server, role, datapath string) error {
	switch role {
	case "salmon":
		svr.HandleCloudEvents(salmonEventReceiver)
		svr.Handle("/websocket", withLog(makeWebSocketHandle(makeSalmonWSReceiver(svr.CloudEventClient()))))
	case "bear":
		svr.HandleCloudEvents(bearEventReceiver)
		svr.Handle("/websocket", withLog(makeWebSocketHandle(makeBearWSReceiver(svr.CloudEventClient()))))
	default:
		return fmt.Errorf("unknown role %q", role)
	}

	svr.Handle("/static/shared/", withLog(http.StripPrefix("/static/shared", http.FileServer(http.Dir(path.Join(datapath, "shared"))))))
	svr.Handle("/static/", withLog(http.StripPrefix("/static", http.FileServer(http.Dir(path.Join(datapath, role))))))
	return nil
}
