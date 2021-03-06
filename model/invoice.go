package model

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
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
	"math"
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
	ivURL       = "https://ranking.numax.com.tw/test/einvoice/api/invoice"
	ivReturnURL = "https://ranking.numax.com.tw/test/einvoice/api/allowance"

	auth_iv = "4/00OB50qLc==rNPi+eE+8+8fWi/i5AK65Mz7NTsxFJem3q" //"A2dC56TZfkpra6TCIuo5UQW870L/0=mazjT0g7=s5Do0K9z"
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
	Signatrue string `json:"-"`
	//No          string    `json:"invoice_no"`
	RandNum     string    `json:"random_number"` //解析字串要用到，回傳前端就當作多餘的吧
	Date        string    `json:"invoice_datetime"`
	SalesAmount float64   `json:"-"`
	TotalAmount int       `json:"amount"`
	BuyerID     string    `json:"buyerID"`  //買家統編
	SellerID    string    `json:"sellerID"` //賣家統編(台灣房屋?)
	Remark      string    `json:"remark"`
	Detail      []*Detail `json:"-"`

	Rid string `json:"rid"`
	Sid string `json:"sid"`
	//Date    time.Time `json:"date"`
	Title     string `json:"title"`
	InvoiceNo string `json:"invoice_no"`
	//Amount    string `json:"amount"`
	//Invoice      string `json:"invoice"`
	Left_qrcode   string //解析復升API字串要用到，回傳前端就當作多餘的吧
	Right_qrcode  string //解析復升API字串要用到，回傳前端就當作多餘的吧
	Status        string `json:"status"`
	InvoiceStatus int    `json:"invoice_status"`
	Branch        string `json:"branch"` //
}

