package model

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"toad/pdf"
	"toad/resource/db"
	"toad/util"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/qr"
	"github.com/tidwall/gjson"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	storeID = "25077808"
	//復升
	ivURL   = "https://ranking.numax.com.tw/test/einvoice/api/invoice"
	auth_iv = "A2dC56TZfkpra6TCIuo5UQW870L/0=mazjT0g7=s5Do0K9z"
)

// type Invoice struct {
// 	Rid     string    `json:"id"`
// 	Date    time.Time `json:"date"`
// 	Title   string    `json:"title"`
// 	GUI     string    `json:"GUI"`
// 	Amount  string    `json:"amount"`
// 	Invoice string    `json:"invoice"`
// }

type Invoice struct {
	Signatrue string `json:"signatrue"`
	//No          string    `json:"invoice_no"`
	RandNum     string    `json:"random_number"`
	Date        string    `json:"invoice_datetime"`
	SalesAmount float64   `json:"sales_amount"`
	TotalAmount int       `json:"total_amount"`
	BuyerID     string    `json:"buyer_uniform"` //買家統編
	SellerID    string    `json:"SellerID"`      //賣家統編(台灣房屋?)
	Remark      string    `json:"remark"`
	Detail      []*Detail `json:"details"`

	Rid string `json:"id"`
	//Date    time.Time `json:"date"`
	Title     string `json:"title"`
	InvoiceNo string `json:"invoice_no"`
	//Amount    string `json:"amount"`
	//Invoice      string `json:"invoice"`
	Left_qrcode  string `json:"left_qrcode"`
	Right_qrcode string `json:"right_qrcode"`
}

type Detail struct {
	//ProductID string `json:"product_id"`
	Name     string `json:"product_name"`
	Quantity int    `json:"quantity"`
	//Unit      string `json:"unit"`
	UnitPrice int `json:"unit_price"`
	//Amount    int    `json:"amount"`
}

var (
	invoiceM *InvoiceModel
)

type InvoiceModel struct {
	imr         interModelRes
	db          db.InterSQLDB
	invoiceList []*Invoice
}

func GetInvoiceModel(imr interModelRes) *InvoiceModel {
	if invoiceM != nil {
		return invoiceM
	}

	invoiceM = &InvoiceModel{
		imr: imr,
	}
	return invoiceM
}

func (invoiceM *InvoiceModel) GetInvoiceData(rid string) *Invoice {

	const qspl = `SELECT rid, invoiceno, buyerID, sellerID, randomnum, title, date, amount, left_qrcode, right_qrcode FROM public.Invoice where rid = '%s';`
	db := invoiceM.imr.GetSQLDB()
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, rid))
	if err != nil {
		return nil
	}
	fmt.Println(fmt.Sprintf(qspl, rid))
	for rows.Next() {
		invoice := &Invoice{}

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&invoice.Rid, &invoice.InvoiceNo, &invoice.BuyerID, &invoice.SellerID, &invoice.RandNum, &invoice.Title, &invoice.Date, &invoice.TotalAmount, &invoice.Left_qrcode, &invoice.Right_qrcode); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		fmt.Println(invoice)
		return invoice
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	return nil
}

func (invoiceM *InvoiceModel) Json() ([]byte, error) {
	return json.Marshal(invoiceM.invoiceList)
}

