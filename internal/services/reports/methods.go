package reports

import (
	"context"
	"fmt"

	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/repo"
)

// Create creates a new report
func (s *Service) Create(ctx context.Context, report *structs.Report) (*structs.Report, error) {
	if err := s.hooks.BeforeCreate(ctx, report); err != nil {
		return nil, fmt.Errorf("before create hook failed: %w", err)
	}

	created, err := s.repo.Create(ctx, *report)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	if err := s.hooks.AfterCreate(ctx, created); err != nil {
		return nil, fmt.Errorf("after create hook failed: %w", err)
	}

	return created, nil
}

// FindByID finds a report by ID
func (s *Service) FindByID(ctx context.Context, id string) (*structs.Report, error) {
	report, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find report: %w", err)
	}
	return report, nil
}

// FindOpen finds all open reports
func (s *Service) FindOpen(ctx context.Context) ([]*structs.Report, error) {
	filters := []repo.Filter{
		{Field: "closed_at", Operator: "IS NULL"},
	}
	reports, err := s.repo.Find(ctx, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find open reports: %w", err)
	}
	return reports, nil
}

// FindClosed finds all closed reports
func (s *Service) FindClosed(ctx context.Context) ([]*structs.Report, error) {
	filters := []repo.Filter{
		{Field: "closed_at", Operator: "IS NOT NULL"},
	}
	reports, err := s.repo.Find(ctx, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find closed reports: %w", err)
	}
	return reports, nil
}

// FindByGuild finds reports by guild
func (s *Service) FindByGuild(ctx context.Context, guildID string) ([]*structs.Report, error) {
	filters := []repo.Filter{
		{Field: "guild_id", Operator: "=", Value: guildID},
	}
	reports, err := s.repo.Find(ctx, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find reports by guild: %w", err)
	}
	return reports, nil
}

// Update updates a report
func (s *Service) Update(ctx context.Context, id string, report *structs.Report) (*structs.Report, error) {
	if err := s.hooks.BeforeUpdate(ctx, report); err != nil {
		return nil, fmt.Errorf("before update hook failed: %w", err)
	}

	report.ChannelID = id
	updated, err := s.repo.Update(ctx, id, *report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	if err := s.hooks.AfterUpdate(ctx, updated); err != nil {
		return nil, fmt.Errorf("after update hook failed: %w", err)
	}

	return updated, nil
}

// Delete deletes a report
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.hooks.BeforeDelete(ctx, id); err != nil {
		return fmt.Errorf("before delete hook failed: %w", err)
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	if err := s.hooks.AfterDelete(ctx, id); err != nil {
		return fmt.Errorf("after delete hook failed: %w", err)
	}

	return nil
}

// Close closes a report
func (s *Service) Close(ctx context.Context, id string, closedByID string, closedAt int64) (*structs.Report, error) {
	report, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	report.ClosedBy = &closedByID
	report.ClosedAt = &closedAt

	return s.Update(ctx, id, report)
}

// FindAll finds all reports
func (s *Service) FindAll(ctx context.Context) ([]*structs.Report, error) {
	reports, err := s.repo.Find(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find all reports: %w", err)
	}
	return reports, nil
}

// Search finds reports based on provided filters
func (s *Service) Search(ctx context.Context, filters []repo.Filter) ([]*structs.Report, error) {
	reports, err := s.repo.Find(ctx, filters, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search reports: %w", err)
	}
	return reports, nil
}
