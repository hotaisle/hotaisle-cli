package client

import (
	"context"
	"net/http"
)

// VirtualMachinesService handles virtual machine-related API operations
type VirtualMachinesService struct {
	client *Client
}

// VirtualMachines returns a new VirtualMachinesService
func (c *Client) VirtualMachines() *VirtualMachinesService {
	return &VirtualMachinesService{client: c}
}

// List retrieves all virtual machines for a team
func (s *VirtualMachinesService) List(ctx context.Context, teamHandle string) ([]VirtualMachineDetails, error) {
	path := buildPath("/teams/{team}/virtual_machines/", map[string]string{
		"team": teamHandle,
	})
	var result []VirtualMachineDetails
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get retrieves detailed information about a specific virtual machine
func (s *VirtualMachinesService) Get(ctx context.Context, teamHandle, vmName string) (*VirtualMachineDetails, error) {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	var result VirtualMachineDetails
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Provision assigns and provisions a virtual machine
func (s *VirtualMachinesService) Provision(ctx context.Context, teamHandle string, specs VirtualMachineSpecs) (*VirtualMachineDetails, error) {
	path := buildPath("/teams/{team}/virtual_machines/", map[string]string{
		"team": teamHandle,
	})
	var result VirtualMachineDetails
	err := s.client.doRequest(ctx, http.MethodPost, path, specs, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a virtual machine's description
func (s *VirtualMachinesService) Update(ctx context.Context, teamHandle, vmName string, update VirtualMachineUpdate) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPatch, path, update, nil)
}

// Delete completely deletes a virtual machine and all its resources
func (s *VirtualMachinesService) Delete(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// GetAvailable retrieves available virtual machine types
func (s *VirtualMachinesService) GetAvailable(ctx context.Context, teamHandle string) ([]AvailableVirtualMachineTypes, error) {
	path := buildPath("/teams/{team}/virtual_machines/available/", map[string]string{
		"team": teamHandle,
	})
	var result []AvailableVirtualMachineTypes
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetState retrieves the current power state of a virtual machine
func (s *VirtualMachinesService) GetState(ctx context.Context, teamHandle, vmName string) (*VirtualMachineState, error) {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/state/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	var result VirtualMachineState
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Start starts a virtual machine that is currently stopped
func (s *VirtualMachinesService) Start(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/start/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// Stop forcefully stops a running virtual machine
func (s *VirtualMachinesService) Stop(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/stop/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// Shutdown sends a graceful shutdown signal to a running virtual machine
func (s *VirtualMachinesService) Shutdown(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/shutdown/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// Reboot gracefully reboots a running virtual machine
func (s *VirtualMachinesService) Reboot(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/reboot/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// HardReset forcefully resets a running virtual machine
func (s *VirtualMachinesService) HardReset(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/hard-reset/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// Rebuild performs a complete rebuild of the virtual machine to its initial state
func (s *VirtualMachinesService) Rebuild(ctx context.Context, teamHandle, vmName string) error {
	path := buildPath("/teams/{team}/virtual_machines/{vm}/rebuild/", map[string]string{
		"team": teamHandle,
		"vm":   vmName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}
