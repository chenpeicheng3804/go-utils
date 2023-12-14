package util

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
)

func TestImportSql(t *testing.T) {
	tool := &ImportSqlTool{
		SqlPath:  "../import.sql",
		Username: "go-fly",
		Password: "go-fly",
		Server:   "127.0.0.1",
		Port:     "3306",
		Database: "go-fly",
	}
	tool.ImportSql()

}

func TestImportSqlBatch(t *testing.T) {
	data := time.Now().Unix()
	tool := &ImportSqlTool{
		SqlPath:  "../import.sql",
		Username: "go-fly",
		Password: "go-fly",
		Server:   "127.0.0.1",
		Port:     "3306",
		Database: "go-fly",
	}
	tool.ImportSqlBatch()
	fmt.Println(time.Now().Unix() - data)
}
