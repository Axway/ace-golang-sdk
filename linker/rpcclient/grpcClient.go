package rpcclient

import (
	"context"
	"fmt"
	"io"

	"time"

	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util"
	"github.com/Axway/ace-golang-sdk/util/logging"
	"github.com/Axway/ace-golang-sdk/util/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Constants
const (
	ErrMsgSending               = "Error in sending message"
	ErrMsgOpenClientRelayStream = "Error attempting to open stream for client relay"
)

var log = logging.Logger()

// ClientRegister makes a grpc call to Registration method on the server represented by host and port inputs
func ClientRegister(host string, port uint16, serviceInfo *rpc.ServiceInfo) (bool, error) {
	hostInfo := fmt.Sprintf("%s:%d", host, port)
	gracefulStop := util.CreateSignalChannel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	go func() {
		<-gracefulStop
		cancel()
		log.Debug("Received system signal, cancelling connection")
	}()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(tracing.GetOpenTracingClientInterceptor()),
	}

	conn, err := grpc.DialContext(ctx, hostInfo, opts...)
	if err != nil {
		log.Error("Error in connecting to registration server",
			zap.String(logging.LogFieldError, err.Error()),
		)
		return false, err
	}
	defer conn.Close()

	client := rpc.NewLinkageClient(conn)

	_, err = client.Registration(context.Background(), serviceInfo)
	if err != nil {
		log.Fatal("Error in client registration",
			zap.String(logging.LogFieldError, err.Error()),
		)
		return false, err
	}
	log.Debug("Received acknowledgment for registration",
		zap.String(logging.LogFieldServiceName, serviceInfo.ServiceName),
		zap.String(logging.LogFieldServiceVersion, serviceInfo.ServiceVersion),
	)
	return true, nil
}

// ClientRelay makes a grpc call to Relay method on the server represented by host and port inputs
func ClientRelay(clientContext context.Context, msg *rpc.Message, host string, port uint16) (bool, error) {
	var hostInfo = fmt.Sprintf("%s:%d", host, port)

	assignTraceContext(clientContext, msg)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(hostInfo, opts...)
	if err != nil {
		log.Fatal("Error in connecting to relay server",
			zap.String(logging.LogFieldServiceHost, host),
			zap.Uint16(logging.LogFieldServicePort, port),
			zap.String(logging.LogFieldError, err.Error()),
		)
		return false, err
	}
	defer conn.Close()

	client := rpc.NewLinkageClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.Relay(ctx)
	if err != nil {
		log.Error(ErrMsgOpenClientRelayStream,
			zap.String(logging.LogFieldError, err.Error()),
		)
		return false, err
	}

	waitc := make(chan struct{})
	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Error("Error in receiving client relay receipts",
					zap.String(logging.LogFieldError, err.Error()),
				)
				return
			}
		}
	}()

	if err := stream.Send(msg); err != nil {
		log.Error(ErrMsgSending,
			zap.String(logging.LogFieldError, err.Error()),
		)
		return false, err
	}
	log.Debug("Sent message, closing stream",
		zap.String(logging.LogFieldMessageID, msg.UUID),
	)
	stream.CloseSend()

	<-waitc
	log.Debug("Received client relay receipt")
	return true, nil
}

// LinkerClientRelay -
type LinkerClientRelay struct {
	SourceMessage *rpc.Message
	deferrables   []func()
	Stream        rpc.Linkage_RelayClient
	sentCount     uint64
	traceContext  context.Context
}

// ClientRelayHousekeeper - contains methods NOT exposed to the client business function
type ClientRelayHousekeeper interface {
	configure(ctx context.Context, host string, port uint16) error

	CloseSend(context.Context)
}

