package permission

const (
	Office  = "office"
	Admin   = "admin"
	Sales   = "sales"
	Manager = "manager"
)

var (
	Backend  = []string{Office, Admin}
	Frontend = []string{Sales, Manager, Admin}
	All      = []string{Sales, Manager, Admin, Office}
)
