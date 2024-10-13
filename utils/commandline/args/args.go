package args

import (
	"flag"
	"os"
	"path"
)

type args struct {
	ExeAbsDir  string
	ExeName    string
	ConfigFile string
}

var (
	Args   = args{}
	isInit = false
)

func init() {
	exe := os.Args[0]

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	Args.ExeAbsDir = pwd
	Args.ExeName = path.Base(exe)
}

func InitConfigFlag(defaultVal string, tips string) {
	if isInit {
		return
	}
	isInit = true

	flag.StringVar(&Args.ConfigFile, "c", defaultVal, tips)
	return
}

func FlagHandle() {
	if !path.IsAbs(Args.ConfigFile) {
		Args.ConfigFile = path.Join(Args.ExeAbsDir, Args.ConfigFile)
	}
}
