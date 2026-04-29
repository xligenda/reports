package reports

import (
	"github.com/jmoiron/sqlx"
	"github.com/xligenda/reports/internal/services/hooks"
	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/repo"
)

// Service handles report business logic
type Service struct {
	repo  *repo.GenericRepository[string, structs.Report]
	hooks hooks.Hooks
}

// NewService creates a new report service
func NewService(db *sqlx.DB, h hooks.Hooks) *Service {
	if h == nil {
		h = &hooks.NoOpHooks{}
	}
	return &Service{
		repo:  repo.NewRepository[string, structs.Report](db, "reports"),
		hooks: h,
	}
}

// NewServiceWithRepository creates a service with custom repository (useful for testing)
func NewServiceWithRepository(r *repo.GenericRepository[string, structs.Report], h hooks.Hooks) *Service {
	if h == nil {
		h = &hooks.NoOpHooks{}
	}
	return &Service{
		repo:  r,
		hooks: h,
	}
}
