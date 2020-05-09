package permission

const (
	Office  = "office"
	Admin   = "admin"
	Sales   = "sales"
	Manager = "manager"
)

var (
	Backend  = []string{Manager, Admin}
	Frontend = []string{Office, Manager, Admin}
	All      = []string{Sales, Manager, Admin, Office}
)
