package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/xwb1989/sqlparser"
)

type ImportSqlTool struct {
	SqlPath                                    string
	Username, Password, Server, Port, Database string
	Db                                         *gorm.DB
}

// ImportSql
// 导入数据库SQL文件
func (this *ImportSqlTool) ImportSql() error {
	// 检查数据库SQL文件是否存在
	_, err := os.Stat(this.SqlPath)
	if os.IsNotExist(err) {
		log.Println(this.SqlPath, "数据库SQL文件不存在:", err)
		return err
	}

	// 根据提供的参数拼接数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.Username, this.Password, this.Server, this.Port, this.Database)
	// 初始计数器
	var count int32
	// 进行数据库连接，如果失败则进行重试
Connect:
	count++
	// 尝试三次
	if count > 3 {
		return errors.New("Database connection is nil")
	}
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
	// 转换win换行符
	convertedContent := strings.ReplaceAll(string(sqls), "\r\n", "\n")
	// 将SQL文件内容按分号分割成数组
	sqlArr := strings.Split(convertedContent+"\n", ";\n")
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

// ImportSqlBatch
// 批量执行SQL文件
func (this *ImportSqlTool) ImportSqlBatch() error {
	// 检查数据库SQL文件是否存在
	_, err := os.Stat(this.SqlPath)
	if os.IsNotExist(err) {
		log.Println(this.SqlPath, "数据库SQL文件不存在:", err)
		return err
	}

	// 根据提供的参数拼接数据库连接字符串
	// golang使用mysq无法执行多条语句
	// 需要加入参数 multiStatements=true
	// 因为 multi statements 可能会增加sql注入的风险
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true", this.Username, this.Password, this.Server, this.Port, this.Database)
	// 初始计数器
	var count int32
	// 进行数据库连接，如果失败则进行重试
Connect:
	count++
	// 尝试三次
	if count > 3 {
		return errors.New("Database connection is nil")
	}
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		//time.Sleep(time.Second)
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
	// sqls, _ := os.ReadFile(this.SqlPath)

	// 去除BOM字符
	// 去除文件开头的BOM字符
	// sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})
	// 转换win换行符
	// convertedContent := strings.ReplaceAll(string(sqls), "\r\n", "\n")
	// 将SQL文件内容按分号分割成数组
	// sqlArr := strings.Split(convertedContent+"\n", ";\n")
	sqlArr := readFileSqlParser(this.SqlPath)
	fifth := len(sqlArr) / 5
	// 每次拼接并打印五分之一的字符串切片
	for i := 0; i < 5; i++ {
		start := i * fifth
		end := (i + 1) * fifth
		tempSlice := sqlArr[start:end]
		if i == 4 {
			tempSlice = sqlArr[start:]
		}

		// 创建正则表达式，用于匹配SQL注释
		re := regexp.MustCompile(`# .*\n|-- .*\n`)

		concatenatedString := ""
		for _, str := range tempSlice {
			// 使用正则表达式替换SQL中的注释
			str = re.ReplaceAllString(str, "")
			// 去除SQL语句两端的空白字符
			str = strings.TrimSpace(str)
			// 如果SQL为空，则跳过本次循环
			if str == "" {
				continue
			}
			concatenatedString += str + ";\n"
		}
		//fmt.Println(concatenatedString)
		err = db.Exec(concatenatedString).Error
		if err != nil {
			log.Println("数据库导入失败:" + err.Error())
		}

		//err = db.Exec(concatenatedString).Error
		//if err != nil {
		//	// 如果执行SQL出错，则打印错误日志
		//	log.Println(this.Database, concatenatedString, "数据库导入失败:"+err.Error())
		//	//log.Println("数据库导入失败:" + err.Error())
		//} else {
		//	// 如果执行SQL成功，则打印成功日志
		//	log.Println(this.Database, concatenatedString, "\t success!")
		//}
	}
	// 执行完所有SQL语句后，返回空值
	return nil

}

// ImportSqlFileWithTransaction
// 使用事务批量导入数据库，SQL文件不解析
// 该方法为一个SQL文件当做一条SQL进行执行。
func (this *ImportSqlTool) ImportSqlFileWithTransaction() error {
	// Db.Begin() 开始事务
	// Db.Commit() 提交事务
	// Db.Rollback() 回滚事务

	// 检查数据库SQL文件是否存在
	_, err := os.Stat(this.SqlPath)
	if os.IsNotExist(err) {
		log.Println(this.SqlPath, "数据库SQL文件不存在:", err)
		return err
	}

	// 根据提供的参数拼接数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.Username, this.Password, this.Server, this.Port, this.Database)
	// 初始计数器
	var count int32
	// 进行数据库连接，如果失败则进行重试
Connect:
	count++
	// 尝试三次
	if count > 3 {
		return errors.New("Database connection is nil")
	}
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

	tx := db.Begin()
	// 去除BOM字符
	// 去除文件开头的BOM字符
	sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})

	// 执行SQL语句，并获取可能的错误
	err = tx.Exec(string(sqls)).Error

	if err != nil {
		// 如果执行SQL出错，则打印错误日志
		log.Println("\nSQL文件：", this.SqlPath, "\n数据库：", this.Database, "\nSQL内容：\n", string(sqls), "数据库导入失败:"+err.Error())
		return err
	} else {
		// 如果执行SQL成功，则打印成功日志
		log.Println("\nSQL文件：", this.SqlPath, "\n数据库：", this.Database, "\nSQL内容：\n", string(sqls), "\t success!")
	}
	// }
	tx.Commit()

	// 执行完所有SQL语句后，返回空值
	return nil
}

