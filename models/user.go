package models

import "time"

type AuthUser struct {
	BaseModel
	Username    string    `json:"username" gorm:"index"`
	LastLogin   time.Time `json:"last_login"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	EmployeeID  string    `json:"employee_id"`
	Type        string    `json:"type" gorm:"default:'user'"`
	IsSuperuser bool      `json:"is_superuser" sql:"default:false"`
	IsApprover  bool      `json:"is_approver" sql:"default:false"`
}
