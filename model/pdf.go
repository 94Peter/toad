package model

import (
	"fmt"

	"github.com/94peter/gopdf"
	"github.com/94peter/toad/resource/db"
)

type PdfModel struct {
	imr interModelRes
	db  db.InterSQLDB
}

const (
	valignTop    = 1
	valignMiddle = 2
	valignBottom = 3
)

const (
	alignLeft   = 4
	alignCenter = 5
	alignRight  = 6
)

var (
	pdfM *PdfModel
)

func GetPdfModel(imr interModelRes) *PdfModel {
	if pdfM != nil {
		return pdfM
	}

	pdfM = &PdfModel{
		imr: imr,
	}
	return pdfM
}

func loadTTF(pdf gopdf.GoPdf, ttfName, ttfPathFile, fontstyle string, fontsize int) (gopdf.GoPdf, error) {
	err := pdf.AddTTFFont(ttfName, ttfPathFile)
	if err != nil {
		return pdf, err
	}
	err = pdf.SetFont(ttfName, fontstyle, fontsize)
	if err != nil {
		return pdf, err
	}
	return pdf, nil
}

func (pdfM *PdfModel) Test() []byte {
	pdf := gopdf.GoPdf{}
	//pdf.Start(gopdf.Config{Unit: "pt", PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A4
	pdf.Start(gopdf.Config{Unit: "pt", PageSize: gopdf.Rect{W: 841.89, H: 595.28}}) // A4(橫向))
	pdf.AddPage()

	pdf, err := loadTTF(pdf, "TW-Medium", "conf/dev/TW-Medium.ttf", "", 14)
	if err != nil {
		fmt.Println("loadTTF:", err.Error())
		return nil
	}
	pdf.SetFont("TW-Medium", "U", 14)
	pdf.Cell(nil, "Hi! This is italic.王")
	pdf.Text("200")
	pdf.SetTextColor(255, 255, 255)
	pdf.Text("200")
	pdf.Br(200)
	pdf.Cell(nil, "255,255,255")

	rectFillColor(&pdf, "Play", 14, 10, 50, 100, 100, 155, 155, 155, alignRight, valignBottom)

	rectFillColor(&pdf, "分店名稱", 14, 10, 170, 60, 35, 155, 155, 155, alignCenter, valignMiddle)

	rectFillColor(&pdf, "Play", 14, 10, 300, 50, 20, 155, 155, 155, alignRight, valignBottom)
	//pdf.WritePdf("italic.pdf")
	return pdf.GetBytesPdf()
}

func rectFillColor(pdf *gopdf.GoPdf,
	text string,
	fontSize int,
	x, y, w, h float64,
	r, g, b uint8,
	align, valign int,
) {
	pdf.SetLineWidth(0.1)
	pdf.SetFillColor(r, g, b) //setup fill color
	pdf.RectFromUpperLeftWithStyle(x, y, w, h, "FD")
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
