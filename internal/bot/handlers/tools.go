package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/xlkv/ilovepdf/internal/bot/keyboards"
	"github.com/xlkv/ilovepdf/internal/engine"
	"github.com/xlkv/ilovepdf/internal/session"
)

// HandleToolSelect responds when user clicks a tool from main menu
func (h *BotHandlers) HandleToolSelect(b *gotgbot.Bot, ctx *ext.Context, toolID string) error {
	userID := ctx.EffectiveUser.Id
	h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)
	h.sm.Reset(userID)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
	}

	var prompt string
	var newState session.State

	switch toolID {
	case "merge":
		newState = session.StateMergeUploading
		prompt = "🧩 **PDF Birlashtirish:**\n\nBirlashtirmoqchi bo'lgan 2 yoki undan ortiq PDF (yoki Word/Office) fayllaringizni birin-ketin yuboring."
	case "split":
		newState = session.StateSplitAwaitFile
		prompt = "✂️ **PDF Ajratish:**\n\nAjratmoqchi bo'lgan PDF faylingizni yuboring."
	case "compress":
		newState = session.StateCompressAwaitFile
		prompt = "📉 **PDF Siqish:**\n\nHajmini kichraytirmoqchi bo'lgan PDF faylni yuboring."
	case "word2pdf":
		newState = session.StateWord2PDFAwaitFile
		prompt = "📝 **Word -> PDF:**\n\nWord hujjatini (`.docx` yoki `.doc`) yuboring."
	case "ppt2pdf":
		newState = session.StatePPT2PDFAwaitFile
		prompt = "📊 **PowerPoint -> PDF:**\n\nPrezentatsiyani (`.pptx` yoki `.ppt`) yuboring."
	case "excel2pdf":
		newState = session.StateExcel2PDFAwaitFile
		prompt = "📈 **Excel -> PDF:**\n\nExcel jadvalini (`.xlsx` yoki `.xls`) yuboring."
	case "pdf2word":
		newState = session.StatePDF2WordAwaitFile
		prompt = "📄 **PDF -> Word:**\n\nWord ga o'girmoqchi bo'lgan PDF faylni yuboring."
	case "pdf2jpg":
		newState = session.StatePDF2JPGAwaitFile
		prompt = "🖼 **PDF -> JPG:**\n\nRasmlarga ajratmoqchi bo'lgan PDF faylni yuboring."
	case "jpg2pdf":
		newState = session.StateJPG2PDFUploading
		prompt = "📷 **JPG -> PDF:**\n\nPDF qilmoqchi bo'lgan 1 yoki undan ortiq rasmlaringizni yuboring."
	case "rotate":
		newState = session.StateRotateAwaitFile
		prompt = "🔄 **PDF Burish:**\n\nBurchagini o'zgartirmoqchi bo'lgan PDF faylingizni yuboring."
	case "protect":
		newState = session.StateProtectAwaitFile
		prompt = "🔒 **PDF Parol Qo'yish:**\n\nParol o'rnatmoqchi bo'lgan PDF faylingizni yuboring."
	case "unlock":
		newState = session.StateUnlockAwaitFile
		prompt = "🔓 **PDF Parolni O'chirish:**\n\nParolini olib tashlamoqchi bo'lgan PDF faylingizni yuboring."
	case "watermark":
		newState = session.StateWatermarkAwaitFile
		prompt = "🏷 **Watermark Qo'shish:**\n\nWatermark qo'shmoqchi bo'lgan PDF faylingizni yuboring."
	case "pagenum":
		newState = session.StatePagenumAwaitFile
		prompt = "🔢 **Sahifa Raqamlarini Qo'shish:**\n\nPDF faylingizni yuboring."
	case "organize":
		newState = session.StateOrganizeAwaitFile
		prompt = "🗂 **PDF Tartiblash:**\n\nSahifalar ketma-ketligini o'zgartirmoqchi bo'lgan PDF faylni yuboring."
	case "html2pdf":
		newState = session.StateHTML2PDFAwaitInput
		prompt = "🌐 **Web -> PDF:**\n\nSayt manzillarini (masalan: `https://example.com`) yoki `.html` faylini yuboring."
	case "ocr":
		newState = session.StateOCRAwaitFile
		prompt = "🔍 **OCR Matnni Tanish:**\n\nScan qilingan PDF faylingizni yuboring."
	default:
		return nil
	}

	h.sm.SetState(userID, newState)
	kb := keyboards.CancelKeyboard()

	if ctx.CallbackQuery != nil {
		msg, err := h.EditMessage(b, ctx.EffectiveChat.Id, ctx.EffectiveMessage.MessageId, prompt, &kb)
		if msg != nil {
			h.sm.AddTempMsg(userID, msg.MessageId)
		}
		return err
	}

	msg, err := b.SendMessage(ctx.EffectiveChat.Id, prompt, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	if msg != nil {
		h.sm.AddTempMsg(userID, msg.MessageId)
	}
	return err
}

