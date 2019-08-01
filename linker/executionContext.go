package linker

import (
	"context"
	"strconv"

	"github.com/Axway/ace-golang-sdk/util/logging"
	"go.uber.org/zap"

	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
)

// messageContext - Represents the current execution context that holds the message and related properties
type messageContext struct {
	ctx              context.Context
	businessMessages []*messaging.BusinessMessage
	configMap        map[string]*rpc.ConfigParameter
	msgProducer      MsgProducer
}

// ExecutionContext - Interface exposed to clients to access the current message context
type ExecutionContext interface {
	GetSpanContext() context.Context
	GetBusinessMessages() []*messaging.BusinessMessage
	GetMsgProducer() MsgProducer
	GetStringConfig(string) string
	GetIntConfig(string) int
	GetBooleanConfig(string) bool
}

// GetSpanContext - Returns the current context
func (msgCtx *messageContext) GetSpanContext() context.Context {
	return msgCtx.ctx
}

// GetBusinessMessages - Returns business messages
func (msgCtx *messageContext) GetBusinessMessages() []*messaging.BusinessMessage {
	return msgCtx.businessMessages
}

// GetMsgProducer - Returns the interfacee for producing messages
func (msgCtx *messageContext) GetMsgProducer() MsgProducer {
	return msgCtx.msgProducer
}

// GetStringConfig - Returns the string confign value
func (msgCtx *messageContext) GetStringConfig(name string) string {
	cfgParam, ok := msgCtx.configMap[name]
	if !ok {
		log.Warn("The requested configuration did not exist, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamTypeRequested, "string"))
		return ""
	}

	if cfgParam.GetType() != "string" {
		log.Warn("The requested configuration is not of type string, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamType, cfgParam.GetType()),
			zap.String(logging.LogConfigParamTypeRequested, "string"))
		return ""
	}

	return cfgParam.GetValue()
}

// GetIntConfig - Returns the int config value
func (msgCtx *messageContext) GetIntConfig(name string) int {
	cfgParam, ok := msgCtx.configMap[name]
	if !ok {
		log.Warn("The requested configuration did not exist, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamTypeRequested, "boolean"))
		return 0
	}

	if cfgParam.GetType() != "int" {
		log.Warn("The requested configuration is not of type int, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamType, cfgParam.GetType()),
			zap.String(logging.LogConfigParamTypeRequested, "int"))
		return 0
	}

	intVal, err := strconv.Atoi(cfgParam.GetValue())
	if err != nil {
		log.Warn("Could not parse the config value to an int, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamType, cfgParam.GetType()),
			zap.String(logging.LogConfigParamTypeRequested, "int"))
		return 0
	}
	return intVal
}

// GetBooleanConfig - Returns the boolean config value
func (msgCtx *messageContext) GetBooleanConfig(name string) bool {
	cfgParam, ok := msgCtx.configMap[name]
	if !ok {
		log.Warn("The requested configuration did not exist, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamTypeRequested, "boolean"))
		return false
	}

	if cfgParam.GetType() != "boolean" {
		log.Warn("The requested configuration is not of type boolean, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamType, cfgParam.GetType()),
			zap.String(logging.LogConfigParamTypeRequested, "boolean"))
		return false
	}

	boolVal, err := strconv.ParseBool(cfgParam.GetValue())
	if err != nil {
		log.Warn("Could not parse the config value to a boolean, returning default for type",
			zap.String(logging.LogConfigParamName, name),
			zap.String(logging.LogConfigParamType, cfgParam.GetType()),
			zap.String(logging.LogConfigParamTypeRequested, "boolean"))
		return false
	}
	return boolVal
}
