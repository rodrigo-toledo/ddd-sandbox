package order

import (
	"testing"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testItems() []Item {
	return []Item{
		{ProductID: "prod-1", Quantity: 2, UnitPrice: shared.MustMoney(1000, "USD")},
		{ProductID: "prod-2", Quantity: 1, UnitPrice: shared.MustMoney(500, "USD")},
	}
}

func TestNewOrder(t *testing.T) {
	o, err := New("order-1", "cust-1", testItems())
	require.NoError(t, err)
	assert.Equal(t, StatusPending, o.Status)
	assert.Equal(t, int64(2500), o.Total.Amount)
	assert.Len(t, o.Events(), 1)
	assert.Equal(t, "order.placed", o.Events()[0].EventName())
}

func TestNewOrderRejectsEmptyItems(t *testing.T) {
	_, err := New("order-1", "cust-1", nil)
	assert.ErrorIs(t, err, ErrEmptyOrder)
}

func TestConfirmOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Confirm())
	assert.Equal(t, StatusConfirmed, o.Status)
}

func TestConfirmOrderTwice(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Confirm())
	assert.ErrorIs(t, o.Confirm(), ErrAlreadyConfirmed)
}

func TestShipOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Confirm())
	require.NoError(t, o.Ship())
	assert.Equal(t, StatusShipped, o.Status)
}

func TestShipUnconfirmedOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	assert.ErrorIs(t, o.Ship(), ErrNotConfirmed)
}

func TestDeliverOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Confirm())
	require.NoError(t, o.Ship())
	require.NoError(t, o.Deliver())
	assert.Equal(t, StatusDelivered, o.Status)
}

func TestDeliverUnshippedOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Confirm())
	assert.ErrorIs(t, o.Deliver(), ErrNotShipped)
}

func TestCancelPendingOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Cancel())
	assert.Equal(t, StatusCancelled, o.Status)
}

func TestCancelConfirmedOrder(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	require.NoError(t, o.Confirm())
	assert.ErrorIs(t, o.Cancel(), ErrCannotCancel)
}

func TestClearEvents(t *testing.T) {
	o, _ := New("order-1", "cust-1", testItems())
	o.ClearEvents()
	assert.Empty(t, o.Events())
}
