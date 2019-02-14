package rpcserver

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util"
	"github.com/Axway/ace-golang-sdk/util/logging"
	"github.com/Axway/ace-golang-sdk/util/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var log = logging.Logger()

// OnRelayHandler defines what will be called in Relay method when payload is received; different implementation
// is used by Linker and by Sidecar
type OnRelayHandler func(ctx context.Context, msg *rpc.Message)

// OnRelayCompleteHandler -
type OnRelayCompleteHandler func(ctx context.Context, msg *rpc.Message)

// OnRelayCompleteErrorHandler -
type OnRelayCompleteErrorHandler func(*rpc.Message, error)

// OnRegistrationCompleteHandler handles registration info in application specific way
type OnRegistrationCompleteHandler func(serviceInfo *rpc.ServiceInfo) error

// Server implements LinkageService interface methods
type Server struct {
	OnRelay                OnRelayHandler
	OnRelayComplete        OnRelayCompleteHandler
	OnRelayCompleteError   OnRelayCompleteErrorHandler
	OnRegistrationComplete OnRegistrationCompleteHandler
	Name                   string
}

// Registration implements LinkageService interface
func (s *Server) Registration(ctx context.Context, serviceInfo *rpc.ServiceInfo) (*rpc.Receipt, error) {
	log.Info("Registering client service",
		zap.String("service.name", serviceInfo.GetServiceName()),
	)

	span, _ := tracing.StartTraceFromContext(ctx, "Registration")
	span.LogStringField("event", "Registration")
	span.LogStringField("ServiceName", serviceInfo.ServiceName)
	defer span.Finish()

	err := s.OnRegistrationComplete(serviceInfo)

	return &rpc.Receipt{IsOk: err == nil}, err
}

// Relay implements LinkageService interface
func (s *Server) Relay(stream rpc.Linkage_RelayServer) error {
	// receive data from stream
	log.Debug(fmt.Sprintf("Relay [%s]: starting Recv on stream", s.Name),
		zap.String("service.name", s.Name),
	)

	var msgCount uint64
	var last *rpc.Message
	for {
		in, recErr := stream.Recv()

		if recErr == io.EOF {
			//done with Recv, so if recErr == nil, sidecar will need to commit
			log.Info(fmt.Sprintf("Relay [%s]: Done receiving total of %d messages, calling OnRelayComplete\n", s.Name, msgCount),
				zap.String("service.name", s.Name),
				zap.Uint64("msg.count", msgCount),
			)
			if last != nil {
				span, ctxWithSpan := spanFromMetadataOrNew(last.GetMetaData(), "Receiving")
				span.LogStringField("event", "end of message receive")
				span.LogIntField("total message count", int(msgCount))
				span.LogStringField("message.UUID", last.UUID)
				span.LogStringField("message.Parent_UUID", last.Parent_UUID)
				span.Finish()

				last.SequenceUpperBound = msgCount

				s.OnRelayComplete(ctxWithSpan, last)
			} else {
				log.Fatal(fmt.Sprintf("Relay [%s]: at io.EOF, it is expected to have at least one message:%v\n", s.Name, last))
			}
			return nil
		} else if recErr != nil {
			log.Error("Relay error",
				zap.String("error", recErr.Error()),
			)
			s.OnRelayCompleteError(last, recErr)
			return recErr
		}

		var receipt = &rpc.Receipt{IsOk: recErr == nil}
		if err := stream.Send(receipt); err != nil {
			log.Error("Relay: error sending receipt",
				zap.String("service.name", s.Name),
				zap.String("error", err.Error()),
			)
		}
		log.Debug("Relay: Done sending receipt",
			zap.String("service.name", s.Name),
		)

		if recErr == nil {
			msg := in
			msgCount++
			msg.SequenceTerm = msgCount
			log.Debug("Relay: received message",
				zap.String("service.name", s.Name),
				zap.String("msg.Parent_UUID", msg.Parent_UUID),
				zap.String("msg.UUID", msg.UUID),
				zap.Uint64("msg.count", msgCount),
			)

			if last != nil {
				span, ctxWithSpan := spanFromMetadataOrNew(last.GetMetaData(), "Receiving")
				span.LogStringField("message.UUID", last.UUID)
				span.LogStringField("message.Parent_UUID", last.Parent_UUID)
				span.Finish()

				s.OnRelay(ctxWithSpan, last)
			}
			last = msg
		}
	}
}

// StartServer creates new grpc LinkageService server returns error to prevent registration to take place
func StartServer(host string, port uint16, server *Server) {
	gracefulStop := util.CreateSignalChannel()

	log.Debug("Starting Server",
		zap.String("server.name", server.Name),
		zap.String("server.host", host),
		zap.Uint16("port", port),
	)

	// create listener
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatal("Server failed to listen",
			zap.Error(err),
		)
	}

	options := []grpc.ServerOption{
		grpc.UnaryInterceptor(tracing.GetOpenTracingServerInterceptor()),
	}
	// create grpc server
	grpcServer := grpc.NewServer(options...)

	rpc.RegisterLinkageServer(grpcServer, server)

	go func() {
		sig := <-gracefulStop
		log.Debug("received system signal, shutting down server...", zap.String("signal", sig.String()))
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve",
			zap.Error(err),
		)
	} else {
		log.Info("server stopped")
	}
}

func spanFromMetadataOrNew(msgMetadata map[string]string, msg string) (tracing.Tracer, context.Context) {
	var span tracing.Tracer
	var ctxWithSpan context.Context
	otSpan, ok := msgMetadata[tracing.OpentracingContext]
	if ok {
		span, _ = tracing.Base64ToTrace(otSpan, msg)
		ctxWithSpan, _ = tracing.ContextWithSpan(context.Background(), span)
	} else {
		span, ctxWithSpan = tracing.StartTraceFromContext(context.Background(), msg)
	}
	return span, ctxWithSpan
}
