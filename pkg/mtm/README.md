# mtm项目简介
MySql数据库表转换Go Struct(mysql to model)简称(mtm)
大型项目，一般采用数据库先行的设计方式，该工具很好的实现了从数据库到Go Struct代码的转换，减少了手动写Struct的时间。
# 使用方法
#### 获取
go get github.com/xianghu1314/mtm
#### 调用
```go
//模型转换
t2s := mtm.CreateTableToStruct(&mtm.Options{
    MySqlUrl:                "XXX",
    FileName:                "Models.go",
    IfOneFile:               true,
    PackageName:             "Models",
    SavePath:                "./Models",
    IfToHump:                true,
    IfJsonTag:               true,
    IfPluralToSingular:      true,
    IfCapitalizeFirstLetter: true,
})
err := t2s.Run()
if err != nil {
    log.Fatal("模型转换：" + err.Error())
}
```
#### 参数说明
```go
type Options struct {
	MySqlUrl                string //数据库地址 DSN (Data Source Name) ：[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	IfOneFile               bool   //多个表是否放在同一文件 true=同一文件 默认false
	FileName                string //文件名 当IfOneFile=true时有效 默认Models.go
	PackageName             string //自定义项目package名称 默认Models
	SavePath                string //保存文件夹 默认./Models
	IfToHump                bool   //是否转换驼峰 true=是 默认false
	IfJsonTag               bool   //是否包含json tag true=是 默认false
	IfPluralToSingular      bool   //是否复数转单数 true=是 默认false
	IfCapitalizeFirstLetter bool   //是否首字母转换大写 true=是 默认false
}
```
#### 生成样例
```sql
CREATE TABLE `back_users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '真实姓名',
  `account` varchar(50) NOT NULL COMMENT '账号',
  `pwd` varchar(50) CHARACTER SET ascii NOT NULL COMMENT '密码',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
CREATE TABLE `entry_forms` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL,
  `tournament_id` int(11) NOT NULL,
  `name` varchar(50) NOT NULL COMMENT '姓名',
  `id_card` varchar(19) CHARACTER SET ascii NOT NULL COMMENT '身份证18位+x',
  `height` double(3,2) NOT NULL COMMENT '身高，单位cm',
  `weight` double(3,2) NOT NULL COMMENT '体重,单位kg',
  `age` tinyint(3) NOT NULL,
  `blood_type` varchar(2) NOT NULL COMMENT '血型',
  `phone` char(11) NOT NULL COMMENT '联系电话',
  `emergency_phone` char(11) NOT NULL COMMENT '紧急联系人电话',
  `gender` enum('male','female') NOT NULL,
  `status` tinyint(4) NOT NULL COMMENT '默认0 报名，-1退赛， 1领取赛事包， 2检录',
  `pay_status` bit(1) NOT NULL,
  `pay_log` varchar(500) NOT NULL,
  `amount` decimal(65,0) NOT NULL,
  `prepay_id` varchar(50) NOT NULL COMMENT '与支付id',
  `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_entryForms_tournaments_1` (`tournament_id`),
  KEY `fk_entry_forms_users_1` (`user_id`),
  CONSTRAINT `fk_entryForms_tournaments_1` FOREIGN KEY (`tournament_id`) REFERENCES `tournaments` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```
```go
package Models

import (
	"time"
)

type BackUser struct {
	Id      int    `json:"Id"`      //
	Name    string `json:"Name"`    //真实姓名
	Account string `json:"Account"` //账号
	Pwd     string `json:"Pwd"`     //密码
}
type EntryForm struct {
	Id             int       `json:"Id"`             //
	UserId         int       `json:"UserId"`         //
	TournamentId   int       `json:"TournamentId"`   //
	Name           string    `json:"Name"`           //姓名
	IdCard         string    `json:"IdCard"`         //身份证18位+x
	Height         float64   `json:"Height"`         //身高，单位cm
	Weight         float64   `json:"Weight"`         //体重,单位kg
	Age            int       `json:"Age"`            //
	BloodType      string    `json:"BloodType"`      //血型
	Phone          string    `json:"Phone"`          //联系电话
	EmergencyPhone string    `json:"EmergencyPhone"` //紧急联系人电话
	Gender         string    `json:"Gender"`         //
	Status         int       `json:"Status"`         //默认0 报名，-1退赛， 1领取赛事包， 2检录
	PayStatus      int       `json:"PayStatus"`      //
	PayLog         string    `json:"PayLog"`         //
	Amount         float64   `json:"Amount"`         //
	PrepayId       string    `json:"PrepayId"`       //与支付id
	CreateAt       time.Time `json:"CreateAt"`       //
}
```