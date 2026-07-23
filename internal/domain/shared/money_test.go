package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMoney(t *testing.T) {
	m, err := NewMoney(1000, "USD")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), m.Amount)
	assert.Equal(t, "USD", m.Currency)
}

func TestNewMoneyRejectsNegative(t *testing.T) {
	_, err := NewMoney(-1, "USD")
	assert.Error(t, err)
}

func TestNewMoneyRejectsEmptyCurrency(t *testing.T) {
	_, err := NewMoney(100, "")
	assert.Error(t, err)
}

func TestMoneyAdd(t *testing.T) {
	a := MustMoney(500, "USD")
	b := MustMoney(300, "USD")
	sum, err := a.Add(b)
	require.NoError(t, err)
	assert.Equal(t, int64(800), sum.Amount)
}

func TestMoneyAddDifferentCurrency(t *testing.T) {
	a := MustMoney(500, "USD")
	b := MustMoney(300, "EUR")
	_, err := a.Add(b)
	assert.Error(t, err)
}

func TestMoneyMultiply(t *testing.T) {
	m := MustMoney(250, "USD")
	result := m.Multiply(3)
	assert.Equal(t, int64(750), result.Amount)
}

func TestMoneyEquals(t *testing.T) {
	a := MustMoney(100, "USD")
	b := MustMoney(100, "USD")
	c := MustMoney(100, "EUR")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
