package hooks

import (
	"context"
)

// Hooks defines the interface for lifecycle hooks
type Hooks interface {
	BeforeCreate(ctx context.Context, entity any) error
	AfterCreate(ctx context.Context, entity any) error
	BeforeUpdate(ctx context.Context, entity any) error
	AfterUpdate(ctx context.Context, entity any) error
	BeforeDelete(ctx context.Context, id any) error
	AfterDelete(ctx context.Context, id any) error
}

// NoOpHooks provides default no-op implementations
type NoOpHooks struct{}

func (h *NoOpHooks) BeforeCreate(ctx context.Context, entity any) error { return nil }
func (h *NoOpHooks) AfterCreate(ctx context.Context, entity any) error  { return nil }
func (h *NoOpHooks) BeforeUpdate(ctx context.Context, entity any) error { return nil }
func (h *NoOpHooks) AfterUpdate(ctx context.Context, entity any) error  { return nil }
func (h *NoOpHooks) BeforeDelete(ctx context.Context, id any) error     { return nil }
func (h *NoOpHooks) AfterDelete(ctx context.Context, id any) error      { return nil }
