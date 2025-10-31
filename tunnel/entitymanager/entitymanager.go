package entitymanager

import (
	"fmt"

	"github.com/go-gost/gost.plus/tunnel"
)

// Types declarations
type EntityIteratorFn func() (NetworkEntityManager, bool)
type EntityGetterFn func(int) tunnel.Tunnel

type EntitySource struct {
	Count     int
	GetEntity EntityGetterFn
}

// Common manager types
type INetworkEntityManager interface {
	DeleteEntity(label string)
}
type NetworkEntityManager struct {
	ID         string
	Entity     *tunnel.Tunnel // interface for the both tunnel.Tunnel and entrypoint.EntryPoint
	Delete     func(string)
	SaveConfig func() error
}

func (em *NetworkEntityManager) DeleteEntity(label string) error {
	// Check if Entity is nil
	if em.Entity == nil {
		errMsg := fmt.Sprintf("%s with ID '%s' not found.\n", label, em.ID)
		return fmt.Errorf(errMsg)
	}

	entity := *em.Entity
	if entity == nil {
		errMsg := fmt.Sprintf("entity is nil for %s with ID '%s'\n", label, em.ID)
		return fmt.Errorf(errMsg)
	}

	// Call Delete if it's set
	if em.Delete != nil {
		em.Delete(em.ID)
	}

	// Ensure SaveConfig is set
	if em.SaveConfig == nil {
		errMsg := "SaveConfig function is not set"
		return fmt.Errorf(errMsg)
	}

	if err := em.SaveConfig(); err != nil {
		errMsg := fmt.Sprintf("Error while saving config after deletion: %v\n", err)
		fmt.Print(errMsg)
		return fmt.Errorf(errMsg)
	}

	name := entity.Name()
	fmt.Printf("%s '%s' (ID: %s) has been deleted.\n", label, name, em.ID)
	return nil
}

func MakeEntityIteratorFrom(source EntitySource) EntityIteratorFn {
	return MakeEntityIterator(source.Count, source.GetEntity)
}

func MakeEntityIterator(count int, getItem EntityGetterFn) EntityIteratorFn {
	i := 0
	return func() (NetworkEntityManager, bool) {
		for i < count {
			entity := getItem(i)
			i++
			if entity == nil {
				continue // skip an iteration
			}
			return NetworkEntityManager{
				ID:     entity.ID(),
				Entity: &entity,
			}, true
		}
		return NetworkEntityManager{}, false // stop iterations
	}
}
