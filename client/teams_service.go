package client

import (
	"context"
	"net/http"
)

// TeamsService handles team-related API operations
type TeamsService struct {
	client *Client
}

// Teams returns a new TeamsService
func (c *Client) Teams() *TeamsService {
	return &TeamsService{client: c}
}

// List retrieves all teams the user belongs to
func (s *TeamsService) List(ctx context.Context) ([]UserTeam, error) {
	var result []UserTeam
	err := s.client.doRequest(ctx, http.MethodGet, "/teams/", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Create creates a new team
func (s *TeamsService) Create(ctx context.Context, team Team) (*UserTeamWithMembers, error) {
	var result UserTeamWithMembers
	err := s.client.doRequest(ctx, http.MethodPost, "/teams/", team, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves detailed information about a specific team
func (s *TeamsService) Get(ctx context.Context, teamHandle string) (*UserTeamDetails, error) {
	path := buildPath("/teams/{team}/", map[string]string{
		"team": teamHandle,
	})
	var result UserTeamDetails
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a team's information
func (s *TeamsService) Update(ctx context.Context, teamHandle string, update TeamUpdate) (*UserTeamWithMembers, error) {
	path := buildPath("/teams/{team}/", map[string]string{
		"team": teamHandle,
	})
	var result UserTeamWithMembers
	err := s.client.doRequest(ctx, http.MethodPatch, path, update, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetInvitations retrieves pending team invitations for the user
func (s *TeamsService) GetInvitations(ctx context.Context) ([]UserTeam, error) {
	var result []UserTeam
	err := s.client.doRequest(ctx, http.MethodGet, "/teams/invitations/", nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// AcceptInvitation accepts a pending team invitation
func (s *TeamsService) AcceptInvitation(ctx context.Context, teamHandle string) (*UserTeamWithMembers, error) {
	path := buildPath("/teams/{team}/accept-invitation/", map[string]string{
		"team": teamHandle,
	})
	var result UserTeamWithMembers
	err := s.client.doRequest(ctx, http.MethodPost, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBalance retrieves team balance information
func (s *TeamsService) GetBalance(ctx context.Context, teamHandle string) (*BalanceInfo, error) {
	path := buildPath("/teams/{team}/balance/", map[string]string{
		"team": teamHandle,
	})
	var result BalanceInfo
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PurchaseCredits creates a Stripe checkout session for purchasing credits
func (s *TeamsService) PurchaseCredits(ctx context.Context, teamHandle string, req PurchaseTeamCreditsRequest) (*PurchaseTeamCreditsResponse, error) {
	path := buildPath("/teams/{team}/purchase-credits/", map[string]string{
		"team": teamHandle,
	})
	var result PurchaseTeamCreditsResponse
	err := s.client.doRequest(ctx, http.MethodPost, path, req, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RequestPaymentApproval requests self-service payment approval for a team
func (s *TeamsService) RequestPaymentApproval(ctx context.Context, teamHandle string, req RequestPaymentApprovalRequest) error {
	path := buildPath("/teams/{team}/request-payment-approval/", map[string]string{
		"team": teamHandle,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, req, nil)
}

// GetTeamInvitations retrieves pending invitations for a team
func (s *TeamsService) GetTeamInvitations(ctx context.Context, teamHandle string) ([]TeamMember, error) {
	path := buildPath("/teams/{team}/members/invitations/", map[string]string{
		"team": teamHandle,
	})
	var result []TeamMember
	err := s.client.doRequest(ctx, http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// InviteMember invites a user to join the team
func (s *TeamsService) InviteMember(ctx context.Context, teamHandle string, req TeamInvitationRequest) error {
	path := buildPath("/teams/{team}/members/invitations/", map[string]string{
		"team": teamHandle,
	})
	return s.client.doRequest(ctx, http.MethodPost, path, req, nil)
}

// UpdateMember updates a team member's roles
func (s *TeamsService) UpdateMember(ctx context.Context, teamHandle, email string, update TeamMemberUpdate) (*TeamMember, error) {
	path := buildPath("/teams/{team}/members/{email}/", map[string]string{
		"team":  teamHandle,
		"email": email,
	})
	var result TeamMember
	err := s.client.doRequest(ctx, http.MethodPatch, path, update, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveMember removes a member from a team
func (s *TeamsService) RemoveMember(ctx context.Context, teamHandle, email string) error {
	path := buildPath("/teams/{team}/members/{email}/", map[string]string{
		"team":  teamHandle,
		"email": email,
	})
	return s.client.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
