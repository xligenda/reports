package structs

type Report struct {
	ChannelID string  `db:"channel"`
	GuildID   string  `db:"guild"`
	IssuerID  string  `db:"issuer"`
	Name      string  `db:"name"`
	Topic     string  `db:"topic"`
	Note      string  `db:"note"`
	Proof     string  `db:"proof"`
	CreatedAt int64   `db:"created_at"`
	ClosedBy  *string `db:"closed_by"`
	ClosedAt  *int64  `db:"closed_at"`
}

func (r Report) GetID() string {
	return r.ChannelID
}