// HandleDocumentUpload handles all document uploads based on current FSM state
func (h *BotHandlers) HandleDocumentUpload(b *gotgbot.Bot, ctx *ext.Context) error {
	doc := ctx.EffectiveMessage.Document
	if doc == nil {
		return nil
	}

	userID := ctx.EffectiveUser.Id
	h.sm.AddTempMsg(userID, ctx.EffectiveMessage.MessageId)
	sess := h.sm.Get(userID)

	if sess.State == session.StateIdle {
		kb := keyboards.MainMenuKeyboard(sess.Language)
		msg, err := b.SendMessage(ctx.EffectiveChat.Id, "ℹ️ Iltimos, avval menyudan kerakli funksiyani tanlang:", &gotgbot.SendMessageOpts{
			ReplyMarkup: kb,
		})
		if msg != nil {
			h.sm.AddTempMsg(userID, msg.MessageId)
		}
		return err
	}

	destPath := filepath.Join(sess.SessionDir, fmt.Sprintf("%d_%s", len(sess.Files)+1, doc.FileName))
	statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ Fayl yuklab olinmoqda...", nil)
	if statusMsg != nil {
		h.sm.AddTempMsg(userID, statusMsg.MessageId)
	}

	err := h.DownloadTelegramFile(b, doc.FileId, destPath)
	if err != nil {
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, fmt.Sprintf("❌ Faylni yuklab olishda xatolik: %v", err), nil)
		return err
	}

	// Auto-convert Office files to PDF if uploaded in Merge mode
	extLower := strings.ToLower(filepath.Ext(destPath))
	if sess.State == session.StateMergeUploading && (extLower == ".doc" || extLower == ".docx" || extLower == ".pptx" || extLower == ".xlsx") {
		outPDF, convErr := h.convertEng.ConvertOfficeToPDF(destPath, sess.SessionDir)
		if convErr == nil && outPDF != "" {
			destPath = outPDF
			doc.FileName = filepath.Base(outPDF)
			fi, _ := os.Stat(outPDF)
			if fi != nil {
				doc.FileSize = fi.Size()
			}
		}
	}

	fileMeta := session.FileMeta{
		ID:       doc.FileId,
		Name:     doc.FileName,
		Path:     destPath,
		Size:     doc.FileSize,
		MimeType: doc.MimeType,
	}
	h.sm.AddFile(userID, fileMeta)

	switch sess.State {
	case session.StateMergeUploading:
		files := h.sm.Get(userID).Files
		msgText := fmt.Sprintf("📥 **Qabul qilingan fayllar (%d):**\n", len(files))
		for i, f := range files {
			msgText += fmt.Sprintf("%d. 📄 `%s` (%s)\n", i+1, f.Name, engine.FormatBytes(f.Size))
		}
		msgText += "\nYana fayl yuborishingiz yoki 'Birlashtirish' tugmasini bosishingiz mumkin:"
		kb := keyboards.MergeFilesKeyboard(len(files))
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, msgText, &kb)
		return err

	case session.StateJPG2PDFUploading:
		files := h.sm.Get(userID).Files
		msgText := fmt.Sprintf("📷 **Qabul qilingan rasmlar (%d):**\n", len(files))
		for i, f := range files {
			msgText += fmt.Sprintf("%d. 🖼 `%s`\n", i+1, f.Name)
		}
		kb := keyboards.JPG2PDFKeyboard(len(files))
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, msgText, &kb)
		return err

	case session.StateSplitAwaitFile:
		h.sm.SetState(userID, session.StateSplitAwaitRange)
		pageCount, _ := h.pdfcpuEng.GetPageCount(destPath)
		h.sm.SetMeta(userID, "page_count", fmt.Sprintf("%d", pageCount))

		prompt := fmt.Sprintf("📄 **Fayl qabul qilindi:** `%s` (%d sahifa)\n\nAjratmoqchi bo'lgan sahifalaringizni kiriting (masalan: `1-3, 5, 8-10`).", doc.FileName, pageCount)
		kb := keyboards.CancelKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateCompressAwaitFile:
		h.sm.SetState(userID, session.StateCompressAwaitLevel)
		prompt := fmt.Sprintf("📉 **Fayl qabul qilindi:** `%s` (%s)\n\nSiqish darajasini tanlang:", doc.FileName, engine.FormatBytes(doc.FileSize))
		kb := keyboards.CompressLevelKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateWord2PDFAwaitFile, session.StatePPT2PDFAwaitFile, session.StateExcel2PDFAwaitFile:
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "⏳ Hujjat PDF ga o'girilmoqda...", nil)
		outPDF, err := h.convertEng.ConvertOfficeToPDF(destPath, sess.SessionDir)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ O'girishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outPDF, "✅ PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StatePDF2WordAwaitFile:
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "⏳ PDF Word (DOCX) ga o'girilmoqda...", nil)
		outDocx := filepath.Join(sess.SessionDir, strings.TrimSuffix(doc.FileName, ".pdf")+".docx")
		err := h.convertEng.ConvertPDFToWord(destPath, outDocx)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ O'girishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outDocx, "✅ Word hujjati tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StatePDF2JPGAwaitFile:
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "⏳ PDF rasmlarga ajratilmoqda...", nil)
		imgDir := filepath.Join(sess.SessionDir, "images")
		_ = os.MkdirAll(imgDir, 0755)
		imgs, err := h.convertEng.ConvertPDFToImages(destPath, imgDir)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil || len(imgs) == 0 {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, "❌ Rasmlarga ajratishda xatolik yuz berdi.", nil)
			return err
		}

		zipPath := filepath.Join(sess.SessionDir, "images.zip")
		_ = createZipArchive(imgs, zipPath)
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, zipPath, fmt.Sprintf("✅ %d ta rasm tayyor (ZIP arxiv)!", len(imgs)), sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StateRotateAwaitFile:
		h.sm.SetState(userID, session.StateRotateAwaitAngle)
		kb := keyboards.RotateKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "🔄 Burish burchagini tanlang:", &kb)
		return err

	case session.StateProtectAwaitFile:
		h.sm.SetState(userID, session.StateProtectAwaitPass)
		kb := keyboards.CancelKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "🔒 PDF uchun parol kiriting:", &kb)
		return err

	case session.StateUnlockAwaitFile:
		h.sm.SetState(userID, session.StateUnlockAwaitPass)
		kb := keyboards.CancelKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "🔓 PDF ning parolini kiriting:", &kb)
		return err

	case session.StateWatermarkAwaitFile:
		h.sm.SetState(userID, session.StateWatermarkAwaitText)
		kb := keyboards.CancelKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "🏷 Watermark matnini kiriting (masalan: `CONFIDENTIAL`):", &kb)
		return err

	case session.StatePagenumAwaitFile:
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "⏳ Sahifa raqamlari qo'shilmoqda...", nil)
		outPDF := filepath.Join(sess.SessionDir, "numbered.pdf")
		err := h.pdfcpuEng.AddPageNumbers(destPath, outPDF)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Raqamlashda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outPDF, "✅ Sahifa raqamlari qo'shildi!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StateOrganizeAwaitFile:
		h.sm.SetState(userID, session.StateOrganizeAwaitPages)
		pageCount, _ := h.pdfcpuEng.GetPageCount(destPath)
		prompt := fmt.Sprintf("🗂 **Jami sahifalar soni:** %d\n\nYangi sahifalar tartibini kiriting (masalan: `3, 1, 2`):", pageCount)
		kb := keyboards.CancelKeyboard()
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateOCRAwaitFile:
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, "🔍 OCR qilinmoqda, kuting...", nil)
		outPDF := filepath.Join(sess.SessionDir, "ocr_searchable.pdf")
		err := h.convertEng.OCRPDF(destPath, outPDF, "eng")
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ OCR xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outPDF, "✅ OCR qilingan PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil
	}

	return nil
}

