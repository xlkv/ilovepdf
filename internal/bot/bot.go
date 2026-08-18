package bot

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"github.com/xlkv/ilovepdf/internal/analytics"
	botHandlers "github.com/xlkv/ilovepdf/internal/bot/handlers"
	"github.com/xlkv/ilovepdf/internal/config"
	"github.com/xlkv/ilovepdf/internal/db"
	"github.com/xlkv/ilovepdf/internal/engine"
	"github.com/xlkv/ilovepdf/internal/matcher"
	"github.com/xlkv/ilovepdf/internal/payments"
	"github.com/xlkv/ilovepdf/internal/scraper"
	"github.com/xlkv/ilovepdf/internal/session"
)

type Bot struct {
	cfg        *config.Config
	b          *gotgbot.Bot
	updater    *ext.Updater
	monitorH   *botHandlers.MonitorHandlers
	payManager *payments.PaymentManager
}

func NewBot(cfg *config.Config) (*Bot, error) {
	b, err := gotgbot.NewBot(cfg.BotToken, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	sm := session.NewManager(cfg.TempDir)
	pdfcpuEng := engine.NewPDFCPUEngine()
	convEng := engine.NewConvertersEngine(cfg.SofficePath, cfg.PythonPath, cfg.TesseractPath)

	h := botHandlers.NewBotHandlers(cfg, sm, pdfcpuEng, convEng)

	// Initialize Avto-E'lon Monitor Engine
	storage, err := db.NewStorage("data")
	if err != nil {
		log.Printf("[WARN] Storage initialization error: %v", err)
	}
	evaluator := analytics.NewPriceEvaluator()
	olxScraper := scraper.NewOLXScraper(evaluator)
	filterMatcher := matcher.NewFilterMatcher(storage)
	payMgr := payments.NewPaymentManager(storage, "SECRET_KEY", "12345", "67890")

	mh := botHandlers.NewMonitorHandlers(storage, evaluator, olxScraper, filterMatcher, payMgr)

	// Register Bot Commands with Telegram Menu
	_, _ = b.SetMyCommands([]gotgbot.BotCommand{
		{Command: "start", Description: "🚀 Asosiy menyu / Main Menu"},
		{Command: "monitor", Description: "🚗 Avto-E'lon Monitor Bot (Arzon E'lonlar)"},
		{Command: "cancel", Description: "❌ Bekor qilish / Cancel"},
		{Command: "lang", Description: "🌐 Tilni o'zgartirish / Language"},
		{Command: "stats", Description: "📊 Admin statistika (Admin only)"},
	}, nil)

	// Set Bot Description ("What can this bot do?" intro screen)
	botDescription := `✨ iLovePDF & Avto-E'lon Monitor

🇺🇿 PDF fayllar bilan ishlash va Real-time Arzon Avto-E'lonlar monitori.
🇷🇺 Удобные инструменты для работы с PDF и Мониторинг авто-объявлений.
🇬🇧 All-in-one PDF tools and Real-time Marketplace Alert Engine.`

	_, _ = b.SetMyDescription(&gotgbot.SetMyDescriptionOpts{
		Description: botDescription,
	})

	botShortDescription := "✨ Smart PDF tools & Avto-E'lon Monitor"
	_, _ = b.SetMyShortDescription(&gotgbot.SetMyShortDescriptionOpts{
		ShortDescription: botShortDescription,
	})

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Printf("[ERROR] Dispatcher error: %v", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	updater := ext.NewUpdater(dispatcher, nil)

	// Command Handlers
	dispatcher.AddHandler(handlers.NewCommand("start", h.HandleStart))
	dispatcher.AddHandler(handlers.NewCommand("monitor", mh.HandleMonitorStart))
	dispatcher.AddHandler(handlers.NewCommand("cancel", h.HandleCancel))
	dispatcher.AddHandler(handlers.NewCommand("lang", h.HandleLangNav))
	dispatcher.AddHandler(handlers.NewCommand("stats", h.HandleStats))
	dispatcher.AddHandler(handlers.NewCommand("broadcast", h.HandleBroadcast))

	// Monitor Callback Query Handlers
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("monitor:"), func(b *gotgbot.Bot, ctx *ext.Context) error {
		return mh.HandleCallbackQuery(b, ctx)
	}))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("flt_"), func(b *gotgbot.Bot, ctx *ext.Context) error {
		return mh.HandleCallbackQuery(b, ctx)
	}))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("vip:"), func(b *gotgbot.Bot, ctx *ext.Context) error {
		return mh.HandleCallbackQuery(b, ctx)
	}))

	// Callback Query Handlers for PDF
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("action:cancel"), h.HandleCancel))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("nav:lang"), h.HandleLangNav))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("nav:main"), h.HandleStart))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("lang:"), func(b *gotgbot.Bot, ctx *ext.Context) error {
		lang := ctx.CallbackQuery.Data[5:]
		return h.HandleLangSelect(b, ctx, lang)
	}))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("tool:"), func(b *gotgbot.Bot, ctx *ext.Context) error {
		toolID := ctx.CallbackQuery.Data[5:]
		return h.HandleToolSelect(b, ctx, toolID)
	}))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.All, func(b *gotgbot.Bot, ctx *ext.Context) error {
		return h.HandleCallbackActions(b, ctx, ctx.CallbackQuery.Data)
	}))

	// File & Document Filters
	dispatcher.AddHandler(handlers.NewMessage(message.Document, h.HandleDocumentUpload))
	dispatcher.AddHandler(handlers.NewMessage(message.Photo, h.HandlePhotoUpload))
	dispatcher.AddHandler(handlers.NewMessage(message.Text, h.HandleTextMessage))

	return &Bot{
		cfg:        cfg,
		b:          b,
		updater:    updater,
		monitorH:   mh,
		payManager: payMgr,
	}, nil
}

func (b *Bot) Start() error {
	log.Printf("🚀 Starting Bot Engine as @%s...", b.b.User.Username)

	// Start Background Marketplace Real-Time Scraper Engine
	b.monitorH.StartBackgroundScraper(b.b)

	// Start Click / Payme HTTP Webhook Server on port 8089
	go func() {
		http.HandleFunc("/payment/click", b.payManager.HandleClickWebhook)
		log.Printf("💳 Click Payment Webhook Server running on http://0.0.0.0:8089/payment/click")
		_ = http.ListenAndServe(":8089", nil)
	}()

	// Start WebApp static HTTP server on port 8088
	go func() {
		fs := http.FileServer(http.Dir("webapp"))
		http.Handle("/", fs)
		log.Printf("🌐 WebApp HTTP Server running on http://0.0.0.0:8088")
		if err := http.ListenAndServe(":8088", nil); err != nil {
			log.Printf("[WARN] WebApp HTTP server stopped: %v", err)
		}
	}()

	err := b.updater.StartPolling(b.b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 10,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to start polling: %w", err)
	}

	b.updater.Idle()
	return nil
}

func (b *Bot) GetBotUser() *gotgbot.User {
	return &b.b.User
}
