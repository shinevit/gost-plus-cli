package entitymanager

import (
	"fmt"

	"github.com/go-gost/gost.plus/tunnel"
)

type EntityGetterFn func(int) tunnel.Tunnel

type EntitySource struct {
	Count     int
	GetEntity EntityGetterFn
}

// Common manager interface for the both tunnel.Tunnel and entrypoint.EntryPoint
type INetworkEntityManager interface {
	DeleteEntity(label string)
}

type NetworkEntityManager struct {
	ID         string
	Entity     *tunnel.Tunnel
	Delete     func(string)
	SaveConfig func() error
}

func (em *NetworkEntityManager) DeleteEntity(label string) error {
	if em.Entity == nil {
		errMsg := fmt.Sprintf("%s with ID '%s' not found.\n", label, em.ID)
		return fmt.Errorf(errMsg)
	}

	entity := *em.Entity
	if entity == nil {
		errMsg := fmt.Sprintf("entity is nil for %s with ID '%s'\n", label, em.ID)
		return fmt.Errorf(errMsg)
	}

	if em.Delete != nil {
		em.Delete(em.ID)
	}

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
	fmt.Printf("\n%s '%s' (ID: %s) has been deleted.\n", label, name, em.ID)
	return nil
}
