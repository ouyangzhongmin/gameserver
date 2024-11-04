package mtm

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

//参考
//https://blog.csdn.net/Charles_Thanks/article/details/80503124

// map for converting mysql type to golang types
var typeForMysqlToGo = map[string]string{
	"int":                "int",
	"integer":            "int",
	"tinyint":            "int",
	"smallint":           "int",
	"mediumint":          "int",
	"bigint":             "int64",
	"int unsigned":       "uint32",
	"integer unsigned":   "uint32",
	"tinyint unsigned":   "uint8",
	"smallint unsigned":  "int",
	"mediumint unsigned": "int",
	"bigint unsigned":    "uint64",
	"bit":                "int8",
	"bool":               "bool",
	"enum":               "string",
	"json":               "mysqlx.JSON",
	"set":                "string",
	"varchar":            "string",
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string",
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "time.Time", // time.Time or string
	"datetime":           "time.Time", // time.Time or string
	"timestamp":          "time.Time", // time.Time or string
	"time":               "time.Time", // time.Time or string
	"float":              "float64",
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string",
	"varbinary":          "string",
}

func CreateTableToStruct(options *Options) *TableToStruct {
	if len(options.MySqlUrl) == 0 {
		log.Fatal("MySqlUrl参数不能为空")
	}
	if len(options.PackageName) == 0 {
		options.PackageName = "Models"
	}
	if len(options.SavePath) == 0 {
		options.SavePath = "./Models"
	}
	if len(options.FileName) == 0 {
		options.FileName = "Models.go"
	}

	t2s := new(TableToStruct)
	t2s.NeedCreateAtTables = make([]string, 0)
	t2s.NeedUpdateAtTables = make([]string, 0)
	if options != nil {
		t2s.MySqlUrl = options.MySqlUrl
		t2s.IfOneFile = options.IfOneFile
		t2s.FileName = options.FileName
		t2s.PackageName = options.PackageName
		t2s.SavePath = options.SavePath
		t2s.IfToHump = options.IfToHump
		t2s.IfJsonTag = options.IfJsonTag
		t2s.IfDbTag = options.IfDbTag
		t2s.IfPluralToSingular = options.IfPluralToSingular
		t2s.IfCapitalizeFirstLetter = options.IfCapitalizeFirstLetter
		if options.NeedCreateAtTables != "" {
			t2s.NeedCreateAtTables = append(t2s.NeedCreateAtTables, strings.Split(options.NeedCreateAtTables, ",")...)
		}
		if options.NeedUpdateAtTables != "" {
			t2s.NeedUpdateAtTables = append(t2s.NeedUpdateAtTables, strings.Split(options.NeedUpdateAtTables, ",")...)
		}

	}
	return t2s
}
func (t2s *TableToStruct) Run() error {
	//1、获取table列表
	db, err := CreateMysqlDb(t2s.MySqlUrl)
	if err != nil {
		return err
	}
	tables, err := db.Query("SELECT table_schema,table_name FROM information_schema.TABLES WHERE table_schema=DATABASE () AND table_type='BASE TABLE'; ")
	if err != nil {
		return err
	}
	defer tables.Close()

	for tables.Next() {
		tableSchema := ""
		structName := ""
		originTableName := ""

		err = tables.Scan(&tableSchema, &structName)
		fmt.Println(tableSchema, structName)
		if structName == "app_use_guide" {
			fmt.Println(tableSchema, structName)
		}
		if err != nil {
			return err
		}
		ttf := new(TableToFile)
		ttf._import = make(map[string]string)
		ttf._struct = structName
		//2、循环获取table Column列表
		columns, err := db.Query("SELECT COLUMN_NAME,DATA_TYPE,IS_NULLABLE,TABLE_NAME,COLUMN_COMMENT FROM information_schema.COLUMNS WHERE table_schema=DATABASE () AND table_name=?;", structName)
		if err != nil {
			return err
		}
		defer columns.Close()

		originTableName = structName
		//3、打印输出格式
		//3.1、输出类名
		if t2s.IfToHump {
			structName = toHump(structName)
		}
		if t2s.IfPluralToSingular {
			structName = toSingular(structName)
		}
		if t2s.IfCapitalizeFirstLetter {
			structName = strFirstToUpper(structName)
		} else {
			structName = strFirstToLower(structName)
		}
		if t2s.IfCapitalizeFirstLetter {
			structName = strFirstToUpper(structName)
		}
		ttf._fileName = structName
		ttf._struct = "type " + structName + " struct {\n"
		//3.2、输出属性
		ttf._property = make([]string, 0)
		for columns.Next() {
			columnName := ""
			dataType := ""
			isNullable := ""
			tableName := ""
			columnComment := ""
			err = columns.Scan(&columnName, &dataType, &isNullable, &tableName, &columnComment)
			if err != nil {
				return err
			}
			_type, ok := typeForMysqlToGo[dataType]
			if !ok {
				_type = "[]byte"
			}
			if _type == "time.Time" {
				ttf._import["time"] = `"time"`
			}
			if _type == "mysqlx.JSON" {
				ttf._import["json"] = `"github.com/ouyangzhongmin/gameserver/pkg/mysqlx"`
			}
			if strings.EqualFold(isNullable, "YES") {
				_type = "*" + _type //指针类型
			}
			GoColumnName := columnName
			if t2s.IfToHump {
				GoColumnName = toHump(GoColumnName)
			}
			if t2s.IfCapitalizeFirstLetter {
				GoColumnName = strFirstToUpper(GoColumnName)
			} else {
				GoColumnName = strFirstToLower(GoColumnName)
			}

			tags := ""
			if t2s.IfJsonTag {
				if (columnName == "create_at" && findStr(t2s.NeedCreateAtTables, originTableName) == -1) || (columnName == "update_at" && findStr(t2s.NeedUpdateAtTables, originTableName) == -1) {
					tags += fmt.Sprintf("json:\"%s\" ", "-")
				} else {
					tags += fmt.Sprintf("json:\"%s\" ", columnName)
				}
			}
			if t2s.IfDbTag {
				tags += fmt.Sprintf("db:\"%s\" ", columnName)
			}
			columnComment = strings.ReplaceAll(columnComment, "\r\t", "")
			columnComment = strings.ReplaceAll(columnComment, "\n", "")
			if tags != "" {
				ttf._property = append(ttf._property, fmt.Sprintf("	%s %s `%s` //%s", GoColumnName, _type, tags, columnComment))
			} else {
				ttf._property = append(ttf._property, fmt.Sprintf("	%s %s //%s", GoColumnName, _type, columnComment))
			}
		}
		t2s.tableToFile = append(t2s.tableToFile, ttf)
	}
	//4、写入文件
	err = t2s.saveToFile()
	if err != nil {
		log.Fatal(err.Error())
	}
	cmd := exec.Command("gofmt", "-w", t2s.SavePath)
	cmd.Run()
	log.Print("模型装换成功")
	return nil
}

