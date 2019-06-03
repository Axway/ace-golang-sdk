package rpcclient

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
)

func Test_copyStepPattern_one_level(t *testing.T) {
	orig := &rpc.StepPattern{
		ServiceName: "test-service-name",
		Child: []*rpc.StepPattern{
			&rpc.StepPattern{
				ServiceName: "child-sn-0",
			},
		},
	}
	target := &rpc.StepPattern{}

	copyPattern(target, orig)

	if target.ServiceName != orig.ServiceName {
		t.Error(fmt.Sprintf("%s did not copy", orig.ServiceName))
	}
	if len(target.Child) != 1 {
		t.Error(fmt.Sprintf("Child did not copy, expecting %v, got %v\n", orig.Child, target.Child))
	}
	if target.Child[0].ServiceName != orig.Child[0].ServiceName {
		t.Error(fmt.Sprintf("first Child did not copy, expecting %v, got %v\n", orig.Child[0], target.Child[0]))
	}
	if len(target.Child[0].Child) != 0 {
		t.Error(fmt.Sprintf("first Child did not copy, expecting %v, got %v\n", orig.Child[0], target.Child[0]))
	}

}
func Test_copyStepPattern_two_levels(t *testing.T) {
	orig := &rpc.StepPattern{
		ServiceName: "sn-0",
		Child: []*rpc.StepPattern{
			&rpc.StepPattern{
				ServiceName: "child-sn-0",
				Child: []*rpc.StepPattern{
					&rpc.StepPattern{
						ServiceName: "child-sn-0-0",
					},
					&rpc.StepPattern{
						ServiceName: "child-sn-0-1",
					},
				},
			},
		},
	}
	target := &rpc.StepPattern{}

	copyPattern(target, orig)

	if !reflect.DeepEqual(orig, target) {
		fmt.Printf("expected target to equal orig: %v\n, got: %v\n", orig, target)
	}

	if target.ServiceName != orig.ServiceName {
		t.Error(fmt.Sprintf("%s did not copy", orig.ServiceName))
	}
	if len(target.Child) != 1 {
		t.Error(fmt.Sprintf("Child did not copy, expecting %v, got %v\n", orig.Child, target.Child))
	}
	if target.Child[0].ServiceName != orig.Child[0].ServiceName {
		t.Error(fmt.Sprintf("first Child did not copy, expecting %v, got %v\n", orig.Child[0], target.Child[0]))
	}
	if len(target.Child[0].Child) != 2 {
		t.Error(fmt.Sprintf("first Child did not copy, expecting %d children of first  top-most child, got %d\n",
			len(orig.Child[0].Child), len(target.Child[0].Child)))
	}

}

func TestBuildResult(t *testing.T) {
	parentMsg := &rpc.Message{
		UUID: "test_UUID",
		Pattern: &rpc.StepPattern{
			ServiceName:    "test-a",
			ServiceVersion: "xxx",
		},
	}
	bMsg := &messaging.BusinessMessage{
		Payload: &messaging.Payload{
			Body: []byte("test"),
		},
	}

	copy := buildResult(parentMsg, bMsg)

	if copy.GetParent_UUID() != parentMsg.GetUUID() {
		t.Errorf("expected UUID of parent msg to become Parent_UUID of a copy, expected: %s got %s", parentMsg.GetUUID(), copy.GetParent_UUID())
	}
	if len(copy.GetUUID()) > 0 {
		t.Errorf("expected UUID of copied message to be empty")
	}

	if copy.Pattern.ServiceName != parentMsg.Pattern.ServiceName || copy.Pattern.ServiceVersion != parentMsg.Pattern.ServiceVersion {
		t.Errorf("Pattern should have been copied")
	}

	if string(copy.BusinessMessage[0].Payload.Body) != "test" {
		t.Errorf("Business message should have been copied")
	}
}
