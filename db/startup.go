package db

import (
	"fmt"
	"github.com/spf13/viper"
)

func Startup() func() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetString("database.port"),
		viper.GetString("database.dbname"),
		viper.GetString("database.args"))

	return MustStartup(
		dsn,
		MaxIdleConns(viper.GetInt("database.max_idle_conns")),
		MaxIdleConns(viper.GetInt("database.max_open_conns")),
		ShowSQL(viper.GetBool("database.show_sql")))
}
