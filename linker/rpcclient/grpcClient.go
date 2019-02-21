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

var log = logging.Logger()

// ClientRegister makes a grpc call to Registration method on the server represented by host and port inputs
func ClientRegister(host string, port uint16, serviceInfo *rpc.ServiceInfo) (bool, error) {
	hostInfo := fmt.Sprintf("%s:%d", host, port)
	log.Info(fmt.Sprintf("calling Registration on: %s", hostInfo),
		zap.String("event", "ClientRegister"),
		zap.String("service.name", serviceInfo.ServiceName),
	)
	gracefulStop := util.CreateSignalChannel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		sig := <-gracefulStop
		cancel()
		log.Debug("received system signal, cancelling connection...", zap.String("signal", sig.String()))
	}()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(tracing.GetOpenTracingClientInterceptor()),
	}

	conn, err := grpc.DialContext(ctx, hostInfo, opts...)
	if err != nil {
		log.Error("cannot connect to registration server",
			zap.String("error", err.Error()),
		)
		return false, err
	}
	defer conn.Close()

	client := rpc.NewLinkageClient(conn)

	span, ctxWithSpan := tracing.StartTraceFromContext(ctx, "ClientRegister")
	span.LogStringField("event", "start client Registration")
	span.LogStringField("ServiceName", serviceInfo.ServiceName)
	defer span.Finish()

	_, err = client.Registration(ctxWithSpan, serviceInfo)
	if err != nil {
		log.Fatal("Error in client.Registration",
			zap.String("error", err.Error()),
		)
		return false, err
	}
	log.Debug("received Receipt for registering",
		zap.String("service.name", serviceInfo.ServiceName),
	)
	return true, nil
}

// ClientRelay makes a grpc call to Relay method on the server represented by host and port inputs
func ClientRelay(clientContext context.Context, msg *rpc.Message, host string, port uint16) (bool, error) {
	var hostInfo = fmt.Sprintf("%s:%d", host, port)

	if spanAsBase64, err := tracing.ContextWithSpanToBase64(clientContext); err == nil {
		msg.MetaData[tracing.OpentracingContext] = spanAsBase64
	} else {
		log.Error("error encoding tracing context", zap.Error(err))
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(hostInfo, opts...)
	if err != nil {
		log.Fatal("cannot connect to relay server",
			zap.String("host.info", hostInfo),
			zap.String("error", err.Error()),
		)
		return false, err
	}
	defer conn.Close()

	client := rpc.NewLinkageClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.Relay(ctx)
	if err != nil {
		log.Error("error attempting to open stream to client.Relay",
			zap.String("error", err.Error()),
		)
		return false, err
	}

	waitc := make(chan struct{})
	go func() {
		log.Debug("ClientRelay: opening stream to receive Relay receipt")
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				log.Debug("ClientRelay, done receiving, EOF")
				close(waitc)
				return
			}
			if err != nil {
				log.Error("ClientRelay, error in stream.Recv",
					zap.Error(err),
				)
				return
			}
			log.Debug("ClientRelay, got payload receipt")
		}
	}()

	if err := stream.Send(msg); err != nil {
		log.Error("error sending",
			zap.Error(err),
		)
		return false, err
	}
	log.Debug("sent message, closing stream",
		zap.String("msg.UUID", msg.UUID),
	)
	stream.CloseSend()

	log.Debug("waiting to receive receipt")

	<-waitc

	log.Debug("exiting client relay")
	return true, nil
}

// LinkerClientRelay -
type LinkerClientRelay struct {
	sourceMessage *rpc.Message
	deferrables   []func()
	stream        rpc.Linkage_RelayClient
}

// ClientRelayHousekeeper - contains methods NOT exposed to the client business function
type ClientRelayHousekeeper interface {
	configure(ctx context.Context, host string, port uint16) error

	CloseSend(context.Context)
}

// BuildClientRelay -
func BuildClientRelay(clientContext context.Context, aceMsg *rpc.Message, host string, port uint16) (*LinkerClientRelay, error) {
	result := LinkerClientRelay{
		sourceMessage: aceMsg,
		deferrables:   make([]func(), 0),
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
		log.Error("cannot connect to server",
			zap.String("host.info", hostInfo),
			zap.String("error", err.Error()),
		)
		return err
	}
	closeConnFunc := func() { conn.Close() }
	lcr.deferrables = append(lcr.deferrables, closeConnFunc)
	//replace this by closure func
	//defer conn.Close()

	client := rpc.NewLinkageClient(conn)

	//ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithCancel(clientContext)
	cancelFunc := func() {
		cancel()
		log.Debug("cancelled ClientRelay receive context")
	}
	lcr.deferrables = append(lcr.deferrables, cancelFunc)
	//defer cancelFunc()

	var clientRelayErr error
	lcr.stream, clientRelayErr = client.Relay(ctx)

	if clientRelayErr != nil {
		log.Error("error attempting to open stream to client.Relay",
			zap.String("error", clientRelayErr.Error()),
		)
		return clientRelayErr
	}
	return nil
}

