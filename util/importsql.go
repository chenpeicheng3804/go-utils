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
	// 检查数据库SQL文件是否存在
	_, err := os.Stat(this.SqlPath)
	if os.IsNotExist(err) {
		log.Println(this.SqlPath, "数据库SQL文件不存在:", err)
		return err
	}

	// 根据提供的参数拼接数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.Username, this.Password, this.Server, this.Port, this.Database)

	// 进行数据库连接，如果失败则进行重试
Connect:
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		time.Sleep(time.Second)
		goto Connect
	}

	// 设置数据库连接参数
	// 设置数据库操作对象的表名是否使用单数形式
	db.SingularTable(true)
	// 设置是否打印SQL语句
	db.LogMode(false)
	// 设置连接池中的最大空闲连接数
	db.DB().SetMaxIdleConns(0)
	// 设置数据库的最大打开连接数
	db.DB().SetMaxOpenConns(0)
	// 设置连接的最大可复用时间
	db.DB().SetConnMaxLifetime(59 * time.Second)
	// 读取SQL文件内容，并忽略错误
	sqls, _ := os.ReadFile(this.SqlPath)

	// 去除BOM字符
	// 去除文件开头的BOM字符
	sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})
	// 将SQL文件内容按分号分割成数组
	sqlArr := strings.Split(string(sqls)+"\n", ";")
	// 打印日志，表示开始执行SQL文件
	log.Println("executing", this.SqlPath)

	for _, sql := range sqlArr {
		// 创建正则表达式，用于匹配SQL注释
		re := regexp.MustCompile(`# .*\n|-- .*\n`)
		// 使用正则表达式替换SQL中的注释
		sql = re.ReplaceAllString(sql, "")
		// 去除SQL语句两端的空白字符
		sql = strings.TrimSpace(sql)
		// 如果SQL为空，则跳过本次循环
		if sql == "" {
			continue
		}
		// 执行SQL语句，并获取可能的错误
		err = db.Exec(sql).Error

		if err != nil {
			// 如果执行SQL出错，则打印错误日志
			log.Println(this.Database, strings.Replace(sql, "\n", "", -1), "数据库导入失败:"+err.Error())

		} else {
			// 如果执行SQL成功，则打印成功日志
			log.Println(this.Database, strings.Replace(sql, "\n", "", -1), "\t success!")

		}
	}
	return nil
	// 执行完所有SQL语句后，返回空值
}