// BuildClientRelay - saves clientContext here so opentraces emitted from sidecar after relay from linker will originate from it
func BuildClientRelay(clientContext context.Context, aceMsg *rpc.Message, host string, port uint16) (*LinkerClientRelay, error) {
	result := LinkerClientRelay{
		SourceMessage: aceMsg,
		deferrables:   make([]func(), 0),
		traceContext:  clientContext,
	}
	err := result.configure(clientContext, host, port)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (lcr *LinkerClientRelay) configure(clientContext context.Context, host string, port uint16) error {
	var hostInfo = fmt.Sprintf("%s:%d", host, port)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial(hostInfo, opts...)
	if err != nil {
		log.Error("Error in connecting to server",
			zap.String(logging.LogFieldSidecarHost, host),
			zap.Uint16(logging.LogFieldSidecarPort, port),
			zap.String(logging.LogFieldError, err.Error()),
		)
		return err
	}
	closeConnFunc := func() { conn.Close() }
	lcr.deferrables = append(lcr.deferrables, closeConnFunc)

	client := rpc.NewLinkageClient(conn)

	ctx, cancel := context.WithCancel(clientContext)
	cancelFunc := func() {
		cancel()
	}
	lcr.deferrables = append(lcr.deferrables, cancelFunc)

	var clientRelayErr error
	lcr.Stream, clientRelayErr = client.Relay(ctx)

	if clientRelayErr != nil {
		log.Error(ErrMsgOpenClientRelayStream,
			zap.String(logging.LogFieldError, clientRelayErr.Error()),
		)
		return clientRelayErr
	}
	return nil
}

func assignTraceContext(ctx context.Context, msg *rpc.Message) {
	if ctx != nil {
		if b64, err := tracing.ContextWithSpanToBase64(ctx); err == nil {
			msg.OpentracingContext = b64
		} else {
			log.Error("Error encoding tracing context", zap.String(logging.LogFieldError, err.Error()))
		}
	}
}

// Send -
func (lcr *LinkerClientRelay) Send(bm *messaging.BusinessMessage) error {
	//combine with sourceMessage
	msg := buildResult(lcr.SourceMessage, bm)
	ctx := lcr.traceContext

	assignTraceContext(ctx, msg)

	if err := lcr.Stream.Send(msg); err != nil {
		log.Error(ErrMsgSending, zap.String(logging.LogFieldError, err.Error()))

		return NewSendingError(err)
	}
	lcr.sentCount++
	util.Show("Response message", msg)
	return nil
}

// SendWithError -
func (lcr *LinkerClientRelay) SendWithError(err error) error {
	msg := lcr.SourceMessage
	ctx := lcr.traceContext

	switch error := err.(type) {
	case ProcessingError:
		msg.ErrorType = rpc.Message_PROCESSING
		msg.ErrorDescription = error.Error()
	default:
		msg.ErrorType = rpc.Message_SYSTEM
		msg.ErrorDescription = error.Error()
	}

	assignTraceContext(ctx, msg)

	if err := lcr.Stream.Send(msg); err != nil {
		log.Error(ErrMsgSending,
			zap.String(logging.LogFieldError, err.Error()),
		)
		return err
	}
	lcr.sentCount++
	util.Show("Error response message", msg)
	return nil
}

// sendWithEmptyPayload - send a BusinessMessage with empty payload over grpc to the sidecar server.
func (lcr *LinkerClientRelay) sendWithEmptyPayload() error {
	//combine with sourceMessage
	emptyBusinessMessage := &messaging.BusinessMessage{
		Payload: &messaging.Payload{},
	}
	return lcr.Send(emptyBusinessMessage)
}

// CloseSend -
//
func (lcr *LinkerClientRelay) CloseSend() {
	if lcr.sentCount == 0 {
		// Send empty message for next child service
		if err := lcr.sendWithEmptyPayload(); err != nil {
			log.Error("Error closing stream with no messages",
				zap.String(logging.LogFieldError, err.Error()),
			)
		}
	}

	waitc := make(chan struct{})
	go func() {
		for {
			_, err := lcr.Stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Error("Error in receiving close messages",
					zap.String(logging.LogFieldError, err.Error()),
				)
				return
			}
		}
	}()
	lcr.Stream.CloseSend()

	<-waitc
}

// UUID of the resulting message is not set, it will be set by kafka producer before placing in queue
func buildResult(parentMsg *rpc.Message, bm *messaging.BusinessMessage) *rpc.Message {
	msg := util.CopyMessage(parentMsg)

	copyStepPattern := rpc.StepPattern{}
	copyPattern(&copyStepPattern, parentMsg.Pattern)

	// Copy over the BusinessMessage MetaData, do not overwrite new values
	for _, bmsg := range parentMsg.GetBusinessMessage() {
		for key, value := range bmsg.GetMetaData() {
			if bm.GetMetaData() == nil {
				bm.MetaData = make(map[string]string)
			}
			if _, ok := bm.GetMetaData()[key]; !ok {
				// key was not in map, add it
				bm.MetaData[key] = value
			}
		}
	}

	// then set/change to reflect it's a child
	msg.Parent_UUID = parentMsg.GetUUID()
	msg.Pattern = &copyStepPattern
	msg.BusinessMessage = append(msg.BusinessMessage, bm)
	return msg
}

func copyPattern(target *rpc.StepPattern, source *rpc.StepPattern) {
	target.ServiceName = source.GetServiceName()
	target.ServiceVersion = source.GetServiceVersion()
	target.Validation = source.GetValidation()
	target.Evaluation = source.GetEvaluation()
	target.Transformation = source.GetTransformation()

	target.Child = make([]*rpc.StepPattern, 0)
	for _, sp := range source.Child {
		spCopy := &rpc.StepPattern{}
		copyPattern(spCopy, sp)
		target.Child = append(target.Child, spCopy)
	}
}
