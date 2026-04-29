package hooks

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/xligenda/reports/internal/discord"
	"github.com/xligenda/reports/internal/services/perms"
	"github.com/xligenda/reports/pkg/kit"
)

const (
	stackCommand = "hooks.PermsHook/command"
	stackField   = "hooks.PermsHook/field"
)

type PermsProvider interface {
	Check(ctx context.Context, id string, action perms.Permission, stack string) (bool, error)
}

type PermsHook struct {
	*kit.EmptyHook
	provider           PermsProvider
	requiredPermission perms.Permission
	fieldsPermissions  map[string]perms.Permission
}

func NewPermsHook(
	provider PermsProvider,
	requiredPermission perms.Permission,
	fieldsPermissions map[string]perms.Permission,
) *PermsHook {
	return &PermsHook{
		provider:           provider,
		requiredPermission: requiredPermission,
		fieldsPermissions:  fieldsPermissions,
	}
}

func (h *PermsHook) Execute(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userID, err := discord.ResolveUserID(i)
	if err != nil {
		return err
	}

	if err := h.checkPerm(ctx, userID, h.requiredPermission, stackCommand); err != nil {
		return err
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		if err := h.checkFieldPerms(ctx, userID, i.ApplicationCommandData().Options); err != nil {
			return err
		}
	}

	return nil
}

func (h *PermsHook) checkPerm(ctx context.Context, userID string, perm perms.Permission, stack string) error {
	hasPerm, err := h.provider.Check(ctx, userID, perm, stack)
	if err != nil {
		return discord.ErrInternal
	}
	if !hasPerm {
		return discord.ErrForbidden
	}
	return nil
}

func (h *PermsHook) checkFieldPerms(ctx context.Context, userID string, opts []*discordgo.ApplicationCommandInteractionDataOption) error {
	for _, opt := range opts {
		perm, exists := h.fieldsPermissions[opt.Name]
		if !exists {
			continue
		}
		if err := h.checkPerm(ctx, userID, perm, stackField); err != nil {
			return err
		}
	}
	return nil
}
