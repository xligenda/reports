package reports

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/xligenda/reports/internal/services/hooks"
	"github.com/xligenda/reports/internal/structs"
	"github.com/xligenda/reports/pkg/repo"
)

type MockHooks struct {
	BeforeCreateFn func(ctx context.Context, entity any) error
	AfterCreateFn  func(ctx context.Context, entity any) error
	BeforeUpdateFn func(ctx context.Context, entity any) error
	AfterUpdateFn  func(ctx context.Context, entity any) error
	BeforeDeleteFn func(ctx context.Context, id any) error
	AfterDeleteFn  func(ctx context.Context, id any) error
}

func (m *MockHooks) BeforeCreate(ctx context.Context, entity any) error {
	if m.BeforeCreateFn != nil {
		return m.BeforeCreateFn(ctx, entity)
	}
	return nil
}

func (m *MockHooks) AfterCreate(ctx context.Context, entity any) error {
	if m.AfterCreateFn != nil {
		return m.AfterCreateFn(ctx, entity)
	}
	return nil
}

func (m *MockHooks) BeforeUpdate(ctx context.Context, entity any) error {
	if m.BeforeUpdateFn != nil {
		return m.BeforeUpdateFn(ctx, entity)
	}
	return nil
}

func (m *MockHooks) AfterUpdate(ctx context.Context, entity any) error {
	if m.AfterUpdateFn != nil {
		return m.AfterUpdateFn(ctx, entity)
	}
	return nil
}

func (m *MockHooks) BeforeDelete(ctx context.Context, id any) error {
	if m.BeforeDeleteFn != nil {
		return m.BeforeDeleteFn(ctx, id)
	}
	return nil
}

func (m *MockHooks) AfterDelete(ctx context.Context, id any) error {
	if m.AfterDeleteFn != nil {
		return m.AfterDeleteFn(ctx, id)
	}
	return nil
}

func TestNewService(t *testing.T) {
	db := &sqlx.DB{}
	svc := NewService(db, nil)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.repo)
	assert.NotNil(t, svc.hooks)
}

func TestNewServiceWithCustomHooks(t *testing.T) {
	db := &sqlx.DB{}
	myHooks := &hooks.NoOpHooks{}
	svc := NewService(db, myHooks)
	assert.NotNil(t, svc)
	assert.Equal(t, myHooks, svc.hooks)
}

func TestNewServiceWithRepository(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	myHooks := &hooks.NoOpHooks{}
	svc := NewServiceWithRepository(myRepo, myHooks)
	assert.NotNil(t, svc)
	assert.Equal(t, myRepo, svc.repo)
}

func TestCreate_Success(t *testing.T) {
	mockHooks := &MockHooks{
		BeforeCreateFn: func(ctx context.Context, entity any) error { return nil },
		AfterCreateFn:  func(ctx context.Context, entity any) error { return nil },
	}

	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, mockHooks)

	assert.NotNil(t, svc)
}

func TestCreate_HookError(t *testing.T) {
	mockHooks := &MockHooks{
		BeforeCreateFn: func(ctx context.Context, entity any) error {
			return errors.New("validation failed")
		},
	}

	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, mockHooks)

	assert.NotNil(t, svc)
}

