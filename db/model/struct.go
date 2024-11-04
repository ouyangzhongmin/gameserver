package model

type AdminRecharge struct {
	Id           int64
	AgentId      int64  `xorm:"not null BIGINT(20) default"`
	AgentName    string `xorm:"not null VARCHAR(32) default"`
	AgentAccount string `xorm:"not null VARCHAR(32) default"`
	AdminId      string `xorm:"not null VARCHAR(32) default"`
	AdminName    string `xorm:"not null VARCHAR(32) default"`
	AdminAccount string `xorm:"not null VARCHAR(32) default"`
	Extra        string `xorm:"not null VARCHAR(255) default"`
	CreateAt     int64  `xorm:"not null BIGINT(20) default"`
	CardCount    int64  `xorm:"not null BIGINT(20) default"`
}

type Agent struct {
	Id             int64
	Name           string `xorm:"not null VARCHAR(32) default"`
	Account        string `xorm:"not null VARCHAR(32) default"`
	Password       string `xorm:"not null VARCHAR(64) default"`
	Phone          string `xorm:"not null VARCHAR(11) default"`
	Wechat         string `xorm:"not null VARCHAR(32) default"`
	Salt           string `xorm:"not null VARCHAR(32) default"`
	Role           int    `xorm:"not null TINYINT(4) default"`
	Status         int    `xorm:"not null TINYINT(4) default"`
	Extra          string `xorm:"not null VARCHAR(255) default"`
	CreateAt       int64  `xorm:"not null BIGINT(20) default"`
	DeleteAt       int64  `xorm:"not null BIGINT(20) default"`
	DeleteAccount  string `xorm:"not null VARCHAR(32) default"`
	CreateAccount  string `xorm:"not null VARCHAR(32) default"`
	ConfirmAccount string `xorm:"not null VARCHAR(32) default"`
	CardCount      int64  `xorm:"not null BIGINT(20) default"`
	Level          int    `xorm:"not null INT(20) default"`
	Discount       int    `xorm:"not null INT(20) default"`
}

type Order struct {
	Id             int64
	OrderId        string `xorm:"not null unique VARCHAR(32)"`
	Type           int    `xorm:"not null TINYINT(1) default 0"`
	AppId          string `xorm:"not null index VARCHAR(32) default"`
	ChannelId      string `xorm:"not null index VARCHAR(32) default"`
	PayPlatform    string `xorm:"not null VARCHAR(32) default"`
	ChannelOrderId string `xorm:"not null VARCHAR(255) default"`
	Currency       string `xorm:"not null VARCHAR(255) default"`
	Extra          string `xorm:"not null VARCHAR(1024) default"`
	Money          int    `xorm:"not null INT(11) default"`
	RealMoney      int    `xorm:"not null INT(11) default"`
	Uid            int64  `xorm:"not null index BIGINT(20) default"`
	RoleId         string `xorm:"not null VARCHAR(255) default"`
	RoleName       string `xorm:"not null VARCHAR(255) default"`
	ServerId       string `xorm:"not null VARCHAR(255) default"`
	ServerName     string `xorm:"not null VARCHAR(255) default"`
	CreatedAt      int64  `xorm:"not null BIGINT(11) default"`
	ProductId      string `xorm:"not null VARCHAR(255) default"`
	ProductCount   int    `xorm:"not null INT(10) default"`
	ProductName    string `xorm:"not null VARCHAR(255) default"`
	ProductExtra   string `xorm:"not null VARCHAR(255) default"`
	NotifyUrl      string `xorm:"not null VARCHAR(2048) default"`
	Status         int    `xorm:"not null TINYINT(2) default 1"`
	Remote         string `xorm:"not null VARCHAR(40) default"`
	Ip             string `xorm:"not null VARCHAR(40) default"`
	Imei           string `xorm:"not null VARCHAR(64) default"`
	Os             string `xorm:"not null VARCHAR(20) default"`
	Model          string `xorm:"not null VARCHAR(20) default"`
}

type Recharge struct {
	Id           int64
	AgentId      string `xorm:"not null VARCHAR(32) default"`
	AgentName    string `xorm:"not null VARCHAR(32) default"`
	AgentAccount string `xorm:"not null VARCHAR(32) default"`
	PlayerId     int64  `xorm:"not null BIGINT(20) default"`
	Extra        string `xorm:"not null VARCHAR(255) default"`
	CreateAt     int64  `xorm:"not null BIGINT(20) default"`
	CardCount    int64  `xorm:"not null BIGINT(20) default"`
}

type Trade struct {
	Id            int64
	OrderId       string `xorm:"not null unique VARCHAR(32) default"`
	PayOrderId    string `xorm:"not null VARCHAR(255) default"`
	PayPlatform   string `xorm:"not null VARCHAR(32) default"`
	PayAt         int64  `xorm:"not null BIGINT(11) default"`
	PayCreateAt   int64  `xorm:"not null BIGINT(11) default"`
	ComsumerId    string `xorm:"not null VARCHAR(128) default"`
	MerchantId    string `xorm:"not null VARCHAR(128) default"`
	ComsumerEmail string `xorm:"not null VARCHAR(64) default"`
	Raw           string `xorm:"not null VARCHAR(2048) default"`
}

type Uuid struct {
	Id        int64
	UidInUse  int64  `xorm:"not null index BIGINT(20) default 0"`
	UidOrigin int64  `xorm:"not null BIGINT(20) default 0"`
	Appid     string `xorm:"not null index VARCHAR(32) default"`
	Uuid      string `xorm:"not null VARCHAR(64) default"`
}