// HandlePhotoUpload handles photo uploads for JPG2PDF
func (h *BotHandlers) HandlePhotoUpload(b *gotgbot.Bot, ctx *ext.Context) error {
	photos := ctx.EffectiveMessage.Photo
	if len(photos) == 0 {
		return nil
	}

	userID := ctx.EffectiveUser.Id
	h.sm.AddTempMsg(userID, ctx.EffectiveMessage.MessageId)
	sess := h.sm.Get(userID)

	if sess.State != session.StateJPG2PDFUploading {
		return nil
	}

	bestPhoto := photos[len(photos)-1]
	destPath := filepath.Join(sess.SessionDir, fmt.Sprintf("photo_%d.jpg", len(sess.Files)+1))

	err := h.DownloadTelegramFile(b, bestPhoto.FileId, destPath)
	if err != nil {
		return err
	}

	h.sm.AddFile(userID, session.FileMeta{
		ID:   bestPhoto.FileId,
		Name: fmt.Sprintf("Photo_%d.jpg", len(sess.Files)+1),
		Path: destPath,
		Size: bestPhoto.FileSize,
	})

	files := h.sm.Get(userID).Files
	msgText := fmt.Sprintf("📷 **Qabul qilingan rasmlar (%d):**\n\nYana rasm yuborishingiz yoki PDF ga o'girish tugmasini bosishingiz mumkin.", len(files))
	kb := keyboards.JPG2PDFKeyboard(len(files))
	msg, err := b.SendMessage(ctx.EffectiveChat.Id, msgText, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	if msg != nil {
		h.sm.AddTempMsg(userID, msg.MessageId)
	}
	return err
}

// HandleTextMessage handles text inputs for password, page ranges, watermark text, etc.
func (h *BotHandlers) HandleTextMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	text := ctx.EffectiveMessage.Text
	if text == "" || strings.HasPrefix(text, "/") {
		return nil
	}

	userID := ctx.EffectiveUser.Id
	h.sm.AddTempMsg(userID, ctx.EffectiveMessage.MessageId)
	sess := h.sm.Get(userID)

	switch sess.State {
	case session.StateSplitAwaitRange:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outDir := filepath.Join(sess.SessionDir, "split_output")
		_ = os.MkdirAll(outDir, 0755)

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ Sahifalar ajratilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		pages := engine.ParsePageRange(text)
		outFiles, err := h.pdfcpuEng.SplitPDF(inFile, outDir, pages)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil || len(outFiles) == 0 {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Ajratishda xatolik: %v", err), nil)
			return err
		}

		if len(outFiles) == 1 {
			_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFiles[0], "✅ Ajratilgan PDF tayyor!", sess.Language)
		} else {
			zipPath := filepath.Join(sess.SessionDir, "split_pages.zip")
			_ = createZipArchive(outFiles, zipPath)
			_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, zipPath, "✅ Ajratilgan sahifalar (ZIP arxiv)!", sess.Language)
		}
		h.sm.Reset(userID)
		return nil

	case session.StateProtectAwaitPass:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "protected.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "🔒 PDF qulflanmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.pdfcpuEng.EncryptPDF(inFile, outFile, text)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Parol o'rnatishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, "🔒 Parol o'rnatilgan PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StateUnlockAwaitPass:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "unlocked.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "🔓 PDF dan parol olib tashlanmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.pdfcpuEng.DecryptPDF(inFile, outFile, text)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, "❌ Parol noto'g'ri yoki xatolik yuz berdi.", nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, "🔓 Parolsiz PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StateWatermarkAwaitText:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "watermarked.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "🏷 Watermark qo'shilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.pdfcpuEng.WatermarkPDF(inFile, outFile, text)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Watermark qo'shishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, "🏷 Watermark qo'shilgan PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StateOrganizeAwaitPages:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "organized.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "🗂 PDF sahifalari tartiblanmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		pages := engine.ParsePageRange(text)
		err := h.pdfcpuEng.OrganizePDF(inFile, outFile, pages)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Tartiblashda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, "🗂 Tartiblangan PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil

	case session.StateHTML2PDFAwaitInput:
		outFile := filepath.Join(sess.SessionDir, "webpage.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ Veb-sahifa PDF ga o'girilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.convertEng.HTMLToPDF(text, outFile)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ O'girishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, "🌐 Web-sahifa PDF tayyor!", sess.Language)
		h.sm.Reset(userID)
		return nil
	}

	return nil
}

