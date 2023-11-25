package greeter

import (
	"context"
	"time"

	"github.com/nokia/srlinux-ndk-go/ndk"
	"github.com/openconfig/gnmic/pkg/api"
	"github.com/openconfig/gnmic/pkg/target"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	ndkSocket            = "localhost:50053"
	grpcServerUnixSocket = "unix:///opt/srlinux/var/run/sr_gnmi_server"
	AppName              = "greeter"
)

type App struct {
	Name  string // Agent name
	AppID uint32

	// ConfigState holds the application configuration and state.
	ConfigState *ConfigState

	gRPCConn     *grpc.ClientConn
	logger       *zerolog.Logger
	retryTimeout time.Duration

	target *target.Target

	// NDK Service clients
	SDKMgrServiceClient       ndk.SdkMgrServiceClient
	NotificationServiceClient ndk.SdkNotificationServiceClient
	TelemetryServiceClient    ndk.SdkMgrTelemetryServiceClient
}

// NewApp creates a new Greeter App instance and connects to NDK socket.
// It also creates the NDK service clients and registers the agent with NDK.
func NewApp(ctx context.Context, logger *zerolog.Logger) *App {
	// connect to NDK socket
	conn, err := connect(ctx, ndkSocket)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("gRPC connect failed")
	}

	// create SDK Manager Client
	sdkMgrClient := ndk.NewSdkMgrServiceClient(conn)
	// create Notification Service Client
	notifSvcClient := ndk.NewSdkNotificationServiceClient(conn)
	// create Telemetry Service Client
	telemetrySvcClient := ndk.NewSdkMgrTelemetryServiceClient(conn)

	logger.Info().Msg("creating gNMI Client")
	target, err := newGNMITarget(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("gNMI target creation failed")
	}

	// register agent
	// http://learn.srlinux.dev/ndk/guide/dev/go/#register-the-agent-with-the-ndk-manager
	r, err := sdkMgrClient.AgentRegister(ctx, &ndk.AgentRegistrationRequest{})
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("Agent registration failed")
	}

	logger.Info().
		Uint32("app-id", r.GetAppId()).
		Str("name", AppName).
		Msg("Application registered successfully!")

	return &App{
		Name:  AppName,
		AppID: r.GetAppId(),

		ConfigState: &ConfigState{},

		logger:       logger,
		retryTimeout: 5 * time.Second,
		gRPCConn:     conn,

		target: target,

		SDKMgrServiceClient:       sdkMgrClient,
		NotificationServiceClient: notifSvcClient,
		TelemetryServiceClient:    telemetrySvcClient,
	}
}

func (a *App) Start(ctx context.Context) {
	configStream := a.StartConfigNotificationStream(ctx)
	for {
		select {
		case cfgStreamResp := <-configStream:
			b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(cfgStreamResp)
			if err != nil {
				a.logger.Info().
					Msgf("Config notification Marshal failed: %+v", err)
				continue
			}

			a.logger.Info().
				Msgf("Received notifications:\n%s", b)

			a.handleConfigNotifications(ctx, cfgStreamResp)

		case <-ctx.Done():
			a.logger.Info().Msg("Got a signal to exit, bye!")
			return
		}
	}
}

// connect attempts connecting to the NDK socket with backoff and retry.
// https://learn.srlinux.dev/ndk/guide/dev/go/#establish-grpc-channel-with-ndk-manager-and-instantiate-an-ndk-client
func connect(ctx context.Context, socket string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(ndkSocket,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	return conn, err
}

func newGNMITarget(ctx context.Context) (*target.Target, error) {
	// create a target
	tg, err := api.NewTarget(
		api.Name("srl"),
		api.Address(grpcServerUnixSocket),
		api.Username("admin"),
		api.Password("NokiaSrl1!"),
		api.Insecure(true),
	)
	if err != nil {
		return nil, err
	}

	// create a gNMI client
	err = tg.CreateGNMIClient(ctx)
	if err != nil {
		return nil, err
	}

	return tg, nil
}

func (a *App) getUptime(ctx context.Context) (string, error) {
	// create a GetRequest
	getReq, err := api.NewGetRequest(
		api.Path("/system/information/last-booted"),
		api.Encoding("proto"))
	if err != nil {
		return "", err
	}

	getResp, err := a.target.Get(ctx, getReq)
	if err != nil {
		return "", err
	}

	a.logger.Info().Msgf("GetResponse: %+v", getResp)

	return getResp.Notification[0].Update[0].GetVal().GetStringVal(), nil
}
