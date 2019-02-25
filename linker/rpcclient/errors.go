package rpcclient

import "fmt"

const (
	// ErrorSystem - metadata key name for rpc.Message.MetaData map to carry description of error encountered during message processing
	// which is not payload-related;
	ErrorSystem string = "ErrorSystem"

	// ErrorProcessing - metadata key name for rpc.Message.MetaData map to carry description of error from payload processing
	ErrorProcessing string = "ErrorProcessing"
)

// MsgErrorInfo -
type MsgErrorInfo struct {
	ErrDescription string
}

// ProcessingError - signals errors encountered when processing payload
type ProcessingError struct {
	ErrorInfo MsgErrorInfo
}

func (e ProcessingError) Error() string {
	return fmt.Sprintf("ProcessingError: %s", e.ErrorInfo.ErrDescription)
}

// SystemError - any other error not related to payload
type SystemError struct {
	ErrorInfo MsgErrorInfo
}

func (e SystemError) Error() string {
	return fmt.Sprintf("SystemError: %s", e.ErrorInfo.ErrDescription)
}

// SendingError -
type SendingError struct {
	ErrorInfo MsgErrorInfo
}

func (e SendingError) Error() string {
	return fmt.Sprintf("SendingError: %s", e.ErrorInfo.ErrDescription)
}

// NewSendingError - constructor function
func NewSendingError(err error) SendingError {
	return SendingError{
		ErrorInfo: MsgErrorInfo{
			ErrDescription: err.Error(),
		},
	}
}
