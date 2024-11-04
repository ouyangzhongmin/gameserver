package utils

import "github.com/DanPlayer/randomname"

func GetRandomHeroName() string {
	return randomname.GenerateName()
}
