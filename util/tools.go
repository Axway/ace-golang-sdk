package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util/logging"
	"go.uber.org/zap"
)

var log = logging.Logger()

// CreateSignalChannel - Create a signal channel
func CreateSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	return signalChannel
}

// CopyStringsMapTo - copies source map to dest which has to be rpc.Message using predefined mapping of keys
// to message fields
// PLEASE NOTE: if mapping of key is not present, the key will be ignored and warning logged
func CopyStringsMapTo(source map[string]string, target *rpc.Message) {
	for key, val := range source {
		switch key {
		case "ID":
			target.ID = val

		default:
			log.Warn("util.CopyStringsMapTo unknown mapping of key to rpc.Message field",
				zap.String("key", key),
				zap.String("val", val),
			)
		}
	}
}

//CopyMessage - shallow copy of source message; DOES NOT set UUID or Parent_UUID
func CopyMessage(source *rpc.Message) *rpc.Message {
	msg := rpc.Message{
		CHN_UUID:           source.GetCHN_UUID(),
		CHX_UUID:           source.GetCHX_UUID(),
		ID:                 source.GetID(),
		TopicName:          source.GetTopicName(),
		OpentracingContext: source.GetOpentracingContext(),

		//SequenceTerm:       seqTerm, TODO: sidecar will need to do that when Send indicates io.EOF
		//SequenceUpperBound: seqUpperBound,
	}
	if source.HasProcessingError {
		msg.HasProcessingError = true
		msg.ProcessingErrorDescription = source.GetProcessingErrorDescription()
	}
	if source.HasSystemError {
		msg.HasSystemError = true
		msg.SystemErrorDescription = source.GetSystemErrorDescription()
	}

	return &msg
}
