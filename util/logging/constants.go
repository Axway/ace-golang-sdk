package logging

// Constants for Log fields
const (
	// General constants
	LogFieldErrorType = "errorType"
	LogFieldError     = "error"
	LogFieldErrorInfo = "errorInfo"
	LogFieldEvent     = "event"

	// Messaging Constants
	LogFieldChoreographyID     = "choreographyId"
	LogFieldExecutionID        = "executionId"
	LogFieldMessageID          = "messageId"
	LogFieldParentMessageID    = "parentMessageId"
	LogFieldMessageCount       = "messageCount"
	LogFieldSequenceID         = "sequenceId"
	LogFieldSequenceTerm       = "sequenceTerm"
	LogFieldSequenceUpperBound = "sequenceUpperBound"

	// Service constants
	LogFieldServiceName    = "serviceName"
	LogFieldServiceType    = "serviceType"
	LogFieldServiceVersion = "serviceVersion"
	LogFieldServiceDesc    = "serviceDesc"

	LogFieldServiceHost = "serviceHost"
	LogFieldServicePort = "servicePort"

	LogFieldSidecarHost = "sidecarHost"
	LogFieldSidecarPort = "sidecarPort"
	LogFieldAgentHost   = "agentHost"
	LogFieldAgentPort   = "agentPort"
	LogFieldSidecarID   = "sidecarId"
)
