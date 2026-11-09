package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-up-monitoring/internal/config"
	"go-up-monitoring/internal/database"
	"go-up-monitoring/internal/modules/auth"
	"go-up-monitoring/internal/modules/heartbeat"
	"go-up-monitoring/internal/modules/incident"
	"go-up-monitoring/internal/modules/monitor"
	"go-up-monitoring/internal/modules/notification"
	"go-up-monitoring/internal/queue"
	"go-up-monitoring/internal/worker"
	"go-up-monitoring/pkg/logger"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	_, err = logger.InitLogger(logger.Config{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSizeMB:  cfg.Log.MaxSizeMB,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAgeDays: cfg.Log.MaxAgeDays,
		Compress:   cfg.Log.Compress,
		Stdout:     true,
	})
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("starting Go Uptime Monitor Worker Engine", zap.String("app", cfg.App.Name))

	// Connect to Database
	db, err := database.ConnectPostgres(cfg.Database)
	if err != nil {
		logger.Warn("database connection failed (will retry or continue in mock)", zap.Error(err))
	} else {
		defer db.Close()
		_ = auth.AutoMigrate(db.DB)
		_ = monitor.AutoMigrate(db.DB)
		_ = heartbeat.AutoMigrate(db.DB)
		_ = incident.AutoMigrate(db.DB)
	}

	// Connect to Redis & Queue
	q := queue.NewQueue(cfg.Redis)
	defer q.Close()

	var monRepo monitor.Repository
	var hbRepo heartbeat.Repository
	var incRepo incident.Repository
	var incSvc incident.Service

	if db != nil {
		monRepo = monitor.NewRepository(db.DB)
		hbRepo = heartbeat.NewRepository(db.DB)
		incRepo = incident.NewRepository(db.DB)

		notifier := incident.NewAsynqAlertNotifier(q.AsynqClient)
		incSvc = incident.NewService(incRepo, notifier)
	}

	prober := heartbeat.NewProber(10 * time.Second)
	processor := worker.NewProcessor(monRepo, hbRepo, prober, incSvc)

	// Configure Telegram notifier and debouncer
	telegramClient := notification.NewTelegramClient(cfg.Telegram.BotToken, &http.Client{Timeout: 10 * time.Second})
	debouncer := notification.NewRedisDebouncer(q.RedisClient)
	processor.SetNotification(telegramClient, debouncer, cfg.Telegram.DefaultChatID)

	// Setup Asynq Server
	concurrency := cfg.Worker.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	srv := asynq.NewServer(
		q.RedisOpt,
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"probes": 10,
				"alerts": 5,
			},
			Logger: logger.NewAsynqLogger(logger.GetLogger()),
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if monRepo != nil {
		sched := worker.NewScheduler(monRepo, q.AsynqClient, 10*time.Second)
		go sched.Start(ctx)
		defer sched.Stop()
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		mux := asynq.NewServeMux()
		mux.HandleFunc(worker.TypeProbeMonitor, processor.ProcessTask)
		mux.HandleFunc(worker.TypeDispatchAlert, processor.ProcessTask)

		if err := srv.Run(mux); err != nil {
			logger.Error("asynq server error", zap.Error(err))
		}
	}()

	logger.Info("worker engine is running. Press CTRL+C to stop.")
	<-stop
	logger.Info("shutting down worker engine...")
	srv.Shutdown()
}
