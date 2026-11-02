package main

import (
	"errors"
	"sync"
)

var (
	ErrTruckNotFound = errors.New("truck not found")
)

type FleetManager interface {
	AddTruck(id string, cargo int) error
	GetTruck(id string) (Truck, error)
	RemoveTruck(id string) error
	UpdateTruckCargo(id string, cargo int) error
}

type Truck struct {
	ID    string
	Cargo int
}

type TruckManager struct {
	trucks map[string]*Truck
	sync.RWMutex
}

func (t *TruckManager) AddTruck(id string, cargo int) error {
	t.Lock()
	defer t.Unlock()

	t.trucks[id] = &Truck{ID: id, Cargo: cargo}
	return nil
}

func (t *TruckManager) GetTruck(id string) (Truck, error) {
	t.RLock()
	defer t.RUnlock()

	if truck, ok := t.trucks[id]; !ok {
		return Truck{}, ErrTruckNotFound
	} else {
		return *truck, nil
	}
}

func (t *TruckManager) RemoveTruck(id string) error {
	t.Lock()
	defer t.Unlock()

	delete(t.trucks, id)
	return nil
}

func (t *TruckManager) UpdateTruckCargo(id string, cargo int) error {
	t.Lock()
	defer t.Unlock()

	t.trucks[id].Cargo = cargo
	return nil
}

func NewTruckManager() TruckManager {
	return TruckManager{
		trucks: make(map[string]*Truck),
	}
}
