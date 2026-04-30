package reports

import (
	"context"
	"fmt"

	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/repo"
)

// Create creates a new report.
func (s *Service) Create(
	ctx context.Context,
	channelID, guildID, issuerID, topic string,
	createdAt int64,
	note, proof *string,
) (*structs.Report, error) {
	if channelID == "" {
		return nil, fmt.Errorf("channel_id is required")
	}
	if guildID == "" {
		return nil, fmt.Errorf("guild_id is required")
	}
	if issuerID == "" {
		return nil, fmt.Errorf("issuer_id is required")
	}
	if topic == "" {
		return nil, fmt.Errorf("topic is required")
	}
	if createdAt <= 0 {
		return nil, fmt.Errorf("created_at is required")
	}

	report := &structs.Report{
		ChannelID: channelID,
		GuildID:   guildID,
		IssuerID:  issuerID,
		Topic:     topic,
		CreatedAt: createdAt,
		Note:      note,
		Proof:     proof,
	}

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

// FindByID finds a report by ID.
func (s *Service) FindByID(ctx context.Context, id string) (*structs.Report, error) {
	report, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find report: %w", err)
	}
	return report, nil
}

// FindAll finds all reports with optional query options.
func (s *Service) FindAll(ctx context.Context, opts *repo.QueryOptions) ([]*structs.Report, error) {
	reports, err := s.repo.Find(ctx, nil, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find all reports: %w", err)
	}
	return reports, nil
}

// Search finds reports based on provided filters and optional query options.
func (s *Service) Search(ctx context.Context, filters []repo.Filter, opts *repo.QueryOptions) ([]*structs.Report, error) {
	reports, err := s.repo.Find(ctx, filters, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search reports: %w", err)
	}
	return reports, nil
}

// FindPaginated retrieves a page of reports matching the filters and returns the total count.
func (s *Service) FindPaginated(ctx context.Context, filters []repo.Filter, page, pageSize int, orderBy repo.OrderBy) ([]*structs.Report, int64, error) {
	reports, total, err := s.repo.FindWithPagination(ctx, filters, page, pageSize, orderBy)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find paginated reports: %w", err)
	}
	return reports, total, nil
}

// Count returns the number of reports matching the given filters.
func (s *Service) Count(ctx context.Context, filters []repo.Filter) (int64, error) {
	count, err := s.repo.Count(ctx, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to count reports: %w", err)
	}
	return count, nil
}

// Exists reports whether any report matches the given filters.
func (s *Service) Exists(ctx context.Context, filters []repo.Filter) (bool, error) {
	exists, err := s.repo.Exists(ctx, filters)
	if err != nil {
		return false, fmt.Errorf("failed to check report existence: %w", err)
	}
	return exists, nil
}

// Update updates a report by ID.
func (s *Service) Update(ctx context.Context, id string, report *structs.Report) (*structs.Report, error) {
	if err := s.hooks.BeforeUpdate(ctx, report); err != nil {
		return nil, fmt.Errorf("before update hook failed: %w", err)
	}

	updated, err := s.repo.Update(ctx, id, *report)
	if err != nil {
		return nil, fmt.Errorf("failed to update report: %w", err)
	}

	if err := s.hooks.AfterUpdate(ctx, updated); err != nil {
		return nil, fmt.Errorf("after update hook failed: %w", err)
	}

	return updated, nil
}

// Delete deletes a report by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.hooks.BeforeDelete(ctx, id); err != nil {
		return fmt.Errorf("before delete hook failed: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	if err := s.hooks.AfterDelete(ctx, id); err != nil {
		return fmt.Errorf("after delete hook failed: %w", err)
	}

	return nil
}

// Close closes a report by setting ClosedBy and ClosedAt.
func (s *Service) Close(ctx context.Context, id string, closedByID string, closedAt int64) (*structs.Report, error) {
	report, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	report.ClosedBy = &closedByID
	report.ClosedAt = &closedAt

	return s.Update(ctx, id, report)
}