// HandleCallbackActions handles non-tool callback queries (merge process, rotate angle, compress level, etc.)
func (h *BotHandlers) HandleCallbackActions(b *gotgbot.Bot, ctx *ext.Context, data string) error {
	userID := ctx.EffectiveUser.Id
	sess := h.sm.Get(userID)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
	}

	if data == "merge:process" {
		if len(sess.Files) < 2 {
			return nil
		}
		var paths []string
		for _, f := range sess.Files {
			paths = append(paths, f.Path)
		}
		outFile := filepath.Join(sess.SessionDir, "merged.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ PDF fayllar birlashtirilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.pdfcpuEng.MergePDFs(paths, outFile)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Birlashtirishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, fmt.Sprintf("✅ %d ta PDF birlashtirildi!", len(paths)), sess.Language)
		h.sm.Reset(userID)
		return nil

	} else if data == "jpg2pdf:process" {
		if len(sess.Files) < 1 {
			return nil
		}
		var paths []string
		for _, f := range sess.Files {
			paths = append(paths, f.Path)
		}
		outFile := filepath.Join(sess.SessionDir, "images_combined.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ Rasmlar PDF ga o'girilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.pdfcpuEng.ImagesToPDF(paths, outFile)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ PDF ga o'girishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, "✅ Rasmlar PDF ga o'girildi!", sess.Language)
		h.sm.Reset(userID)
		return nil

	} else if strings.HasPrefix(data, "compress:level:") {
		level := strings.TrimPrefix(data, "compress:level:")
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "compressed.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ PDF siqilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.convertEng.CompressPDF(inFile, outFile, level)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Siqishda xatolik: %v", err), nil)
			return err
		}

		inStat, _ := os.Stat(inFile)
		outStat, _ := os.Stat(outFile)
		cap := fmt.Sprintf("📉 **PDF Siqildi!**\nBoshlang'ich: `%s` -> Yangi: `%s`", engine.FormatBytes(inStat.Size()), engine.FormatBytes(outStat.Size()))
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, cap, sess.Language)
		h.sm.Reset(userID)
		return nil

	} else if strings.HasPrefix(data, "rotate:angle:") {
		angleStr := strings.TrimPrefix(data, "rotate:angle:")
		angle := engine.ParseInt(angleStr)
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "rotated.pdf")
		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, "⏳ PDF burilmoqda...", nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}
		err := h.pdfcpuEng.RotatePDF(inFile, outFile, angle)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Burishda xatolik: %v", err), nil)
			return err
		}
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, fmt.Sprintf("🔄 PDF %d° ga burildi!", angle), sess.Language)
		h.sm.Reset(userID)
		return nil
	}

	return nil
}

// Zip helper
func createZipArchive(files []string, zipPath string) error {
	newZipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		w, err := zipWriter.Create(filepath.Base(file))
		if err == nil {
			_, _ = io.Copy(w, f)
		}
		_ = f.Close()
	}
	return nil
}
