package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReserve(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	require.NoError(t, p.Reserve("order-1", 3))
	assert.Equal(t, 7, p.Available())
	assert.Len(t, p.Reservations, 1)
	assert.Len(t, p.Events(), 1)
}

func TestReserveInsufficientStock(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 2)
	assert.ErrorIs(t, p.Reserve("order-1", 5), ErrInsufficientStock)
}

func TestReserveAccountsForExistingReservations(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	require.NoError(t, p.Reserve("order-1", 7))
	assert.ErrorIs(t, p.Reserve("order-2", 5), ErrInsufficientStock)
}

func TestReserveDuplicateOrder(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	require.NoError(t, p.Reserve("order-1", 3))
	assert.ErrorIs(t, p.Reserve("order-1", 2), ErrAlreadyReserved)
}

func TestRelease(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	require.NoError(t, p.Reserve("order-1", 3))
	require.NoError(t, p.Release("order-1"))
	assert.Equal(t, 10, p.Available())
	assert.Empty(t, p.Reservations)
}

func TestReleaseNotFound(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	assert.ErrorIs(t, p.Release("order-99"), ErrReservationNotFound)
}

func TestConfirmReservation(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	require.NoError(t, p.Reserve("order-1", 3))
	require.NoError(t, p.ConfirmReservation("order-1"))
	assert.Equal(t, 7, p.Stock)
	assert.Empty(t, p.Reservations)
}

func TestConfirmReservationNotFound(t *testing.T) {
	p := NewProduct("prod-1", "Widget", 10)
	assert.ErrorIs(t, p.ConfirmReservation("order-99"), ErrReservationNotFound)
}
