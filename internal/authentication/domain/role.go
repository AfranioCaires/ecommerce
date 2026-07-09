package domain

import "errors"

var ErrInvalidRole = errors.New("the role is invalid.")

type Role string

const (
	RoleCustomer      Role = "CUSTOMER"
	RoleAdministrator Role = "ADMIN"
	RoleSupport       Role = "SUPPORT"
)

func (role Role) IsValid() bool {
	switch role {
	case RoleCustomer, RoleAdministrator, RoleSupport:
		return true
	default:
		return false
	}
}
