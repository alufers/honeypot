package main

import "gorm.io/gorm"

type CredentialUsage struct {
	gorm.Model
	Protocol string `json:"protocol"`
	Username string `json:"username"`
	Password string `json:"password"`
}
