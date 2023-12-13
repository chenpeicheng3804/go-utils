package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type ImportSqlTool struct {
	SqlPath                                    string
	Username, Password, Server, Port, Database string
}

func (this *ImportSqlTool) ImportSql() error {
	_, err := os.Stat(this.SqlPath)
	if os.IsNotExist(err) {
		log.Println(this.SqlPath, "数据库SQL文件不存在:", err)
		return err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.Username, this.Password, this.Server, this.Port, this.Database)
Connect:
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		//log.Println("数据库连接失败，数据库地址：", this.Server, "数据库名称：", this.Database, err)
		//panic("数据库连接失败!")
		//return err
		time.Sleep(time.Second)
		goto Connect
	}
	db.SingularTable(true)
	db.LogMode(false)
	db.DB().SetMaxIdleConns(0)
	db.DB().SetMaxOpenConns(0)
	db.DB().SetConnMaxLifetime(59 * time.Second)

	sqls, _ := os.ReadFile(this.SqlPath)
	sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})
	sqlArr := strings.Split(string(sqls)+"\n", ";")
	log.Println("executing", this.SqlPath)
	for _, sql := range sqlArr {
		re := regexp.MustCompile(`^# .*\n|^-- .*\n`)
		sql = re.ReplaceAllString(sql, "")
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		err = db.Exec(sql).Error
		if err != nil {
			log.Println(this.Database, strings.Replace(sql, "\n", "", -1), "数据库导入失败:"+err.Error())

		} else {
			log.Println(this.Database, strings.Replace(sql, "\n", "", -1), "\t success!")
		}
	}
	return nil
}
