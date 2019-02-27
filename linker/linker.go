package linker

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/Axway/ace-golang-sdk/config"
	"github.com/Axway/ace-golang-sdk/linker/rpcclient"
	"github.com/Axway/ace-golang-sdk/linker/rpcserver"
	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util"
	"github.com/Axway/ace-golang-sdk/util/logging"
	"github.com/Axway/ace-golang-sdk/util/tracing"
)

// AceLinkerConfig - represents configuration attributes for linker
type AceLinkerConfig struct {
	ServerHost         string `cfg:"SERVER_HOST" cfg_default:"0.0.0.0"`
	ServerPort         uint16 `cfg:"SERVER_PORT" cfg_default:"50006"`
	SidecarHost        string `cfg:"SIDECAR_HOST" cfg_default:"0.0.0.0"`
	SidecarPort        uint16 `cfg:"SIDECAR_PORT" cfg_default:"50005"`
	ServiceName        string `cfg:"SERVICE_NAME"`
	ServiceVersion     string `cfg:"SERVICE_VERSION"`
	ServiceDescription string `cfg:"SERVICE_DESCRIPTION"`
}

// Link combines what is specific to Linker functionality
type Link struct {
	cfg          AceLinkerConfig
	name         string
	version      string
	description  string
	MsgProcessor interface{}
}

var link *Link
var traceLogging tracing.TraceLogging
var log = logging.Logger()

// MsgProducer - what is exposed to client business function
type MsgProducer interface {
	Send(context.Context, *messaging.BusinessMessage) error
}

// BusinessMessageProcessor type of 'business' function used to process paylod relayed to Linker
type BusinessMessageProcessor func(context.Context, *messaging.BusinessMessage, MsgProducer) error

// Register -Registers the business service with linker
func Register(name, version, description string, fn BusinessMessageProcessor) (*Link, error) {
	if link != nil {
		return link, fmt.Errorf("Service registration already initialized")
	}

	link = &Link{}
	var cfg AceLinkerConfig
	config.ReadConfigFromEnv(&cfg)

	link.cfg = cfg
	link.name = name
	link.version = version
	link.description = description
	link.MsgProcessor = fn

	if len(link.name) == 0 {
		if len(link.cfg.ServiceName) == 0 {
			return nil, fmt.Errorf("Incomplete registration. Provide service name argument or setup SERVICE_NAME environment variable")
		}
		link.name = cfg.ServiceName
	}

	if len(link.version) == 0 {
		if len(link.cfg.ServiceVersion) == 0 {
			return nil, fmt.Errorf("Incomplete registration. Provide service version argument or setup SERVICE_VERSION environment variable")
		}
		link.version = cfg.ServiceVersion
	}

	if len(link.description) == 0 {
		link.description = cfg.ServiceDescription
	}

	if fn == nil {
		return nil, fmt.Errorf("Incomplete registration. Provide the business service callback method")
	}

	display := fmt.Sprintf("%s_%s", cfg.ServiceName, cfg.ServiceVersion)
	traceLogging = tracing.InitTracing(display)

	return link, nil
}

var (
	sidecarHost string
	sidecarPort uint16
)

// OnRelay Server method implements linker's role in message processing; it applies business function fn
// specified at Linker's creation
func (link Link) OnRelay(aceMessage *rpc.Message) {
	util.Show("linker.onRelay msg:", aceMessage)

	ctxWithSpan := tracing.IssueTrace(aceMessage.GetOpentracingContext(), "agent message receive", aceMessage.UUID, aceMessage.Parent_UUID)

	switch msgProcessor := link.MsgProcessor.(type) {
	case BusinessMessageProcessor:
		clientRelay, buildErr := rpcclient.BuildClientRelay(ctxWithSpan, aceMessage, sidecarHost, sidecarPort)
		if buildErr != nil {
			log.Fatal("agent unable to BuildClientRelay", zap.Error(buildErr))
			return
		}
		defer func() {
			clientRelay.CloseSend(ctxWithSpan)
		}()
		err := msgProcessor(ctxWithSpan, aceMessage.GetBusinessMessage(), clientRelay)
		if err != nil {
			tracing.IssueErrorTrace(
				aceMessage.GetOpentracingContext(),
				err,
				"error processing business message",
				aceMessage.UUID, aceMessage.Parent_UUID)

			switch error := err.(type) {
			case rpcclient.SendingError: // log it as we don't want to send again
				log.Error("SendingError", zap.Error(fmt.Errorf(error.Error())))
			default:
				clientRelay.SendWithError(ctxWithSpan, error)
				log.Error("message processor error", zap.Error(err))
			}
		}
	default:
		panic(fmt.Sprintf("MsgProcessor of %s agent is not of expected BusinessMessageProcessor type, actual: %T\n",
			link.name, msgProcessor))
	}

}

