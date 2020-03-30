package pdf

import (
	"fmt"

	"dforcepro.com/report"
)

type DataTable struct {
	Caption     string
	ColumnLen   int
	ColumnWidth []float64
	RawData     []*TableStyle
}

type TableStyle struct {
	Text   string
	Bg     report.Color
	Front  report.Color
	Align  int
	Valign int
}

type ReportType int32

func ReportToString(mtype int) string {
	switch mtype {
	case BranchSalary:
		return "各店薪資表"
	case Commission:
		return "佣金明細表"
	case SalerSalary:
		return "個人薪資明細"
	case SalarCommission:
		return "個人佣金明細"
	default:
		return "UNKNOWN"
	}
}

var (
	BranchSalary    = 1 //1.各店薪資表
	NHI             = 3 // 二代健保
	Commission      = 4
	AgentSign       = 5
	SR              = 6
	SalerSalary     = 7  //7.個人薪資明細 (要轉傳給員工的)
	SalarCommission = 8  //8.個人佣金明細 (要轉傳給員工的)
	Pocket          = 10 //10. 零用金
	Prepay          = 11 //11. 代支費用
	Amortization    = 12 //12. 設立成本分攤表

	//
	NHITableHeader = []string{"員工姓名", "健保投保薪資", "薪資", "績效", "獎金", "兼職(加保工會)", "合計薪資", "薪資差額(G-B)", "累計差額(H+上月H)", "4倍獎金", "補充保費((I-J)＞0=H)", "4倍獎金補充保費", "員工應扣4倍獎金補充保費", "員工應扣兼職補充保費", "代扣稅款5%", "職稱", "備註"}

	// For AgentSign 5 (很像個人傭金明細)
	AgentSignTableHeader = []string{"收款項目", "收款金額", "應扣費用", "姓名", "比例", "實績", "獎金", "備註", "店別", "%"}
	// For SalarCommission 8
	SalarCommissionTableHeader = []string{"收款項目", "收款金額", "應扣費用", "姓名", "比例", "實績", "獎金", "備註"}
	// For Commission 4
	CommissionTableHeader = []string{"入帳日期", "發票號碼", "收款項目", "收款金額", "應扣費用", "姓名", "比例", "實績", "獎金", "備註", "店別", "%", "說明", "票號"}
	// For BranchSalary 1 & SalerSalary 7
	SalaryTableHeader = []string{"NO.", "姓名", "底薪", "+績效獎金", "+領導獎金", "-出勤", "薪資總額", "-代收(補充保費)", "-代收(所得稅)", "-代收(勞保費)", "-代收(健保費)", "-福利金", "-商耕費", "-其他", "-轉帳", "備註"}

	// For SR 6
	SRTableHeader = []string{"姓名", "實績", "獎金"}
	// For 10
	PocketTableHeader = []string{"日期", "分店別", "科目", "摘要", "收入", "支出", "結餘"}
	// For 11
	PrepayTableHeader = []string{"日期", "科目", "摘要", "支出金額"} //後續的靠Hard code客製化
	// For 12
	AmortizationTableHeader = []string{"項目名稱", "取得日期", "取得成本", "攤提年限", "每月攤提金額", "已攤提金額", "未攤提金額"}
)

//收款項目		 應扣費用 	  				備註

/*
*@param mtype is Table header type
 */
func GetDataTable(mtype int) *DataTable {

	header := getTableHeader(mtype)
	DataTable := &DataTable{
		Caption:   "??",
		ColumnLen: len(header),
		RawData:   header,
	}
	//fmt.Println("header:", header)
	//fmt.Println("header:", len(header))
	initWidth := []float64{}
	for i := 0; i < len(header); i++ {
		initWidth = append(initWidth, TextWidth)
	}
	DataTable.ColumnWidth = initWidth

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

	return DataTable
}

func ResizeWidth(table *DataTable, new float64, index int) {
	if table.ColumnWidth[index] < new {
		//fmt.Println("ResizeWidth:", table.ColumnWidth[index], " to new ", new)
		table.ColumnWidth[index] = new + 20
	}
}

func getTableHeader(mType int) []*TableStyle {
	var h_data []string
	fmt.Println("getTableHeader:...", mType)
	switch mType {
	case NHI:
		h_data = NHITableHeader
		break
	case BranchSalary:
		h_data = SalaryTableHeader
		fmt.Println("BranchSalary:")
		break
	case SalerSalary:
		h_data = SalaryTableHeader
		fmt.Println("SalerSalary:")
		break
	case Commission:
		h_data = CommissionTableHeader
		fmt.Println("Commission:")
		break
	case SalarCommission:
		h_data = SalarCommissionTableHeader
		fmt.Println("SalarCommission:")
		break
	case AgentSign:
		h_data = AgentSignTableHeader
		break
	case SR:
		h_data = SRTableHeader
		break
	case Amortization:
		h_data = AmortizationTableHeader
		break
	case Pocket:
		h_data = PocketTableHeader
		break
	case Prepay:
		h_data = PrepayTableHeader
		break
	default:
		fmt.Println("default:", h_data)
		break
	}
	//fmt.Println("h_data:", h_data)
	header := []*TableStyle{}
	for i := 0; i < len(h_data); i++ {
		var vs = &TableStyle{
			Text:  h_data[i],
			Bg:    report.ColorWhite,
			Front: report.ColorTableLine,
		}
		header = append(header, vs)
	}

	// var vs = &TableStyle{
	// 	Text:  "t1",
	// 	Bg:    report.ColorWhite,
	// 	Front: report.ColorTableLine,
	// }
	// header = append(header, vs)

	return header
}
