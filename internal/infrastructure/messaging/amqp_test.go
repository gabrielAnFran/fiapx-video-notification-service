package messaging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDial_InvalidURL_ReturnsError(t *testing.T) {
	conn, err := Dial("amqp://127.0.0.1:1/")
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestNewEvent_UnmarshalablePayload_ReturnsError(t *testing.T) {
	_, err := NewEvent("SomeEvent", "corr", "", func() {})
	assert.Error(t, err)
}