type InvoiceConfig struct {
	SellerID string `json:"sellerID"`
	Auth     string `json:"auth"`
	Branch   string `json:"branch"`
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

func (invoiceM *InvoiceModel) GetInvoiceData(rid, dbname string) *Invoice {

	const qspl = `SELECT rid, sid, invoiceno, buyerID, sellerID, randomnum, title, date, amount, left_qrcode, right_qrcode FROM public.Invoice where rid = '%s';`
	db := invoiceM.imr.GetSQLDBwithDbname(dbname)
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
		if err := rows.Scan(&invoice.Rid, &invoice.Sid, &invoice.InvoiceNo, &invoice.BuyerID, &invoice.SellerID, &invoice.RandNum, &invoice.Title, &invoice.Date, &invoice.TotalAmount, &invoice.Left_qrcode, &invoice.Right_qrcode); err != nil {
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
func (invoiceM *InvoiceModel) GetInvoiceDataByArid(arid, dbname string) []*Invoice {

	const qspl = `SELECT iv.rid, iv.sid, invoiceno, buyerid, sellerid, randomnum, title, iv.date, iv.amount, left_qrcode, right_qrcode, iv.status, iv.invoice_status
				FROM public.invoice iv
				inner join public.receipt r on iv.rid = r.rid
				inner join public.ar ar on r.arid = ar.arid and ar.arid = '%s';`

	db := invoiceM.imr.GetSQLDBwithDbname(dbname)
	rows, err := db.SQLCommand(fmt.Sprintf(qspl, arid))
	if err != nil {
		fmt.Println("GetInvoiceDataByArid:", err)
		return nil
	}

	ivList := []*Invoice{}

	for rows.Next() {
		invoice := &Invoice{}

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&invoice.Rid, &invoice.Sid, &invoice.InvoiceNo, &invoice.BuyerID, &invoice.SellerID, &invoice.RandNum, &invoice.Title, &invoice.Date, &invoice.TotalAmount, &invoice.Left_qrcode, &invoice.Right_qrcode, &invoice.Status, &invoice.InvoiceStatus); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		ivList = append(ivList, invoice)
	}

	// out, err := json.Marshal(arList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	return ivList
}
func (invoiceM *InvoiceModel) Json() ([]byte, error) {
	return json.Marshal(invoiceM.invoiceList)
}

func (invoiceM *InvoiceModel) UpdateInvoiceStatus() {
	const qspl = `SELECT rid, sid, invoiceno, buyerID, sellerID, randomnum, title, date, amount, branch , status , invoice_status FROM public.Invoice ;`
	db := invoiceM.imr.GetSQLDBwithDbname("toad")
	sqldb, err := db.ConnectSQLDB()
	if err != nil {
		return
	}
	rows, err := sqldb.Query(fmt.Sprintf(qspl))
	if err != nil {
		return
	}

	invoiceList := []*Invoice{}
	for rows.Next() {
		invoice := &Invoice{}

		// if err := rows.Scan(&r.ARid, &s); err != nil {
		// 	fmt.Println("err Scan " + err.Error())
		// }
		if err := rows.Scan(&invoice.Rid, &invoice.Sid, &invoice.InvoiceNo, &invoice.BuyerID, &invoice.SellerID, &invoice.RandNum, &invoice.Title, &invoice.Date, &invoice.TotalAmount, &invoice.Branch,
			&invoice.Status, &invoice.InvoiceStatus); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		invoiceList = append(invoiceList, invoice)
	}
	// out, err := json.Marshal(invoiceList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(out))

	ivconfigList, _ := invoiceM.GetInvoiceConfig("%", "toad")
	for _, ivConfig := range ivconfigList {
		invoiceM.UpdateInvoiceStatusFromAPI(ivConfig, sqldb)
	}

	for _, data := range invoiceList {
		if data.Status == "-4" {
			for _, ivConfig := range ivconfigList {
				fmt.Println(ivConfig)
				if data.Branch == ivConfig.Branch {
					invoiceM.DeleteInvoiceFromAPI(data, ivConfig.SellerID, ivConfig.Auth)
					break
				}
			}
		} else if data.Status == "-7" {
			for _, ivConfig := range ivconfigList {
				if data.Branch == ivConfig.Branch {
					invoiceM.ReturnsInvoiceFromAPI(data, ivConfig.SellerID, ivConfig.Auth)
					break
				}
			}
		}
	}
	defer sqldb.Close()
}

func (invoiceM *InvoiceModel) GetInvoiceConfig(branch, dbname string) ([]*InvoiceConfig, error) {
	const invoiceSql = `Select sellerid, auth, branch from public.invoiceconfig where branch like $1;`

	interdb := invoiceM.imr.GetSQLDBwithDbname(dbname)
	sqldb, _ := interdb.ConnectSQLDB()

	rows, _ := sqldb.Query(invoiceSql, branch)
	InvoiceConfigList := []*InvoiceConfig{}
	for rows.Next() {
		var InvoiceConfig InvoiceConfig
		if err := rows.Scan(&InvoiceConfig.SellerID, &InvoiceConfig.Auth, &InvoiceConfig.Branch); err != nil {
			fmt.Println("err Scan " + err.Error())
			return nil, err
		}
		InvoiceConfigList = append(InvoiceConfigList, &InvoiceConfig)
	}
	defer sqldb.Close()
	return InvoiceConfigList, nil
}

func (invoiceM *InvoiceModel) GetCommissionBranchByRid(rid, dbname string) ([]*Commission, error) {
	const invoiceSql = `SELECT c.sid, cs.branch, c.sr from public.commission c
	inner join public.configSaler cs on c.sid = cs.sid
	 where rid = $1;`

	interdb := invoiceM.imr.GetSQLDBwithDbname(dbname)
	sqldb, _ := interdb.ConnectSQLDB()

	rows, _ := sqldb.Query(invoiceSql, rid)
	var cDataList []*Commission

	for rows.Next() {
		var c Commission
		if err := rows.Scan(&c.Sid, &c.Branch, &c.SR); err != nil {
			fmt.Println("err Scan " + err.Error())
		}
		cDataList = append(cDataList, &c)

	}
	defer sqldb.Close()
	return cDataList, nil
}

func (invoiceM *InvoiceModel) CreateInvoiceConfig(inputInvoiceConfig *InvoiceConfig, dbname string) error {

	InvoiceConfigList, err := invoiceM.GetInvoiceConfig(inputInvoiceConfig.Branch, dbname)
	for _, data := range InvoiceConfigList {
		if data.Auth != "" && data.Branch == inputInvoiceConfig.Branch {
			return errors.New("delete old setting first")
		}
	}

	const invoiceSql = `INSERT INTO public.invoiceconfig(auth, SellerID ,Branch) 
		VALUES ($1, $2, $3);`

	interdb := invoiceM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()

	res, err := sqldb.Exec(invoiceSql, inputInvoiceConfig.Auth, inputInvoiceConfig.SellerID, inputInvoiceConfig.Branch)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[ERROR CreateInvoiceConfig]", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("CreateInvoiceConfig Error")
	}
	defer sqldb.Close()
	return nil
}

func (invoiceM *InvoiceModel) updateInvoiceStatusByIvNo(invoiceno, status string, sqldb *sql.DB) (err error) {

	const sql = `UPDATE public.invoice set status = $1 where invoiceno = $2 ;`

	_, err = sqldb.Exec(sql, status, invoiceno)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("updateInvoiceStatus:", err)
	}

	return nil
}
func (invoiceM *InvoiceModel) updateInvoiceStatusInvoiceByIvNo(invoiceno string, invoice_status int, sqldb *sql.DB) (err error) {

	const sql = `UPDATE public.invoice set invoice_status = $1 where invoiceno = $2 ;`

	_, err = sqldb.Exec(sql, invoice_status, invoiceno)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("updateInvoiceStatusInvoiceByIvNo:", err)
	}

	return nil
}
func (invoiceM *InvoiceModel) UpdateInvoiceStatusFromAPI(iv *InvoiceConfig, sqldb *sql.DB) {
	fmt.Println("UpdateInvoiceStatusFromAPI")

	client := &http.Client{}
	auth_token := iv.Auth //"fJXg+7Z4I5e1QOg+9cQ28wa971/8b8qC1=IsJ+5x8xi+4P="
	url := ivURL + "?seller=" + iv.SellerID + "&invoice_start_date=" + "2021-03-01"
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
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
		return
	}

	//fmt.Println(string(result))
	result = bytes.TrimPrefix(result, []byte("\xef\xbb\xbf"))

	var data map[string]interface{}
	json.Unmarshal([]byte(result), &data)

	//InvoiceData := data["result"].(map[string]interface{})
	//fmt.Println("Out\n", InvoiceData)
	//fmt.Println("Out result data\n", InvoiceData["datas"])
	//var invoice *Invoice
	var invoiceList []*Invoice

	value := gjson.Get(string(result), "result")
	datas := value.Get("datas")
	//fmt.Println("Out value\n", value.Get("datas"))

	err = json.Unmarshal([]byte(datas.String()), &invoiceList)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		// out, err := json.Marshal(invoiceList)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		// fmt.Println("Out2\n" + string(out))
		for _, data := range invoiceList {
			invoiceM.updateInvoiceStatusInvoiceByIvNo(data.InvoiceNo, data.InvoiceStatus, sqldb)
		}
	}

	return
}

