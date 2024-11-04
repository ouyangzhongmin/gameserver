package env

import (
	"fmt"
	"os"
	"strings"
)

const (
	ENV_PRODUCTION    = "PRODUCTION"
	ENV_DEVELOP       = "DEVELOP"
	ENV_TEST          = "TEST"
	ENV_OVERSEAS      = "OVERSEAS"
	ENV_OVERSEAS_TEST = "OVERSEAS_TEST"
	ENV_KEY           = "ENV_GAME"
)

var env = strings.TrimSpace(os.Getenv(ENV_KEY))

func init() {
	fmt.Println("env::::", env)
	LogEnv()
}

// 是否是开发环境
func IsDevelopEnv() bool {
	if strings.TrimSpace(env) == "" {
		env = strings.TrimSpace(os.Getenv(ENV_KEY))
	}
	return env == "" || env == ENV_DEVELOP
}

// 是否是测试环境
func IsTestEnv() bool {
	return IsCNTestEnv() || IsOverseasTestEnv()
}

// 国内测试环境
func IsCNTestEnv() bool {
	if strings.TrimSpace(env) == "" {
		env = strings.TrimSpace(os.Getenv(ENV_KEY))
	}
	return env == ENV_TEST
}

// 是否是生产环境,包括国内国外
func IsProductionEnv() bool {
	return IsCNProductionEnv() || IsOverseasEnv()
}

// 是否是国内生产环境
func IsCNProductionEnv() bool {
	if strings.TrimSpace(env) == "" {
		env = strings.TrimSpace(os.Getenv(ENV_KEY))
	}
	return env == ENV_PRODUCTION
}

// 是否是海外
func IsOverseasEnv() bool {
	if strings.TrimSpace(env) == "" {
		env = strings.TrimSpace(os.Getenv(ENV_KEY))
	}
	return env == ENV_OVERSEAS
}

// 是否是海外测试
func IsOverseasTestEnv() bool {
	if strings.TrimSpace(env) == "" {
		env = strings.TrimSpace(os.Getenv(ENV_KEY))
	}
	return env == ENV_OVERSEAS_TEST
}

// 非线上环境
func IsJustTestEnv() bool {
	return !(IsProductionEnv() || IsOverseasEnv())
}

// 指定一个环境变量值，指定后系统变更时需要重启应用才会使用新的环境变量，
// 不会影响系统环境变量
func SetEnv(v string) {
	env = v
	fmt.Println("指定环境变量:", env)
	LogEnv()
}

func LogEnv() {
	if IsCNProductionEnv() {
		fmt.Println("当前是国内生产环境")
	} else if IsOverseasEnv() {
		fmt.Println("当前是海外生产环境")
	} else if IsCNTestEnv() {
		fmt.Println("当前是国内TEST环境")
	} else if IsOverseasTestEnv() {
		fmt.Println("当前是海外TEST环境")
	} else {
		fmt.Println("当前是DEVELOP环境")
	}
}
