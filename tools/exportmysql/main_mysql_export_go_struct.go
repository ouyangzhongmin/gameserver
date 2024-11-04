//mysql表导出go struct工具

package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"

	"github.com/ouyangzhongmin/gameserver/pkg/mtm"
)

type MysqlCnf struct {
	UserName  string `json:"-"`
	Password  string `json:"-"`
	IpAddrees string
	Port      int
	DbName    string
}

func (m *MysqlCnf) GetDsn() string {
	var charset = "utf8mb4"
	//mycat json读取报错问题：Error 1023: program err;java.lang.Index0utofBoundsException: Index: 1, Size: 1
	//原因是mycat prepare json处理问题，可以通过改为客户端prepare跳过mycat 预处理，方法为dsn加上：&interpolateParams=true，但是加上后写入json数据的地方要用sqlx.JSONText类型
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local", m.UserName, m.Password, m.IpAddrees, m.Port, m.DbName, charset)
	return dsn
}

func main() {
	dns := &MysqlCnf{
		UserName:  "root",
		Password:  "root",
		IpAddrees: "172.16.2.8",
		Port:      3306,
		DbName:    "jsmx",
	}
	dir := filepath.Join(GetCurrentDir(), "../../db/model")
	//模型转换
	t2s := mtm.CreateTableToStruct(&mtm.Options{
		MySqlUrl:                dns.GetDsn(), //数据库地址 DSN (Data Source Name) ：[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
		FileName:                "models.go",  //文件名 当IfOneFile=true时有效 默认Models.go
		IfOneFile:               true,         //多个表是否放在同一文件 true=同一文件 默认false
		PackageName:             "model",      //自定义项目package名称 默认Models
		SavePath:                dir,          //保存文件夹 默认./Models
		IfToHump:                true,         //是否转换驼峰 true=是 默认false
		IfJsonTag:               true,         //是否包含json tag true=是 默认false
		IfDbTag:                 true,
		IfPluralToSingular:      false, //是否复数转单数 true=是 默认false
		IfCapitalizeFirstLetter: true,  //是否首字母转换大写 true=是 默认false
	})
	err := t2s.Run()
	if err != nil {
		log.Fatal("模型转换：" + err.Error())
	}
	t2s.PrintDBColumns("jsmx")
}

// 获取当前文件夹（此函数不建议在执行次数比较频繁地方使用，可以在启动过程或者脚本里使用）
func GetCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
