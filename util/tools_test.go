package util

import (
	"testing"

	"github.com/Axway/ace-golang-sdk/rpc"
)

func TestCopyMessageNoError(t *testing.T) {
	sourceMsg := &rpc.Message{
		UUID:               "test-UUID",
		Parent_UUID:        "test-Parent-UUID",
		ErrorType:          rpc.Message_NONE,
		CHN_UUID:           "test-CHN_UUID",
		CHX_UUID:           "test-CHX_UUID",
		Consumption_ID:     "test-Consumption_ID",
		TopicName:          "test-TopicName",
		OpentracingContext: "OpentracingContext",
	}

	res := CopyMessage(sourceMsg)

	if len(res.UUID) > 0 {
		t.Errorf("CopyMessage should ignore UUID, but did not: %s", res.GetUUID())
	}
	if len(res.Parent_UUID) > 0 {
		t.Errorf("CopyMessage should ignore Parent_UUID, but did not: %s", res.GetParent_UUID())
	}
	if res.ErrorType != rpc.Message_NONE {
		t.Errorf("CopyMessage copied ErrorType incorrectly, expected: %s got %s", sourceMsg.GetErrorType(), res.GetErrorType())
	}
	if res.GetCHN_UUID() != sourceMsg.GetCHN_UUID() {
		t.Errorf("CopyMessage copied CHN_UUID incorrectly, expected: %s got %s", sourceMsg.GetCHN_UUID(), res.GetCHN_UUID())
	}
	if res.GetCHX_UUID() != sourceMsg.GetCHX_UUID() {
		t.Errorf("CopyMessage copied CHX_UUID incorrectly, expected: %s got %s", sourceMsg.GetCHX_UUID(), res.GetCHX_UUID())
	}
	if res.GetConsumption_ID() != sourceMsg.GetConsumption_ID() {
		t.Errorf("CopyMessage copied Consumption_ID incorrectly, expected: %s got %s", sourceMsg.GetConsumption_ID(), res.GetConsumption_ID())
	}
	if res.GetTopicName() != sourceMsg.GetTopicName() {
		t.Errorf("CopyMessage copied TopicName incorrectly, expected: %s got %s", sourceMsg.GetTopicName(), res.GetTopicName())
	}
	if res.GetOpentracingContext() != sourceMsg.GetOpentracingContext() {
		t.Errorf("CopyMessage copied OpentracingContext incorrectly, expected: %s got %s", sourceMsg.GetOpentracingContext(), res.GetOpentracingContext())
	}

}
func TestCopyMessageWithError(t *testing.T) {
	sourceMsg := &rpc.Message{
		ErrorType:        rpc.Message_PROCESSING,
		ErrorDescription: "test-error-description",
	}

	res := CopyMessage(sourceMsg)

	if res.ErrorType != rpc.Message_PROCESSING {
		t.Errorf("CopyMessage copied ErrorType incorrectly, expected: %s got %s", sourceMsg.GetErrorType(), res.GetErrorType())
	}
	if res.GetErrorDescription() != sourceMsg.GetErrorDescription() {
		t.Errorf("CopyMessage copied ErrorDescription incorrectly, expected: %s got %s", sourceMsg.GetErrorDescription(), res.GetErrorDescription())
	}

}
