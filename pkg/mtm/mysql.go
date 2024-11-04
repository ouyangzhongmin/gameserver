//参考 https://yq.aliyun.com/articles/178898?utm_content=m_29337
package mtm

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB = nil

//sql.DB的设计就是用来作为长连接使用的。不要频繁Open, Close
func CreateMysqlDb(mysqlServer string) (*sql.DB, error) {

	if Db != nil {
		return Db, nil
	}
	_db, err := sql.Open("mysql", mysqlServer)
	//都适用默认配置，有需要在配置
	//来限制连接池中空闲连接的数量，但是这并不会限制连接池的大小。连接回收(recycle)的很快，通过设置一个较大的N，可以在连接池中保留一些空闲连接，供快速复用(reuse)。但保持连接空闲时间过久可能会引发其他问题，比如超时。设置N=0则可以避免连接空闲太久。
	//_db.SetMaxIdleConns(0)
	////来限制连接池中打开的连接数量。
	//_db.SetMaxOpenConns(0)
	////来限制连接的生命周期。连接超时后，会在需要时惰性回收复用。
	//_db.SetConnMaxLifetime(60 * time.Second)
	if err != nil {
		return nil, nil
	} else {
		return _db, nil
	}
}

//sql.DB的设计就是用来作为长连接使用的。不要频繁Open, Close
//func (this *MySql) Close() {
//	this.Db.Close()
//}
