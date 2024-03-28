package util

import (
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// 创建连接
func (this *ImportSqlTool) CreateDb() {
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
	this.Db = db
}

// DbExec
// 执行器
// DB.Exec() 方法用于执行原始 SQL 语句或可执行的命令。它不返回查询结果，而是返回一个 *sql.Result
// 对象，其中包含了执行结果的信息，例如受影响的行数。 DB.Exec() 方法通常用于执行 INSERT、UPDATE、DELETE
// 等修改操作，并用于判断操作是否成功。
func (this *ImportSqlTool) DbExec(sql string) (err error) {
DBnil:
	if this.Db == nil {
		this.CreateDb()
		goto DBnil
		//return errors.New("Database connection is nil")
	}
	err = this.Db.Exec(sql).Error
	return err
	//result := this.Db.Exec("UPDATE users SET name = ? WHERE id = ?", "John", 1)
	//rowsAffected, err := result.RowsAffected()
}

// DbRaw
// 查询器
// DB.Raw() 方法用于执行原始 SQL 查询语句或可执行的命令。它可以执行任意的 SQL 语句，并返回查询结果或影响的行数。
// DB.Raw() 方法返回的是 *sql.Rows 结果集对象，通过调用 .Scan() 方法可以将查询结果映射到相应的结构体中。
// 由于直接执行原始 SQL，所以需要手动处理 SQL 注入、参数绑定和结果集映射等问题。
func (this *ImportSqlTool) DbRaw(sql string) ([]byte, error) {
DBnil:
	if this.Db == nil {
		this.CreateDb()
		goto DBnil
		//return nil, errors.New("Database connection is nil")
	}

	rows, err := this.Db.Raw(sql).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowData[col] = string(b)
			} else {
				rowData[col] = val
			}
		}
		results = append(results, rowData)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
