package permission

const (
	Office     = "office" //店秘書
	Admin      = "admin"
	Sales      = "sales"
	Manager    = "manager"    //店長
	Accountant = "accountant" //會計
)

var (
	Backend  = []string{Manager, Admin}
	Frontend = []string{Office, Manager, Admin}
	All      = []string{Accountant, Sales, Manager, Admin, Office}
)
