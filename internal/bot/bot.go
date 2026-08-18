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

	botHandlers "github.com/xlkv/ilovepdf/internal/bot/handlers"
	"github.com/xlkv/ilovepdf/internal/config"
	"github.com/xlkv/ilovepdf/internal/engine"
	"github.com/xlkv/ilovepdf/internal/session"
)

type Bot struct {
	cfg     *config.Config
	b       *gotgbot.Bot
	updater *ext.Updater
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

	// Register Bot Commands with Telegram Menu for iLovePDF
	_, _ = b.SetMyCommands([]gotgbot.BotCommand{
		{Command: "start", Description: "🚀 Asosiy menyu / Main Menu"},
		{Command: "cancel", Description: "❌ Bekor qilish / Cancel"},
		{Command: "lang", Description: "🌐 Tilni o'zgartirish / Language"},
		{Command: "stats", Description: "📊 Admin statistika (Admin only)"},
	}, nil)

	// Set Bot Description
	botDescription := `✨ iLovePDF — Smart PDF Tools

🇺🇿 PDF fayllar bilan ishlash va konvertatsiya qilish.
🇷🇺 Удобные инструменты для работы и конвертации PDF.
🇬🇧 All-in-one PDF tools and converter.`

	_, _ = b.SetMyDescription(&gotgbot.SetMyDescriptionOpts{
		Description: botDescription,
	})

	botShortDescription := "✨ Smart PDF tools / PDF fayllar boti"
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
	dispatcher.AddHandler(handlers.NewCommand("cancel", h.HandleCancel))
	dispatcher.AddHandler(handlers.NewCommand("lang", h.HandleLangNav))
	dispatcher.AddHandler(handlers.NewCommand("stats", h.HandleStats))
	dispatcher.AddHandler(handlers.NewCommand("broadcast", h.HandleBroadcast))

	// Callback Query Handlers
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
		cfg:     cfg,
		b:       b,
		updater: updater,
	}, nil
}

func (b *Bot) Start() error {
	log.Printf("🚀 Starting iLovePDF Bot as @%s...", b.b.User.Username)

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
