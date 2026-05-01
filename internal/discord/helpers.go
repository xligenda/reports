package discord

import (
	"fmt"
	"hash/fnv"
	"time"

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

func errorHash(e error) string {
	h := fnv.New32a()
	h.Write(fmt.Appendf(nil, "%v%d", e, time.Now().UnixNano()))

	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	sum := h.Sum32()

	result := make([]byte, 5)
	for i := range result {
		result[i] = chars[sum%uint32(len(chars))]
		sum /= uint32(len(chars))
	}

	return string(result)
}
