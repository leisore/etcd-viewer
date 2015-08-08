package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const htmlHead = `<html> 
					<head> 
						<style type="text/css"> 
							p { 
								line-height: 50%; 
								font-family: Menlo, monospace; 
								font-size: 14px;
							} 
						</style> 
					</head> 
				<body>
				<h3><p>etcd viewer</p></h3><hr>
				`
const htmlTail = `</body></html>`

var etcd = "http://192.168.1.143:4001/v2/keys"
var localAddr string
var remoteAddr string

func init() {
	addrs, _ := net.InterfaceAddrs()
	localAddr = fmt.Sprintf("%s:%s", "0.0.0.0", "5555")
	remoteAddr = fmt.Sprintf("%s:%s", addrs[0].String(), "5555")

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("etcdv <etcd http address>")
		os.Exit(0)
	}
	etcd = os.Args[1]
	fmt.Printf("etcd address: %s\n", etcd)
	fmt.Printf("listen on %s ......", "5555")

	startEtcdViewerServer()
}

func startEtcdViewerServer() {
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		reqUri := fmt.Sprintf("%s/%s", etcd, req.RequestURI)
		if rawInfo, err := reqRawEtcdInfo(reqUri); err != nil {
			res.Write([]byte(err.Error()))
		} else {
			htmlInfo := toHtml(rawInfo)
			res.Write([]byte(htmlInfo))
		}
	})
	http.ListenAndServe(localAddr, nil)
}

func toHtml(rawEtcdInfo []byte) string {
	htmlInfo := string(rawEtcdInfo)
	htmlInfo = regexp.MustCompile("(.*)(\")(/.*)(\")(.*)").ReplaceAllString(htmlInfo, fmt.Sprintf("${1}${2}<a href=http://%s${3}>${3}</a>${4}", remoteAddr))
	htmlInfo = regexp.MustCompile("    ").ReplaceAllString(htmlInfo, "&nbsp;&nbsp;&nbsp;&nbsp;")
	htmlInfo = strings.Replace(htmlInfo, "\n", "<p>", -1)
	htmlInfo = fmt.Sprintf("%s%s%s", htmlHead, htmlInfo, htmlTail)
	return htmlInfo
}

func reqRawEtcdInfo(reqUri string) ([]byte, error) {
	res, err := http.Get(reqUri)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var prettyJson bytes.Buffer
	rawJson, err := ioutil.ReadAll(res.Body)
	json.Indent(&prettyJson, rawJson, "", "    ")
	return prettyJson.Bytes(), nil
}
