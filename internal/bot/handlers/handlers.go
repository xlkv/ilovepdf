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

// ClearOldMessages deletes all temporary accumulated messages for clean UI
func (h *BotHandlers) ClearOldMessages(b *gotgbot.Bot, chatID int64, userID int64) {
	msgIDs := h.sm.PopTempMsgs(userID)
	for _, id := range msgIDs {
		if id > 0 {
			_, _ = b.DeleteMessage(chatID, id, nil)
		}
	}
}

// EditMessage is a helper to safely edit both Telegram text & photo caption messages
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
	if err != nil {
		// Fallback to EditMessageCaption if editing a Photo/Banner message
		capOpts := &gotgbot.EditMessageCaptionOpts{
			Caption:   text,
			ChatId:    chatID,
			MessageId: messageID,
			ParseMode: "Markdown",
		}
		if replyMarkup != nil {
			capOpts.ReplyMarkup = *replyMarkup
		}
		msg, _, err = b.EditMessageCaption(capOpts)
	}
	return msg, err
}

// HandleStart responds to /start or main menu callback
func (h *BotHandlers) HandleStart(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	h.sm.TrackUser(userID)
	h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)
	sess := h.sm.Get(userID)

	// Step 1: If user hasn't selected language yet, present Language Selection immediately
	if !sess.LanguageSelected {
		langPrompt := "🌐 **Iltimos, muloqot tilini tanlang:**\n🌐 **Пожалуйста, выберите язык:**\n🌐 **Please select your language:**"
		kb := keyboards.LanguageKeyboard()

		if ctx.CallbackQuery != nil {
			_, _ = ctx.CallbackQuery.Answer(b, nil)
			_, err := h.EditMessage(b, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, langPrompt, &kb)
			if err == nil {
				return nil
			}
		}

		bannerPath := "assets/banner.png"
		if f, err := os.Open(bannerPath); err == nil {
			defer f.Close()
			_, sendErr := b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByReader("banner.png", f), &gotgbot.SendPhotoOpts{
				Caption:     langPrompt,
				ParseMode:   "Markdown",
				ReplyMarkup: kb,
			})
			if sendErr == nil {
				return nil
			}
		}

		_, err := b.SendMessage(ctx.EffectiveChat.Id, langPrompt, &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	// Step 2: Language is selected -> Present Main Menu in chosen language
	h.sm.Reset(userID)
	sess = h.sm.Get(userID)

	msgText := "✨ **iLovePDF Bot** — Kerakli bo'limni tanlang:"
	if sess.Language == "ru" {
		msgText = "✨ **iLovePDF Bot** — Выберите нужный раздел:"
	} else if sess.Language == "en" {
		msgText = "✨ **iLovePDF Bot** — Select a tool below:"
	}

	kb := keyboards.MainMenuKeyboard(sess.Language)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, msgText, &kb)
		if err == nil {
			return nil
		}
	}

	bannerPath := "assets/banner.png"
	if f, err := os.Open(bannerPath); err == nil {
		defer f.Close()
		_, sendErr := b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByReader("banner.png", f), &gotgbot.SendPhotoOpts{
			Caption:     msgText,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		if sendErr == nil {
			return nil
		}
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, msgText, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return err
}

// HandleStats shows analytics data to admin
func (h *BotHandlers) HandleStats(b *gotgbot.Bot, ctx *ext.Context) error {
	totalUsers, activeToday, totalFiles, toolStats := h.sm.GetStats()

	toolStatsText := ""
	for tool, count := range toolStats {
		toolStatsText += fmt.Sprintf("• `%s`: %d marta\n", tool, count)
	}
	if toolStatsText == "" {
		toolStatsText = "• Hozircha uskunalar statistikasi yo'q\n"
	}

	statsMsg := fmt.Sprintf(`📊 **iLovePDF Bot Analitika va Statistika**

👥 **Foydalanuvchilar:**
• Jami ro'yxatdan o'tganlar: **%d ta**
• Bugun faol foydalanuvchilar: **%d ta**
• Qayta ishlangan jami fayllar: **%d ta**

🛠 **Eng ko'p ishlatilgan uskunalar:**
%s`, totalUsers, activeToday, totalFiles, toolStatsText)

	_, err := b.SendMessage(ctx.EffectiveChat.Id, statsMsg, &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	})
	return err
}

// HandleBroadcast sends announcement message to all users
func (h *BotHandlers) HandleBroadcast(b *gotgbot.Bot, ctx *ext.Context) error {
	if len(ctx.EffectiveMessage.Text) <= 11 {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "⚠️ **Foydalanish:** `/broadcast <xabar matni>`", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
		return err
	}

	broadcastMsg := ctx.EffectiveMessage.Text[11:]
	users := h.sm.GetAllUsers()

	statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("⏳ **Xabar %d ta foydalanuvchiga yuborilmoqda...**", len(users)), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})

	successCount := 0
	failCount := 0

	for _, userID := range users {
		_, err := b.SendMessage(userID, broadcastMsg, &gotgbot.SendMessageOpts{
			ParseMode: "Markdown",
		})
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	resultText := fmt.Sprintf(`✅ **Xabar tarqatish yakunlandi!**

• Muvaffaqiyatli yetib bordi: **%d ta**
• Etib bormadi (bloklangan/o'chirilgan): **%d ta**`, successCount, failCount)

	if statusMsg != nil {
		_, _, _ = b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:    ctx.EffectiveChat.Id,
			MessageId: statusMsg.MessageId,
			Text:      resultText,
			ParseMode: "Markdown",
		})
	}

	return nil
}

// HandleCancel handles cancellation
func (h *BotHandlers) HandleCancel(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)
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
	userID := ctx.EffectiveUser.Id
	sess := h.sm.Get(userID)
	sess.LanguageSelected = false
	return h.HandleStart(b, ctx)
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
