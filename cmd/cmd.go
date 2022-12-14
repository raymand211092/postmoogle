package main

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/etke.cc/go/logger"
	"gitlab.com/etke.cc/linkpearl"
	lpcfg "gitlab.com/etke.cc/linkpearl/config"

	"gitlab.com/etke.cc/postmoogle/bot"
	"gitlab.com/etke.cc/postmoogle/config"
	"gitlab.com/etke.cc/postmoogle/smtp"
)

var (
	mxb      *bot.Bot
	smtpserv *smtp.Server
	log      *logger.Logger
)

func main() {
	quit := make(chan struct{})

	cfg := config.New()
	log = logger.New("postmoogle.", cfg.LogLevel)

	log.Info("#############################")
	log.Info("Postmoogle")
	log.Info("Matrix: true")
	log.Info("#############################")

	log.Debug("starting internal components...")
	initSentry(cfg)
	initBot(cfg)
	initSMTP(cfg)
	initShutdown(quit)
	defer recovery()

	go startBot(cfg.StatusMsg)

	if err := smtpserv.Start(); err != nil {
		//nolint:gocritic
		log.Fatal("SMTP server crashed: %v", err)
	}

	<-quit
}

func initSentry(cfg *config.Config) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.Sentry.DSN,
		AttachStacktrace: true,
	})
	if err != nil {
		log.Fatal("cannot initialize sentry: %v", err)
	}
}

func initBot(cfg *config.Config) {
	db, err := sql.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		log.Fatal("cannot initialize SQL database: %v", err)
	}
	mxlog := logger.New("matrix.", cfg.LogLevel)
	lp, err := linkpearl.New(&lpcfg.Config{
		Homeserver:   cfg.Homeserver,
		Login:        cfg.Login,
		Password:     cfg.Password,
		DB:           db,
		Dialect:      cfg.DB.Dialect,
		NoEncryption: cfg.NoEncryption,
		LPLogger:     mxlog,
		APILogger:    logger.New("api.", cfg.LogLevel),
		StoreLogger:  logger.New("store.", cfg.LogLevel),
		CryptoLogger: logger.New("olm.", cfg.LogLevel),
	})
	if err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("cannot initialize matrix bot: %v", err)
	}

	mxb, err = bot.New(lp, mxlog, cfg.Prefix, cfg.Domain, cfg.Admins)
	if err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("cannot start matrix bot: %v", err)
	}
	log.Debug("bot has been created")
}

func initSMTP(cfg *config.Config) {
	smtpserv = smtp.NewServer(&smtp.Config{
		Domain:      cfg.Domain,
		Port:        cfg.Port,
		TLSCert:     cfg.TLS.Cert,
		TLSKey:      cfg.TLS.Key,
		TLSPort:     cfg.TLS.Port,
		TLSRequired: cfg.TLS.Required,
		LogLevel:    cfg.LogLevel,
		MaxSize:     cfg.MaxSize,
		Bot:         mxb,
	})
}

func initShutdown(quit chan struct{}) {
	listener := make(chan os.Signal, 1)
	signal.Notify(listener, os.Interrupt, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-listener
		defer close(quit)

		shutdown()
	}()
}

func startBot(statusMsg string) {
	log.Debug("starting matrix bot: %s...", statusMsg)
	err := mxb.Start(statusMsg)
	if err != nil {
		//nolint:gocritic
		log.Fatal("cannot start the bot: %v", err)
	}
}

func shutdown() {
	log.Info("Shutting down...")
	smtpserv.Stop()
	mxb.Stop()

	sentry.Flush(5 * time.Second)
	log.Info("Postmoogle has been stopped")
	os.Exit(0)
}

func recovery() {
	defer shutdown()
	err := recover()
	// no problem just shutdown
	if err == nil {
		sentry.CurrentHub().Recover(err)
	}
}
