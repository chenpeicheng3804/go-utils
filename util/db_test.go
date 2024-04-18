package util

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"testing"
)

// SysBrowserVersion
type SysBrowserVersion struct {
	//ID      string `json:"id" gorm:"id"`
	Version      string `json:"version" gorm:"version"`           // 版本号
	Downloadurl  string `json:"downLoadUrl" gorm:"downLoadUrl"`   // 下载URL
	Isneedupdate string `json:"isNeedUpdate" gorm:"isNeedUpdate"` // 必须更新，0-否，1-是
	Updatetype   string `json:"updateType" gorm:"updateType"`     // 升级类型，0-静默，1-非静默
	Createdate   string `json:"createDate" gorm:"createDate"`     // 创建日期（当前数据日期）
	Appid        string `json:"appid" gorm:"appid"`
}
type Version struct {
	Version string `json:"version" gorm:"version"` // 版本号
}

type Update struct {
	UpEnvir    string `json:"upenvir" gorm:"upenvir"`       // 环境名称
	UpdateType string `json:"updatetype" gorm:"updatetype"` // 升级类型
	Version    string `json:"version" gorm:"version"`       // 版本号
}

var (
	test = &ImportSqlTool{
		Username: "root",
		Password: "demo",
		Server:   "127.0.0.1",
		Port:     "3306",
		Database: "demo9",
	}
	prod = &ImportSqlTool{
		Username: "root",
		Password: "demo",
		Server:   "127.0.0.1",
		Port:     "3306",
		Database: "ops",
	}
	testVersions, prodVersions []Version
)

// rows, _ := tool.Db.Raw("SELECT id,version,isNeedUpdate,updateType FROM sys_browser_version WHERE createDate > DATE_SUB(CURDATE(), INTERVAL 30 DAY) ORDER BY version DESC;").Rows()
// demo9 测试数据库
// ops 生产数据
func TestDB(t *testing.T) {
	r := gin.Default()
	r.GET("/api/jenkins/rpa/Version", Run)
	r.POST("/api/jenkins/rpa/Version", Run1)
	r.Run()
}
func Run1(c *gin.Context) {
	update := Update{}
	c.BindJSON(&update)
	//fmt.Println(update.UpEnvir)
	//fmt.Println(update.Version)
	if len(update.Version) == 0 || len(update.UpdateType) == 0 || len(update.UpEnvir) == 0 {
		c.Writer.Write([]byte(`传参不正确`))
		return
	}

	Versions := strings.Replace(update.Version, ",", "','", -1)

	sql := fmt.Sprintf("SELECT * FROM sys_browser_version WHERE version IN  ('%s') ORDER BY createDate ASC;", Versions)
	testJsonByte, _ := test.DbRaw(sql)
	//查生产库
	prodJsonByte, _ := prod.DbRaw(sql)
	if string(prodJsonByte) != "null" {
		//fmt.Println(string(prodJsonByte))
		c.Writer.Write([]byte(`已存在\n` + string(prodJsonByte)))
		return
	}

	var sysbrowserversion []SysBrowserVersion
	//fmt.Println(string(testJsonByte))
	err := json.Unmarshal(testJsonByte, &sysbrowserversion)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
	var Insert string
	for _, v := range sysbrowserversion {
		//fmt.Println(v)
		// 下载地址预处理
		downloadurl := strings.Replace(v.Downloadurl, "/test/", "/pro/", -1)
		//	转成插入SQL
		Insert = Insert + "('" + v.Version + "','" + downloadurl + "','" + v.Isneedupdate + "','" + v.Updatetype + "','" + v.Createdate + "','" + v.Appid + "'),"
	}
	// 行位替换,
	Inserts := strings.TrimSuffix(Insert, ",") + ";"
	sqlInserts := "INSERT INTO sys_browser_version (version, downLoadUrl, isNeedUpdate, updateType, createDate, appid) VALUES "
	//fmt.Println(sqlInserts)
	//fmt.Println(Inserts)
	err = prod.DbExec(sqlInserts + Inserts)
	if err != nil {
		c.Writer.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}
	//Versions := strings.Split(update.Version, ",")
	//for _, version := range Versions {
	//	fmt.Println(version)
	//}
	c.Writer.Write([]byte("ok"))
}
func Run(c *gin.Context) {

	// prod 查询结果需转成map
	prodJsonByte, _ := prod.DbRaw("SELECT version FROM sys_browser_version WHERE appid = 'jcszBrowser' and createDate > DATE_SUB(CURDATE(), INTERVAL 30 DAY) ORDER BY version DESC;")
	// json to struct
	err := json.Unmarshal(prodJsonByte, &prodVersions)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
	// struct to map
	prodVersionMap := make(map[string]Version)
	for _, v := range prodVersions {
		prodVersionMap[v.Version] = v
	}

	testJsonByte, _ := test.DbRaw("SELECT version FROM sys_browser_version WHERE createDate > DATE_SUB(CURDATE(), INTERVAL 30 DAY) ORDER BY version DESC;")
	//fmt.Println(string(testJsonByte))
	// test 查询结果转成结构体
	err = json.Unmarshal(testJsonByte, &testVersions)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
	//遍历test结构体差值prod map
	var DifferenceValue []Version
	for _, v := range testVersions {
		//fmt.Println("Version:", v.Version)
		_, ok := prodVersionMap[v.Version]
		if !ok {
			DifferenceValue = append(DifferenceValue, v)
		}
	}
	// fmt.Println(DifferenceValue)
	c.JSON(http.StatusOK, DifferenceValue)
}

// var results []SysBrowserVersion
// for rows.Next() {
//     var version string
//     if err := rows.Scan(&version); err != nil {
//         return nil, err
//     }

//     sysBrowserVersion := SysBrowserVersion{
//         Version: version,
//     }
//     results = append(results, sysBrowserVersion)
// }
