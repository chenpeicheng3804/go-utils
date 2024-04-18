package util

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func TestReadFile(t *testing.T) {
	SqlPath := "/home/demo/Documents/tmp/sql/server-jchl/biz/bovms/dstr_bovms_upgrade.sql"
	sqls, _ := os.ReadFile(SqlPath)
	fmt.Println(string(sqls))
	sqls = bytes.TrimPrefix(sqls, []byte{0xef, 0xbb, 0xbf})
	fmt.Println(string(sqls))
}
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
		SqlPath:  "/home/demo/Downloads/浙里信12w.sql",
		Username: "root",
		Password: "demo",
		Server:   "127.0.0.1",
		Port:     "3306",
		Database: "demo1",
	}
	tool.ImportSqlBatch()
	fmt.Println(time.Now().Unix() - data)
}
func TestImportSqlBatchWithTransaction(t *testing.T) {
	data := time.Now().Unix()
	tool := &ImportSqlTool{
		SqlPath:  "/tmp/ttk_data_prd_2",
		Username: "ttk_admin",
		Password: "goY4Qbby_aGXaht9_vKO",
		Server:   "192.168.8.232",
		Port:     "32061",
		Database: "ttk_data_prd_0001",
	}
	tool.ImportSqlBatchWithTransaction()
	fmt.Println(time.Now().Unix() - data)
}

// 执行存储过程
func TestStoreProcedure(t *testing.T) {
	data := time.Now().Unix()
	tool := &ImportSqlTool{

		SqlPath: "/home/demo/Documents/tmp/sql/server-jchl/biz/bovms/dstr_bovms_upgrade.sql",
		// demo
		Username: "root",
		Password: "demo",
		Server:   "127.0.0.1",
		Port:     "3306",
		Database: "demo9",
		// staging
		//Username: "ttk_admin",
		//Password: "cxZwAta5L4MGMoR6_cCz",
		//Server:   "rm-2ze9w29r886763jrv.mysql.rds.aliyuncs.com",
		//Port:     "3306",
		//Database: "ttk_data_staging_0001",

		//test
		//Username: "jcsz_admin",
		//Password: "cxZwAta5L4MGMoR6_cCz",
		//Server:   "rm-2zeyqrz49xmrr3jwa.mysql.rds.aliyuncs.com",
		//Port:     "3306",
		//Database: "ttk_data_test_0002",
	}
	tool.ImportSqlBatchWithTransaction()
	fmt.Println(time.Now().Unix() - data)
}
