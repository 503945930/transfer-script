package main

import (
	"fmt"
	"os"
	"wei.com/utils"
	"log"
	"net/http"
	"bytes"
	"io/ioutil"
	"regexp"
	"io"
)

var authHttpUrl string = "http://localhost:4001/internal/migrate/users?_type=User"

func main() {
	start()
}

func start() {
	httpPost()
}

func httpPost(httpUrl string, data string) io.Reader {

	//var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", httpUrl, bytes.NewBuffer([]byte(data)))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	utils.CheckErr(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	utils.CheckErr(err)
	//fmt.Println(string(body))
	r, _ := regexp.Compile("error_code")
	if r.MatchString(string(body)) {
		logFile(data)
	}
	utils.CheckErr(err)

	return body
	//var result interface{}
	//err = json.Unmarshal(body, &result)

}

func logFile(str string) {
	f, err := os.OpenFile("./log/logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	utils.CheckErr(err)
	defer f.Close()
	log.SetOutput(f)
	log.Println(str)
}
