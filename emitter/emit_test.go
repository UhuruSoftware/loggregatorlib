package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockLoggregatorClient struct {
	received chan *[]byte
}

func (m MockLoggregatorClient) Send(data []byte) {
	m.received <- &data
}

func (m MockLoggregatorClient) Emit() instrumentation.Context {
	return instrumentation.Context{}
}

func (m MockLoggregatorClient) IncLogStreamRawByteCount(uint64) {

}

func (m MockLoggregatorClient) IncLogStreamPbByteCount(uint64) {

}

func TestEmitter(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewEmitter("localhost:3456", "ROUTER", "42", nil)
	e.lc = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedMessage := getBackendMessage(t, <-received)

	assert.Equal(t, receivedMessage.GetMessage(), []byte("foo"))
	assert.Equal(t, receivedMessage.GetAppId(), "appid")
	assert.Equal(t, receivedMessage.GetSourceId(), "42")
}

func TestInvalidSourcetype(t *testing.T) {
	_, err := NewEmitter("server", "FOOSERVER", "42", nil)
	assert.Error(t, err)
}

func TestValidSourcetype(t *testing.T) {
	_, err := NewEmitter("localhost:38452", "ROUTER", "42", nil)
	assert.NoError(t, err)
}

func TestEmptyAppIdDoesNotEmit(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewEmitter("localhost:3456", "ROUTER", "42", nil)
	e.lc = &MockLoggregatorClient{received}

	e.Emit("", "foo")
	select {
	case <-received:
		t.Error("This message should not have been emitted since it does not have an AppId")
	default:
		// success
	}

	e.Emit("    ", "foo")
	select {
	case <-received:
		t.Error("This message should not have been emitted since it does not have an AppId")
	default:
		// success
	}
}

func TestEmptyMessageDoesNotEmit(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewEmitter("localhost:3456", "ROUTER", "42", nil)
	e.lc = &MockLoggregatorClient{received}

	e.Emit("appId", "")
	select {
	case <-received:
		t.Error("This message should not have been emitted since it does not have a message")
	default:
		// success
	}

	e.Emit("appId", "   ")
	select {
	case <-received:
		t.Error("This message should not have been emitted since it does not have a message")
	default:
		// success
	}
}

func getBackendMessage(t *testing.T, data *[]byte) *logmessage.LogMessage {
	receivedMessage := &logmessage.LogMessage{}

	err := proto.Unmarshal(*data, receivedMessage)

	if err != nil {
		t.Fatalf("Message invalid. %s", err)
	}
	return receivedMessage
}
