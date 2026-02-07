package auth

// Role represents a user or device role in the system.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// ValidRoles is the set of all valid roles.
var ValidRoles = map[Role]bool{
	RoleCustomer: true,
	RoleAdmin:    true,
}

// Permission represents a specific action that can be authorized.
type Permission string

const (
	// Public / Customer permissions
	PermOrdersCreate  Permission = "orders:create"
	PermOrdersReadOwn Permission = "orders:read:own"

	// POS permissions
	PermPOSAccess       Permission = "pos:access"
	PermPOSOrdersCreate Permission = "pos:orders:create"
	PermPOSPayments     Permission = "pos:payments"

	// Station permissions
	PermStationAccess Permission = "station:access"
	PermStationVerify Permission = "station:verify"
	PermStationRedeem Permission = "station:redeem"

	// Admin permissions
	PermAdminAccess       Permission = "admin:access"
	PermProductsWrite     Permission = "products:write"
	PermOrdersReadAll     Permission = "orders:read:all"
	PermOrdersExport      Permission = "orders:export"
	PermOrdersStatusWrite Permission = "orders:status:write"
	PermCategoriesWrite   Permission = "categories:write"
	PermMenusWrite        Permission = "menus:write"
	PermStationsManage    Permission = "stations:manage"
	PermPOSManage         Permission = "pos:manage"
	PermUsersRead         Permission = "users:read"
	PermUsersRoleSet      Permission = "users:role:set"
	PermDevicesApprove    Permission = "devices:approve"
	PermDevicesRevoke     Permission = "devices:revoke"
	PermInvitesManage     Permission = "invites:manage"
)

// RolePermissions maps each role to its set of permissions.
// This is the single source of truth for authorization.
var RolePermissions = map[Role]map[Permission]bool{
	RoleCustomer: {
		PermOrdersCreate:  true,
		PermOrdersReadOwn: true,
	},
	RoleAdmin: {
		// Full access
		PermOrdersCreate:      true,
		PermOrdersReadOwn:     true,
		PermPOSAccess:         true,
		PermPOSOrdersCreate:   true,
		PermPOSPayments:       true,
		PermStationAccess:     true,
		PermStationVerify:     true,
		PermStationRedeem:     true,
		PermAdminAccess:       true,
		PermProductsWrite:     true,
		PermOrdersReadAll:     true,
		PermOrdersExport:      true,
		PermOrdersStatusWrite: true,
		PermCategoriesWrite:   true,
		PermMenusWrite:        true,
		PermStationsManage:    true,
		PermPOSManage:         true,
		PermUsersRead:         true,
		PermUsersRoleSet:      true,
		PermDevicesApprove:    true,
		PermDevicesRevoke:     true,
		PermInvitesManage:     true,
	},
}

// DevicePermissions maps device types to their permitted actions.
// Device permissions are separate from user roles.
var DevicePermissions = map[DeviceType]map[Permission]bool{
	DeviceTypePOS: {
		PermPOSAccess:       true,
		PermPOSOrdersCreate: true,
		PermPOSPayments:     true,
	},
	DeviceTypeStation: {
		PermStationAccess: true,
		PermStationVerify: true,
		PermStationRedeem: true,
	},
}

// HasPermission checks if a role has a specific permission.
func HasPermission(role Role, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	return perms[perm]
}

// GetPermissions returns all permissions for a role.
func GetPermissions(role Role) []Permission {
	perms, ok := RolePermissions[role]
	if !ok {
		return nil
	}
	result := make([]Permission, 0, len(perms))
	for p := range perms {
		result = append(result, p)
	}
	return result
}

// HasDevicePermission checks if a device type has a specific permission.
func HasDevicePermission(deviceType DeviceType, perm Permission) bool {
	perms, ok := DevicePermissions[deviceType]
	if !ok {
		return false
	}
	return perms[perm]
}

// IsValidRole checks if a role string is a valid role.
func IsValidRole(role string) bool {
	return ValidRoles[Role(role)]
}
