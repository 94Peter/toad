package pdf

import (
	"fmt"
	"os"
	"strconv"

	"dforcepro.com/report"
	"github.com/94peter/gopdf"
	"github.com/94peter/toad/util"
)

type Pdf struct {
	myPDF *gopdf.GoPdf
	//pageSize *Rect
	pageSize gopdf.Rect
}

const (
	valignTop    = 1
	valignMiddle = 2
	valignBottom = 3

	alignLeft   = 4
	alignCenter = 5
	alignRight  = 6

	//595.28, 841.89 = A4 (210 × 297 cm)
	A4_W = 841.89
	A4_H = 595.28
	//1000.155, 708.333 = B4 (250 × 353 cm)
	B4_W = 1000.637
	B4_H = 708.333
	//1,000.637, 1,416.313 = B3 (353 × 500 cm)
	B3_W = 1416.313
	B3_H = 1000.637

	TW_Medium_PATH = "resource/ttf/TW-Medium.ttf"

	NewPdf = true
	OriPdf = false
)

// type Rect struct {
// 	W float64
// 	H float64
// }

var (
	p          *Pdf
	TextWidth  = float64(60)
	TextHeight = float64(25)
)

func GetNewPDF(things ...interface{}) *Pdf {
	// if p != nil {
	// 	return p
	// }
	myPDFPage := PageSizeB3 // default page
	for _, it := range things {
		myPDFPage = it.(gopdf.Rect)
	}

	fmt.Println(p)

	p = &Pdf{
		myPDF: &gopdf.GoPdf{},
		// pageSize: &Rect{
		// 	W: B3_W,
		// 	H: B3_H,
		// },
		pageSize: myPDFPage,
	}

	//595.28, 841.89 = A4 (210 × 297 cm)
	//1000.155, 708.333 = B4 (250 × 353 cm)
	//1,000.637, 1,416.313 = B3 (353 × 500 cm)
	pdf := p.myPDF
	// pdf.Start(gopdf.Config{
	// 	Unit:     "pt",
	// 	PageSize: p.pageSize, //gopdf.Rect{W: p.pageSize.W, H: p.pageSize.H},
	// 	Protection: gopdf.PDFProtectionConfig{
	// 		UseProtection: true,
	// 		Permissions:   gopdf.PermissionsPrint | gopdf.PermissionsCopy | gopdf.PermissionsModify,
	// 		OwnerPass:     []byte("000000"),
	// 		UserPass:      []byte("123456")},
	// }) // B4(1000.155, 708.333)

	pdf.Start(gopdf.Config{
		Unit:     "pt",
		PageSize: p.pageSize, //gopdf.Rect{W: p.pageSize.W, H: p.pageSize.H},
	}) // B4(1000.155, 708.333)

	pdf.AddPage()
	err := p.LoadTTF("TW-Medium", TW_Medium_PATH, "", 14)
	if err != nil {
		fmt.Println("LoadTTF:", err.Error())
		return nil
	}
	pdf.SetFont("TW-Medium", "", 10)

	fmt.Println("Coming GetNewPDF Coming GetNewPDF Coming GetNewPDF Coming GetNewPDF AddPage AddPage AddPage AddPage")

	return p
}

func GetOriPDF() *Pdf {
	if p != nil {
		fmt.Println("Direct P return")
		return p
	}
	fmt.Println("GetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDFGetNewPDF")
	return GetNewPDF()
}

func (p *Pdf) PutImage(fn string, x, y float64) {
	pdf := p.myPDF

	pdf.Image(util.PdfDir+fn, x, y, nil) //print image
}

