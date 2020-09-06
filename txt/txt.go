package txt

import (
	"bufio"
	"fmt"
	"os"
	"toad/util"
)

func Write(outstr, filename string) {

	err := os.MkdirAll(util.PdfDir, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	outputf, err := os.OpenFile(util.PdfDir+filename+".txt", os.O_CREATE|os.O_WRONLY, 0664)

	//
	// 是用 OpenFile,不是只用Open,因為還要設定模式. 建立檔案 只有寫入  UNIX檔案權限
	if err != nil {
		fmt.Println("open file error!")
		return
	}
	defer outputf.Close()
	// ^^^^
	// 離開時關檔

	outputWriter := bufio.NewWriter(outputf)
	//^^^^^^^^^^^^
	// 建立緩衝輸出物件

	_, err = outputWriter.WriteString(outstr)
	if err != nil {
		fmt.Println(err)
	}

	outputWriter.Flush()
	//             ^^^^^^
	// 因為是緩衝式,最後要強制寫入

}
