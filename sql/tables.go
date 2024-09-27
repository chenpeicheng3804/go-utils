package sql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// show tables in database
func ShowTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		getCreateTableStmt(db, table)
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

// 获取指定表的创建语句
func getCreateTableStmt(db *sql.DB, tableName string) {
	var createStmt string

	rows, _ := db.Query("SHOW CREATE TABLE " + tableName)
	// 非常重要：关闭rows释放持有的数据库链接
	defer rows.Close()
	// 循环读取结果集中的数据
	for rows.Next() {
		rows.Scan(&tableName, &createStmt)
		fmt.Println(createStmt)
	}

}