func (p *Pdf) LoadTTF(ttfName, ttfPathFile, fontstyle string, fontsize int) error {
	pdf := p.myPDF
	err := pdf.AddTTFFont(ttfName, ttfPathFile)
	if err != nil {
		return err
	}
	err = pdf.SetFont(ttfName, fontstyle, fontsize)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pdf) DrawTablePDF(dataTable *DataTable) {
	//salaryTable :=

	//fmt.Println("dataTable:", dataTable)

	pdf := p.myPDF

	//595.28, 841.89 = A4 (210 × 297 cm)
	// pdf.Start(gopdf.Config{Unit: "pt", PageSize: gopdf.Rect{W: 1000.155, H: 708.333}}) // B4(1000.155, 708.333)
	// pdf.AddPage()

	// err := p.loadTTF("TW-Medium", "conf/dev/TW-Medium.ttf", "", 14)
	// if err != nil {
	// 	fmt.Println("loadTTF:", err.Error())
	// 	return nil
	// }
	// pdf.SetFont("TW-Medium", "", 10)

	// pdf.Cell(nil, "Hi! This is italic.王")
	// pdf.Text("200")
	// pdf.SetTextColor(0, 0, 0)
	// pdf.Text("200")
	// pdf.Br(200)
	// pdf.Cell(nil, "255,255,255")
	//fmt.Println("pdf.GetX() init", pdf.GetX())

	if dataTable != nil {
		data := dataTable.RawData
		splitSize := dataTable.ColumnLen
		drawData := data[:splitSize]
		fmt.Println("splitSize:", splitSize)

		//fmt.Println("len(data):", len(data))
		for data != nil {

			//for i := 0; i < len(drawData); i++ {
			//fmt.Println("len(drawData)", len(drawData))
			for i := 0; i < len(drawData); i++ {
				TextWidth = dataTable.ColumnWidth[i]
				//fmt.Println("i:", i)
				//fmt.Println("data:", drawData[i].Text, " width:", dataTable.ColumnWidth[i])
				p.DrawRectangle(TextWidth, TextHeight, report.ColorWhite, "FD")
				//fmt.Println("DrawRectangle")
				p.FillText(drawData[i].Text, 12, report.ColorTableLine, alignCenter, valignMiddle, TextWidth, TextHeight)
				//fmt.Println("FillText")
				pdf.SetX(pdf.GetX() + TextWidth)
				//fmt.Println("SetX")
			}
			//fmt.Println("****")
			pdf.Br(25)

			//fmt.Println("new line")
			data = data[splitSize:]
			if len(data) == 0 {
				break
			} else if len(data) < dataTable.ColumnLen {
				splitSize = len(data)
			}
			drawData = data[:splitSize]
			if (p.pageSize.H - pdf.GetY()) < 50 {
				p.NewPage()
			}
		}
		// for i := 0; i < len(drawData); i++ {
		// 	p.DrawRectangle(textWidth, textHeight, report.ColorWhite, "FD")
		// 	p.FillText("aaa", 12, report.ColorTableLine, pdf.GetX(), pdf.GetY(), alignCenter, valignBottom, textWidth, textHeight)
		// 	pdf.SetX(pdf.GetX() + textWidth)
		// }
	}
	//p.rectFillColor("Play", 14, p.x, p.y, 100, 100, report.ColorWhite, alignRight, valignBottom)

	//rectFillColor(&pdf, "分店名11111111111111111111", 14, p.x, 170, 60, 35, report.ColorWhite, alignCenter, valignMiddle)

	//rectFillColor(&pdf, "Play", 14, p.x, 300, 50, 20, report.ColorWhite, alignRight, valignBottom)

	//pdf.WritePdf("italic.pdf")

	return //pdf.GetBytesPdf()
}

func (p *Pdf) GetTextWidth(text string) float64 {
	pdf := p.myPDF
	textw, _ := pdf.MeasureTextWidth(text)
	return textw
}

func (p *Pdf) GetBytesPdf() []byte {
	pdf := p.myPDF
	return pdf.GetBytesPdf()
}
func (p *Pdf) WriteFile(fname string) {
	pdf := p.myPDF
	os.MkdirAll(util.PdfDir, os.ModePerm)
	pdf.WritePdf(util.PdfDir + fname + ".pdf")
	fmt.Println("WriteFile:", fname)

}

