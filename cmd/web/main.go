package main

import (
	"fmt"
	"github.com/google/gops/agent"
	"github.com/ouyangzhongmin/gameserver/db"
	"github.com/ouyangzhongmin/gameserver/pkg/async"
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

	"github.com/ouyangzhongmin/gameserver/internal/web"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	// base application info
	app.Name = "web server"
	app.Author = "Oyzm"
	app.Version = "0.0.1"
	app.Copyright = "oyzm team reserved"
	app.Usage = "web server"

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

	async.Run(func() {
		// service connections
		//gops工具接入,若需要远程访问，可配置 agent.Options{Addr: "0.0.0.0:6060"}，否则默认仅允许本地访问
		fmt.Println("gops listen.....")
		if err := agent.Listen(agent.Options{}); err != nil {
			panic(err)
		}
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		// setup database
		closer := db.Startup()
		defer closer()
		web.Startup()
	}() // 开启web服务器

	wg.Wait()
	return nil
}

func getCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
