package main

import (
	"os"
	"wei.com/utils"
	"log"
	"fmt"
	"net/http"
	"bytes"
	"io/ioutil"
	"regexp"
	"database/sql"
	"encoding/json"
)

var authHttpUrlBefore string = "http://staging-yedian.chinacloudapp.cn:3003/internal/query/users"
var authHttpUrlAfter string = "http://localhost:4001/internal/migrate/users?_type=Operator"

// Wechat model
type Wechat struct {
	Openid     string
	Avatar_url sql.NullString
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

func main() {
	start()
}

type HttpResult struct {
	Data struct {
		Users struct {
			Total int
			List  []Operator
		}
	}
}

func start() {
	//jsonStr := `"query":"query{users(from: 5, size: 3){total,list{id}}"`
	jsonStr := `{"query":"query {\n users(from: 0, size: 10) {\n total\n  list {\n  id\n displayName\n  role\n  lat\n lng\n  email\n mobile {\n   profile{\n   mobile\n  }\n  }\n wechat {\n  profile {\n openid\n  nickname\n sex\n city\n province\n country\n headimgurl\n }\n }\n  }\n }\n}"}`

	result, err := httpPost(authHttpUrlBefore, jsonStr)
	utils.CheckErr(err)

	for _, v := range result.Data.Users.List {
		fmt.Println(v.Id)
		fmt.Println(v.DisplayName)
		fmt.Println(v.Role)
		fmt.Println(v.Mobile.Profile.Mobile)
	}

	//result, err := httpPost(authHttpUrlBefore, jsonStr)

	//err := json.NewDecoder(result).Decode(&transforms)
	//body, err := simplejson.NewJson(result)
	//utils.CheckErr(err)
	//fmt.Printf(string(result))
	//total, err := result.Get("data").Get("users").Get("total").Int()
	//list, err := result.Get("data").Get("users").Get("list").Array()

	//var operator []Operator
	//for _, value := range list {
	//	//var operator Operator
	//	fmt.Println(value.([]interface{}))["echat"]
	//	//json.Unmarshal(, &operator)
	//
	//	//id, _ := value.(map[string]interface{})["id"]
	//	//
	//	//json.Unmarshal(value.(map[string]interface{})[]byte(value), &operator)
	//	//fmt.Println(id)
	//}
	//data, err := result.Get("data").Get("list").MustArray()
	//fmt.Printf("v%", total)
}

func httpPost(httpUrl string, data string) (HttpResult, error) {

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

	fmt.Println(string(body))
	var httpResult HttpResult
	err = json.Unmarshal(body, &httpResult)
	fmt.Println(httpResult.Data.Users.Total)
	utils.CheckErr(err)

	return httpResult, nil
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
