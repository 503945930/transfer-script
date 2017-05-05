package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"wei.com/utils"
	"database/sql"
	"net/http"
	"bytes"
	"io/ioutil"
	"text/template"
	"os"
	"log"
	"regexp"
	"math"
	"time"
)

// Wechat model
type Wechat struct {
	Openid     string
	Avatar_url sql.NullString
}

//User model
type User struct {
	Id           string
	Display_name sql.NullString
	Password     sql.NullString
	Role         sql.NullString
	Email        sql.NullString
	Mobile       sql.NullString
	Lat          sql.NullString
	Lng          sql.NullString
	Wechat       Wechat
}

type Result struct {
	Count int
}

//var sqlConnect = "stagingmysql1%staging:agargh2IUHiu@tcp(stagingmysql1.mysqldb.chinacloudapi.cn:3306)/yedian-abicloud"

var sqlConnect = "root:root@/yedian-abicloud"
var httpUrl string = "http://localhost:4001/internal/migrate/users?_type=User"
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

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		//fmt.Println("worker", id, "processing job", j)
		offset = count * j
		start(offset, count)
		results <- j
	}
}

func start(offset int, count int) {
	fmt.Println("offset:", offset)
	db, err := gorm.Open("mysql", sqlConnect)
	utils.CheckErr(err)
	defer db.Close()

	// Delete an existing record

	var sqlQuery string = "SELECT id,display_name,password,role,email,mobile,lat,lng,openid,avatar_url FROM ac_platform_user limit ?,?"

	// Raw SQL
	rows, err := db.Raw(sqlQuery, offset, count).Rows() //
	utils.CheckErr(err)
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Display_name, &user.Password, &user.Role,
			&user.Email, &user.Mobile, &user.Lat, &user.Lng, &user.Wechat.Openid, &user.Wechat.Avatar_url)
		utils.CheckErr(err)
		//fmt.Println(user)
		users = append(users, user)
	}

	for _, value := range users {
		//fmt.Println()
		jsonStr := `{"id":"{{.Id}}","displayName":"{{.Display_name.String}}","password":"{{.Password.String}}","role":"{{.Role.String}}",
		"email":"{{.Email.String}}","mobile":"{{.Mobile.String}}","locationLat":"{{.Lat.String}}","locationLon":"{{.Lng.String}}",
		"wechat":{"openid":"{{.Wechat.Openid}}","nickName":"{{.Display_name.String}}","headimgurl":"{{.Wechat.Avatar_url.String}}"}}
`
		var data bytes.Buffer
		tmpl, err := template.New("tem").Parse(jsonStr) //建立一个模板

		err = tmpl.Execute(&data, value) //将struct与模板合成，合成结果放到os.Stdout里
		utils.CheckErr(err)

		//var jsonStr = []byte(json.Marshal(value))
		//fmt.Println(data.String())
		httpPost(httpUrl, data.String())
	}
}

func httpPost(httpUrl string, data string) {

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
	//var result interface{}
	//err = json.Unmarshal(body, &result)

}

func getCount() {
	db, err := gorm.Open("mysql", sqlConnect)
	utils.CheckErr(err)
	defer db.Close()
	var result Result
	var sqlQuery string = "SELECT count(*) as count  FROM ac_platform_user"
	db.Raw(sqlQuery).Scan(&result)
	total = int(math.Ceil(float64(result.Count) / float64(count))) //page总数
}

func logFile(str string) {
	f, err := os.OpenFile("./log/logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	utils.CheckErr(err)
	defer f.Close()
	log.SetOutput(f)
	log.Println(str)
}
