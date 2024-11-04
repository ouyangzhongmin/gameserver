package main

import (
	"fmt"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/internal/master"
	"github.com/ouyangzhongmin/gameserver/pkg/env"
	"github.com/ouyangzhongmin/gameserver/pkg/fileutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	// base application info
	app.Name = "master server"
	app.Author = "Oyzm"
	app.Version = "0.0.1"
	app.Copyright = "oyzm team reserved"
	app.Usage = "master server"

	// flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "",
			Usage: "load configuration from `FILE`",
		},
		cli.BoolFlag{
			Name:  "cpuprofile",
			Usage: "enable cpu profile",
		},
	}

	app.Action = serve
	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	cfgpth := c.String("config")
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	workPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if strings.TrimSpace(cfgpth) == "" {
		if !env.IsDevelopEnv() {
			panic("需要指定配置文件路径....")
		}
		tomlName := "configs/config.toml"
		cfgpth = filepath.Join(appPath, tomlName)
		if !fileutil.FileExists(cfgpth) {
			cfgpth = filepath.Join(workPath, tomlName)
			if !fileutil.FileExists(cfgpth) {
				cfgpth = filepath.Join(getCurrentPath(), tomlName)
			}
		}
		log.Println("使用工程配置:", cfgpth)
	} else {
		log.Println("使用指定配置：", cfgpth)
	}
	if !fileutil.FileExists(cfgpth) {
		panic(fmt.Sprintf("%s配置不存在", cfgpth))
	}
	viper.SetConfigType("toml")
	viper.SetConfigFile(cfgpth)
	viper.ReadInConfig()

	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	if viper.GetBool("core.debug") {
		log.SetLevel(log.DebugLevel)
	}

	if c.Bool("cpuprofile") {
		filename := fmt.Sprintf("cpuprofile-%d.pprof", time.Now().Unix())
		f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		// setup database
		closer := db.Startup()
		defer closer()
		master.Startup()
	}() // 开启master

	wg.Wait()
	return nil
}

func getCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
