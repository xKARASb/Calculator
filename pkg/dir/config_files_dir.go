package dir

import "os"

func config_files() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir += `\config\`
	return dir
}

func Json_file() string {
	res := config_files() + `config.json`
	return res
}

func Env_file() string {
	res := config_files() + `.env`
	return res
}
