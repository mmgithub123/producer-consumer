package us3


import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"strings"
	"time"

	ufsdk "github.com/ufilesdk-dev/ufile-gosdk"
)

const (
	delim byte = '\n'
)

func downloadFile(filename string, req *ufsdk.UFileRequest) (buf bytes.Buffer) {
	// 下载文件
	log.Println("下载文件: ")
	reqUrl := req.GetPrivateURL(filename, 10*time.Second)
	err := req.Download(reqUrl)
	if err != nil {
		log.Fatalln(string(req.DumpResponse(true)))
	}
	log.Printf("下载文件成功！")
	log.Println("print file context:")
	//log.Println(string(req.LastResponseBody))
	us3ResponseBodyReader := strings.NewReader(string(req.LastResponseBody))
	gzipReader, err := gzip.NewReader(us3ResponseBodyReader)
	if err != nil {
		log.Println("get gzipreader fail")
		return
	}
	defer gzipReader.Close()
	_, err = io.Copy(&buf, gzipReader)
	if err != nil {
		log.Println("io copy gzipreader fail")
		return
	}
	//log.Println(buf.String())
	return buf
}

func sendFileLog(buf bytes.Buffer, kafkaMsg chan string) {
	go func() {
		for {
			if len(buf.String()) == 0 {
				break
			}
			logLine, err := buf.ReadString(delim)
			if err != nil {
				log.Println("read log file err")
				log.Println(err)
			}
			kafkaMsg <- logLine
			continue
		}
	}()
}

func getBucketLog(req *ufsdk.UFileRequest, bucketPrefixName string, kafkaMsg chan string) {
	marker := ""
	var buf bytes.Buffer
	for {
		//req.PrefixFileList must have marker param, and if it is "",it will get data from begin.
		//so if the NextMarker is "",we assign the current filename to the marker
		//and we just get one file per request
		myPrefixFileList, err := req.PrefixFileList(bucketPrefixName, marker, 1)
		if err != nil {
			log.Println("DumpResponse：", string(req.DumpResponse(true)))
		}

		if myPrefixFileList.NextMarker == "" {
			log.Println(marker)
			log.Println()
			buf = downloadFile(myPrefixFileList.DataSet[0].FileName, req)
			sendFileLog(buf, kafkaMsg)
			marker = myPrefixFileList.DataSet[0].FileName
			log.Println("last file marker change")
			time.Sleep(360 * time.Second) //ucloud dump ulb log to us3 Every five minutes
			continue
		} else {
			log.Println(marker)
			log.Println()
			buf = downloadFile(myPrefixFileList.DataSet[0].FileName, req)
			sendFileLog(buf, kafkaMsg)
			marker = myPrefixFileList.NextMarker
		}

	}

}

func Run(ConfigFile string, kafkaMsg chan string) {
	// 准备下载请求与要下载的文件
	config, err := ufsdk.LoadConfig(ConfigFile)
	if err != nil {
		panic(err.Error())
	}
	req, err := ufsdk.NewFileRequest(config, nil)
	if err != nil {
		panic(err.Error())
	}

	bucketPrefixName := [4]string{"ulb-x", "ulb-p", "ulb-h0", "ulb-1"}
	var i int
	for i = 0; i < 4; i++ {
		log.Println(bucketPrefixName[i])
		go getBucketLog(req, bucketPrefixName[i], kafkaMsg)
	}

}
