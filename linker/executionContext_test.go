package linker

import (
	"context"
	"strconv"
	"testing"

	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
)

// var testProcBusMsg *messaging.BusinessMessage

var testExecutionContext ExecutionContext

func testBusinessProcess(ec ExecutionContext) error {
	testExecutionContext = ec
	return nil
}

type testMsgProducer struct{}

// Send -
func (p testMsgProducer) Send(bm *messaging.BusinessMessage) error {
	return nil
}

func TestExecutionContext(t *testing.T) {
	ctx := context.Background()
	msgProducer := testMsgProducer{}
	configMap := make(map[string]*rpc.ConfigParameter)

	// Add all the config map values we will need
	configMap["string-param"] = &rpc.ConfigParameter{
		Name:  "string-param",
		Type:  "string",
		Value: "string-value",
	}
	configMap["boolean-param"] = &rpc.ConfigParameter{
		Name:  "boolean-param",
		Type:  "boolean",
		Value: "true",
	}
	configMap["int-param"] = &rpc.ConfigParameter{
		Name:  "int-param",
		Type:  "int",
		Value: "123",
	}
	configMap["int-param-misconfig"] = &rpc.ConfigParameter{
		Name:  "int-param",
		Type:  "int",
		Value: "abc",
	}

	msgContext := messageContext{
		ctx:         ctx,
		msgProducer: msgProducer,
		configMap:   configMap,
	}

	testBusinessProcess(&msgContext)

	if ctx != testExecutionContext.GetSpanContext() {
		t.Error("incorrect Context received from GetSpanContext")
	}

	// Test GetStringConfig
	if testExecutionContext.GetStringConfig("string-param") != configMap["string-param"].Value {
		t.Errorf("The GetStringConfig method returned %s but we expected %s", testExecutionContext.GetStringConfig("string-param"), configMap["string-param"])
	}

	if testExecutionContext.GetStringConfig("int-param") != "" {
		t.Error("The GetStringConfig method returned a value for an int type but we expected \"\"")
	}

	if testExecutionContext.GetStringConfig("param-does-not-exist") != "" {
		t.Error("The GetStringConfig method returned a value for a non existent params but we expected \"\"")
	}

	// Test GetIntConfig
	intVal, _ := strconv.Atoi(configMap["int-param"].Value)
	if testExecutionContext.GetIntConfig("int-param") != intVal {
		t.Errorf("The GetIntConfig method returned %d but we expected %s", testExecutionContext.GetIntConfig("int-param"), configMap["int-param"])
	}

	if testExecutionContext.GetIntConfig("string-param") != 0 {
		t.Error("The GetIntConfig method returned a value for a string type but we expected 0")
	}

	if testExecutionContext.GetIntConfig("param-does-not-exist") != 0 {
		t.Error("The GetIntConfig method returned a value for a non existent params but we expected 0")
	}

	if testExecutionContext.GetIntConfig("int-param-misconfig") != 0 {
		t.Error("The GetIntConfig method returned a value for a non parsable value but we expected 0")
	}

	// Test GetBooleanConfig
	boolVal, _ := strconv.ParseBool(configMap["boolean-param"].Value)
	if testExecutionContext.GetBooleanConfig("boolean-param") != boolVal {
		t.Errorf("The GetBooleanConfig method returned %t but we expected %s", testExecutionContext.GetBooleanConfig("boolean-param"), configMap["boolean-param"])
	}

	if testExecutionContext.GetBooleanConfig("string-param") != false {
		t.Error("The GetBooleanConfig method returned a value for a string type but we expected false")
	}

	if testExecutionContext.GetBooleanConfig("param-does-not-exist") != false {
		t.Error("The GetBooleanConfig method returned a value for a non existent params but we expected false")
	}

	if testExecutionContext.GetBooleanConfig("boolean-param-misconfig") != false {
		t.Error("The GetBooleanConfig method returned a value for a non parsable value but we expected false")
	}
}