func (p *Pdf) SetPdf_XY(X, Y float64) {
	pdf := p.myPDF
	if X >= 0 {
		pdf.SetX(X)
	}
	if Y >= 0 {
		pdf.SetX(Y)
	}
}

func (p *Pdf) NewLine(Height float64) {
	pdf := p.myPDF
	pdf.Br(Height)
}

func (p *Pdf) NewPage() {
	pdf := p.myPDF
	pdf.AddPage()
}

func (p *Pdf) DrawRectangle(w, h float64, color report.Color, rectType string) {
	pdf := p.myPDF
	pdf.SetFillColor(color.R, color.G, color.B)
	pdf.RectFromUpperLeftWithStyle(pdf.GetX(), pdf.GetY(), w, h, rectType)
}

func (p *Pdf) FillText(text string, floatFontSize float64, color report.Color, align, valign int, w, h float64) float64 {
	pdf := p.myPDF
	ox := pdf.GetX()
	oy := pdf.GetY()
	x := ox
	y := oy

	pdf.SetFillColor(color.R, color.G, color.B)
	if align == alignCenter {
		textw, _ := pdf.MeasureTextWidth(text)
		x = x + (w / 2) - (textw / 2)
	} else if align == alignRight {
		textw, _ := pdf.MeasureTextWidth(text)
		x = x + w - textw
	}
	pdf.SetX(x)

	if valign == valignMiddle {
		y = y + (h / 2) - (floatFontSize / 2)
	} else if valign == valignBottom {
		y = y + h - (floatFontSize)
	}

	pdf.SetY(y)
	pdf.Cell(nil, text)
	pdf.SetX(ox)
	pdf.SetY(oy)
	endX := ox + w
	return endX
}

//pdf *gopdf.GoPdf,
func (p *Pdf) rectFillColor(
	text string,
	fontSize int,
	x, y, w, h float64,
	color report.Color,
	align, valign int,
) {
	pdf := p.myPDF
	pdf.SetLineWidth(0.1)
	pdf.SetFillColor(color.R, color.G, color.B) //setup fill color
	pdf.RectFromUpperLeftWithStyle(x, y, w, h, "FD")
	//pdf.SetTextColor(0, 0, 0)
	pdf.SetFillColor(0, 0, 0)

	if align == alignCenter {
		textw, _ := pdf.MeasureTextWidth(text)
		x = x + (w / 2) - (textw / 2)
	} else if align == alignRight {
		textw, _ := pdf.MeasureTextWidth(text)
		x = x + w - textw
	}

	pdf.SetX(x)

	if valign == valignMiddle {
		y = y + (h / 2) - (float64(fontSize) / 2)
	} else if valign == valignBottom {
		y = y + h - float64(fontSize)
	}

	pdf.SetY(y)
	pdf.Cell(nil, text)

}

func SayHelloTo(s string) string {
	str := "Hello " + s + "!"
	return str
}

//1
func (p *Pdf) CustomizedBranchSalary(table *DataTable, T_Salary, T_Pbonus, T_Abonus, T_Lbonus, T_Total, T_SP, T_Tax, T_LaborFee, T_HealthFee, T_Welfare, T_CommercialFee, T_TAmount, T_Other int) {
	//init PDFX is 10
	pdfx := 10.0
	//1 is 姓名欄位
	for i := 0; i < 1; i++ {
		pdfx += table.ColumnWidth[i]
	}

	textw := table.ColumnWidth[1]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("合計", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[2]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Salary), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[3]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Pbonus), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[4]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Lbonus), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[5]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Abonus), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[6]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Total), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[7]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_SP), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[8]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Tax), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[9]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_LaborFee), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[10]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_HealthFee), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[11]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Welfare), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[12]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_CommercialFee), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[13]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Other), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[14]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_TAmount), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

}