func (invoiceM *InvoiceModel) DeleteInvoiceFromAPI(iv *Invoice, sellerID, auth string) (*Invoice, error) {
	fmt.Println("DeleteInvoiceFromAPI")

	//auth := "fJXg+7Z4I5e1QOg+9cQ28wa971/8b8qC1=IsJ+5x8xi+4P="
	//r := rm.GetReceiptDataByID(sqldb, iv.Rid)

	tmap := make(map[string]interface{})

	tmap["seller"] = sellerID //"25077808"
	tmap["invoice_no"] = iv.InvoiceNo
	tmap["cancel_reason"] = "折讓款刪除"

	//tmap["string"] = "Value 01"

	out, err := json.MarshalIndent(tmap, "", "  ")
	if err != nil {
		log.Fatalln(err)
	}
	//fmt.Println(string(out))

	client := &http.Client{}
	//auth_token := "A2dC56TZfkpra6TCIuo5UQW870L/0=mazjT0g7=s5Do0K9z"
	//url := "https://ranking.numax.com.tw/test/einvoice/api/invoice"
	// detail := Detail{
	// 	Name:      "項目一",
	// 	Quantity:  2,
	// 	UnitPrice: 15000,
	// }

	req, err := http.NewRequest("DELETE", ivURL, bytes.NewBuffer(out))
	if err != nil {
		// handle error
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	//復升API回傳的Body
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//預防亂碼轉換
	fmt.Println("復升:", string(result))
	result = bytes.TrimPrefix(result, []byte("\xef\xbb\xbf"))

	//解析至map[string]
	var data map[string]interface{}
	json.Unmarshal([]byte(result), &data)
	//(just for print)從 map[string] 取 result 的json資料
	if data["result"] != nil {
		//InvoiceData := data["result"].(map[string]interface{})
		//fmt.Println("Out\n", InvoiceData)
	} else {
		return nil, errors.New("api from 復升 failed:" + data["msg"].(string))
	}

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
		// out, err := json.Marshal(invoice)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return nil, err
		// }
		// fmt.Println("Out2\n" + string(out))

	}

	return invoice, nil
}

