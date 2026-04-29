package perms

type Permission string

func (p Permission) String() string {
	return string(p)
}
