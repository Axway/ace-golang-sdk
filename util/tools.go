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

// CopyStringsMap - returns copy of the map passed as parameter
func CopyStringsMap(m map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range m {
		result[key] = value
	}

	return result
}

// CopyStringsMapTo - copies source map to dest which has to be aceMsg in order to test for uninitialized MetaData
// PLEASE NOTE: if key is present in target metadata, it will be overwritten
func CopyStringsMapTo(source map[string]string, aceMsgTarget *rpc.Message) {
	if aceMsgTarget.MetaData == nil {
		aceMsgTarget.MetaData = make(map[string]string)
	}
	for key, val := range source {
		oldVal, exists := aceMsgTarget.MetaData[key]
		if exists {
			log.Sugar().Debugf("CopyStringsMapTo will overwrite existing key/value pair: %s=%s with new value: %s\n", key, oldVal, val)
		}
		aceMsgTarget.MetaData[key] = val
	}
}
