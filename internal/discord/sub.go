package discord

import (
	"github.com/bwmarrin/discordgo"
)

func ResolveUserID(i *discordgo.InteractionCreate) (string, error) {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID, nil
	}
	if i.User != nil {
		return i.User.ID, nil
	}
	return "", ErrBadRequest
}
