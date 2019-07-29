package linker

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/Axway/ace-golang-sdk/config"
	"github.com/Axway/ace-golang-sdk/linker/rpcclient"
	"github.com/Axway/ace-golang-sdk/linker/rpcserver"
	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
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
	ServiceType        string `cfg:"SERVICE_TYPE"`
}

// RPCClient .
type RPCClient interface {
	BuildClientRelay(clientContext context.Context, aceMsg *rpc.Message, host string, port uint16) (*rpcclient.LinkerClientRelay, error)
	ClientRegister(host string, port uint16, serviceInfo *rpc.ServiceInfo) (bool, error)
}

type rpcClient struct {
}

// rpcClient implements RPCClient interface
func (c *rpcClient) BuildClientRelay(clientContext context.Context, aceMsg *rpc.Message, host string, port uint16) (*rpcclient.LinkerClientRelay, error) {
	return rpcclient.BuildClientRelay(clientContext, aceMsg, host, port)
}

// rpcClient implements RPCClient interface
func (c *rpcClient) ClientRegister(host string, port uint16, serviceInfo *rpc.ServiceInfo) (bool, error) {
	return rpcclient.ClientRegister(host, port, serviceInfo)
}

// Link combines what is specific to Linker functionality
type Link struct {
	cfg          AceLinkerConfig
	name         string
	version      string
	description  string
	serviceType  string
	MsgProcessor interface{}
	client       RPCClient
}

var link *Link
var traceLogging tracing.TraceLogging
var log = logging.Logger()
var serviceConfigParamTemplates []*rpc.ConfigParameter

// MsgProducer - what is exposed to client business function
type MsgProducer interface {
	Send(*messaging.BusinessMessage) error
}

// BusinessMessageProcessor type of 'business' function used to process paylod relayed to Linker
type BusinessMessageProcessor func(context.Context, []*messaging.BusinessMessage, MsgProducer) error

// Add config parameter to list
func addConfigParam(configParam *rpc.ConfigParameter) {
	serviceConfigParamTemplates = append(serviceConfigParamTemplates, configParam)
}

// AddStringConfigParam - Add String config parameter for the service
func AddStringConfigParam(name, defaultValue string, required bool) error {
	stringConfigParam := &rpc.ConfigParameter{
		Name:         name,
		Type:         "string",
		DefaultValue: defaultValue,
		IsRequired:   required,
	}
	addConfigParam(stringConfigParam)
	return nil
}

// AddIntConfigParam - Add integer config parameter for the service
func AddIntConfigParam(name string, defaultValue int, required bool) error {
	intConfigParam := &rpc.ConfigParameter{
		Name:         name,
		Type:         "int",
		DefaultValue: strconv.Itoa(defaultValue),
		IsRequired:   required,
	}
	addConfigParam(intConfigParam)
	return nil
}

// AddBooleanConfigParam - Add boolean config parameter for the service
func AddBooleanConfigParam(name string, defaultValue bool) error {
	intConfigParam := &rpc.ConfigParameter{
		Name:         name,
		Type:         "boolean",
		DefaultValue: strconv.FormatBool(defaultValue),
		IsRequired:   true,
	}
	addConfigParam(intConfigParam)
	return nil
}

