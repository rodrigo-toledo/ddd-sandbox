package processmanager

import (
	"time"

	"github.com/rodrigotoledo/ddd-sandbox/internal/domain/shared"
)

type State string

const (
	StateWaitingForShipment State = "waiting_for_shipment"
	StateShipped            State = "shipped"
	StateDelivered          State = "delivered"
	StateReturnRequested    State = "return_requested"
	StateCompleted          State = "completed"
)

const ReturnWindowDuration = 30 * 24 * time.Hour

type FulfillmentState struct {
	OrderID      string
	State        State
	DeliveredAt  time.Time
	ReturnDeadline time.Time
}

type StateStore interface {
	Save(fs *FulfillmentState) error
	FindByOrderID(orderID string) (*FulfillmentState, error)
}

type InMemoryStateStore struct {
	states map[string]*FulfillmentState
}

func NewInMemoryStateStore() *InMemoryStateStore {
	return &InMemoryStateStore{states: make(map[string]*FulfillmentState)}
}

func (s *InMemoryStateStore) Save(fs *FulfillmentState) error {
	s.states[fs.OrderID] = fs
	return nil
}

func (s *InMemoryStateStore) FindByOrderID(orderID string) (*FulfillmentState, error) {
	fs, ok := s.states[orderID]
	if !ok {
		return nil, nil
	}
	return fs, nil
}

type FulfillmentPM struct {
	store StateStore
	clock shared.Clock
}

func NewFulfillmentPM(store StateStore, clock shared.Clock) *FulfillmentPM {
	return &FulfillmentPM{store: store, clock: clock}
}

func (pm *FulfillmentPM) HandleOrderConfirmed(orderID string) error {
	fs := &FulfillmentState{
		OrderID: orderID,
		State:   StateWaitingForShipment,
	}
	return pm.store.Save(fs)
}

func (pm *FulfillmentPM) HandleOrderShipped(orderID string) error {
	fs, err := pm.store.FindByOrderID(orderID)
	if err != nil || fs == nil {
		return err
	}
	if fs.State != StateWaitingForShipment {
		return nil
	}
	fs.State = StateShipped
	return pm.store.Save(fs)
}

func (pm *FulfillmentPM) HandleOrderDelivered(orderID string) error {
	fs, err := pm.store.FindByOrderID(orderID)
	if err != nil || fs == nil {
		return err
	}
	if fs.State != StateShipped {
		return nil
	}
	now := pm.clock.Now()
	fs.State = StateDelivered
	fs.DeliveredAt = now
	fs.ReturnDeadline = now.Add(ReturnWindowDuration)
	return pm.store.Save(fs)
}

func (pm *FulfillmentPM) HandleReturnRequested(orderID string) error {
	fs, err := pm.store.FindByOrderID(orderID)
	if err != nil || fs == nil {
		return err
	}
	if fs.State != StateDelivered {
		return nil
	}
	if pm.clock.Now().After(fs.ReturnDeadline) {
		return nil
	}
	fs.State = StateReturnRequested
	return pm.store.Save(fs)
}

func (pm *FulfillmentPM) CheckReturnWindow(orderID string) error {
	fs, err := pm.store.FindByOrderID(orderID)
	if err != nil || fs == nil {
		return err
	}
	if fs.State != StateDelivered {
		return nil
	}
	if pm.clock.Now().After(fs.ReturnDeadline) {
		fs.State = StateCompleted
		return pm.store.Save(fs)
	}
	return nil
}

func (pm *FulfillmentPM) HandleReturnCompleted(orderID string) error {
	fs, err := pm.store.FindByOrderID(orderID)
	if err != nil || fs == nil {
		return err
	}
	if fs.State != StateReturnRequested {
		return nil
	}
	fs.State = StateCompleted
	return pm.store.Save(fs)
}