// ImportSqlBatchWithTransaction
// 使用事务批量导入数据库SQL文件
func (this *ImportSqlTool) ImportSqlBatchWithTransaction() error {
	// Db.Begin() 开始事务
	// Db.Commit() 提交事务
	// Db.Rollback() 回滚事务

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
		//fmt.Println("mysql连接错误")
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
	// sqls, _ := os.ReadFile(this.SqlPath)

	tx := db.Begin()
	defer tx.Commit()
	// // 去除BOM字符
	// // 去除文件开头的BOM字符
	// sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})
	// // 转换win换行符
	// convertedContent := strings.ReplaceAll(string(sqls), "\r\n", "\n")
	// // 将SQL文件内容按分号分割成数组
	// sqlArr := strings.Split(convertedContent+"\n", ";\n")
	sqlArr := readFileSqlParser(this.SqlPath)
	// 打印日志，表示开始执行SQL文件
	//log.Println("executing", this.SqlPath)
	// 创建正则表达式，用于匹配SQL注释
	// re := regexp.MustCompile(`# .*\n|-- .*\n`)
	for _, sql := range sqlArr {

		// 使用正则表达式替换SQL中的注释
		// sql = re.ReplaceAllString(sql, "")
		// 去除SQL语句两端的空白字符
		// sql = strings.TrimSpace(sql)
		// 如果SQL为空，则跳过本次循环
		// if sql == "" {
		// 	continue
		// }
		//fmt.Println(sql)
		// 执行SQL语句，并获取可能的错误
		err = tx.Exec(sql).Error

		if err != nil {
			// 如果执行SQL出错，则打印错误日志
			log.Println("\nSQL文件：", this.SqlPath, "\n数据库：", this.Database, "\nSQL内容：\n", sql, "数据库导入失败:"+err.Error())
		} else {
			// 如果执行SQL成功，则打印成功日志
			log.Println("\nSQL文件：", this.SqlPath, "\n数据库：", this.Database, "\nSQL内容：\n", sql, "\n success!")
		}
	}
	// 执行完所有SQL语句后，返回空值
	return nil
}

// BatchImportSql
// 读取SQL文件，分批并发导入数据库
func (this *ImportSqlTool) BatchImportSql() error {
	// 检查数据库SQL文件是否存在
	_, err := os.Stat(this.SqlPath)
	if os.IsNotExist(err) {
		log.Println(this.SqlPath, "数据库SQL文件不存在:", err)
		return err
	}
	// 根据提供的参数拼接数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.Username, this.Password, this.Server, this.Port, this.Database)
	// 初始计数器
	var count int32
	// 进行数据库连接，如果失败则进行重试
Connect:
	count++
	// 尝试三次
	if count > 3 {
		return errors.New("Database connection is nil")
	}
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		log.Println("连接失败")
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
	// // 读取SQL文件内容，并忽略错误
	// sqls, _ := os.ReadFile(this.SqlPath)

	// // 去除BOM字符
	// // 去除文件开头的BOM字符
	// sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})
	// // 转换win换行符
	// convertedContent := strings.ReplaceAll(string(sqls), "\r\n", "\n")
	// // 将SQL文件内容按分号分割成数组
	// sqlArr := strings.Split(convertedContent+"\n", ";\n")
	sqlArr := readFileSqlParser(this.SqlPath)
	var wg sync.WaitGroup
	// 将sqlArr切割5000条为一组，并发执行
	//fmt.Println("sqlArr", len(sqlArr))
	//fmt.Println("sqlArr", sqlArr)
	for i := 0; i < len(sqlArr); i += 5000 {
		// 获取当前组的SQL语句
		sqlBatch := sqlArr[i:min(i+5000, len(sqlArr))]
		wg.Add(1)
		// 创建一个协程，执行SQL语句
		go func(sqlBatchs []string, wg *sync.WaitGroup, db *gorm.DB) {
			//fmt.Println("sqlBatch", len(sqlBatch))
			// 创建一个事务
			tx := db.Begin()
			defer tx.Commit()
			defer wg.Done()
			for _, sqlB := range sqlBatchs {
				if sqlB == "\n" {
					//log.Println("长度", len(sqlB))
					continue
				}
				err = tx.Exec(sqlB).Error
				if err != nil {
					// 如果执行SQL出错，则打印错误日志
					log.Println("\nSQL文件：", this.SqlPath, "\n数据库：", this.Database, "\nSQL内容：\n", sqlB, "数据库导入失败:"+err.Error())
					//} else {
					//	// 如果执行SQL成功，则打印成功日志
					//	log.Println("\nSQL文件：", this.SqlPath, "\n数据库：", this.Database, "\nSQL内容：\n", sqlB, "\n success!")
				}
			}
		}(sqlBatch, &wg, db)
	}
	wg.Wait()
	return nil
}

// 读取SQL文件内容 返回SQL语句
func readFileSqlParser(file string) (sqls []string) {
	// 读取整个文件内容
	content, err := os.ReadFile(file)
	if err != nil {
		log.Println("Error reading file: ", err)
		os.Exit(1)
	}
	r := strings.NewReader(string(content))
	tokens := sqlparser.NewTokenizer(r)
	for {
		stmt, err := sqlparser.ParseNext(tokens)
		//fmt.Println(stmt)
		if err == io.EOF {
			break
		}
		if stmt != nil {
			// Do something with stmt or err.
			sqls = append(sqls, sqlparser.String(stmt))
		}
	}
	return sqls
}