// OnSidecarRegistrationComplete can perform Linker-specific post registration actions
func (link Link) OnSidecarRegistrationComplete(serviceInfo *rpc.ServiceInfo) error {
	log.Info("linker OnSidecarRegistrationComplete",
		zap.String("sidecar ID", serviceInfo.GetServiceName()),
	)

	if ok, err := link.registerWithSidecar(); !ok {
		log.Fatal("unable to register agent in response to sidecar registration",
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Start linker server first to make sure there is something to process relay'ed messages, then
// call linkage client code to register itself with sidecar as ready to begin processing
func (link Link) Start() {

	sidecarHost = link.cfg.SidecarHost
	sidecarPort = link.cfg.SidecarPort

	var server = rpcserver.Server{
		OnRelay:                link.OnRelay,
		OnRelayComplete:        link.onRelayComplete,
		OnRegistrationComplete: link.OnSidecarRegistrationComplete,
		Name:                   fmt.Sprintf("%s-%s", link.name, link.version),
	}

	log.Info("Starting linker and registering it with sidecar",
		zap.String("link.name", link.name),
		zap.String("link.version", link.version),
		zap.String("link.cfg.ServerHost", link.cfg.ServerHost),
		zap.Uint16("link.cfg.ServerPort", link.cfg.ServerPort),
		zap.String("sidecar.host", sidecarHost),
		zap.Uint16("sidecar.port", sidecarPort),
	)

	waitc := make(chan struct{})
	go func() {
		rpcserver.StartServer(link.cfg.ServerHost, link.cfg.ServerPort, &server)
		close(waitc)
	}()

	if ok, err := link.registerWithSidecar(); !ok {
		log.Fatal("unable to Start agent",
			zap.Error(err),
		)
	}

	<-waitc

	traceLogging.Close()
}

const timeFormat = "15:04:05.9999"

// in linker, onRelayComplete does not need to be different from onRelay
func (link Link) onRelayComplete(msg *rpc.Message) {
	link.OnRelay(msg)
}

func (link Link) registerWithSidecar() (bool, error) {
	serviceInfo := rpc.ServiceInfo{
		ServiceName:        link.name,
		ServiceVersion:     link.version,
		ServiceDescription: link.description,
		ServiceHost:        link.cfg.ServerHost,
		ServicePort:        uint32(link.cfg.ServerPort),
	}
	if ok, err := rpcclient.ClientRegister(sidecarHost, sidecarPort, &serviceInfo); !ok {
		log.Error("ClientRegistration of agent error",
			zap.Error(err),
		)
		return false, err
	}
	return true, nil
}

// NewProcessingError - convenience function to create ProcessingError
func NewProcessingError(err error) rpcclient.ProcessingError {
	return rpcclient.ProcessingError{
		ErrorInfo: rpcclient.MsgErrorInfo{
			ErrDescription: err.Error(),
		},
	}
}

// NewSystemError - convenience function to create SystemError
func NewSystemError(err error) rpcclient.SystemError {
	return rpcclient.SystemError{
		ErrorInfo: rpcclient.MsgErrorInfo{
			ErrDescription: err.Error(),
		},
	}
}
