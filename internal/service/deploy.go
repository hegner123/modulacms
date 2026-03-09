package service

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// DeployService wraps the deploy package functions with dependency injection.
type DeployService struct {
	driver db.DbDriver
	mgr    *config.Manager
}

// NewDeployService creates a DeployService.
func NewDeployService(driver db.DbDriver, mgr *config.Manager) *DeployService {
	return &DeployService{driver: driver, mgr: mgr}
}