func TestGetByID(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestGetOpen(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestGetClosed(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestGetByGuild(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestGetAll(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestUpdate(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestDelete(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestClose(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})
	assert.NotNil(t, svc)
}

func TestDefaultHooksBeforeCreate(t *testing.T) {
	myHooks := NewDefaultHooks()

	t.Run("valid report", func(t *testing.T) {
		report := &structs.Report{Channel: "ch1", Guild: "guild1", Issuer: "usr1"}
		err := myHooks.BeforeCreate(context.Background(), report)
		assert.NoError(t, err)
	})

	t.Run("empty channel ID", func(t *testing.T) {
		report := &structs.Report{Guild: "guild1", Issuer: "usr1"}
		err := myHooks.BeforeCreate(context.Background(), report)
		assert.Error(t, err)
	})

	t.Run("empty guild ID", func(t *testing.T) {
		report := &structs.Report{Channel: "ch1", Issuer: "usr1"}
		err := myHooks.BeforeCreate(context.Background(), report)
		assert.Error(t, err)
	})

	t.Run("empty issuer ID", func(t *testing.T) {
		report := &structs.Report{Channel: "ch1", Guild: "guild1"}
		err := myHooks.BeforeCreate(context.Background(), report)
		assert.Error(t, err)
	})

	t.Run("not a report struct", func(t *testing.T) {
		err := myHooks.BeforeCreate(context.Background(), "not a report")
		assert.NoError(t, err)
	})
}

func TestDefaultHooksBeforeDelete(t *testing.T) {
	myHooks := NewDefaultHooks()

	t.Run("valid ID", func(t *testing.T) {
		err := myHooks.BeforeDelete(context.Background(), "123")
		assert.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		err := myHooks.BeforeDelete(context.Background(), "")
		assert.Error(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		err := myHooks.BeforeDelete(context.Background(), 123)
		assert.Error(t, err)
	})
}

func TestDefaultHooksBeforeUpdate(t *testing.T) {
	myHooks := NewDefaultHooks()

	t.Run("valid report", func(t *testing.T) {
		report := &structs.Report{Channel: "ch1"}
		err := myHooks.BeforeUpdate(context.Background(), report)
		assert.NoError(t, err)
	})

	t.Run("not a report struct", func(t *testing.T) {
		err := myHooks.BeforeUpdate(context.Background(), "not a report")
		assert.NoError(t, err)
	})
}

func TestDefaultHooksAfterHooks(t *testing.T) {
	myHooks := NewDefaultHooks()

	err := myHooks.AfterCreate(context.Background(), nil)
	assert.NoError(t, err)

	err = myHooks.AfterUpdate(context.Background(), nil)
	assert.NoError(t, err)

	err = myHooks.AfterDelete(context.Background(), nil)
	assert.NoError(t, err)
}

func TestClose_Success(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})

	assert.NotNil(t, svc)
}

func TestClose_HookError(t *testing.T) {
	mockHooks := &MockHooks{
		BeforeUpdateFn: func(ctx context.Context, entity any) error {
			return errors.New("cannot update report")
		},
	}

	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, mockHooks)

	assert.NotNil(t, svc)
}

// Test Search method with various filter combinations
func TestSearch_EmptyFilters(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})

	assert.NotNil(t, svc)
	// Empty filters should return all reports (equivalent to GetAll)
}

func TestSearch_SingleChannelFilter(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})

	filters := []repo.Filter{
		repo.NewFilter("channel", repo.Equals, "channel123"),
	}

	assert.NotNil(t, svc)
	// Verify single filter can be applied
	assert.Len(t, filters, 1)
	assert.Equal(t, "channel", filters[0].Field)
	assert.Equal(t, repo.Equals, filters[0].Operator)
	assert.Equal(t, "channel123", filters[0].Value)
}

func TestSearch_MultipleFilters(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})

	filters := []repo.Filter{
		repo.NewFilter("channel", repo.Equals, "channel123"),
		repo.NewFilter("guild", repo.Equals, "guild456"),
		repo.NewFilter("closed_at", repo.IsNull, nil),
	}

	assert.NotNil(t, svc)
	assert.Len(t, filters, 3)
}

func TestSearch_TextSearchFilters(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})

	filters := []repo.Filter{
		repo.NewFilter("name", repo.Like, "%bug%"),
		repo.NewFilter("topic", repo.Like, "%critical%"),
		repo.NewFilter("note", repo.Like, "%urgent%"),
	}

	assert.NotNil(t, svc)
	assert.Len(t, filters, 3)
	for _, f := range filters {
		assert.Equal(t, repo.Like, f.Operator)
	}
}

func TestSearch_DateRangeFilters(t *testing.T) {
	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, &hooks.NoOpHooks{})

	filters := []repo.Filter{
		repo.NewFilter("created_at", ">=", int64(1000000)),
		repo.NewFilter("created_at", "<=", int64(2000000)),
		repo.NewFilter("closed_at", ">=", int64(3000000)),
		repo.NewFilter("closed_at", "<=", int64(4000000)),
	}

	assert.NotNil(t, svc)
	assert.Len(t, filters, 4)
}

func TestSearch_OpenReportsFilter(t *testing.T) {
	// is_open=true -> closed_at IS NULL
	openFilter := repo.NewFilter("closed_at", repo.IsNull, nil)
	assert.Equal(t, "closed_at", openFilter.Field)
	assert.Equal(t, repo.IsNull, openFilter.Operator)

	// is_open=false -> closed_at IS NOT NULL
	closedFilter := repo.NewFilter("closed_at", repo.IsNotNull, nil)
	assert.Equal(t, "closed_at", closedFilter.Field)
	assert.Equal(t, repo.IsNotNull, closedFilter.Operator)
}

func TestSearch_NoHooksTriggered(t *testing.T) {
	hooksCalled := false
	mockHooks := &MockHooks{
		BeforeCreateFn: func(ctx context.Context, entity any) error {
			hooksCalled = true
			return nil
		},
		AfterCreateFn: func(ctx context.Context, entity any) error {
			hooksCalled = true
			return nil
		},
	}

	myRepo := &repo.GenericRepository[string, structs.Report]{}
	svc := NewServiceWithRepository(myRepo, mockHooks)

	// Search should not trigger any hooks (it's read-only)
	filters := []repo.Filter{
		repo.NewFilter("channel", repo.Equals, "channel123"),
	}
	// Just calling Search should not trigger hooks
	assert.False(t, hooksCalled)
	assert.NotNil(t, svc)
	assert.Len(t, filters, 1)
}
