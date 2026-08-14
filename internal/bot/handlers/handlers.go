package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/xlkv/ilovepdf/internal/bot/keyboards"
	"github.com/xlkv/ilovepdf/internal/config"
	"github.com/xlkv/ilovepdf/internal/engine"
	"github.com/xlkv/ilovepdf/internal/session"
)

type BotHandlers struct {
	cfg        *config.Config
	sm         *session.Manager
	pdfcpuEng  *engine.PDFCPUEngine
	convertEng *engine.ConvertersEngine
}

func NewBotHandlers(cfg *config.Config, sm *session.Manager, pdfcpu *engine.PDFCPUEngine, conv *engine.ConvertersEngine) *BotHandlers {
	return &BotHandlers{
		cfg:        cfg,
		sm:         sm,
		pdfcpuEng:  pdfcpu,
		convertEng: conv,
	}
}

// EditMessage is a helper to safely edit Telegram messages
func (h *BotHandlers) EditMessage(b *gotgbot.Bot, chatID int64, messageID int64, text string, replyMarkup *gotgbot.InlineKeyboardMarkup) (*gotgbot.Message, error) {
	opts := &gotgbot.EditMessageTextOpts{
		Text:      text,
		ChatId:    chatID,
		MessageId: messageID,
		ParseMode: "Markdown",
	}
	if replyMarkup != nil {
		opts.ReplyMarkup = *replyMarkup
	}
	msg, _, err := b.EditMessageText(opts)
	return msg, err
}

// HandleStart responds to /start or main menu callback
func (h *BotHandlers) HandleStart(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	h.sm.Reset(userID)
	sess := h.sm.Get(userID)

	msgText := "👋 **iLovePDF Telegram Botiga xush kelibsiz!**\n\nQuyidagi tugmalardan birini tanlang:"
	if sess.Language == "ru" {
		msgText = "👋 **Добро пожаловать в iLovePDF Telegram Bot!**\n\nВыберите нужную функцию:"
	} else if sess.Language == "en" {
		msgText = "👋 **Welcome to iLovePDF Telegram Bot!**\n\nSelect a tool below:"
	}

	kb := keyboards.MainMenuKeyboard(sess.Language)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, msgText, &kb)
		if err == nil {
			return nil
		}
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, msgText, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return err
}

// HandleCancel handles cancellation
func (h *BotHandlers) HandleCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	h.sm.Reset(userID)
	sess := h.sm.Get(userID)

	msgText := "❌ Operatsiya bekor qilindi. Bosh menyu:"
	if sess.Language == "ru" {
		msgText = "❌ Операция отменена. Главное меню:"
	} else if sess.Language == "en" {
		msgText = "❌ Operation cancelled. Main menu:"
	}

	kb := keyboards.MainMenuKeyboard(sess.Language)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Cancelled"})
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, msgText, &kb)
		return err
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, msgText, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return err
}

// HandleLangNav handles language menu
func (h *BotHandlers) HandleLangNav(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
		kb := keyboards.LanguageKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, "🌐 **Tilni tanlang / Select Language:**", &kb)
		return err
	}
	return nil
}

// HandleLangSelect updates user language
func (h *BotHandlers) HandleLangSelect(b *gotgbot.Bot, ctx *ext.Context, lang string) error {
	userID := ctx.EffectiveUser.Id
	h.sm.SetLanguage(userID, lang)
	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{Text: "Language updated!"})
	}
	return h.HandleStart(b, ctx)
}

// DownloadTelegramFile downloads document or photo from Telegram to local path
func (h *BotHandlers) DownloadTelegramFile(b *gotgbot.Bot, fileID, destPath string) error {
	f, err := b.GetFile(fileID, nil)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileURL := f.URL(b, nil)
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download file stream: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SendDocumentResponse sends output file to user with Back to Main Menu button attached
func (h *BotHandlers) SendDocumentResponse(b *gotgbot.Bot, chatID int64, filePath, caption, lang string) error {
	fileData, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer fileData.Close()

	fileName := filepath.Base(filePath)
	kb := keyboards.BackToMenuKeyboard(lang)
	_, err = b.SendDocument(chatID, gotgbot.InputFileByReader(fileName, fileData), &gotgbot.SendDocumentOpts{
		Caption:     caption,
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return err
}
