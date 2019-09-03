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