//折讓發票
func (invoiceM *InvoiceModel) ReturnsInvoiceFromAPI(iv *Invoice, sellerID, auth string) (*Invoice, error) {
	fmt.Println("ReturnsInvoiceFromAPI")

	//auth := "fJXg+7Z4I5e1QOg+9cQ28wa971/8b8qC1=IsJ+5x8xi+4P="
	//r := rm.GetReceiptDataByID(sqldb, iv.Rid)

	//r := rm.GetReceiptDataByID(sqldb, iv.Rid)

	tmap := make(map[string]interface{})
	deatils := make([]map[string]interface{}, 0, 0)
	var deatil = make(map[string]interface{})
	deatil["invoice_no"] = iv.InvoiceNo
	deatil["quantity"] = 5
	deatil["product_name"] = "鋼筆"
	deatil["unit"] = "件"
	deatil["unit_price"] = 100
	deatil["amount"] = 500
	deatil["tax"] = 495
	deatil["tax_type"] = 1
	deatil["tax_tate"] = 0

	deatils = append(deatils, deatil)

	tmap["seller"] = sellerID
	tmap["invoice_to"] = "B"
	tmap["is_exchange"] = "1"
	tmap["allowance_no"] = ""
	tmap["allowance_datetime"] = ""
	tmap["buyer_name"] = "testName"
	tmap["buyer_uniform"] = "31245875"
	tmap["tax_amount"] = 15
	tmap["total_amount"] = 315
	tmap["details"] = deatils

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

	req, err := http.NewRequest("POST", ivReturnURL, bytes.NewBuffer(out))
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	req.Header.Set("Authorization", auth)

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
	fmt.Println("復升:", string(result))
	result = bytes.TrimPrefix(result, []byte("\xef\xbb\xbf"))

	//解析至map[string]
	var data map[string]interface{}
	json.Unmarshal([]byte(result), &data)
	//(just for print)從 map[string] 取 result 的json資料
	if data["result"] != nil {
		InvoiceData := data["result"].(map[string]interface{})
		fmt.Println("Out\n", InvoiceData)
	} else {
		return nil, errors.New("api from 復升 failed:" + data["msg"].(string))
	}

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
		// out, err := json.Marshal(invoice)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return nil, err
		// }
		// fmt.Println("Out2\n" + string(out))

	}

	return invoice, nil
}

