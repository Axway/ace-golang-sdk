package linker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/Axway/ace-golang-sdk/linker/rpcclient"
	"github.com/Axway/ace-golang-sdk/messaging"
	"github.com/Axway/ace-golang-sdk/rpc"
	"google.golang.org/grpc/metadata"
)

var testProcBusMsg *messaging.BusinessMessage

func testProc(c context.Context, bm *messaging.BusinessMessage, mp MsgProducer) error {
	testProcBusMsg = bm
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

type mockClient struct {
	buildClientRelayCalled bool
	clientRegisterCalled   bool
}
type mockStream struct {
	closeSendCalled bool
}

func (m mockStream) CloseSend() error {
	m.closeSendCalled = true
	return nil
}

/*
type Linkage_RelayClient interface {
	Send(*Message) error
	Recv() (*Receipt, error)
	grpc.ClientStream
}
*/
type linkageRelayClient struct{}

var closeSendCalled bool

func (l linkageRelayClient) CloseSend() error { //grpc.ClientStream interface method
	fmt.Printf("CloseSend called")
	closeSendCalled = true
	return nil
}
func (l linkageRelayClient) Context() context.Context     { return context.Background() }
func (l linkageRelayClient) Header() (metadata.MD, error) { return nil, nil }
func (l linkageRelayClient) Trailer() metadata.MD         { return nil }
func (l linkageRelayClient) Send(*rpc.Message) error {
	fmt.Printf("Send of rpc.Message called")
	return nil
}
func (l linkageRelayClient) Recv() (*rpc.Receipt, error) { return nil, io.EOF }
func (l linkageRelayClient) SendMsg(m interface{}) error { return nil }
func (l linkageRelayClient) RecvMsg(m interface{}) error { return nil }

// rpcClient implements RPCClient interface
func (mc *mockClient) BuildClientRelay(clientContext context.Context, aceMsg *rpc.Message, host string, port uint16) (*rpcclient.LinkerClientRelay, error) {
	// real one does this:
	//return rpcclient.BuildClientRelay(clientContext, aceMsg, host, port)
	mc.buildClientRelayCalled = true
	lcr := rpcclient.LinkerClientRelay{
		Stream:        linkageRelayClient{},
		SourceMessage: aceMsg,
	}
	return &lcr, nil
}

// rpcClient implements RPCClient interface
func (mc *mockClient) ClientRegister(host string, port uint16, serviceInfo *rpc.ServiceInfo) (bool, error) {
	// real one:
	//return rpcclient.ClientRegister(host, port, serviceInfo)

	mc.clientRegisterCalled = true
	return true, nil
}

func TestOnSidecarRegistrationComplete(t *testing.T) {
	mockClient := &mockClient{}

	link := &Link{
		client: mockClient,
	}

	serviceInfo := rpc.ServiceInfo{
		ServiceName:    "abc",
		ServiceVersion: "1.0.0",
	}
	err := link.OnSidecarRegistrationComplete(&serviceInfo)

	if err != nil {
		t.Errorf("no error was expected, but got: %v", err)
	}

	if !mockClient.clientRegisterCalled {
		t.Errorf("OnSidecarRegistrationComplete should have called ClientRegister")
	}
}

func TestOnRelayNoErrors(t *testing.T) {
	mockClient := &mockClient{}

	link := &Link{
		client:       mockClient,
		MsgProcessor: BusinessMessageProcessor(testProc),
	}

	aceMsg := &rpc.Message{
		Pattern: &rpc.StepPattern{
			ServiceName:    "abc",
			ServiceVersion: "1.0.0",
		},
		BusinessMessage: &messaging.BusinessMessage{
			Payload: &messaging.Payload{
				Body: []byte("test"),
			},
		},
	}

	link.OnRelay(aceMsg)

	if !closeSendCalled {
		t.Errorf("expected CloseSend on LinkerClientRelay to have been called")
	}

	//was the testProc called with aceMsg.GetBusinessMessage()?
}

func testProcReturnProcessingError(c context.Context, bm *messaging.BusinessMessage, mp MsgProducer) error {
	testProcBusMsg = bm
	return NewProcessingError(errors.New("test-processing-error"))
}
func TestOnRelayProcessingError(t *testing.T) {
	mockClient := &mockClient{}

	link := &Link{
		client:       mockClient,
		MsgProcessor: BusinessMessageProcessor(testProcReturnProcessingError),
	}

	aceMsg := &rpc.Message{
		Pattern: &rpc.StepPattern{
			ServiceName:    "abc",
			ServiceVersion: "1.0.0",
		},
		BusinessMessage: &messaging.BusinessMessage{
			Payload: &messaging.Payload{
				Body: []byte("test"),
			},
		},
	}

	link.OnRelay(aceMsg)

	if !closeSendCalled {
		t.Errorf("expected CloseSend on LinkerClientRelay to have been called")
	}
}
