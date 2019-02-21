package rpcclient

import "fmt"

// MsgErrorInfo -
type MsgErrorInfo struct {
	ErrDescription string
	MsgUUID        string
	MsgParentUUID  string
}

// ProcessingError - signals errors encountered when processing payload
type ProcessingError struct {
	ErrorInfo MsgErrorInfo
}

func (e ProcessingError) Error() string {
	return fmt.Sprintf("ProcessingError: %s MsgUUID: %s MstParentUUID: %s", e.ErrorInfo.ErrDescription, e.ErrorInfo.MsgUUID, e.ErrorInfo.MsgParentUUID)
}

// SystemError - any other error not related to payload
type SystemError struct {
	ErrorInfo MsgErrorInfo
}

func (e SystemError) Error() string {
	return fmt.Sprintf("SystemError: %s MsgUUID: %s MstParentUUID: %s", e.ErrorInfo.ErrDescription, e.ErrorInfo.MsgUUID, e.ErrorInfo.MsgParentUUID)
}

// SendingError -
type SendingError struct {
	ErrorInfo MsgErrorInfo
}

func (e SendingError) Error() string {
	return fmt.Sprintf("SendingError: %s MsgUUID: %s MstParentUUID: %s", e.ErrorInfo.ErrDescription, e.ErrorInfo.MsgUUID, e.ErrorInfo.MsgParentUUID)
}

// NewSendingError - constructor function
func NewSendingError(err error) SendingError {
	return SendingError{
		ErrorInfo: MsgErrorInfo{
			ErrDescription: err.Error(),
		},
	}
}
