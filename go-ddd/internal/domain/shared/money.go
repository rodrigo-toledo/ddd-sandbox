package shared

import (
	"errors"
	"fmt"
)

type Money struct {
	Amount   int64
	Currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("money amount cannot be negative")
	}
	if currency == "" {
		return Money{}, errors.New("currency cannot be empty")
	}
	return Money{Amount: amount, Currency: currency}, nil
}

func MustMoney(amount int64, currency string) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add %s to %s", other.Currency, m.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

func (m Money) Multiply(factor int) Money {
	return Money{Amount: m.Amount * int64(factor), Currency: m.Currency}
}

func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.Amount, m.Currency)
}