// Register -Registers the business service with linker
func Register(name, version, description, serviceType string, fn BusinessMessageProcessor) (*Link, error) {
	if link != nil {
		return link, fmt.Errorf("Service registration already initialized")
	}

	link = &Link{
		client: &rpcClient{},
	}
	var cfg AceLinkerConfig
	config.ReadConfigFromEnv(&cfg)

	link.cfg = cfg
	link.name = name
	link.version = version
	link.description = description
	link.serviceType = serviceType
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
	ctxWithSpan := tracing.IssueTrace(aceMessage, "Agent message receive")

	switch msgProcessor := link.MsgProcessor.(type) {
	case BusinessMessageProcessor:
		clientRelay, buildErr := link.client.BuildClientRelay(ctxWithSpan, aceMessage, sidecarHost, sidecarPort)
		if buildErr != nil {
			log.Fatal("Unable to initialize relay client", zap.String(logging.LogFieldError, buildErr.Error()))
			return
		}
		log.Debug("Relay client created, executing service callback to process message",
			zap.String(logging.LogFieldChoreographyID, aceMessage.GetCHN_UUID()),
			zap.String(logging.LogFieldExecutionID, aceMessage.GetCHX_UUID()),
			zap.String(logging.LogFieldMessageID, aceMessage.GetUUID()),
			zap.String(logging.LogFieldParentMessageID, aceMessage.GetParent_UUID()))

		defer func() {
			log.Info("Service callback execution completed",
				zap.String(logging.LogFieldChoreographyID, aceMessage.GetCHN_UUID()),
				zap.String(logging.LogFieldExecutionID, aceMessage.GetCHX_UUID()),
				zap.String(logging.LogFieldMessageID, aceMessage.GetUUID()),
				zap.String(logging.LogFieldParentMessageID, aceMessage.GetParent_UUID()))
			clientRelay.CloseSend()
		}()

		err := msgProcessor(ctxWithSpan, aceMessage.GetBusinessMessage(), clientRelay)
		if err != nil {
			tracing.IssueErrorTrace(
				aceMessage,
				err,
				"Error processing business message")

			switch error := err.(type) {
			case rpcclient.SendingError: // log it as we don't want to send again
				log.Error("SendingError", zap.String(logging.LogFieldError, error.Error()))
			default:
				clientRelay.SendWithError(error)
				log.Error("Error in message processing", zap.String(logging.LogFieldError, err.Error()))
			}
		}
	default:
		panic(fmt.Sprintf("MsgProcessor of %s agent is not of expected BusinessMessageProcessor type, actual: %T\n",
			link.name, msgProcessor))
	}
}

// OnSidecarRegistrationComplete can perform Linker-specific post registration actions
func (link Link) OnSidecarRegistrationComplete(serviceInfo *rpc.ServiceInfo) error {
	log.Info("Processing sidecar registration",
		zap.String(logging.LogFieldSidecarID, serviceInfo.GetServiceName()),
	)

	if ok, err := link.registerWithSidecar(); !ok {
		log.Fatal("Unable to register agent in response to sidecar registration",
			zap.String(logging.LogFieldError, err.Error()),
		)
		return err
	}

	log.Info("Sidecar registration completed",
		zap.String(logging.LogFieldSidecarID, serviceInfo.GetServiceName()),
	)
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

	log.Info("Starting agent and registering it with sidecar",
		zap.String(logging.LogFieldServiceName, link.name),
		zap.String(logging.LogFieldServiceVersion, link.version),
		zap.String(logging.LogFieldServiceHost, link.cfg.ServerHost),
		zap.Uint16(logging.LogFieldServicePort, link.cfg.ServerPort),
		zap.String(logging.LogFieldSidecarHost, sidecarHost),
		zap.Uint16(logging.LogFieldSidecarPort, sidecarPort),
	)

	waitc := make(chan struct{})
	go func() {
		rpcserver.StartServer(link.cfg.ServerHost, link.cfg.ServerPort, &server)
		close(waitc)
	}()

	if ok, err := link.registerWithSidecar(); !ok {
		log.Fatal("Unable to start agent",
			zap.String(logging.LogFieldError, err.Error()),
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
		ServiceName:           link.name,
		ServiceVersion:        link.version,
		ServiceDescription:    link.description,
		ServiceType:           rpc.ServiceInfo_ServiceType(rpc.ServiceInfo_ServiceType_value[link.serviceType]),
		ServiceHost:           link.cfg.ServerHost,
		ServicePort:           uint32(link.cfg.ServerPort),
		ServiceConfigTemplate: serviceConfigParamTemplates,
	}

	log.Info("Registering with sidecar",
		zap.String(logging.LogFieldEvent, "Client register"),
		zap.String(logging.LogFieldSidecarHost, sidecarHost),
		zap.Uint16(logging.LogFieldSidecarPort, sidecarPort),
		zap.String(logging.LogFieldServiceName, link.name),
		zap.String(logging.LogFieldServiceVersion, link.version))

	if ok, err := link.client.ClientRegister(sidecarHost, sidecarPort, &serviceInfo); !ok {
		log.Error("Error in registering agent",
			zap.String(logging.LogFieldError, err.Error()),
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