func (invoiceM *InvoiceModel) CreateInvoice(inputInvoice *Invoice) (string, error) {
	Rid := inputInvoice.Rid

	fmt.Println(Rid)

	receipt := rm.GetReceiptDataByRid(Rid)
	fmt.Println("get receipt")
	if receipt == nil {
		return "", errors.New("Invalid operation, maybe not found the receipt")
	}

	if receipt.InvoiceNo != "" {
		return "", errors.New("Invalid operation, receipt already bind invoiceNo")
	}

	fmt.Println("Allow to CreateInvoice")

	invoice, err := invoiceM.CreateInvoiceDataFromAPI(inputInvoice)
	if err != nil {
		fmt.Println("[CreateInvoiceDataFromAPI ERR:", err)
		return "", err
	}
	if invoice.InvoiceNo == "" {
		fmt.Println(err)
		return "", errors.New("第三方API 建立發票失敗")
	}

	// const sql = `Update public.receipt set invoiceNo = $1 where Rid = $2;`
	// fmt.Println("%s %s", invoice.InvoiceNo, Rid)
	// interdb := invoiceM.imr.GetSQLDB()
	// sqldb, err := interdb.ConnectSQLDB()
	// if err != nil {
	// 	return "", err
	// }
	// res, err := sqldb.Exec(sql, invoice.InvoiceNo, Rid)
	// //res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return "", err
	// }
	// id, err := res.RowsAffected()
	// if err != nil {
	// 	fmt.Println("PG Affecte Wrong: ", err)
	// 	return "", err
	// }
	// fmt.Println(id)

	// if id == 0 {
	// 	return "", errors.New("Invalid operation, maybe not found the receipt")
	// }
	////////////////////////////////////////////////////////////////////////////////

	const invoiceSql = `INSERT INTO public.invoice(
		rid, invoiceno, buyerid, sellerid, randomnum, title, date, amount, left_qrcode, right_qrcode)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`

	interdb := invoiceM.imr.GetSQLDB()
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		return "", err
	}
	fmt.Println(Rid, invoice.InvoiceNo, inputInvoice.BuyerID, storeID, invoice.RandNum, inputInvoice.Title, invoice.Date, invoice.TotalAmount, invoice.Left_qrcode, invoice.Right_qrcode)
	res, err := sqldb.Exec(invoiceSql, Rid, invoice.InvoiceNo, inputInvoice.BuyerID, storeID, invoice.RandNum, inputInvoice.Title, invoice.Date, invoice.TotalAmount, invoice.Left_qrcode, invoice.Right_qrcode)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[ERROR CreateInvoice]", err)
		return "", err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return "", err
	}
	fmt.Println(id)

	if id == 0 {
		return "", errors.New("CreateInvoice Error")
	}

	return invoice.InvoiceNo, nil
}

