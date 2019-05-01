package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util/logging"
)

var log = logging.Logger()

// CreateSignalChannel - Create a signal channel
func CreateSignalChannel() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	return signalChannel
}

//CopyMessage - shallow copy of source message; DOES NOT set UUID or Parent_UUID
func CopyMessage(source *rpc.Message) *rpc.Message {
	msg := rpc.Message{
		CHN_UUID:           source.GetCHN_UUID(),
		CHX_UUID:           source.GetCHX_UUID(),
		Consumption_ID:     source.GetConsumption_ID(),
		TopicName:          source.GetTopicName(),
		OpentracingContext: source.GetOpentracingContext(),

		//SequenceTerm:       seqTerm, TODO: sidecar will need to do that when Send indicates io.EOF
		//SequenceUpperBound: seqUpperBound,
	}
	if source.ErrorType != rpc.Message_NONE {
		msg.ErrorType = source.GetErrorType()
		msg.ErrorDescription = source.GetErrorDescription()
	}

	// Copy over the MetaData
	msg.MetaData = make(map[string]string)
	for key, val := range source.MetaData {
		msg.MetaData[key] = val
	}

	return &msg
}
