package constants

import "flag"

var ConfigPath string

func init() {
	flag.StringVar(&ConfigPath, "configPath", "configs/development", "Specify app startup configs path.")
}
