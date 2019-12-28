package util

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const PdfDir = "MyPdf/"

func CompressZip(fname string) {

	//獲取原始檔列表
	f, err := ioutil.ReadDir(PdfDir)
	if err != nil {
		fmt.Println(err)
	}
	fzip, _ := os.Create(PdfDir + fname + ".zip")
	w := zip.NewWriter(fzip)
	defer w.Close()
	for _, file := range f {
		if strings.Contains(file.Name(), ".pdf") || strings.Contains(file.Name(), ".PDF") {
			fw, _ := w.Create(file.Name())
			filecontent, err := ioutil.ReadFile(PdfDir + file.Name())
			if err != nil {
				fmt.Println(err)
			}
			n, err := fw.Write(filecontent)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(n)
		}
	}

}

func DeleteAllFile() {
	fmt.Printf("DeleteAllFile:" + PdfDir)
	os.RemoveAll(PdfDir)
}
