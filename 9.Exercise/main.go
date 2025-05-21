package main

import "errors"

var (
	ErrTruckNotFound = errors.New("truck not found")
)

type FleetManager interface {
	AddTruck(id string, cargo int) error
	GetTruck(id string) (*Truck, error)
	RemoveTruck(id string) error
	UpdateTruckCargo(id string, cargo int) error
}

type Truck struct {
	ID    string
	Cargo int
}

type TruckManager struct {
	trucks map[string]*Truck
}

func (t *TruckManager) AddTruck(id string, cargo int) error {
	t.trucks[id] = &Truck{ID: id, Cargo: cargo}
	return nil
}

func (t *TruckManager) GetTruck(id string) (*Truck, error) {
	if truck, ok := t.trucks[id]; !ok {
		return nil, ErrTruckNotFound
	} else {
		return truck, nil
	}
}

func (t *TruckManager) RemoveTruck(id string) error {
	if _, err := t.GetTruck(id); err != nil {
		return ErrTruckNotFound
	}
	delete(t.trucks, id)
	return nil
}

func (t *TruckManager) UpdateTruckCargo(id string, cargo int) error {
	if truck, err := t.GetTruck(id); err != nil {
		return ErrTruckNotFound
	} else {
		truck.Cargo = cargo
		t.trucks[id] = truck
	}
	return nil
}

func NewTruckManager() TruckManager {
	return TruckManager{
		trucks: make(map[string]*Truck),
	}
}
