package payment

import (
	"testing"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorize(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	assert.Equal(t, StatusAuthorized, p.Status)
	assert.Len(t, p.Events(), 1)
}

func TestCapture(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	require.NoError(t, p.Capture())
	assert.Equal(t, StatusCaptured, p.Status)
}

func TestCaptureWithoutAuthorization(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	assert.ErrorIs(t, p.Capture(), ErrNotAuthorized)
}

func TestVoid(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	require.NoError(t, p.Void())
	assert.Equal(t, StatusVoided, p.Status)
}

func TestVoidAfterCapture(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	require.NoError(t, p.Capture())
	assert.ErrorIs(t, p.Void(), ErrAlreadyCaptured)
}

func TestVoidWithoutAuthorization(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	assert.ErrorIs(t, p.Void(), ErrNotAuthorized)
}

func TestRefund(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	require.NoError(t, p.Capture())
	require.NoError(t, p.Refund(shared.MustMoney(1000, "USD")))
	assert.Equal(t, int64(1000), p.Refunded.Amount)
	assert.Equal(t, StatusCaptured, p.Status)
}

func TestFullRefund(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	require.NoError(t, p.Capture())
	require.NoError(t, p.Refund(shared.MustMoney(2500, "USD")))
	assert.Equal(t, StatusRefunded, p.Status)
}

func TestRefundExceedsCaptured(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	require.NoError(t, p.Capture())
	assert.ErrorIs(t, p.Refund(shared.MustMoney(3000, "USD")), ErrRefundExceeds)
}

func TestRefundWithoutCapture(t *testing.T) {
	p := New("pay-1", "order-1", shared.MustMoney(2500, "USD"))
	require.NoError(t, p.Authorize())
	assert.ErrorIs(t, p.Refund(shared.MustMoney(1000, "USD")), ErrNotCaptured)
}