// 打印数据库每个表的字段名称
func (t2s *TableToStruct) PrintDBColumns(dbName string) error {
	log.Println("ExportDBColumns")
	db, err := CreateMysqlDb(t2s.MySqlUrl)
	if err != nil {
		return err
	}
	rows, err := db.Query(fmt.Sprintf("SELECT  TABLE_NAME, GROUP_CONCAT(COLUMN_NAME SEPARATOR ',')  FROM information_schema.`COLUMNS`  WHERE TABLE_SCHEMA = '%s'   group by TABLE_NAME", dbName))
	if err != nil {
		return err
	}

	for rows.Next() {
		tableName := ""
		columns := ""
		err = rows.Scan(&tableName, &columns)
		if err != nil {
			return err
		}
		log.Println("table:", tableName, columns)
	}
	return nil
}

type Options struct {
	MySqlUrl                string //数据库地址 DSN (Data Source Name) ：[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	IfOneFile               bool   //多个表是否放在同一文件 true=同一文件 默认false
	FileName                string //文件名 当IfOneFile=true时有效 默认Models.go
	PackageName             string //自定义项目package名称 默认Models
	SavePath                string //保存文件夹 默认./Models
	IfToHump                bool   //是否转换驼峰 true=是 默认false
	IfJsonTag               bool   //是否包含json tag true=是 默认false
	IfDbTag                 bool   //是否包含db tag true=是 默认false
	IfPluralToSingular      bool   //是否复数转单数 true=是 默认false
	IfCapitalizeFirstLetter bool   //是否首字母转换大写 true=是 默认false
	NeedCreateAtTables      string //需要tag保留createAt的表, 多个逗号隔开
	NeedUpdateAtTables      string //需要tag保留updateAt的表, 多个逗号隔开
}
type TableToStruct struct {
	MySqlUrl                string
	IfOneFile               bool
	FileName                string
	PackageName             string
	SavePath                string
	IfToHump                bool
	IfJsonTag               bool
	IfDbTag                 bool
	IfPluralToSingular      bool
	IfCapitalizeFirstLetter bool
	tableToFile             []*TableToFile
	NeedCreateAtTables      []string
	NeedUpdateAtTables      []string
}

