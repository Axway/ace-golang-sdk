package logging

// Constants for Log fields
const (
	// General constants
	LogFieldError = "error"
	LogFieldEvent = "event"

	// Messaging Constants
	LogFieldMessageID       = "messageId"
	LogFieldParentMessageID = "parentMessageId"
	LogFieldMessageCount    = "messageCount"

	// Service constants
	LogFieldServiceName    = "serviceName"
	LogFieldServiceType    = "serviceType"
	LogFieldServiceVersion = "serviceVersion"

	LogFieldServiceHost = "serviceHost"
	LogFieldServicePort = "servicePort"

	LogFieldSidecarHost = "sidecarHost"
	LogFieldSidecarPort = "sidecarPort"
	LogFieldSidecarID   = "sidecarId"
)
