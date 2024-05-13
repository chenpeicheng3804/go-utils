package util

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// 实现从csv中读取数据，并写入到数据库中

// 约定:cvs首行为表字段名，表字段名与数据库表字段名对应

// 大致思路
// 1.获取表数据类型
// 2.读取csv数据
// 2.1.格式化每行每一列数据类型与数据库表对应
// 特殊处理: 科学计数法数据转换为int
// 2.2.生成sql插入语句
// 3.将数据插入数据库

type CsvData struct {
	sqlxdb    *sqlx.DB
	csvfile   string
	tablename string

	// 记录开始时间
	starttime time.Time
}

type field struct {
	FieldName string `db:"COLUMN_NAME"`
	FieldType string `db:"DATA_TYPE"`
}

// CsvDataInfo
type CsvDataInfo struct {
	Username  string
	Password  string
	Server    string
	Port      string
	Database  string
	CsvFile   string
	TableName string
}

// 创建CsvData对象
func NewCsvData(username, password, server, port, database, csvfile, tablename string) *CsvData {
	return &CsvData{
		sqlxdb:    sqlx.MustConnect("mysql", username+":"+password+"@tcp("+server+":"+port+")/"+database+"?charset=utf8mb4&parseTime=True&loc=Local"),
		csvfile:   csvfile,
		tablename: tablename,
	}
}

// 初始化CsvDataInfo信息
func (c *CsvDataInfo) InitCsvDataInfo() *CsvData {
	return &CsvData{
		sqlxdb:    sqlx.MustConnect("mysql", c.Username+":"+c.Password+"@tcp("+c.Server+":"+c.Port+")/"+c.Database+"?charset=utf8mb4&parseTime=True&loc=Local"),
		csvfile:   c.CsvFile,
		tablename: c.TableName,
	}
}

// 读取CSV文件并将其内容转换为批量插入的SQL语句
// 运行程序
func (c *CsvData) Run() {
	c.starttime = time.Now()
	defer c.sqlxdb.Close()
	// 开启事务
	tx, err := c.sqlxdb.Beginx()
	if err != nil {
		log.Panicln("begin trans failed, err:", err)
	}
	defer func() {
		if p := recover(); p != nil {
			// 回滚
			tx.Rollback()
			log.Panicln(p) // re-throw panic after Rollback
		} else if err != nil {

			tx.Rollback() // err is non-nil; don't change it
			log.Panicln("rollback")
		} else {
			tx.Commit() // err is nil; if Commit returns error update err
			log.Println("commit")
		}
	}()
	var totalAffected int64
	fields, err := c.describeTable()
	//fmt.Println(fields)
	if err != nil {
		log.Panicln(err)
	}
	fieldsMap := fieldsToMap(fields)
	file, err := os.Open(c.csvfile)
	if err != nil {
		log.Panicln(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 自动检测字段数量

	records, err := reader.ReadAll()
	if err != nil {
		log.Panicln(err)
	}

	headers := records[0]
	values := make([]interface{}, len(headers))

	sqlValues := make([]string, len(values))
	for i := range values {
		sqlValues[i] = "?"
	}
	sqlValuesStr := strings.Join(sqlValues, ",")

	insertSql := fmt.Sprintf("INSERT INTO %s (`%s`) VALUES (%s)", c.tablename, strings.Join(headers, "`,`"), sqlValuesStr)
	//fmt.Println(sql)
	// var rs sql.Result
	for _, record := range records[1:] { // 跳过表头

		for i, value := range record {
			fieldType := fieldsMap[headers[i]].FieldType
			convertedValue := convertCSVValue(fieldType, value)

			values[i] = convertedValue
		}
		rs, err := tx.Exec(insertSql, values...)
		if err != nil {
			log.Panicln("Error inserting record: ", err)
			//} else {
			//	fmt.Println("Inserted record successfully")
		}
		// 获取插入行数
		n, err := rs.RowsAffected()
		if err != nil {
			log.Panicln("Error getting rows affected: ", err)
		}
		totalAffected += n
	}

	log.Println("插入行数 ", totalAffected)
	// 打印结束时间
	log.Println("执行完毕，耗时:", time.Since(c.starttime))
}

// 根据属性类型返回对应sql空值类型
func convertMySQLTypeToGo(mysqlType string) interface{} {
	switch mysqlType {
	case "int", "tinyint", "smallint", "mediumint", "bigint":
		return sql.NullInt64{Valid: false}
	case "decimal", "float", "double":
		return sql.NullFloat64{Valid: false}
	case "varchar", "char", "text", "mediumtext", "longtext":
		return sql.NullString{Valid: false}
	case "datetime", "timestamp", "date", "time":
		return sql.NullTime{Valid: false}
	default:
		return sql.NullString{Valid: false}
	}
}

// convertCSVValue 根据数据库字段类型将CSV值转换为适当的Go类型。
func convertCSVValue(fieldType string, csvValue string) interface{} {

	// 判断长度是否为0
	if len(csvValue) == 0 {
		// 返回mysql null值
		return convertMySQLTypeToGo(fieldType)
	}
	switch fieldType {
	case "int", "tinyint", "smallint", "mediumint", "bigint":
		var intVal int64
		if strings.ContainsAny(csvValue, "eE") {
			intVal, _ = convertScientificToInt(csvValue)
			return intVal
		}
		intVal, _ = strconv.ParseInt(csvValue, 10, 64)
		return intVal
	case "decimal", "float", "double":
		floatVal, _ := strconv.ParseFloat(csvValue, 64)
		return floatVal
	case "varchar", "char", "text", "mediumtext", "longtext":
		return csvValue
	case "datetime", "timestamp", "date", "time":
		//fmt.Println("datetime", csvValue)
		// 检查日期时间字符串是否是无效的"1900-01-00"
		parsedTime, err := time.Parse("2006-01-02 15:04:05", csvValue)
		if err != nil {
			return sql.NullTime{}
		}
		return parsedTime // 假设已格式化为Go的时间格式，实际应用中需要转换
	default:
		return csvValue
	}
}

// convertScientificToInt 尝试将科学计数法表示的字符串转换为int。
// 注意：此转换可能会导致精度损失，特别是对于非常大或非常小的数值。
func convertScientificToInt(scientificStr string) (int64, error) {
	floatVal, err := strconv.ParseFloat(scientificStr, 64)
	if err != nil {
		return 0, err
	}
	intVal := int64(floatVal)
	// 检查转换是否有精度损失
	if float64(intVal) != floatVal {
		return 0, fmt.Errorf("precision loss converting scientific notation to int: %s", scientificStr)
	}
	return intVal, nil
}

// describeTable 获取MySQL数据库中指定表的所有字段及其类型。
func (c *CsvData) describeTable() ([]field, error) {
	var fields []field
	err := c.sqlxdb.Select(&fields, `
		SELECT COLUMN_NAME, DATA_TYPE
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`, c.tablename)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// 将[]field切片结构体转换为map 以FieldName为键。
func fieldsToMap(fields []field) map[string]field {
	fieldMap := make(map[string]field)
	for _, field := range fields {
		fieldMap[field.FieldName] = field
	}
	return fieldMap
}