func (invoiceM *InvoiceModel) DeleteInvoiceConfig(branch, dbname string) error {
	const sql = `Delete FROM public.invoiceconfig where branch = $1`

	interdb := invoiceM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	res, err := sqldb.Exec(sql, branch)
	//res, err := sqldb.Exec(sql, unix_time, receivable.Date, receivable.CNo, receivable.Sales)
	if err != nil {
		fmt.Println("[ERROR DeleteInvoiceConfig]", err)
		return err
	}
	id, err := res.RowsAffected()
	if err != nil {
		fmt.Println("PG Affecte Wrong: ", err)
		return err
	}
	fmt.Println(id)

	if id == 0 {
		return errors.New("DeleteInvoiceConfig Error")
	}
	defer sqldb.Close()
	return nil
}

func (invoiceM *InvoiceModel) CreateInvoice(inputInvoice *Invoice, dbname string) (string, error) {
	Rid := inputInvoice.Rid

	interdb := invoiceM.imr.GetSQLDBwithDbname(dbname)
	sqldb, err := interdb.ConnectSQLDB()
	if err != nil {
		fmt.Println("[ERROR CreateInvoice ConnectSQLDB]", err)
		return "", err
	}

	fmt.Println(Rid)
	result := ""
	receipt := rm.GetReceiptDataByID(sqldb, Rid)

	if receipt == nil {
		fmt.Println("receipt is null")
		return "", errors.New("Invalid operation, maybe not found the receipt")
	}

	// if receipt.InvoiceNo != "" {
	// 	return "", errors.New("Invalid operation, receipt already bind invoiceNo")
	// }
	if len(receipt.InvoiceData) > 0 {
		return "", errors.New("Invalid operation, receipt already bind invoiceNo")
	}

	fmt.Println("Allow to CreateInvoice")
	dataList, _ := invoiceM.GetCommissionBranchByRid(Rid, dbname)
	ivconfigList, _ := invoiceM.GetInvoiceConfig("%", dbname)
	count := 0
	mybranch := ""
	for _, data := range dataList {
		mybranch += data.Branch + " "
		for _, config := range ivconfigList {
			if data.Branch == config.Branch && config.Auth != "" && config.SellerID != "" {
				count++
			}
		}
	}
	fmt.Println("count:", count)

	if count != len(dataList) {
		return "", errors.New("invoice setting error, plesse check " + mybranch + " setting")
	}

	for _, data := range dataList {

		auth := ""
		sellerID := ""
		for _, config := range ivconfigList {
			if data.Branch == config.Branch && config.Auth != "" && config.SellerID != "" {
				auth = config.Auth
				sellerID = config.SellerID
				break
			}
		}
		//invoice, err := invoiceM.CreateInvoiceDataFromAPI(inputInvoice, data.Branch, dbname)
		invoice, err := invoiceM.CreateInvoiceDataFromAPI_V2(sqldb, inputInvoice, data, sellerID, auth, dbname)
		if err != nil {
			fmt.Println("[CreateInvoiceDataFromAPI ERR:", err)
			return "", err
		}
		if invoice.InvoiceNo == "" {
			fmt.Println(err)
			return "", errors.New("第三方API 建立發票失敗")
		}

		const invoiceSql = `INSERT INTO public.invoice(
		rid, invoiceno, buyerid, sellerid, randomnum, title, date, amount, left_qrcode, right_qrcode, sid, branch)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`

		if err != nil {
			return "", err
		}
		fmt.Println(Rid, invoice.InvoiceNo, inputInvoice.BuyerID, storeID, invoice.RandNum, inputInvoice.Title, invoice.Date, invoice.TotalAmount, invoice.Left_qrcode, invoice.Right_qrcode)
		res, err := sqldb.Exec(invoiceSql, Rid, invoice.InvoiceNo, inputInvoice.BuyerID, storeID, invoice.RandNum, inputInvoice.Title, invoice.Date, invoice.TotalAmount, invoice.Left_qrcode, invoice.Right_qrcode, data.Sid, data.Branch)
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
		result += invoice.InvoiceNo + ";"
		defer sqldb.Close()
	}

	return result, nil
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
func (invoiceM *InvoiceModel) GetInvoicePDF(rid, dbname string, p *pdf.Pdf) {

	Invoice := invoiceM.GetInvoiceData(rid, dbname)
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

func (invoiceM *InvoiceModel) CreateInvoiceDataFromAPI(sqldb *sql.DB, iv *Invoice, branch, dbname string) (*Invoice, error) {
	fmt.Println("CreateInvoiceDataFromAPI")
	if rm == nil {
		fmt.Println("rm is null")
		return nil, errors.New("rm is null")
	}
	r := rm.GetReceiptDataByID(sqldb, iv.Rid)
	ivdvList, err := invoiceM.GetInvoiceConfig(branch, dbname)

	if len(ivdvList) == 0 {
		return nil, errors.New("invoice setting error, cannot find " + branch)
	}

	ivdv := ivdvList[0]
	if ivdv.Auth == "" {
		return nil, errors.New("invoice setting error")
	}

	out1, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
		return nil, err

	}
	fmt.Println("Out 1\n" + string(out1))
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
	tmap["seller"] = ivdv.SellerID
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

	req.Header.Set("Authorization", ivdv.Auth)

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
	if data["result"] != nil {
		InvoiceData := data["result"].(map[string]interface{})
		fmt.Println("Out\n", InvoiceData)
	} else {
		return nil, errors.New("api from 復升 failed:" + data["msg"].(string))
	}

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
		fmt.Println("Out v1\n" + string(out))

	}

	return invoice, nil
}