//8
func (p *Pdf) CustomizedSalerCommission(table *DataTable, SName string, T_Bonus, T_SR int) {
	fmt.Println("CustomizedSalerCommission")
	//init PDFX is 10
	pdfx := 10.0
	//2 is 應扣費用
	for i := 0; i < 2; i++ {
		pdfx += table.ColumnWidth[i]
	}
	//應扣費用 姓名 比例
	textw := table.ColumnWidth[2] + table.ColumnWidth[3] + table.ColumnWidth[4]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(SName, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[5]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_SR), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[6]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Bonus), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//p.myPDF.AddPage()
	p.NewPage()
}

func (p *Pdf) CustomizedAgentSign(table *DataTable, T_Bonus, T_SR float64) (Total_SR, Total_Bonus float64) {
	fmt.Println("CustomizedAgentSign")
	//init PDFX is 10
	p.NewLine(25)
	pdfx := 10.0
	//3 is 姓名
	for i := 0; i < 4; i++ {
		pdfx += table.ColumnWidth[i]
	}
	Total_SR, Total_Bonus = 0.0, 0.0
	// //姓名
	// textw := table.ColumnWidth[3]

	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(table.ColumnWidth[4]+table.ColumnWidth[5]+table.ColumnWidth[6], TextHeight, report.ColorWhite, "FD")
	// p.FillText(SName, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	// pdfx += textw

	textw := table.ColumnWidth[4]
	p.SetPdf_XY(pdfx, -1)
	//p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("合計", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

	textw = table.ColumnWidth[5]
	p.SetPdf_XY(pdfx, -1)
	//p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	text := fmt.Sprintf("%.1f", T_SR)
	Total_SR += T_SR
	p.FillText(text, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

	textw = table.ColumnWidth[6]
	p.SetPdf_XY(pdfx, -1)
	text = fmt.Sprintf("%.1f", T_Bonus)
	Total_Bonus += T_Bonus
	//p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(text, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

	//p.myPDF.AddPage()
	//p.NewPage()
	p.NewLine(25)
	//p.NewLine(25)
	return
}

func (p *Pdf) CustomizedAmortizationTitle(table *DataTable, title string) {
	fmt.Println("CustomizedAmortizationTitle")
	//init PDFX is 10
	pdfx := 10.0
	textw := 0.0
	for i := 0; i < len(table.ColumnWidth); i++ {
		textw += table.ColumnWidth[i]
	}
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(title, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdf := p.myPDF
	pdf.Br(25)
}

func (p *Pdf) CustomizedAmortization(table *DataTable, T_Month, T_Has, T_not int) {
	fmt.Println("CustomizedAmortization")
	//init PDFX is 10
	pdfx := 10.0

	textw := table.ColumnWidth[0] + table.ColumnWidth[1] + table.ColumnWidth[2] + table.ColumnWidth[3]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("合計", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[4]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Month), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[5]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Has), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[6]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_not), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

}