type TableToFile struct {
	_import   map[string]string
	_struct   string
	_fileName string
	_property []string
}

func (t *TableToFile) _importToStr() string {
	im := ""
	for _, value := range t._import {
		im += value + "\n"
	}
	return im
}
func (t *TableToFile) _propertyToStr() string {
	return strings.Join(t._property, "\n")
}

func (t *TableToStruct) saveToFile() error {
	if !t.IfOneFile {
		for _, v := range t.tableToFile {
			//4、写入文件
			file := "package " + t.PackageName + "\n" + "import (\n" + v._importToStr() + ")\n" + v._struct + v._propertyToStr() + "\n}\n"
			err := t.save(v._fileName+".go", file)
			if err != nil {
				return err
			}
		}
	} else {
		content := ""
		importStr := "import (\n"
		for _, v := range t.tableToFile {
			//4、写入文件
			importStr += v._importToStr()
			content += v._struct + v._propertyToStr() + "\n}\n"
		}
		importStr += ")\n"
		file := "package " + t.PackageName + "\n" + importStr + content
		err := t.save(t.FileName, file)
		if err != nil {
			return err
		}
	}
	return nil
}
func (t *TableToStruct) save(fileName string, content string) error {
	//4、写入文件
	//容错
	if t.SavePath[len(t.SavePath)-1] != '/' {
		t.SavePath += "/"
	}
	//创建目录
	os.MkdirAll(t.SavePath, 0777)
	//创建文件
	filePath := t.SavePath + fileName
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	f.WriteString(content)
	return nil

}

// Convert The First Letter To Capitalize
func strFirstToUpper(str string) string {
	if len(str) < 1 {
		return ""
	}
	//if unicode.IsUpper([]rune(str)[0]) {
	//	return str
	//}
	strArry := []rune(str)
	if strArry[0] >= 97 && strArry[0] <= 122 {
		strArry[0] -= 32
	}
	return string(strArry)
}

// Convert The First Letter To Capitalize
func strFirstToLower(str string) string {
	if len(str) < 1 {
		return ""
	}
	//if unicode.IsLower([]rune(str)[0]) {
	//	return str
	//}
	strArry := []rune(str)
	if strArry[0] >= 65 && strArry[0] <= 90 {
		strArry[0] += 32
	}
	return string(strArry)
}

// Convert The Plural To Singular
func toSingular(word string) string {
	plural1, _ := regexp.Compile("([^aeiou])ies$")
	plural2, _ := regexp.Compile("([aeiou]y)s$")
	plural3, _ := regexp.Compile("([sxzh])es$")
	plural4, _ := regexp.Compile("([^sxzhyu])s$")
	if plural1.Match([]byte(word)) {
		return word[0:len(word)-3] + "y"
	} else if plural2.Match([]byte(word)) {
		return word[0 : len(word)-1]
	} else if plural3.Match([]byte(word)) {
		return word[0 : len(word)-2]
	} else if plural4.Match([]byte(word)) {
		return word[0 : len(word)-1]
	}
	return word
}

// 转换驼峰
func toHump(c string) string {
	cg := strings.Split(c, "_")
	p := ""
	for _, v := range cg {
		p += strFirstToUpper(v)
	}
	return p
}

func findStr(arr []string, c string) int {
	for i, v := range arr {
		if v == c {
			return i
		}
	}
	return -1
}
