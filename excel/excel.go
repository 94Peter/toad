package excel

import (
	"fmt"
	"os"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/94peter/toad/util"
)

type Excel struct {
	File *excelize.File
}

type Alignment struct {
	Horizontal string `json:"horizontal,omitempty"` // 水平对齐方式
	Vertical   string `json:"vertical,omitempty"`   // 垂直对齐方式
	WrapText   bool   `json:"wrap_text,omitempty"`  // 自动换行设置
}

var (
	excel *Excel
)

func GetNewExcel() *Excel {
	excel = &Excel{
		File: excelize.NewFile(),
	}

	return excel
}

func GetOriExcel() *Excel {
	return excel
}

func (excel *Excel) ClearExcel() {
	excel = nil
}

func (excel *Excel) SaveFile(fn string) {
	//fakeId := fmt.Sprintf("%d", time.Now().Unix())

	os.MkdirAll(util.PdfDir, os.ModePerm)

	ex := excel.File
	err := ex.SaveAs(util.PdfDir + fn + ".xlsx")
	if err != nil {
		fmt.Println(err)
	}
	excel.ClearExcel()
}

func (excel *Excel) FillText(datatable []*DataTable) {
	f := excel.File

	//style, _ := f.NewStyle(`{"alignment":{"horizontal":"right","Vertical":"center"},"font":{"bold":true},"border":[{"type":"right","color":"FF0000","style":1}],"fill":{"type":"pattern","color":["#CCFFFF"],"pattern":1}}`)
	rightStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"right"}}`)
	//centerStyle, _ := f.NewStyle(`{"alignment":{"horizontal":"center"}}`)

	for _, element := range datatable {
		f.NewSheet(element.SheetName)
		f.SetCellStyle(element.SheetName, "D1", "D300", rightStyle)

	}
	f.DeleteSheet("Sheet1") //預設表格 刪除

	//f.SetCellValue("薪轉戶", "B2", 100)
	// f.SetColWidth("薪轉戶", "C", "C", 35)
	// f.SetColWidth("薪轉戶", "D", "E", 60)

	for _, element := range datatable {
		data := element.RawData
		//fmt.Println("SheetName:", element.SheetName)
		for key, value := range data {
			//fmt.Println("Key:", key, "Value:", value)
			f.SetCellValue(element.SheetName, key, value)
			f.SetColWidth(element.SheetName, key[:1], key[:1], 35)
		}
	}

	/*置中*/
	// style, err := f.NewStyle(`{"alignment":{"horizontal":"center","ident":1,"justify_last_line":true,"reading_order":0,"relative_indent":1,"shrink_to_fit":true,"text_rotation":45,"vertical":"","wrap_text":true}}`)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = f.SetCellStyle(datatable.SheetName, "H9", "H9", style)
	/*背景顏色*/
	// style2, err := f.NewStyle(`{"fill":{"type":"pattern","color":["#E0EBF5"],"pattern":1}}`)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = f.SetCellStyle(datatable.SheetName, "H10", "H10", style2)
	/*文字*/
	// style, err := f.NewStyle(`{"font":{"bold":true,"italic":true,"family":"Times New Roman","size":36,"color":"#777777"}}`)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = f.SetCellStyle(datatable.SheetName, "H9", "H9", style)

}
