package mq

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConsumer(t *testing.T) {
	uri := fmt.Sprintf("amqp://guest:guest@%s:%d/", "139.219.1.102", 5672)
	exchange := "global_fa34664e_1551750913902413534"
	queue := "bitxhub-6"
	cons, err := NewConsumer(WithURI(uri), WithExchange(exchange), WithQueueName(queue), WithExchangeType("direct"))
	require.Nil(t, err)
	err = cons.Start()
	time.Sleep(10 * time.Second)
	require.Nil(t, err)
}
