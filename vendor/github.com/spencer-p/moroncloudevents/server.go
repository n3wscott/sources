package moroncloudevents

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventsclient "github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
)

// ServerConfig is a struct for options when constructing a new HTTP server.
type ServerConfig struct {
	// Port is the port that serves both HTTP handlers and CloudEvent receiving.
	// Defaults to 80.
	// +optional
	Port string

	// CloudEventReceivePath is the path reserved for CloudEvents.
	// If omitted, defaults to "/".
	// +optional
	CloudEventReceivePath string

	// CloudEventTargets is a slice of targets that the client will send CloudEvents on.
	// +optional
	CloudEventTargets []string

	// ConvertFn is a function to convert non-CloudEvent requests to the CloudEventReceivePath into CloudEvents.
	// +optional
	ConvertFn cloudevents.ConvertFn

	// TransportOptions are forwarded directly to CloudEvent transport construction.
	// +optional
	TransportOptions []cloudeventshttp.Option

	// ClientOptions are forwarded directly to CloudEvent client construction.
	// +optional
	ClientOptions []cloudeventsclient.Option
}

func (conf *ServerConfig) setDefaults() {
	if conf.Port == "" {
		conf.Port = "80"
	}
	if conf.CloudEventReceivePath == "" {
		conf.CloudEventReceivePath = "/"
	}
}

// Server allows you to simply serve HTTP handlers and a CloudEvent receiver side-by-side.
type Server struct {
	*http.ServeMux
	cetransport *cloudevents.HTTPTransport
	ceclient    cloudevents.Client
	cehandler   cloudeventsclient.ReceiveFull
	shutdown    context.CancelFunc
}

func NewServer(conf *ServerConfig) (*Server, error) {
	conf.setDefaults()

	portint, err := strconv.Atoi(conf.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	tOpts := append(conf.TransportOptions, []cloudeventshttp.Option{
		cloudevents.WithPath(conf.CloudEventReceivePath),
		cloudevents.WithPort(portint),
	}...)
	for _, target := range conf.CloudEventTargets {
		tOpts = append(tOpts, cloudevents.WithTarget(target))
	}

	transport, err := cloudevents.NewHTTPTransport(tOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to construct transport: %v", err)
	}

	transport.Client = &http.Client{
		Transport: &ochttp.Transport{
			Propagation: &b3.HTTPFormat{},
		},
	}

	transport.Handler = http.NewServeMux()

	cOps := append(conf.ClientOptions, []cloudeventsclient.Option{
		cloudevents.WithUUIDs(),
		cloudevents.WithTimeNow(),
	}...)
	if conf.ConvertFn != nil {
		cOps = append(cOps, cloudevents.WithConverterFn(conf.ConvertFn))
	}
	client, err := cloudevents.NewClient(transport, cOps...)

	return &Server{
		ServeMux:    transport.Handler,
		cetransport: transport,
		ceclient:    client,
	}, nil
}

// HandleCloudEvents sets the handler for CloudEvent receiveing. There can only be one.
func (s *Server) HandleCloudEvents(handler cloudeventsclient.ReceiveFull) {
	s.cehandler = handler
}

// CloudEventClient returns the Server's client for CloudEvents.
func (s *Server) CloudEventClient() cloudevents.Client {
	return s.ceclient
}

// Shutdown will call the cancel function for the server if it is already listening and serving.
func (s *Server) Shutdown() {
	if s.shutdown != nil {
		s.shutdown()
	}
}

// ListenAndServe starts serving the HTTP handlers and CloudEvent receiver, blocking until termination.
func (s *Server) ListenAndServe() error {
	ctx, shutdown := context.WithCancel(context.Background())
	s.shutdown = shutdown
	return s.ceclient.StartReceiver(ctx, s.cehandler)
}
