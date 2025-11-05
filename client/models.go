package client

import "time"

// User represents a registered user in the system
type User struct {
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
}

// UserUpdate represents the updateable fields for a user
type UserUpdate struct {
	Name string `json:"name"`
}

// GetUserResponse represents information about the authenticated user and their teams
type GetUserResponse struct {
	User  User       `json:"user"`
	Teams []UserTeam `json:"teams"`
}

// Team represents a team
type Team struct {
	Handle                    string `json:"handle"`
	Name                      string `json:"name"`
	Description               string `json:"description,omitempty"`
	SelfServicePaymentEnabled bool   `json:"self_service_payment_enabled,omitempty"`
	MaximumVirtualMachines    int64  `json:"maximum_virtual_machines,omitempty"`
	MaximumBareMetalServers   int64  `json:"maximum_bare_metal_servers,omitempty"`
}

// TeamUpdate represents a team update request
type TeamUpdate struct {
	Handle      string `json:"handle"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UserTeam represents a team that the user belongs to
type UserTeam struct {
	Team
	Roles          []string `json:"roles"`
	EffectiveRoles []string `json:"effective_roles"`
	Invitation     bool     `json:"invitation,omitempty"`
}

// UserTeamWithMembers represents a team with detailed member information
type UserTeamWithMembers struct {
	UserTeam
	Members []TeamMember `json:"members,omitempty"`
}

// UserTeamDetails represents a team with detailed information including members and resources
type UserTeamDetails struct {
	UserTeamWithMembers
	BareMetalServers []BareMetalServer `json:"bare_metal_servers,omitempty"`
	VirtualMachines  []VirtualMachine  `json:"virtual_machines,omitempty"`
}

// TeamMember represents a team member or pending invitation
type TeamMember struct {
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Created    time.Time `json:"created"`
	Roles      []string  `json:"roles"`
	Invitation bool      `json:"invitation,omitempty"`
}

// TeamMemberUpdate represents the data needed to update a team member's roles
type TeamMemberUpdate struct {
	Roles []string `json:"roles"`
}

// TeamInvitationRequest contains the information needed to invite a user to a team
type TeamInvitationRequest struct {
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

// BalanceInfo contains balance and estimated time until depletion
type BalanceInfo struct {
	AvailableBalance    int64      `json:"available_balance"`
	HourlyRate          int64      `json:"hourly_rate"`
	EstimatedRunoutTime *time.Time `json:"estimated_runout_time,omitempty"`
	MinimumBalance      int64      `json:"minimum_balance,omitempty"`
}

// PurchaseTeamCreditsRequest represents a request to purchase credits for a team
type PurchaseTeamCreditsRequest struct {
	Cents int64 `json:"cents"`
}

// PurchaseTeamCreditsResponse represents the response to a credits purchase request
type PurchaseTeamCreditsResponse struct {
	CheckoutURL string    `json:"checkout_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// RequestPaymentApprovalRequest represents a request for self-service payment approval
type RequestPaymentApprovalRequest struct {
	Message string `json:"message"`
}

// SSHKey represents an SSH public key associated with a user
type SSHKey struct {
	Type        string `json:"type"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment,omitempty"`
}

// SSHKeyRequest represents the data needed to add a new SSH key
type SSHKeyRequest struct {
	AuthorizedKey string `json:"authorized_key"`
}

// UserAPIKey represents a user's API key
type UserAPIKey struct {
	Prefix   string       `json:"prefix,omitempty"`
	Label    string       `json:"label,omitempty"`
	UserRole string       `json:"user_role"`
	Teams    []APIKeyTeam `json:"teams,omitempty"`
}

// APIKeyTeam represents a team that the API Key has access to
type APIKeyTeam struct {
	Team
	Roles []string `json:"roles"`
}

// UserAPIKeyRequest represents the data needed to create or update an API key
type UserAPIKeyRequest struct {
	Label    string                `json:"label,omitempty"`
	UserRole string                `json:"user_role,omitempty"`
	Teams    []UserAPIKeyTeamRoles `json:"teams,omitempty"`
}

// UserAPIKeyTeamRoles defines the permissions for an API key on a specific team
type UserAPIKeyTeamRoles struct {
	Team  string   `json:"team"`
	Roles []string `json:"roles"`
}

// UserAPIKeyWithToken represents an API key with its full token value
type UserAPIKeyWithToken struct {
	UserAPIKey
	Token string `json:"token,omitempty"`
}

// Components represents common attributes for hardware components
type Components struct {
	Count        uint64 `json:"count"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}

// CPUs represents processor information
type CPUs struct {
	Components
	Cores     uint64 `json:"cores"`
	Frequency uint64 `json:"frequency"`
}

// Disks represents storage device information
type Disks struct {
	Components
	Type     string `json:"type"`
	Capacity uint64 `json:"capacity"`
}

// GPUs represents graphics processing unit information
type GPUs struct {
	Components
}

// MemoryModules represents memory module information
type MemoryModules struct {
	Components
	Capacity uint64 `json:"capacity"`
}

// ExternalService represents information about an externally accessible service
type ExternalService struct {
	IPAddress string `json:"ip_address"`
	Port      int64  `json:"port"`
	DNSName   string `json:"dns_name,omitempty"`
}

// BareMetalServer represents a bare metal server
type BareMetalServer struct {
	Name                 string           `json:"name"`
	IPAddress            string           `json:"ip_address"`
	Manufacturer         string           `json:"manufacturer"`
	Model                string           `json:"model"`
	Description          string           `json:"description,omitempty"`
	SSHAccess            *ExternalService `json:"ssh_access,omitempty"`
	SupportAccessEnabled bool             `json:"support_access_enabled,omitempty"`
}

// BareMetalServerSpecs contains the hardware specifications of a server
type BareMetalServerSpecs struct {
	CPUCores      uint64          `json:"cpu_cores"`
	RAMCapacity   uint64          `json:"ram_capacity"`
	DiskCapacity  uint64          `json:"disk_capacity"`
	CPUs          []CPUs          `json:"cpus,omitempty"`
	Disks         []Disks         `json:"disks,omitempty"`
	GPUs          []GPUs          `json:"gpus,omitempty"`
	MemoryModules []MemoryModules `json:"memory_modules,omitempty"`
}

// BareMetalServerlOSStatus represents the current state of OS installation on a server
type BareMetalServerlOSStatus struct {
	OSSelection       string    `json:"os_selection"`
	OSStatus          string    `json:"os_install_status"`
	LastImagingUpdate time.Time `json:"last_imaging_update"`
}

// BareMetalServerDetails represents a bare metal server with detailed hardware specifications
type BareMetalServerDetails struct {
	BareMetalServer
	BareMetalServerSpecs
	OSStatus *BareMetalServerlOSStatus `json:"os_status,omitempty"`
}

// BareMetalServerUpdate represents a bare metal server update request
type BareMetalServerUpdate struct {
	Description string `json:"description,omitempty"`
}

// BareMetalServerReservation represents a request to reserve a bare metal server
type BareMetalServerReservation struct {
	Description string               `json:"description,omitempty"`
	Specs       BareMetalServerSpecs `json:"specs"`
}

// BareMetalServerReservationResponse represents the response after successfully reserving a server
type BareMetalServerReservationResponse struct {
	BareMetalServer
	CPUCores      uint64                    `json:"cpu_cores"`
	RAMCapacity   uint64                    `json:"ram_capacity"`
	DiskCapacity  uint64                    `json:"disk_capacity"`
	CPUs          []CPUs                    `json:"cpus,omitempty"`
	Disks         []Disks                   `json:"disks,omitempty"`
	GPUs          []GPUs                    `json:"gpus,omitempty"`
	MemoryModules []MemoryModules           `json:"memory_modules,omitempty"`
	OSStatus      *BareMetalServerlOSStatus `json:"os_status,omitempty"`
}

// BareMetalServerPowerState represents the current power status of a server
type BareMetalServerPowerState struct {
	State string `json:"state"`
}

// BareMetalServerConsoleURL has a temporary URL that can be used to access a local console
type BareMetalServerConsoleURL struct {
	URL string `json:"url"`
}

// AvailableBareMetalTypes is how many bare metal servers of a given type are available to be reserved
type AvailableBareMetalTypes struct {
	Quantity                  int64                `json:"Quantity"`
	MinimumReservationMinutes int64                `json:"MinimumReservationMinutes"`
	OnDemandPrice             int64                `json:"OnDemandPrice,omitempty"`
	Specs                     BareMetalServerSpecs `json:"Specs,omitempty"`
}

// VirtualMachine represents a virtual machine instance
type VirtualMachine struct {
	Name        string           `json:"name"`
	IPAddress   string           `json:"ip_address"`
	Description string           `json:"description,omitempty"`
	SSHAccess   *ExternalService `json:"ssh_access,omitempty"`
}

// VirtualMachineSpecs contains the specifications of a virtual machine
type VirtualMachineSpecs struct {
	CPUCores     uint64 `json:"cpu_cores"`
	RAMCapacity  uint64 `json:"ram_capacity"`
	DiskCapacity uint64 `json:"disk_capacity"`
	CPUs         *CPUs  `json:"cpus,omitempty"`
	GPUs         []GPUs `json:"gpus,omitempty"`
}

// VirtualMachineDetails represents a virtual machine with detailed specifications
type VirtualMachineDetails struct {
	VirtualMachine
	VirtualMachineSpecs
}

// VirtualMachineUpdate represents a virtual machine update request
type VirtualMachineUpdate struct {
	Description string `json:"description,omitempty"`
}

// VirtualMachineState represents the current state of a virtual machine
type VirtualMachineState struct {
	State string `json:"state"`
	Host  string `json:"host"`
}

// AvailableVirtualMachineTypes is how many virtual machines of a given type are available to be deployed
type AvailableVirtualMachineTypes struct {
	Quantity                  int64               `json:"Quantity"`
	MinimumReservationMinutes int64               `json:"MinimumReservationMinutes"`
	OnDemandPrice             int64               `json:"OnDemandPrice,omitempty"`
	Specs                     VirtualMachineSpecs `json:"Specs,omitempty"`
}
