package perms

type Permission string

func (p Permission) String() string {
	return string(p)
}

const (
	None         Permission = ""
	ViewExtended Permission = "view_extended"
)
