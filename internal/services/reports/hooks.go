package reports

import (
	"context"
	"fmt"

	"github.com/xligenda/reports/internal/services/hooks"
	"github.com/xligenda/reports/internal/structs"
)

// DefaultHooks provides default validation hooks for reports
type DefaultHooks struct{}

func NewDefaultHooks() hooks.Hooks {
	return &DefaultHooks{}
}

func (h *DefaultHooks) BeforeCreate(ctx context.Context, entity any) error {
	report, ok := entity.(*structs.Report)
	if !ok {
		return nil
	}

	if report.Channel == "" {
		return fmt.Errorf("report channel cannot be empty")
	}
	if report.Guild == "" {
		return fmt.Errorf("report guild cannot be empty")
	}
	if report.Issuer == "" {
		return fmt.Errorf("report issuer cannot be empty")
	}

	return nil
}

func (h *DefaultHooks) AfterCreate(ctx context.Context, entity any) error {
	return nil
}

func (h *DefaultHooks) BeforeUpdate(ctx context.Context, entity any) error {
	return nil
}

func (h *DefaultHooks) AfterUpdate(ctx context.Context, entity any) error {
	return nil
}

func (h *DefaultHooks) BeforeDelete(ctx context.Context, id any) error {
	reportID, ok := id.(string)
	if !ok {
		return fmt.Errorf("invalid report ID type")
	}
	if reportID == "" {
		return fmt.Errorf("report ID cannot be empty")
	}
	return nil
}

func (h *DefaultHooks) AfterDelete(ctx context.Context, id any) error {
	return nil
}