func (p *Pdf) CustomizedPrepayTitle(table *DataTable, title string, branch []string) {
	fmt.Println("CustomizedPrepayTitle")

	// header := []*TableStyle{}
	// for i := 0; i < len(branch); i++ {
	// 	var vs = &TableStyle{
	// 		Text:  branch[i],
	// 		Bg:    report.ColorWhite,
	// 		Front: report.ColorTableLine,
	// 	}
	// 	header = append(header, vs)
	// }

	// Text Title here
	//init PDFX is 10
	pdf := p.myPDF
	pdfx := 10.0
	textw := 0.0
	for i := 0; i < len(table.ColumnWidth); i++ {
		textw += table.ColumnWidth[i]
	}
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(title, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdf.Br(25)

	//日期------支出金額
	for i := 0; i < 4; i++ {
		p.SetPdf_XY(pdfx, -1)
		textw = table.ColumnWidth[i]
		p.DrawRectangle(textw, TextHeight*2, report.ColorWhite, "FD")
		p.FillText(table.RawData[i].Text, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight*2)
		pdfx += table.ColumnWidth[i]
	}
	p.SetPdf_XY(pdfx, -1)
	//分攤金額
	textw = 0
	for i := 4; i < len(table.ColumnWidth); i++ {
		textw += table.ColumnWidth[i]
	}
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("分攤金額", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdf.Br(25)
	//
	//各店家欄位
	for i := 0; i < len(branch); i++ {
		p.SetPdf_XY(pdfx, -1)
		textw = table.ColumnWidth[i+4]
		p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
		p.FillText(branch[i], 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
		pdfx += textw
	}
	pdf.Br(25)
}

func (p *Pdf) CustomizedPrepay(table *DataTable, Total []int) {
	fmt.Println("CustomizedPrepay")

	// header := []*TableStyle{}
	// for i := 0; i < len(branch); i++ {
	// 	var vs = &TableStyle{
	// 		Text:  branch[i],
	// 		Bg:    report.ColorWhite,
	// 		Front: report.ColorTableLine,
	// 	}
	// 	header = append(header, vs)
	// }

	// Text Title here
	//init PDFX is 10

	pdfx := 10.0
	textw := 0.0
	for i := 0; i < 3; i++ {
		textw += table.ColumnWidth[i]
	}
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("合計金額", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

	//各店家總額
	for i := 0; i < len(Total); i++ {
		p.SetPdf_XY(pdfx, -1)
		textw = table.ColumnWidth[i+3]
		p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
		p.FillText(strconv.Itoa(Total[i]), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
		pdfx += textw
	}

}

func (p *Pdf) CustomizedPocketTitle(table *DataTable, title string) {
	fmt.Println("CustomizedPocketTitle")
	//init PDFX is 10
	pdfx := 10.0
	textw := 0.0
	for i := 0; i < len(table.ColumnWidth); i++ {
		textw += table.ColumnWidth[i]
	}
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(title, 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdf := p.myPDF
	pdf.Br(25)
}

func (p *Pdf) CustomizedPocket(table *DataTable, T_Income, T_Fee, T_Balance int) {
	fmt.Println("CustomizedPocket")
	//init PDFX is 10
	pdfx := 10.0

	textw := table.ColumnWidth[0] + table.ColumnWidth[1] + table.ColumnWidth[2] + table.ColumnWidth[3]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("總計金額", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[4]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Income), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[5]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Fee), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[6]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Balance), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw

}

func (p *Pdf) CustomizedNHI(table *DataTable, T_PayrollBracket, T_Salary, T_Pbonus, T_Bonus, T_Total, T_Balance, T_PTSP,
	T_PD, T_FourBouns, T_SP, T_FourSP, T_Tax, T_SPB int) {
	fmt.Println("CustomizedNHI")
	//init PDFX is 10
	pdfx := 10.0

	textw := table.ColumnWidth[0]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("合計", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[1]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_PayrollBracket), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[2]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Salary), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	textw = table.ColumnWidth[3]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Pbonus), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[4]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Bonus), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[5]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText("0", 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//合計
	textw = table.ColumnWidth[6]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Total), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[7]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Balance), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[8]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_PD), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[9]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_FourBouns), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//補充保費薪資差額
	textw = table.ColumnWidth[10]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_SPB), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[11]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_FourSP), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[12]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_FourSP), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[13]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_PTSP), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
	textw = table.ColumnWidth[14]
	p.SetPdf_XY(pdfx, -1)
	p.DrawRectangle(textw, TextHeight, report.ColorWhite, "FD")
	p.FillText(strconv.Itoa(T_Tax), 12, report.ColorTableLine, alignCenter, valignMiddle, textw, TextHeight)
	pdfx += textw
	//
}
