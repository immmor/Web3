package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	DataSource string `json:",default=file:user.db?cache=shared&_fk=1"`
}
