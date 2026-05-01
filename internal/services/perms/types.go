package perms

type Permission string

func (p Permission) String() string {
	return string(p)
}

const (
	None                Permission = ""
	SaveReports         Permission = "save_reports"
	ViewReports         Permission = "view_reports"
	CloseReports        Permission = "close_reports"
	ViewReportsExtended Permission = "view_reports_extended"
	DeleteReports       Permission = "delete_reports"
)
