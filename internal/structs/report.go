package structs

type Report struct {
	Channel   string  `db:"id"`
	Guild     string  `db:"guild"`
	Issuer    string  `db:"issuer"`
	Topic     string  `db:"topic"`
	Note      *string `db:"note"`
	Proof     *string `db:"proof"`
	CreatedAt int64   `db:"created_at"`
	ClosedBy  *string `db:"closed_by"`
	ClosedAt  *int64  `db:"closed_at"`
}

func (r Report) GetID() string {
	return r.Channel
}
