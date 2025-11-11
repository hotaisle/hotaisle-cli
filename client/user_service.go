package client

import (
	"context"
	"net/http"
)

// UserService handles user-related API operations
type UserService struct {
	client *Client
}

// User returns a new UserService
func (c *Client) User() *UserService {
	return &UserService{client: c}
}

// Get retrieves information about the currently authenticated user
func (s *UserService) Get(ctx context.Context) (*GetUserResponse, error) {
	var result GetUserResponse
	err := s.client.doRequest(ctx, http.MethodGet, "/user/", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates the currently authenticated user profile
func (s *UserService) Update(ctx context.Context, update UserUpdate) (*User, error) {
	var result User
	err := s.client.doRequest(ctx, http.MethodPatch, "/user/", update, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetSSHKeys retrieves all SSH keys for the user
func (s *UserService) GetSSHKeys(ctx context.Context) ([]SSHKey, error) {
	var result []SSHKey
	err := s.client.doRequest(ctx, http.MethodGet, "/user/ssh_keys/", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AddSSHKey adds a new SSH key to the user's account
func (s *UserService) AddSSHKey(ctx context.Context, key SSHKeyRequest) (*SSHKey, error) {
	var result SSHKey
	err := s.client.doRequest(ctx, http.MethodPost, "/user/ssh_keys/", key, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSSHKey permanently deletes an SSH key from the user's account
func (s *UserService) DeleteSSHKey(ctx context.Context, fingerprint string) error {
	path := buildPath("/user/ssh_keys/{fingerprint}/", map[string]string{
		"fingerprint": fingerprint,
	})
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
}

// GetAPIKeys retrieves all API keys for the user
func (s *UserService) GetAPIKeys(ctx context.Context) ([]UserAPIKey, error) {
	var result []UserAPIKey
	err := s.client.doRequest(ctx, http.MethodGet, "/user/api_keys/", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAPIKey retrieves detailed information about a specific API key
func (s *UserService) GetAPIKey(ctx context.Context, prefix string) (*UserAPIKey, error) {
	path := buildPath("/user/api_keys/{prefix}/", map[string]string{
		"prefix": prefix,
	})
	var result UserAPIKey
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateAPIKey creates a new API key with specified permissions
func (s *UserService) CreateAPIKey(ctx context.Context, req UserAPIKeyRequest) (*UserAPIKeyWithToken, error) {
	var result UserAPIKeyWithToken
	err := s.client.doRequest(ctx, http.MethodPost, "/user/api_keys/", req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateAPIKey updates an existing API key
func (s *UserService) UpdateAPIKey(ctx context.Context, prefix string, req UserAPIKeyRequest) (*UserAPIKey, error) {
	path := buildPath("/user/api_keys/{prefix}/", map[string]string{
		"prefix": prefix,
	})
	var result UserAPIKey
	err := s.client.doRequest(ctx, http.MethodPatch, path, req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteAPIKey permanently deletes an API key
func (s *UserService) DeleteAPIKey(ctx context.Context, prefix string) error {
	path := buildPath("/user/api_keys/{prefix}/", map[string]string{
		"prefix": prefix,
	})
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