func (invoiceM *InvoiceModel) MakeQRcodeImageFile(Invoice *Invoice) error {
	//图片的宽度
	dx := 250
	//图片的高度
	dy := 140

	img := image.NewNRGBA(image.Rect(0, 0, dx, dy))

	//设置每个点的 RGBA (Red,Green,Blue,Alpha(设置透明度))
	for y := 0; y < dy; y++ {
		for x := 0; x < dx; x++ {
			//设置一块 白色(255,255,255)不透明的背景
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	///產生 發票資訊
	// Invoice, err := invoiceM.GetInvoiceDataFromAPI(invoiceNo)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil
	// }

	//**************code39 一維條碼********************
	content_code39 := Invoice.InvoiceNo

	fmt.Println("Generating code39 barcode for : ", content_code39)

	// see https://godoc.org/github.com/boombuler/barcode/code39
	bcode, err := code39.Encode(content_code39, false, false)

	if err != nil {
		fmt.Println("String %s cannot be encoded", content_code39)
		fmt.Println(err)
	}
	fmt.Println("Scale")
	// scale to 300x20
	code39, err := barcode.Scale(bcode, 200, 30)

	if err != nil {
		fmt.Println("Code39 scaling error : ", err)
		fmt.Println(err)
	}
	fmt.Println("code39 一維條碼 start")
	//**************code39 一維條碼 end********************

	// 演示base64编码
	// encodeString := base64.StdEncoding.EncodeToString(input)
	// fmt.Println(encodeString)
	fmt.Println("Invoice.Date:", Invoice.Date)
	///產生QR Code
	//TWyear, err := strconv.Atoi(Invoice.Date[0:4])

	// // 4碼 加密驗證資訊
	// ciphertext := invoiceM.get24Encrypted(AES_CBC_Encrypted([]byte(Invoice.InvoiceNo + Invoice.RandNum)))
	// // 發票字軌(10) + 開
	// content := Invoice.InvoiceNo + TW_Date + Invoice.RandNum + fmt.Sprintf("%08x", Invoice.SalesAmount) + fmt.Sprintf("%08x", Invoice.TotalAmount)
	// content += invoiceM.getBuyerID(Invoice.BuyerID) + storeID + ciphertext
	// content += ":" + invoiceM.getRemark(Invoice.Remark)
	// content += invoiceM.getDetail(Invoice.Detail)

	qrsize := 100
	//左邊的qrcode
	qrCode, err := qr.Encode(Invoice.Left_qrcode, qr.M, qr.Auto)
	if err != nil {
		return nil
	}

	// Scale the barcode to 200x200 pixels
	qrCode, err = barcode.Scale(qrCode, qrsize, qrsize)
	if err != nil {
		return nil
	}
	//右邊的qrcode
	qrCode_asterisk, err := qr.Encode(Invoice.Right_qrcode, qr.M, qr.Auto)
	if err != nil {
		return nil
	}

	// Scale the barcode to 200x200 pixels
	qrCode_asterisk, err = barcode.Scale(qrCode_asterisk, qrsize, qrsize)
	if err != nil {
		return nil
	}
	//把水印写在左下角，并向0坐标
	offset := image.Pt(10, img.Bounds().Dy()-qrCode.Bounds().Dy())
	b := img.Bounds()
	//根据b画布的大小新建一个新图像
	m := image.NewRGBA(b)
	//https://studygolang.com/articles/12049
	//image.ZP代表Point结构体，目标的源点，即(0,0)
	//draw.Src源图像透过遮罩后，替换掉目标图像
	//draw.Over源图像透过遮罩后，覆盖在目标图像上（类似图层）
	draw.Draw(m, b, img, image.ZP, draw.Src) // 底圖
	draw.Draw(m, qrCode.Bounds().Add(offset), qrCode, image.ZP, draw.Over)
	//把水印写在右下角，并向0坐标
	offset = image.Pt(img.Bounds().Dx()-qrCode_asterisk.Bounds().Dx()-40, img.Bounds().Dy()-qrCode_asterisk.Bounds().Dy())
	draw.Draw(m, qrCode_asterisk.Bounds().Add(offset), qrCode_asterisk, image.ZP, draw.Over)
	offset_code39 := image.Pt(0, 0)
	draw.Draw(m, code39.Bounds().Add(offset_code39), code39, image.ZP, draw.Over)

	//生成新图片new.jpg,并设置图片质量
	os.MkdirAll(util.PdfDir, os.ModePerm)

	imgw, err := os.Create(util.PdfDir + Invoice.InvoiceNo + ".png")
	png.Encode(imgw, m)
	defer imgw.Close()

	fmt.Println("添加水印图片结束请查看")
	return nil
}

func (invoiceM *InvoiceModel) CustomizedInvoice(p *pdf.Pdf, Invoice *Invoice) {

	width := 130.0
	p.SetPdf_XY(10, -1)
	fontsize := 14.0
	err := p.LoadTTF("TW-Medium", pdf.TW_Medium_PATH, "", int(fontsize))
	if err != nil {
		fmt.Println("loadTTF:", err.Error())
	}

	p.FillText("台 灣 房 屋", fontsize, pdf.ColorTableLine, pdf.AlignCenter, pdf.ValignMiddle, width, pdf.TextHeight)
	p.NewLine(15)
	p.FillText("電子發票證明聯", fontsize, pdf.ColorTableLine, pdf.AlignCenter, pdf.ValignMiddle, width, pdf.TextHeight)
	TWDate, _ := util.ADtoROC(Invoice.Date, "invoice")
	p.NewLine(15)
	p.FillText(TWDate, fontsize, pdf.ColorTableLine, pdf.AlignCenter, pdf.ValignMiddle, width, pdf.TextHeight)
	p.NewLine(15)
	p.FillText(Invoice.InvoiceNo, fontsize, pdf.ColorTableLine, pdf.AlignCenter, pdf.ValignMiddle, width, pdf.TextHeight)
	p.NewLine(15)

	fontsize = 8.0
	err = p.LoadTTF("TW-Medium", pdf.TW_Medium_PATH, "", int(fontsize))
	if err != nil {
		fmt.Println("loadTTF:", err.Error())
	}
	p.FillText(Invoice.Date, fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, width, pdf.TextHeight)
	p.NewLine(10)
	p.FillText("隨機碼"+Invoice.RandNum, fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, width, pdf.TextHeight)
	p.FillText("總計"+strconv.Itoa(Invoice.TotalAmount), fontsize, pdf.ColorTableLine, pdf.AlignRight, pdf.ValignMiddle, width, pdf.TextHeight)
	p.NewLine(10)
	p.FillText("賣方:"+Invoice.SellerID, fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, width, pdf.TextHeight)
	if Invoice.BuyerID != "" {
		p.FillText("買方"+Invoice.BuyerID, fontsize, pdf.ColorTableLine, pdf.AlignRight, pdf.ValignMiddle, width, pdf.TextHeight)
	}
	p.NewLine(120)
	detail_w := 220.0
	detail_h := 140.0
	fontsize = 11.0
	gap := 13.0
	err = p.LoadTTF("TW-Medium", pdf.TW_Medium_PATH, "", int(fontsize))
	if err != nil {
		fmt.Println("loadTTF:", err.Error())
	}
	p.SetPdf_XY(20, -1)
	p.DrawRectangle(detail_w, detail_h, pdf.ColorWhite, "FD")
	p.NewLine(gap)
	detail_w = 200.0
	p.SetPdf_XY(25, -1)
	p.FillText("營業人統編:"+storeID, fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap)
	p.SetPdf_XY(25, -1)
	p.FillText("測試", fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap)
	p.SetPdf_XY(25, -1)
	p.FillText("02-XXXXYYYY", fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap + 10)
	//
	p.SetPdf_XY(25, -1)
	p.FillText("商品名稱", fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap)
	//
	detail_w = 125
	p.SetPdf_XY(75, -1)
	p.FillText("單價", fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.FillText("訂購量", fontsize, pdf.ColorTableLine, pdf.AlignCenter, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.FillText("小計", fontsize, pdf.ColorTableLine, pdf.AlignRight, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap)
	//
	p.SetPdf_XY(25, -1)
	p.FillText("服務費", fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap)
	//
	amount := strconv.Itoa(Invoice.TotalAmount)
	p.SetPdf_XY(75, -1)
	p.FillText(amount, fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.FillText("1", fontsize, pdf.ColorTableLine, pdf.AlignCenter, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	sp := message.NewPrinter(language.English)
	amount = sp.Sprintf("%sTX", amount)
	p.FillText(amount, fontsize, pdf.ColorTableLine, pdf.AlignRight, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.NewLine(gap)
	p.SetPdf_XY(25, -1)
	p.FillText("銷售額", fontsize, pdf.ColorTableLine, pdf.AlignLeft, pdf.ValignMiddle, detail_w, pdf.TextHeight)
	p.FillText(fmt.Sprintf("%.f", round(float64(Invoice.TotalAmount)/1.05, 0)), fontsize, pdf.ColorTableLine, pdf.AlignRight, pdf.ValignMiddle, detail_w+50, pdf.TextHeight)

}
func (invoiceM *InvoiceModel) GetInvoicePDF(rid string, p *pdf.Pdf) {

	Invoice := invoiceM.GetInvoiceData(rid)
	if Invoice == nil {
		fmt.Println("Invoice is null")
		return
	}
	invoiceM.MakeQRcodeImageFile(Invoice)
	//p = pdf.GetOriPDF()
	invoiceM.CustomizedInvoice(p, Invoice)
	p.PutImage(Invoice.InvoiceNo+".png", 10, 110) //print image

	//pdf.Image("3.png", 80, 80, nil)       //print image

	//pdf.Image("qrcode3.png", 200, 200, nil) //print image
	//pdf.Cell(nil, "AA")
	//pdf.Cell(nil, "AAA,您好")

	//Write 寫檔案後，pdf物件資料會釋放掉。
	//pdf.WritePdf("qrcode.pdf")
	return
}

func (invoiceM *InvoiceModel) CreateInvoiceDataFromAPI(iv *Invoice) (*Invoice, error) {
	fmt.Println("CreateInvoiceDataFromAPI")
	if rm == nil {
		fmt.Println("rm is null")
		return nil, errors.New("rm is null")
	}
	r := rm.GetReceiptDataByID(iv.Rid)

	out2, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
		return nil, err

	}
	fmt.Println("Out2\n" + string(out2))
	if r.Rid == "" {
		fmt.Println("not found receipt")
		return nil, errors.New("not found receipt")
	}

	tmap := make(map[string]interface{})
	deatils := make([]map[string]interface{}, 0, 0)
	var deatil = make(map[string]interface{})
	deatil["product_id"] = r.CNo
	deatil["quantity"] = 1
	deatil["product_name"] = r.CaseName
	deatil["unit"] = "件"
	deatil["unit_price"] = r.Amount
	deatil["amount"] = r.Amount

	deatils = append(deatils, deatil)
	tmap["details"] = deatils
	tmap["seller"] = storeID
	tmap["buyer_name"] = "(" + r.CustomerType + "家)" + r.Name
	tmap["buyer_uniform"] = iv.BuyerID
	tmap["sales_amount"] = round(float64(r.Amount)/1.05, 0)    //對第一位小數 四捨五入
	tmap["tax_amount"] = round(float64(r.Amount)/1.05*0.05, 0) //對第一位小數 四捨五入
	tmap["total_amount"] = r.Amount
	tmap["tax_type"] = "1"    //應稅
	tmap["is_exchange"] = "1" //需要到財政部確認發票
	tmap["carrier_type"] = "" //
	tmap["carrier_id1"] = ""  //
	tmap["carrier_id2"] = ""  //
	tmap["remark"] = iv.Title //

	//tmap["string"] = "Value 01"

	out, err := json.MarshalIndent(tmap, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(out))

	client := &http.Client{}
	//auth_token := "A2dC56TZfkpra6TCIuo5UQW870L/0=mazjT0g7=s5Do0K9z"
	//url := "https://ranking.numax.com.tw/test/einvoice/api/invoice"
	// detail := Detail{
	// 	Name:      "項目一",
	// 	Quantity:  2,
	// 	UnitPrice: 15000,
	// }

	req, err := http.NewRequest("POST", ivURL, bytes.NewBuffer(out))
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	req.Header.Set("Authorization", auth_iv)

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	defer resp.Body.Close()

	//復升API回傳的Body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}
	//預防亂碼轉換
	fmt.Println(string(result))
	result = bytes.TrimPrefix(result, []byte("\xef\xbb\xbf"))

	//解析至map[string]
	var data map[string]interface{}
	json.Unmarshal([]byte(result), &data)
	//(just for print)從 map[string] 取 result 的json資料
	InvoiceData := data["result"].(map[string]interface{})
	fmt.Println("Out\n", InvoiceData)

	//發票資料結構
	var invoice *Invoice
	// mapstructure.Decode(InvoiceData, &invoice)
	// out, err := json.Marshal(invoice)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// fmt.Println("Out1\n" + string(out))

	//從body取出key為result的json物件
	value := gjson.Get(string(result), "result")
	//將資料轉換為struct
	err = json.Unmarshal([]byte(value.String()), &invoice)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		out, err := json.Marshal(invoice)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println("Out2\n" + string(out))
	}

	return invoice, nil
}

func round(v float64, decimals int) float64 {
	var pow float64 = 1
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int((v*pow)+0.5)) / pow
}

func (invoiceM *InvoiceModel) GetInvoiceDataFromAPI(invoiceNo string) (*Invoice, error) {
	// var systemBranchDataList []*SystemBranch
	// var s1, s2, s3, s4 SystemBranch
	// s1.Branch = "北京店"
	// s2.Branch = "東京店"
	// s3.Branch = "西京店"
	// s4.Branch = "南京店"
	// systemBranchDataList = append(systemBranchDataList, &s1)
	// systemBranchDataList = append(systemBranchDataList, &s2)
	// systemBranchDataList = append(systemBranchDataList, &s3)
	// systemBranchDataList = append(systemBranchDataList, &s4)

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))
	// systemM.systemBranchList = systemBranchDataList

	client := &http.Client{}
	auth_token := "A2dC56TZfkpra6TCIuo5UQW870L/0=mazjT0g7=s5Do0K9z"
	url := "https://ranking.numax.com.tw/test/einvoice/api/invoice?seller=" + storeID
	req, err := http.NewRequest("GET", url+"&invoice_no="+invoiceNo, nil)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	req.Header.Set("Authorization", auth_token)

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	fmt.Println(string(result))
	result = bytes.TrimPrefix(result, []byte("\xef\xbb\xbf"))

	var data map[string]interface{}
	json.Unmarshal([]byte(result), &data)

	InvoiceData := data["result"].(map[string]interface{})
	fmt.Println("Out\n", InvoiceData)

	var invoice *Invoice
	// mapstructure.Decode(InvoiceData, &invoice)
	// out, err := json.Marshal(invoice)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return nil, err
	// }
	// fmt.Println("Out1\n" + string(out))

	value := gjson.Get(string(result), "result")

	err = json.Unmarshal([]byte(value.String()), &invoice)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		out, err := json.Marshal(invoice)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println("Out2\n" + string(out))
	}

	return invoice, nil
}

func (invoiceM *InvoiceModel) getBuyerID(id string) string {
	if id == "" {
		return "00000000"
	}
	return id
}

func (invoiceM *InvoiceModel) getRemark(remark string) string {
	if remark == "" {
		return "**********"
	}
	return remark
}

func (invoiceM *InvoiceModel) getDetail(details []*Detail) string {
	var result = ""
	for _, Detail := range details {
		result += ":" + Detail.Name + ":" + strconv.Itoa(Detail.Quantity) + ":" + strconv.Itoa(Detail.UnitPrice)
	}

	return result
}

func (invoiceM *InvoiceModel) get24Encrypted(text string) string {

	data, err := hex.DecodeString(text)
	if err != nil {
		println(err)
		return ""
	}

	sEnc := base64.StdEncoding.EncodeToString(data)
	fmt.Println("sEnc:", sEnc)
	return sEnc
}

func PKCS7Padding(ciphertext []byte) []byte {
	padding := aes.BlockSize - len(ciphertext)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(plantText []byte) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	return plantText[:(length - unpadding)]
}

func AES_CBC_Encrypted(plaintext []byte) string {
	iv_base64 := "Dt8lyToo17X/XkXaQvihuA=="
	aes_key := "CB211F126E1E12C2ACE4BC3145085A50"
	p, err := base64.StdEncoding.DecodeString(iv_base64)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	h := hex.EncodeToString(p)
	fmt.Println("iv:" + h)            // prints 415256494e
	fmt.Println("aes_key:" + aes_key) // prints 415256494e

	key, _ := hex.DecodeString(aes_key)
	iv, _ := hex.DecodeString(hex.EncodeToString(p))
	//plaintext := []byte(invoiceNum)

	// -------- 加密开始---------
	plaintext = PKCS7Padding(plaintext)
	ciphertext := make([]byte, len(plaintext))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)
	fmt.Printf("ciphertext:%x\n", ciphertext) //ehU8DinWePXYxFyQXuZf8g==

	ciphertext_str := fmt.Sprintf("%x", ciphertext)
	return ciphertext_str
	// ----------------解密开始---------

	// mode = cipher.NewCBCDecrypter(block, iv)
	// mode.CryptBlocks(ciphertext, ciphertext)

	// ciphertext = PKCS7UnPadding(ciphertext)
	// fmt.Printf("%s\n", ciphertext)

	//

}
