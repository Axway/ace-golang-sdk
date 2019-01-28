package rpcclient

import (
	"fmt"
	"reflect"
	"testing"

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
