package linker

import (
	"context"
	"testing"

	"github.com/Axway/ace-golang-sdk/messaging"
)

func testProc(context.Context, *messaging.BusinessMessage, MsgProducer) error {

	return nil
}

func TestRegister(t *testing.T) {

	testServiceName := "test-service"
	testServiceVersion := "test-version"
	testDescription := "test-descr"

	link, err := Register(testServiceName, testServiceVersion, testDescription, testProc)

	if link == nil && err != nil {
		t.Errorf("did not expect error from call to Register with valid parameters")
	}
	if link.MsgProcessor == nil {
		t.Errorf("message processor not assigned")
	}

	if link.name != testServiceName {
		t.Errorf("incorrect description, expected %s got %s", testServiceName, link.name)
	}
	if link.version != testServiceVersion {
		t.Errorf("incorrect description, expected %s got %s", testServiceVersion, link.version)
	}
	if link.description != testDescription {
		t.Errorf("incorrect description, expected %s got %s", testDescription, link.description)
	}
}
