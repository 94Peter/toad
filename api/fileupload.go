package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"toad/model"
	"toad/permission"

	"github.com/axgle/mahonia"
)

type FileuploadAPI bool

const uploadPath = "/tmp"

//const uploadPath = "C:\Users\Oswin\AppData\Local\Packages\CanonicalGroupLimited.Ubuntu16.04onWindows_79rhkp1fndgsc\LocalState\rootfs\tmp"

func (api FileuploadAPI) Enable() bool {
	return bool(api)
}

func (api FileuploadAPI) GetAPIs() *[]*APIHandler {
	return &[]*APIHandler{
		&APIHandler{Path: "/v1/upload", Next: api.uploadFile, Method: "POST", Auth: true, Group: permission.All},
	}
}

func (api *FileuploadAPI) uploadFile(w http.ResponseWriter, r *http.Request) {
	dbname := r.Header.Get("dbname")
	r.ParseMultipartForm(100)
	mForm := r.MultipartForm
	am := model.GetARModel(di)
	rt := model.GetRTModel(di)

	datas := make([]map[string]interface{}, 0, 0)
	f := false
	for k, _ := range mForm.File {
		// k is the key of file part
		file, fileHeader, err := r.FormFile(k)
		if err != nil {
			fmt.Println("inovke FormFile error:", err)
			return
		}
		defer file.Close()
		fmt.Printf("the uploaded file: name[%s], size[%d], header[%#v]\n",
			fileHeader.Filename, fileHeader.Size, fileHeader.Header)

		// store uploaded file into local path
		localFileName := uploadPath + "/" + fileHeader.Filename
		out, err := os.Create(localFileName)
		if err != nil {
			fmt.Printf("failed to open the file %s for writing", localFileName, " err:", err)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			fmt.Printf("copy file err:%s\n", err)
			return
		}
		fmt.Printf("file %s uploaded ok, localFileName %s\n", fileHeader.Filename, localFileName)

		records := readCsvFile(localFileName)
		fmt.Println(records)

		sqlDB := am.GetSqlDB(dbname)
		for _, row := range records {
			fmt.Println(len(row))
			if len(row) <= 7 {
				f = false
				break
			}
			if f {
				name := SubString(row[6], 4, 3)
				branch := SubString(row[6], 0, 2)
				mtype := SubString(row[6], 2, 2)
				fmt.Println(mtype, name, branch, dbname)
				mresult := am.CheckARExist(mtype, name, branch, sqlDB)
				if mresult == "" {
					fmt.Println("this data is not contain:", row)
					var data = make(map[string]interface{})
					data["date"] = row[0]
					data["time"] = row[1]
					data["bank"] = row[2]
					data["order"] = row[3]
					data["sender"] = row[4]
					data["amount"] = row[5]
					data["item"] = row[6]
					data["desc"] = row[7]
					data["reason"] = "找不到相關應收款 或 存在重複應收款"
					datas = append(datas, data)
				} else {
					row[5] = strings.Replace(row[5], ",", "", -1)

					amount, _ := strconv.Atoi(row[5])

					ymd := strings.Split(row[0], "/")
					date := ymd[0] + "-"
					//fmt.Println("ymd:", ymd, "----", len(ymd[0]), len(ymd[1]), len(ymd[2]))
					if len(ymd[1]) == 1 {
						date += "0" + ymd[1]
					} else {
						date += ymd[1]
					}
					date += "-"
					if len(ymd[2]) == 1 {
						date += "0" + ymd[2]
					} else {
						date += ymd[2]
					}
					date += "T" + row[1] + "+08:00"

					t, _ := time.Parse(time.RFC3339, date)

					receipt := &model.Receipt{
						ARid:        mresult, //"1617727588", //mresult,
						Amount:      amount,
						Description: row[7],
						Item:        row[6],
						Fee:         0,
						Date:        t,
						// t := time.Now().Unix()
						// res, err := sqldb.Exec(sql, t, rt.Date, rt.Amount, rt.ARid, rt.Fee, rt.Item, rt.Description)
					}
					//fmt.Println(receipt)
					err := rt.CreateReceipt(receipt, dbname, sqlDB, &t)
					if err != nil {
						fmt.Println("this data is duplicate:", row)
						var data = make(map[string]interface{})
						data["date"] = row[0]
						data["time"] = row[1]
						data["bank"] = row[2]
						data["order"] = row[3]
						data["sender"] = row[4]
						data["amount"] = row[5]
						data["item"] = row[6]
						data["desc"] = row[7]
						data["reason"] = "[匯入錯誤]" + err.Error()
						datas = append(datas, data)
					}
					//rt *Receipt, dbname string, sqldb *sql.DB, idTime *time.Time)
				}
			}
			if row[6] == "附言" {
				f = true
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if f {
		json.NewEncoder(w).Encode(datas)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("file format error"))
	}
}

func SubString(str string, begin, length int) string {

	rs := []rune(str)
	lth := len(rs)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}
	return string(rs[begin:end])
}

func readCsvFile(filePath string) [][]string {

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	// csvReader := csv.NewReader(f)
	// records, err := csvReader.ReadAll()
	// if err != nil {
	// 	fmt.Printf("Unable to parse file as CSV for "+filePath, err)
	// }

	decoder := mahonia.NewDecoder("big5") // 把原来ANSI格式的文本文件里的字符，用gbk进行解码。
	// r := csv.NewReader(file)
	r := csv.NewReader(decoder.NewReader(f)) // 这样，最终返回的字符串就是utf-8了。（go只认utf8）
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("Error when read csv in GetCSV():", err)
		return nil
	}

	// for _, row := range records {
	// 	for _, data := range row {
	// 		k := UseNewEncoder(data, "big5", "utf-8")
	// 		fmt.Println(k)
	// 	}
	// }

	return records
}

func UseNewEncoder(src string, oldEncoder string, newEncoder string) string {
	srcDecoder := mahonia.NewDecoder(oldEncoder)
	desDecoder := mahonia.NewDecoder(newEncoder)
	resStr := srcDecoder.ConvertString(src)
	_, resBytes, _ := desDecoder.Translate([]byte(resStr), true)
	return string(resBytes)
}
