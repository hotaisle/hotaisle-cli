package client

import (
	"context"
	"net/http"
)

// BareMetalService handles bare metal server-related API operations
type BareMetalService struct {
	client *Client
}

// BareMetal returns a new BareMetalService
func (c *Client) BareMetal() *BareMetalService {
	return &BareMetalService{client: c}
}

// List retrieves all bare metal servers for a team
func (s *BareMetalService) List(ctx context.Context, teamHandle string) ([]BareMetalServerDetails, error) {
	path := buildPath("/teams/{team}/bare_metal/", map[string]string{
		"team": teamHandle,
	})
	var result []BareMetalServerDetails
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get retrieves detailed information about a specific bare metal server
func (s *BareMetalService) Get(ctx context.Context, teamHandle, serverName string) (*BareMetalServerDetails, error) {
	path := buildPath("/teams/{team}/bare_metal/{server}/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	var result BareMetalServerDetails
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Reserve reserves a bare metal server for the team
func (s *BareMetalService) Reserve(ctx context.Context, teamHandle string, req BareMetalServerReservation) (*BareMetalServerReservationResponse, error) {
	path := buildPath("/teams/{team}/bare_metal/", map[string]string{
		"team": teamHandle,
	})
	var result BareMetalServerReservationResponse
	err := s.client.doRequest(ctx, http.MethodPost, path, req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a bare metal server's description
func (s *BareMetalService) Update(ctx context.Context, teamHandle, serverName string, update BareMetalServerUpdate) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPatch, path, update, nil)
}

// Delete releases a bare metal server back to the available pool
func (s *BareMetalService) Delete(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// GetAvailable retrieves available bare metal server types
func (s *BareMetalService) GetAvailable(ctx context.Context, teamHandle string) ([]AvailableBareMetalTypes, error) {
	path := buildPath("/teams/{team}/bare_metal/available/", map[string]string{
		"team": teamHandle,
	})
	var result []AvailableBareMetalTypes
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetPowerState retrieves the current power state of a server
func (s *BareMetalService) GetPowerState(ctx context.Context, teamHandle, serverName string) (*BareMetalServerPowerState, error) {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	var result BareMetalServerPowerState
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PowerOn turns on a server
func (s *BareMetalService) PowerOn(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/power_on/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// GracefulShutdown sends an ACPI signal to initiate a clean shutdown
func (s *BareMetalService) GracefulShutdown(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/graceful_shutdown/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// ForceShutdown immediately powers off the server
func (s *BareMetalService) ForceShutdown(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/force_shutdown/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// WarmReboot reboots the system without turning the power off completely
func (s *BareMetalService) WarmReboot(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/warm_reboot/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// ColdReboot turns off and then reboots the system
func (s *BareMetalService) ColdReboot(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/cold_reboot/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// ACReset performs a complete AC reset of the server
func (s *BareMetalService) ACReset(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/power/ac_reset/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, nil, nil)
}

// Reinstall resets BIOS settings, wipes all disks, and reinstalls the OS
func (s *BareMetalService) Reinstall(ctx context.Context, teamHandle, serverName string) (*BareMetalServerDetails, error) {
	path := buildPath("/teams/{team}/bare_metal/{server}/reinstall/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	var result BareMetalServerDetails
	err := s.client.doRequest(ctx, http.MethodPost, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConsoleURL generates a URL for bare metal console access
func (s *BareMetalService) GetConsoleURL(ctx context.Context, teamHandle, serverName string) (*BareMetalServerConsoleURL, error) {
	path := buildPath("/teams/{team}/bare_metal/{server}/console/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	var result BareMetalServerConsoleURL
	err := s.client.doRequest(ctx, http.MethodPost, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// EnableSupportAccess enables Hot Aisle support staff to access the server
func (s *BareMetalService) EnableSupportAccess(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/support_access_enable/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodPut, path, nil, nil)
}

// DisableSupportAccess revokes Hot Aisle support staff access to the server
func (s *BareMetalService) DisableSupportAccess(ctx context.Context, teamHandle, serverName string) error {
	path := buildPath("/teams/{team}/bare_metal/{server}/support_access_enable/", map[string]string{
		"team":   teamHandle,
		"server": serverName,
	})
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
