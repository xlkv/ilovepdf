package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	u := ctx.EffectiveUser
	userID := u.Id
	username := ""
	if u.Username != "" {
		username = "@" + u.Username
	}

	sess := h.sm.Get(userID)
	h.sm.TrackUserFull(userID, username, u.FirstName, u.LastName, sess.Language)
	h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

	// Step 1: Prompt Language selection if first time
	if !sess.LanguageSelected {
		langPrompt := "🌐 **Iltimos, tilni tanlang:**\n🌐 **Пожалуйста, выберите язык:**\n🌐 **Please select your language:**"
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

	// Step 2: Language is selected -> Present Main Menu
	h.sm.Reset(userID)
	sess = h.sm.Get(userID)

	msgText := "✨ **iLovePDF** — Kerakli bo'limni tanlang:"
	if sess.Language == "ru" {
		msgText = "✨ **iLovePDF** — Выберите нужный раздел:"
	} else if sess.Language == "en" {
		msgText = "✨ **iLovePDF** — Select a tool below:"
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

// HandleStats shows detailed analytics data to admin only
func (h *BotHandlers) HandleStats(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	if userID != 8958346579 && userID != h.cfg.AdminUserID {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "🚫 **Ushbu buyruq faqat bot admini uchun ruxsat berilgan.**", &gotgbot.SendMessageOpts{
			ParseMode: "Markdown",
		})
		return err
	}

	stats := h.sm.GetDetailedStats()

	// Language breakdown percentages
	total := stats.TotalUsers
	uzPct, ruPct, enPct := 0.0, 0.0, 0.0
	if total > 0 {
		uzPct = (float64(stats.LangCounts["uz"]) / float64(total)) * 100
		ruPct = (float64(stats.LangCounts["ru"]) / float64(total)) * 100
		enPct = (float64(stats.LangCounts["en"]) / float64(total)) * 100
	}

	// Top tools text
	topToolsText := ""
	limit := 5
	if len(stats.TopTools) < limit {
		limit = len(stats.TopTools)
	}
	for i := 0; i < limit; i++ {
		t := stats.TopTools[i]
		topToolsText += fmt.Sprintf("  %d. `%s`: **%d** marta\n", i+1, t.ToolID, t.Count)
	}
	if topToolsText == "" {
		topToolsText = "  • Hozircha uskunalar statistikasi mavjud emas\n"
	}

	// Recent users text
	recentText := ""
	for i, u := range stats.RecentUsers {
		name := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if name == "" {
			name = "Foydalanuvchi"
		}
		userTag := u.Username
		if userTag == "" {
			userTag = fmt.Sprintf("ID: %d", u.UserID)
		}
		timeShort := u.LastActive
		if len(timeShort) >= 16 {
			timeShort = timeShort[11:16] + " (" + timeShort[8:10] + "-" + timeShort[5:7] + ")"
		}
		recentText += fmt.Sprintf("  %d. %s (%s) — `%s` | **%d** op\n", i+1, userTag, name, timeShort, u.TotalOperations)
	}
	if recentText == "" {
		recentText = "  • Foydalanuvchilar ro'yxati yo'q\n"
	}

	statsMsg := fmt.Sprintf(`📊 **iLovePDF — Admin Analitika va Statistika**

👥 **Foydalanuvchilar Ko'rsatkichlari:**
• Jami ro'yxatdan o'tganlar: **%d** ta
• Bugun faol bo'lganlar: **%d** ta
• Bugun yangi qo'shilganlar: **%d** ta
• Jami bajarilgan fayl operatsiyalari: **%d** ta

🌐 **Tillar bo'yicha ulushi:**
• 🇺🇿 O'zbek tili: **%d** (%0.1f%%)
• 🇷🇺 Rus tili: **%d** (%0.1f%%)
• 🇬🇧 Ingliz tili: **%d** (%0.1f%%)

🛠 **Eng ko'p ishlatilgan Top-5 Uskunalar:**
%s
👥 **So'nggi faol 10 ta Foydalanuvchilar:**
%s`,
		stats.TotalUsers, stats.ActiveToday, stats.NewToday, stats.TotalFiles,
		stats.LangCounts["uz"], uzPct,
		stats.LangCounts["ru"], ruPct,
		stats.LangCounts["en"], enPct,
		topToolsText, recentText)

	_, err := b.SendMessage(ctx.EffectiveChat.Id, statsMsg, &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	})
	return err
}

// HandleBroadcast sends announcement message to all users (Admin only)
func (h *BotHandlers) HandleBroadcast(b *gotgbot.Bot, ctx *ext.Context) error {
	userID := ctx.EffectiveUser.Id
	if userID != 8958346579 && userID != h.cfg.AdminUserID {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "🚫 **Ushbu buyruq faqat bot admini uchun ruxsat berilgan.**", &gotgbot.SendMessageOpts{
			ParseMode: "Markdown",
		})
		return err
	}

	if len(ctx.EffectiveMessage.Text) <= 11 {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "⚠️ **Foydalanish:** `/broadcast <xabar matni>`", &gotgbot.SendMessageOpts{ParseMode: "Markdown"})
		return err
	}

	broadcastMsg := ctx.EffectiveMessage.Text[11:]
	users := h.sm.GetAllUserIDs()

	statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("⏳ **Xabar %d ta foydalanuvchiga yuborilmoqda...**", len(users)), &gotgbot.SendMessageOpts{ParseMode: "Markdown"})

	successCount := 0
	failCount := 0

	for _, targetID := range users {
		_, err := b.SendMessage(targetID, broadcastMsg, &gotgbot.SendMessageOpts{
			ParseMode: "Markdown",
		})
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	resultText := fmt.Sprintf(`✅ **Xabar tarqatish yakunlandi!**

• Muvaffaqiyatli yetib bordi: **%d**
• Xatolik (bloklaganlar): **%d**`, successCount, failCount)

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

	msgText := "❌ Operatsiya bekor qilindi."
	if sess.Language == "ru" {
		msgText = "❌ Операция отменена."
	} else if sess.Language == "en" {
		msgText = "❌ Operation cancelled."
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
	u := ctx.EffectiveUser
	h.sm.SetLanguage(u.Id, lang)
	username := ""
	if u.Username != "" {
		username = "@" + u.Username
	}
	h.sm.TrackUserFull(u.Id, username, u.FirstName, u.LastName, lang)

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
