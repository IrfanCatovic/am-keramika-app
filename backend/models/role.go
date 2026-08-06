package models

const (
	RoleDeveloper = "developer"
	RoleBoss      = "sef"
	RoleManager   = "menadzer"
	RoleWorker    = "radnik"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleDeveloper, RoleBoss, RoleManager, RoleWorker:
		return true
	default:
		return false
	}
}

// IsAssignableUserRole su uloge koje smiju biti dodijeljene kroz POST/PUT /users.
// Developer se kreira samo bootstrapom, ne kroz obični user-management API.
func IsAssignableUserRole(role string) bool {
	switch role {
	case RoleBoss, RoleManager, RoleWorker:
		return true
	default:
		return false
	}
}

func CanViewSensitiveProductFields(role string) bool {
	return role == RoleDeveloper || role == RoleBoss || role == RoleManager
}

func CanAccessFinancialReports(role string) bool {
	return role == RoleDeveloper || role == RoleBoss || role == RoleManager
}
