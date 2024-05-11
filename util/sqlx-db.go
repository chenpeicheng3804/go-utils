package util

import (
	"github.com/jmoiron/sqlx"
)

type SqlxDb struct {
	SqlPath                                    string
	Username, Password, Server, Port, Database string
	Db                                         *sqlx.DB
}

func NewSqlxDb(sqlPath, username, password, server, port, database string) *SqlxDb {
	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, server, port, database)
	return &SqlxDb{
		SqlPath:  sqlPath,
		Username: username,
		Password: password,
		Server:   server,
		Port:     port,
		Database: database,
		Db:       sqlx.MustConnect("mysql", username+":"+password+"@tcp("+server+":"+port+")/"+database+"?charset=utf8mb4&parseTime=True&loc=Local"),
	}
}

//// BatchInsert
//// 批量插入
//func (s *SqlxDb) BatchInsert() error {
//	query, args, _ := sqlx.In(
//		"INSERT INTO user (name, age) VALUES (?), (?), (?)",
//		users..., // 如果arg实现了 driver.Valuer, sqlx.In 会通过调用 Value()来展开它
//	)
//	fmt.Println(query) // 查看生成的querystring
//	fmt.Println(args)  // 查看生成的args
//	_, err := s.Db.Exec(query, args...)
//	return err
//}

//func (s *SqlxDb) createSqlxDb() {
//	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", s.Username, s.Password, s.Server, s.Port, s.Database)
//	//dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", this.Username, this.Password, this.Server, this.Port, this.Database)
//	// 也可以使用MustConnect连接不成功就panic
//	// var err error
//	s.Db = sqlx.MustConnect("mysql", dsn)
//	// s.Db, err = sqlx.Connect("mysql", dsn)
//	// if err != nil {
//	// 	fmt.Printf("connect DB failed, err:%v\n", err)
//	// 	time.Sleep(time.Second)
//	// 	goto Connect
//	// }
//	// 设置数据库连接的最大限制
//	// 此段代码为配置数据库连接池的参数，以优化数据库操作的性能和资源利用。
//	// SetMaxOpenConns 设置数据库的最大打开连接数为20。
//	// 这是数据库可以同时打开的最多连接数，用于控制数据库的并发访问量。
//	s.Db.SetMaxOpenConns(20)
//
//	// SetMaxIdleConns 设置数据库的最大空闲连接数为10。
//	// 这是数据库中保持的最小活跃连接数，用于快速响应后续的数据库访问需求。
//	s.Db.SetMaxIdleConns(10)
//	// SetConnMaxLifetime 设置数据库连接的最大生命周期。
//	// 通过设置连接的最大生命周期来避免使用过旧的连接，提升数据库的安全性和稳定性。
//	// s.Db.SetConnMaxLifetime()
//
//	// SetConnMaxIdleTime 设置数据库最大空闲时间。
//	// 通过设置连接的最大空闲时间来控制连接的回收，避免因长时间不使用导致的资源浪费。
//	// s.Db.SetConnMaxIdleTime()
//}

//// QueryMultiRow
//// 查询多条数据示例
//func (s *SqlxDb) QueryMultiRow(sqlStr string, record any) {
//	if s.Db == nil {
//		s.createSqlxDb()
//	}
//	//fmt.Println(len(record))
//	var records []interface{}
//	err := s.Db.Select(&records, sqlStr)
//	if err != nil {
//		fmt.Printf("query failed, err:%v\n", err)
//		return
//	}
//	// 序列化 []interface{} 为 JSON 字符串
//	jsonStr, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(records)
//	if err != nil {
//		panic(err)
//	}
//	err = jsoniter.Unmarshal(jsonStr, &record)
//	if err != nil {
//		panic(err)
//	}
//}

//// QueryRow
//// 查询单条数据示例
//func (s *SqlxDb) QueryRow(sqlStr string, record interface{}) {
//	if s.Db == nil {
//		s.createSqlxDb()
//	}
//
//	err := s.Db.Get(&record, sqlStr)
//	if err != nil {
//		fmt.Printf("get failed, err:%v\n", err)
//		return
//	}
//
//}
