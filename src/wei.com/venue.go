package main

import (
	"os"
	"wei.com/utils"
	"log"
	"net/http"
	"bytes"
	"io/ioutil"
	"regexp"
	"encoding/json"
	"fmt"
	"time"
	"math"
	"strconv"
	"text/template"
)

// Wechat model
type Wechat struct {
	Openid     string
	Headimgurl string
	Nickname   string
	Sex        string
	Language   string
	City       string
	Province   string
	Country    string
}
type Operator struct {
	Id          string
	DisplayName string
	Role        string
	Mobile struct {
		Profile struct {
			Mobile string
		}
	}
	Ktvid string
	Lat   float64
	Lng   float64
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

var authHttpUrlBefore string = "http://prod-api.chinacloudapp.cn:3000/ktvCore/ktvs/846/users/graphql"
//var authHttpUrlAfter string = "http://user.prod-v1.ye-dian.com/auth/users?_type=VenuesManager"
var authHttpUrlAfter string = "http://127.0.0.1:4000/auth/users?_type=VenuesManager"
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
	soffset := strconv.Itoa(offset)
	scount := strconv.Itoa(count)
	//jsonStr := `"query":"query{users(from: 5, size: 3){total,list{id}}"`
	jsonStr := `{"query":"query {\n users(from:` + soffset + `, size: ` + scount + `) {\n total\n  list {\n  id\n displayName\n  role\n ktvid\n  lat\n lng\n  \n mobile {\n   profile{\n   mobile\n  }\n  }\n wechat {\n  profile {\n openid\n  nickname\n  headimgurl\n  }\n }\n  }\n }\n}"}`

	fmt.Println("offset", jsonStr)
	result, err := httpPost(authHttpUrlBefore, jsonStr)
	var httpResult HttpResult
	err = json.Unmarshal(result, &httpResult)
	utils.CheckErr(err)
	//fmt.Println("httpResult", httpResult)
	/* 创建集合 */
	role := make(map[string]string)

	/* map 插入 key-value 对，各个国家对应的首都 */
	role["ktv_manager"] = "VenuesManager"
	role["ktv_admin_manager"] = "VenuesAdmin"

	for _, v := range httpResult.Data.Users.List {
		//fmt.Println(v.Id)
		//fmt.Println(v.DisplayName)
		//fmt.Println(v.Role)
		//fmt.Println(v.Mobile.Profile.Mobile)
		v.Role = role[v.Role]
		var jsonStr string = ""
		if v.DisplayName == "" {
			v.DisplayName = strconv.FormatInt(time.Now().Unix(), 10)
		}
		//fmt.Println()
		if v.Wechat.Profile.Openid != "" {
			jsonStr = `{"displayName":"{{.DisplayName}}","role":"{{.Role}}","subscribe":1,"venuesId":"{{.Ktvid}}",
				"mobile":"{{.Mobile.Profile.Mobile}}","locationLat":"{{.Lat}}","locationLon":"{{.Lng}}",
				"wechat":{"openid":"{{.Wechat.Profile.Openid}}","nickName":"{{.DisplayName}}","sex":"0","headimgurl":"not found"}}`
		} else {
			jsonStr = `{"displayName":"{{.DisplayName}}","role":"{{.Role}}","subscribe":1,"venuesId":"{{.Ktvid}}",
				"mobile":"{{.Mobile.Profile.Mobile}}","locationLat":"{{.Lat}}","locationLon":"{{.Lng}}"}`
		}

		var data bytes.Buffer
		tmpl, err := template.New("tem").Parse(jsonStr) //建立一个模板

		err = tmpl.Execute(&data, v) //将struct与模板合成，合成结果放到os.Stdout里
		utils.CheckErr(err)

		//var jsonStr = []byte(json.Marshal(value))
		//fmt.Println("post:jsonStr", data.String())
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
	//fmt.Println("httpPost", httpUrl)
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
	fmt.Println(httpUrl+"return body", string(body))
	r, _ := regexp.Compile("error_code")
	if r.MatchString(string(body)) {
		logFile(string(body))
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
	//fmt.Println("getCount", httpResult)
	total = int(math.Ceil(float64(httpResult.Data.Users.Total) / float64(count))) //page总数
}