func (invoiceM *InvoiceModel) CreateInvoiceDataFromAPI_V2(sqldb *sql.DB, iv *Invoice, commission *Commission, sellerID, auth, dbname string) (*Invoice, error) {
	fmt.Println("CreateInvoiceDataFromAPI_V2")

	if auth == "" {
		return nil, errors.New("auth invoice setting error")
	}
	if sellerID == "" {
		return nil, errors.New("sellerID invoice setting error")
	}

	r := rm.GetReceiptDataByID(sqldb, iv.Rid)

	tmap := make(map[string]interface{})
	deatils := make([]map[string]interface{}, 0, 0)
	var deatil = make(map[string]interface{})
	deatil["product_id"] = r.CNo
	deatil["quantity"] = 1
	deatil["product_name"] = r.CaseName
	deatil["unit"] = "件"
	deatil["unit_price"] = commission.SR
	deatil["amount"] = commission.SR

	deatils = append(deatils, deatil)
	tmap["details"] = deatils
	tmap["seller"] = sellerID
	tmap["buyer_name"] = "(" + r.CustomerType + "家)" + r.Name
	tmap["buyer_uniform"] = iv.BuyerID
	tmap["sales_amount"] = round(float64(commission.SR)/1.05, 0)    //對第一位小數 四捨五入
	tmap["tax_amount"] = round(float64(commission.SR)/1.05*0.05, 0) //對第一位小數 四捨五入
	tmap["total_amount"] = commission.SR
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

	req.Header.Set("Authorization", auth)

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
	if data["result"] != nil {
		InvoiceData := data["result"].(map[string]interface{})
		fmt.Println("Out\n", InvoiceData)
	} else {
		return nil, errors.New("api from 復升 failed:" + data["msg"].(string))
	}

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

func round(x float64, decimals int) float64 {
	return float64(int(math.Floor(x + 0.5000000001)))
}

//原文網址：https://kknews.cc/news/val5oky.html
//四捨五入 取精度
// func round(f float64, places int) float64 {
// 	shift := math.Pow(10, float64(places))
// 	fv := 0.0000000001 + f //對浮點數產生.xxx999999999 計算不准進行處理
// 	return math.Floor(fv*shift+.5) / shift
// }

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
