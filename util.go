package suture

import (
	"errors"
	"sync"
	"time"
)

var supervisorList []*Supervisor
var mu sync.Mutex

func addSupervisor(supervisor *Supervisor) {
	mu.Lock()
	defer mu.Unlock()
	supervisorList = append(supervisorList, supervisor)
}

func removeSupervisor(supervisor *Supervisor) (err error) {
	mu.Lock()
	defer mu.Unlock()
	for i, s := range supervisorList {
		if s.id == supervisor.id {
			// cut out the supervisor, disregard slice ordering
			supervisorList[i] = supervisorList[len(supervisorList)-1]
			supervisorList = supervisorList[:len(supervisorList)-1]
			return
		}
	}
	err = errors.New("Could not find supervisor")
	return
}

// FindService locates a named service from all known supervisors. Returns
// nil if no service is found.
func FindService(id string) Service {
	mu.Lock()
	defer mu.Unlock()
	for _, supervisor := range supervisorList {
		for _, service := range supervisor.services {
			if service.name == id {
				return service.Service
			}
		}
	}
	return nil
}

// WaitForServices waits until the supplied services are available.
// If the full set of services supplied in the Map are found and running
// normally (ServiceNormal), the function returns boolean true. If the
// services are not found in normal state before the time period, the
// function returns false.
func WaitForServices(services map[string]bool, waitDuration time.Duration) bool {
	if services == nil || len(services) == 0 {
		return true
	}
	servicesCopy := make(map[string]bool, len(services))
	for k := range services {
		servicesCopy[k] = true
	}
	timeout := time.Now().Add(waitDuration)
	for {
		for key := range servicesCopy {
			service := FindService(key)
			if service != nil && service.State() == ServiceNormal {
				delete(servicesCopy, key)
			}
		}
		if len(servicesCopy) == 0 {
			return true
		}
		if time.Now().Unix() > timeout.Unix() {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
}
