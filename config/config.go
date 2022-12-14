package config

import (
	"gitlab.com/etke.cc/go/env"
)

const prefix = "postmoogle"

// New config
func New() *Config {
	env.SetPrefix(prefix)

	cfg := &Config{
		Homeserver:   env.String("homeserver", defaultConfig.Homeserver),
		Login:        env.String("login", defaultConfig.Login),
		Password:     env.String("password", defaultConfig.Password),
		Prefix:       env.String("prefix", defaultConfig.Prefix),
		Domain:       env.String("domain", defaultConfig.Domain),
		Port:         env.String("port", defaultConfig.Port),
		NoEncryption: env.Bool("noencryption"),
		MaxSize:      env.Int("maxsize", defaultConfig.MaxSize),
		StatusMsg:    env.String("statusmsg", defaultConfig.StatusMsg),
		Admins:       env.Slice("admins"),
		TLS: TLS{
			Cert:     env.String("tls.cert", defaultConfig.TLS.Cert),
			Key:      env.String("tls.key", defaultConfig.TLS.Key),
			Required: env.Bool("tls.required"),
			Port:     env.String("tls.port", defaultConfig.TLS.Port),
		},
		Sentry: Sentry{
			DSN: env.String("sentry.dsn", defaultConfig.Sentry.DSN),
		},
		LogLevel: env.String("loglevel", defaultConfig.LogLevel),
		DB: DB{
			DSN:     env.String("db.dsn", defaultConfig.DB.DSN),
			Dialect: env.String("db.dialect", defaultConfig.DB.Dialect),
		},
	}

	return cfg
}
