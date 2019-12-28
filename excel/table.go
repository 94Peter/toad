package excel

type DataTable struct {
	SheetName string
	RawData   map[string]string
}

type RawData struct {
	Text string
	Pos  string
	//Bg    report.Color
	//Front report.Color
}

type ReportType int32

var (
	PayrollTransfer = 2 // 薪轉表
	IncomeTaxReturn = 9 // 年度所得申報
	//
	PayrollTransferHeader = []string{"姓名", "身分證字號", "帳號", "轉帳金額"}

	IncomeTaxReturnHeader = []string{"姓名", "身分證字號", "戶籍地址", "1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"}

	AZTable = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
)

//收款項目		 應扣費用 	  				備註

/*
*@param mtype is Table header type
 */
func GetDataTable(mtype int) *DataTable {

	data := map[string]string{}
	header := GetHeader(mtype)

	for i := 0; i < len(header); i++ {
		data[AZTable[i]+"1"] = header[i]
	}

	table := &DataTable{
		SheetName: "none",
		RawData:   data,
	}
	// header := []TextBlockStyle{}
	// for i := 0; i < len(salaryTableHeader); i++ {
	// 	var tbs = TextBlockStyle{
	// 		Text:        salaryTableHeader[i],
	// 		Color:       report.ColorTableLine,
	// 		ColumnWidth: 30,
	// 	}
	// 	header = append(header, tbs)
	// }
	// data := [5][]TextBlockStyle{}
	// tmp := []TextBlockStyle{}
	// for i := 0; i < len(salaryTableHeader); i++ {
	// 	var tbs = TextBlockStyle{
	// 		Text:        salaryTableHeader[i],
	// 		Color:       report.ColorTableLine,
	// 		ColumnWidth: 30,
	// 	}
	// 	tmp = append(tmp, tbs)
	// }
	// data[0] = tmp

	// ts := &TableStyle{
	// 	Header: header,
	// 	Data:   data,
	// }
	return table
}

func GetHeader(mtype int) []string {
	switch mtype {
	case PayrollTransfer:
		return PayrollTransferHeader
		break
	case IncomeTaxReturn:
		return IncomeTaxReturnHeader
	default:
		break
	}
	return nil
}
