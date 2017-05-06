package main

import (
	"os"
	"wei.com/utils"
	"log"
	"net/http"
	"bytes"
	"io/ioutil"
	"regexp"
	"text/template"
	"encoding/json"
	"fmt"
	"time"
	"math"
)

// Wechat model
type Wechat struct {
	Openid     string
	Headimgurl string
	Nickname   string
}

type Operator struct {
	Id          string
	DisplayName string
	Role        string
	Email       string
	Mobile struct {
		Profile struct {
			Mobile string
		}
	}
	Lat float64
	Lng float64
	Wechat struct {
		Profile Wechat
	}
}

type HttpResult struct {
	Data struct {
		Users struct {
			Total int
			List  []Operator
		}
	}
}

var authHttpUrlBefore string = "http://staging-yedian.chinacloudapp.cn:3003/internal/query/users"
var authHttpUrlAfter string = "http://localhost:4001/internal/migrate/users?_type=Operator"
var maxRoutineNum int = 100
var count int = 20 //每页数量
var offset int = 0
var total int = 0 // 总页数

func main() {
	//记录开始时间
	start := time.Now()
	//ch := make(chan int, maxRoutineNum)

	getCount() //总页数
	fmt.Print("总页数:")
	fmt.Println(total)

	//定义俩个管道，任务channel和结果channel
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	//启动3个协程，因为jobs里面没有数据，他们都会阻塞。
	for w := 1; w <= maxRoutineNum; w++ {
		go worker(w, jobs, results)
	}

	//发送9个任务到jobs channel，然后关闭管道channel。
	for j := 0; j <= total; j++ {
		jobs <- j
	}
	close(jobs)

	//然后收集所有的结果。
	for a := 0; a <= total; a++ {
		<-results
	}

	//记录时间
	elapsed := time.Since(start)
	fmt.Println("App elapsed: ", elapsed)
}

func start(offset int, count int) {
	//jsonStr := `"query":"query{users(from: 5, size: 3){total,list{id}}"`
	jsonStr := `{"query":"query {\n users(from: 0, size: 10) {\n total\n  list {\n  id\n displayName\n  role\n  lat\n lng\n  email\n mobile {\n   profile{\n   mobile\n  }\n  }\n wechat {\n  profile {\n openid\n  nickname\n sex\n city\n province\n country\n headimgurl\n }\n }\n  }\n }\n}"}`

	result, err := httpPost(authHttpUrlBefore, jsonStr)
	var httpResult HttpResult
	err = json.Unmarshal(result, &httpResult)
	utils.CheckErr(err)

	for _, v := range httpResult.Data.Users.List {
		//fmt.Println(v.Id)
		//fmt.Println(v.DisplayName)
		//fmt.Println(v.Role)
		//fmt.Println(v.Mobile.Profile.Mobile)

		//fmt.Println()
		jsonStr := `{"id":"{{.Id}}","displayName":"{{.DisplayName}}","role":"{{.Role}}",
		"email":"{{.Email}}","mobile":"{{.Mobile.Profile.Mobile}}","locationLat":"{{.Lat}}","locationLon":"{{.Lng}}",
		"wechat":{"openid":"{{.Wechat.Profile.Openid}}","nickName":"{{.Wechat.Profile.Nickname}}","headimgurl":"{{.Wechat.Profile.Headimgurl}}"}}`
		var data bytes.Buffer
		tmpl, err := template.New("tem").Parse(jsonStr) //建立一个模板

		err = tmpl.Execute(&data, v) //将struct与模板合成，合成结果放到os.Stdout里
		utils.CheckErr(err)

		//var jsonStr = []byte(json.Marshal(value))
		fmt.Println(data.String())
		// 添加到新的 operator
		httpPost(authHttpUrlAfter, data.String())
	}

}
func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		//fmt.Println("worker", id, "processing job", j)
		offset = count * j
		start(offset, count)
		results <- j
	}
}

func httpPost(httpUrl string, data string) ([]byte, error) {

	//var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	req, err := http.NewRequest("POST", httpUrl, bytes.NewBuffer([]byte(data)))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Content-Type", "application/graphql")

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

	return body, nil
	// fmt.Println(string(body))
	//var httpResult HttpResult
	//err = json.Unmarshal(body, &httpResult)
	//fmt.Println(httpResult.Data.Users.Total)
	//utils.CheckErr(err)
	//
	//return httpResult, nil
	//var operator Operator
	//fmt.Println(value.([]interface{}))["echat"]
	//json.Unmarshal(, &operator)
	//js, err := simplejson.NewJson(body)
	///fmt.Println(string(body))

	//total, err := js.Get("data").Get("users").Get("total").Int()
	//fmt.Printf(string(total))
	///return js, nil

	//return body
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

func getCount() {
	jsonStr := `{"query":"query {\n users(from: 0, size: 10) {\n total\n  }\n}"}`
	result, err := httpPost(authHttpUrlBefore, jsonStr)
	var httpResult HttpResult
	err = json.Unmarshal(result, &httpResult)
	utils.CheckErr(err)
	total = int(math.Ceil(float64(httpResult.Data.Users) / float64(count))) //page总数
}
