package models

const (
	RoleBoss    = "sef"
	RoleManager = "menadzer"
	RoleWorker  = "radnik"
)

func IsValidRole(role string) bool {
	switch role {
	case RoleBoss, RoleManager, RoleWorker:
		return true
	default:
		return false
	}
}

func CanViewSensitiveProductFields(role string) bool {
	return role == RoleBoss || role == RoleManager
}

func CanAccessFinancialReports(role string) bool {
	return role == RoleBoss || role == RoleManager
}
