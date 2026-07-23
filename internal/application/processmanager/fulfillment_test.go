package processmanager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time { return c.now }
func (c *fakeClock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func TestFulfillmentLifecycle(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := NewInMemoryStateStore()
	pm := NewFulfillmentPM(store, clock)

	require.NoError(t, pm.HandleOrderConfirmed("order-1"))
	fs, _ := store.FindByOrderID("order-1")
	assert.Equal(t, StateWaitingForShipment, fs.State)

	require.NoError(t, pm.HandleOrderShipped("order-1"))
	fs, _ = store.FindByOrderID("order-1")
	assert.Equal(t, StateShipped, fs.State)

	require.NoError(t, pm.HandleOrderDelivered("order-1"))
	fs, _ = store.FindByOrderID("order-1")
	assert.Equal(t, StateDelivered, fs.State)
	assert.Equal(t, clock.now.Add(ReturnWindowDuration), fs.ReturnDeadline)
}

func TestReturnWithinWindow(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := NewInMemoryStateStore()
	pm := NewFulfillmentPM(store, clock)

	pm.HandleOrderConfirmed("order-1")
	pm.HandleOrderShipped("order-1")
	pm.HandleOrderDelivered("order-1")

	clock.Advance(10 * 24 * time.Hour)
	require.NoError(t, pm.HandleReturnRequested("order-1"))
	fs, _ := store.FindByOrderID("order-1")
	assert.Equal(t, StateReturnRequested, fs.State)

	require.NoError(t, pm.HandleReturnCompleted("order-1"))
	fs, _ = store.FindByOrderID("order-1")
	assert.Equal(t, StateCompleted, fs.State)
}

func TestReturnWindowExpires(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := NewInMemoryStateStore()
	pm := NewFulfillmentPM(store, clock)

	pm.HandleOrderConfirmed("order-1")
	pm.HandleOrderShipped("order-1")
	pm.HandleOrderDelivered("order-1")

	clock.Advance(31 * 24 * time.Hour)
	require.NoError(t, pm.CheckReturnWindow("order-1"))
	fs, _ := store.FindByOrderID("order-1")
	assert.Equal(t, StateCompleted, fs.State)
}

func TestReturnRejectedAfterWindow(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := NewInMemoryStateStore()
	pm := NewFulfillmentPM(store, clock)

	pm.HandleOrderConfirmed("order-1")
	pm.HandleOrderShipped("order-1")
	pm.HandleOrderDelivered("order-1")

	clock.Advance(31 * 24 * time.Hour)
	require.NoError(t, pm.HandleReturnRequested("order-1"))
	fs, _ := store.FindByOrderID("order-1")
	assert.Equal(t, StateDelivered, fs.State)
}

func TestInvalidTransitionsIgnored(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	store := NewInMemoryStateStore()
	pm := NewFulfillmentPM(store, clock)

	pm.HandleOrderConfirmed("order-1")

	pm.HandleOrderDelivered("order-1")
	fs, _ := store.FindByOrderID("order-1")
	assert.Equal(t, StateWaitingForShipment, fs.State)
}