// Send -
func (lcr *LinkerClientRelay) Send(ctx context.Context, bm *messaging.BusinessMessage) error {
	//combine with sourceMessage
	msg := buildResult(lcr.sourceMessage, bm)

	if b64, err := tracing.ContextWithSpanToBase64(ctx); err == nil {
		msg.MetaData[tracing.OpentracingContext] = b64
	} else { //log it and continue
		log.Error("error encoding tracing context", zap.Error(err))
	}

	if err := lcr.stream.Send(msg); err != nil {
		log.Error("error sending", zap.String("error", err.Error()))

		return NewSendingError(err)
	}
	util.Show("LinkerClientRelay sent msg:\n", msg)
	return nil
}

// SendWithError -
func (lcr *LinkerClientRelay) SendWithError(ctx context.Context, err error) error {
	var msg *rpc.Message

	switch error := err.(type) {
	case ProcessingError:
		// buildResult makes a copy of lcr.sourceMessage and assigns to Parent_UUID value of sourceMessage.UUID
		// so if the message is to be committed, which in case of ProcessingError, it will be, we want to call
		// buildResult to make a copy;
		msg := buildResult(lcr.sourceMessage, nil)
		msg.MetaData["TODO-key for ProcessingError"] = error.Error()
	default:
		// treat everything else as SystemError in case the 'business function' did not catch & convert whatever error they encountered to ours;
		// so if it is a SystemError then no commit so use sourceMessage to send back
		// it will not be a SendingError because that was already filtered out in linker.OnRelay
		msg := lcr.sourceMessage
		//TODO: add to metadata about SystemError
		msg.MetaData["TODO-key for SystemError"] = error.Error()
	}

	if b64, err := tracing.ContextWithSpanToBase64(ctx); err == nil {
		msg.MetaData[tracing.OpentracingContext] = b64
	} else {
		log.Error("error encoding tracing context", zap.Error(err))
	}

	if err := lcr.stream.Send(msg); err != nil {
		log.Error("error sending",
			zap.String("error", err.Error()),
		)
		return err
	}
	util.Show("LinkerClientRelay sent msg:\n", msg)
	return nil
}

// CloseSend -
//
func (lcr *LinkerClientRelay) CloseSend(ctx context.Context) {
	waitc := make(chan struct{})
	go func() {
		log.Debug("LinkerClientRelay: openining stream to receive Relay receipt(s)")
		for {
			_, err := lcr.stream.Recv()
			if err == io.EOF {
				log.Debug("LinkerClientRelay, done receiving receipt(s), EOF")
				close(waitc)
				return
			}
			if err != nil {
				log.Error("LinkerClientRelay, fatal error in stream.Recv",
					zap.String("error", err.Error()),
				)
				return
			}
			log.Debug("LinkerClientRelay, got payload receipt")
		}
	}()
	lcr.stream.CloseSend()

	<-waitc
	//wait for receipt func to complete
	log.Debug("LinkerClientRelay.CloseSend completed")
}

func buildResult(parentMsg *rpc.Message, bm *messaging.BusinessMessage) *rpc.Message {
	//copy needed attributes
	// then set/change to reflect it's a child
	copyStepPattern := rpc.StepPattern{}
	copyPattern(&copyStepPattern, parentMsg.Pattern)

	msg := rpc.Message{
		Parent_UUID: parentMsg.GetUUID(),
		CHN_UUID:    parentMsg.GetCHN_UUID(),
		CHX_UUID:    parentMsg.GetCHX_UUID(),
		Pattern:     &copyStepPattern,

		// UUID:            uuid.New().String(), do not set UUID as it will be set by kafka producer before placed in queue
		BusinessMessage: bm,
		//SequenceTerm:       seqTerm, TODO: sidecar will need to do that when Send indicates io.EOF
		//SequenceUpperBound: seqUpperBound,
	}

	msg.MetaData = util.CopyStringsMap(parentMsg.GetMetaData())

	return &msg
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
