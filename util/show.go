package util

import (
	"github.com/Axway/ace-golang-sdk/rpc"
	"github.com/Axway/ace-golang-sdk/util/logging"
	"go.uber.org/zap"
)

var zlog = logging.Logger()

// Show - display entire ACE message using 'log' package
func Show(descr string, aceMsg *rpc.Message) {
	zlog.Info(descr,
		zap.String(logging.LogFieldChoreographyID, aceMsg.GetCHN_UUID()),
		zap.String(logging.LogFieldExecutionID, aceMsg.GetCHX_UUID()),
		zap.String(logging.LogFieldMessageID, aceMsg.GetUUID()),
		zap.String(logging.LogFieldParentMessageID, aceMsg.GetParent_UUID()),
		zap.Uint64(logging.LogFieldSequenceTerm, aceMsg.GetSequenceTerm()),
		zap.Uint64(logging.LogFieldSequenceUpperBound, aceMsg.GetSequenceUpperBound()),
		zap.String(logging.LogFieldErrorType, rpc.Message_ErrorType_name[int32(aceMsg.GetErrorType())]),
		zap.String(logging.LogFieldError, aceMsg.GetErrorDescription()),
	)
}
