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
	u := ctx.EffectiveUser
	userID := u.Id
	h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)
	h.sm.Reset(userID)

	username := ""
	if u.Username != "" {
		username = "@" + u.Username
	}

	sess := h.sm.Get(userID)
	lang := sess.Language
	h.sm.TrackUserFull(userID, username, u.FirstName, u.LastName, lang)
	h.sm.TrackToolUsageFull(userID, toolID)

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
	}

	var prompt string
	var newState session.State

	switch toolID {
	case "merge":
		newState = session.StateMergeUploading
		prompt = "🧩 **PDF Birlashtirish:**\n\nBirlashtirmoqchi bo'lgan 2 yoki undan ortiq fayllaringizni yuboring."
		if lang == "ru" {
			prompt = "🧩 **Объединение PDF:**\n\nОтправьте 2 или более файла для объединения."
		} else if lang == "en" {
			prompt = "🧩 **Merge PDF:**\n\nUpload 2 or more PDF files you wish to merge."
		}

	case "split":
		newState = session.StateSplitAwaitFile
		prompt = "✂️ **PDF Ajratish:**\n\nAjratmoqchi bo'lgan PDF faylingizni yuboring."
		if lang == "ru" {
			prompt = "✂️ **Разделение PDF:**\n\nОтправьте PDF файл для разделения."
		} else if lang == "en" {
			prompt = "✂️ **Split PDF:**\n\nUpload the PDF file you wish to split."
		}

	case "compress":
		newState = session.StateCompressAwaitFile
		prompt = "📉 **PDF Siqish:**\n\nHajmini kichraytirmoqchi bo'lgan PDF faylni yuboring."
		if lang == "ru" {
			prompt = "📉 **Сжатие PDF:**\n\nОтправьте PDF файл для сжатия."
		} else if lang == "en" {
			prompt = "📉 **Compress PDF:**\n\nUpload the PDF file you wish to compress."
		}

	case "word2pdf":
		newState = session.StateWord2PDFAwaitFile
		prompt = "📝 **Word -> PDF:**\n\nWord hujjatini (`.docx` yoki `.doc`) yuboring."
		if lang == "ru" {
			prompt = "📝 **Word в PDF:**\n\nОтправьте документ Word (`.docx` или `.doc`)."
		} else if lang == "en" {
			prompt = "📝 **Word to PDF:**\n\nUpload your Word document (`.docx` or `.doc`)."
		}

	case "ppt2pdf":
		newState = session.StatePPT2PDFAwaitFile
		prompt = "📊 **PowerPoint -> PDF:**\n\nPrezentatsiyani (`.pptx` yoki `.ppt`) yuboring."
		if lang == "ru" {
			prompt = "📊 **PowerPoint в PDF:**\n\nОтправьте презентацию (`.pptx` или `.ppt`)."
		} else if lang == "en" {
			prompt = "📊 **PowerPoint to PDF:**\n\nUpload your presentation (`.pptx` or `.ppt`)."
		}

	case "excel2pdf":
		newState = session.StateExcel2PDFAwaitFile
		prompt = "📈 **Excel -> PDF:**\n\nExcel jadvalini (`.xlsx` yoki `.xls`) yuboring."
		if lang == "ru" {
			prompt = "📈 **Excel в PDF:**\n\nОтправьте таблицу Excel (`.xlsx` или `.xls`)."
		} else if lang == "en" {
			prompt = "📈 **Excel to PDF:**\n\nUpload your Excel spreadsheet (`.xlsx` or `.xls`)."
		}

	case "pdf2word":
		newState = session.StatePDF2WordAwaitFile
		prompt = "📄 **PDF -> Word:**\n\nWord ga o'girmoqchi bo'lgan PDF faylni yuboring."
		if lang == "ru" {
			prompt = "📄 **PDF в Word:**\n\nОтправьте PDF файл для конвертации в Word."
		} else if lang == "en" {
			prompt = "📄 **PDF to Word:**\n\nUpload the PDF file you wish to convert to Word."
		}

	case "pdf2jpg":
		newState = session.StatePDF2JPGAwaitFile
		prompt = "🖼 **PDF -> JPG:**\n\nRasmlarga ajratmoqchi bo'lgan PDF faylni yuboring."
		if lang == "ru" {
			prompt = "🖼 **PDF в JPG:**\n\nОтправьте PDF файл для извлечения изображений."
		} else if lang == "en" {
			prompt = "🖼 **PDF to JPG:**\n\nUpload the PDF file you wish to convert to images."
		}

	case "jpg2pdf":
		newState = session.StateJPG2PDFUploading
		prompt = "📷 **JPG -> PDF:**\n\nPDF qilmoqchi bo'lgan rasmlaringizni yuboring."
		if lang == "ru" {
			prompt = "📷 **JPG в PDF:**\n\nОтправьте изображения для конвертации в PDF."
		} else if lang == "en" {
			prompt = "📷 **JPG to PDF:**\n\nUpload the images you wish to convert to PDF."
		}

	case "rotate":
		newState = session.StateRotateAwaitFile
		prompt = "🔄 **PDF Burish:**\n\nBurchagini o'zgartirmoqchi bo'lgan PDF faylingizni yuboring."
		if lang == "ru" {
			prompt = "🔄 **Поворот PDF:**\n\nОтправьте PDF файл для поворота."
		} else if lang == "en" {
			prompt = "🔄 **Rotate PDF:**\n\nUpload the PDF file you wish to rotate."
		}

	case "protect":
		newState = session.StateProtectAwaitFile
		prompt = "🔒 **PDF Parol Qo'yish:**\n\nParol o'rnatmoqchi bo'lgan PDF faylingizni yuboring."
		if lang == "ru" {
			prompt = "🔒 **Защита PDF:**\n\nОтправьте PDF файл для установки пароля."
		} else if lang == "en" {
			prompt = "🔒 **Protect PDF:**\n\nUpload the PDF file you wish to password protect."
		}

	case "unlock":
		newState = session.StateUnlockAwaitFile
		prompt = "🔓 **PDF Parolni O'chirish:**\n\nParolini olib tashlamoqchi bo'lgan PDF faylingizni yuboring."
		if lang == "ru" {
			prompt = "🔓 **Снятие пароля PDF:**\n\nОтправьте защищенный PDF файл."
		} else if lang == "en" {
			prompt = "🔓 **Unlock PDF:**\n\nUpload the protected PDF file."
		}

	case "watermark":
		newState = session.StateWatermarkAwaitFile
		prompt = "🏷 **Watermark Qo'shish:**\n\nWatermark qo'shmoqchi bo'lgan PDF faylingizni yuboring."
		if lang == "ru" {
			prompt = "🏷 **Водяной знак:**\n\nОтправьте PDF файл для добавления водяного знака."
		} else if lang == "en" {
			prompt = "🏷 **Add Watermark:**\n\nUpload the PDF file you wish to watermark."
		}

	case "pagenum":
		newState = session.StatePagenumAwaitFile
		prompt = "🔢 **Sahifa Raqamlari:**\n\nSahifa raqamlarini qo'shmoqchi bo'lgan PDF faylni yuboring."
		if lang == "ru" {
			prompt = "🔢 **Нумерация страниц:**\n\nОтправьте PDF файл для нумерации страниц."
		} else if lang == "en" {
			prompt = "🔢 **Page Numbers:**\n\nUpload the PDF file to add page numbers."
		}

	case "organize":
		newState = session.StateOrganizeAwaitFile
		prompt = "🗂 **PDF Tartiblash:**\n\nSahifalar ketma-ketligini o'zgartirmoqchi bo'lgan PDF faylni yuboring."
		if lang == "ru" {
			prompt = "🗂 **Организация PDF:**\n\nОтправьте PDF файл для изменения порядка страниц."
		} else if lang == "en" {
			prompt = "🗂 **Organize PDF:**\n\nUpload the PDF file to reorder pages."
		}

	case "html2pdf":
		newState = session.StateHTML2PDFAwaitInput
		prompt = "🌐 **Web -> PDF:**\n\nSayt manzilini (masalan: `https://example.com`) yoki `.html` faylini yuboring."
		if lang == "ru" {
			prompt = "🌐 **Web в PDF:**\n\nОтправьте URL сайта (`https://example.com`) или файл `.html`."
		} else if lang == "en" {
			prompt = "🌐 **Web to PDF:**\n\nSend a webpage URL (`https://example.com`) or an `.html` file."
		}

	case "ocr":
		newState = session.StateOCRAwaitFile
		prompt = "🔍 **OCR Matnni Tanish:**\n\nScan qilingan PDF faylingizni yuboring."
		if lang == "ru" {
			prompt = "🔍 **OCR Распознавание:**\n\nОтправьте отсканированный PDF файл."
		} else if lang == "en" {
			prompt = "🔍 **OCR Text Recognition:**\n\nUpload your scanned PDF document."
		}

	default:
		return nil
	}

	h.sm.SetState(userID, newState)
	kb := keyboards.CancelKeyboard(lang)

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
	lang := sess.Language

	if sess.State == session.StateIdle {
		kb := keyboards.MainMenuKeyboard(lang)
		promptMsg := "ℹ️ Iltimos, avval menyudan kerakli funksiyani tanlang:"
		if lang == "ru" {
			promptMsg = "ℹ️ Пожалуйста, сначала выберите нужную функцию из меню:"
		} else if lang == "en" {
			promptMsg = "ℹ️ Please select a tool from the main menu first:"
		}

		msg, err := b.SendMessage(ctx.EffectiveChat.Id, promptMsg, &gotgbot.SendMessageOpts{
			ReplyMarkup: kb,
		})
		if msg != nil {
			h.sm.AddTempMsg(userID, msg.MessageId)
		}
		return err
	}

	destPath := filepath.Join(sess.SessionDir, fmt.Sprintf("%d_%s", len(sess.Files)+1, doc.FileName))
	waitMsg := "⏳ Fayl yuklab olinmoqda..."
	if lang == "ru" {
		waitMsg = "⏳ Скачивание файла..."
	} else if lang == "en" {
		waitMsg = "⏳ Downloading file..."
	}

	statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, waitMsg, nil)
	if statusMsg != nil {
		h.sm.AddTempMsg(userID, statusMsg.MessageId)
	}

	err := h.DownloadTelegramFile(b, doc.FileId, destPath)
	if err != nil {
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, fmt.Sprintf("❌ Error: %v", err), nil)
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
		if lang == "ru" {
			msgText = fmt.Sprintf("📥 **Полученные файлы (%d):**\n", len(files))
		} else if lang == "en" {
			msgText = fmt.Sprintf("📥 **Received files (%d):**\n", len(files))
		}

		for i, f := range files {
			msgText += fmt.Sprintf("%d. 📄 `%s` (%s)\n", i+1, f.Name, engine.FormatBytes(f.Size))
		}
		kb := keyboards.MergeFilesKeyboard(len(files), lang)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, msgText, &kb)
		return err

	case session.StateJPG2PDFUploading:
		files := h.sm.Get(userID).Files
		msgText := fmt.Sprintf("📷 **Qabul qilingan rasmlar (%d):**\n", len(files))
		if lang == "ru" {
			msgText = fmt.Sprintf("📷 **Полученные изображения (%d):**\n", len(files))
		} else if lang == "en" {
			msgText = fmt.Sprintf("📷 **Received images (%d):**\n", len(files))
		}

		for i, f := range files {
			msgText += fmt.Sprintf("%d. 🖼 `%s`\n", i+1, f.Name)
		}
		kb := keyboards.JPG2PDFKeyboard(len(files), lang)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, msgText, &kb)
		return err

	case session.StateSplitAwaitFile:
		h.sm.SetState(userID, session.StateSplitAwaitRange)
		pageCount, _ := h.pdfcpuEng.GetPageCount(destPath)
		h.sm.SetMeta(userID, "page_count", fmt.Sprintf("%d", pageCount))

		prompt := fmt.Sprintf("📄 `%s` (%d sahifa)\n\nAjratmoqchi bo'lgan sahifalaringizni kiriting (masalan: `1-3, 5`):", doc.FileName, pageCount)
		if lang == "ru" {
			prompt = fmt.Sprintf("📄 `%s` (%d страниц)\n\nУкажите страницы для разделения (например: `1-3, 5`):", doc.FileName, pageCount)
		} else if lang == "en" {
			prompt = fmt.Sprintf("📄 `%s` (%d pages)\n\nEnter the page range to split (e.g.: `1-3, 5`):", doc.FileName, pageCount)
		}

		kb := keyboards.CancelKeyboard(lang)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateCompressAwaitFile:
		h.sm.SetState(userID, session.StateCompressAwaitLevel)
		prompt := fmt.Sprintf("📉 `%s` (%s)\n\nSiqish darajasini tanlang:", doc.FileName, engine.FormatBytes(doc.FileSize))
		if lang == "ru" {
			prompt = fmt.Sprintf("📉 `%s` (%s)\n\nВыберите уровень сжатия:", doc.FileName, engine.FormatBytes(doc.FileSize))
		} else if lang == "en" {
			prompt = fmt.Sprintf("📉 `%s` (%s)\n\nSelect compression level:", doc.FileName, engine.FormatBytes(doc.FileSize))
		}

		kb := keyboards.CompressLevelKeyboard(lang)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateWord2PDFAwaitFile, session.StatePPT2PDFAwaitFile, session.StateExcel2PDFAwaitFile:
		procMsg := "⏳ Hujjat PDF ga o'girilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Конвертация документа в PDF..."
		} else if lang == "en" {
			procMsg = "⏳ Converting document to PDF..."
		}
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, procMsg, nil)

		outPDF, err := h.convertEng.ConvertOfficeToPDF(destPath, sess.SessionDir)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "✅ PDF tayyor!"
		if lang == "ru" {
			succMsg = "✅ PDF готов!"
		} else if lang == "en" {
			succMsg = "✅ PDF is ready!"
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outPDF, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StatePDF2WordAwaitFile:
		procMsg := "⏳ PDF Word (DOCX) ga o'girilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Конвертация PDF в Word..."
		} else if lang == "en" {
			procMsg = "⏳ Converting PDF to Word..."
		}
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, procMsg, nil)

		outDocx := filepath.Join(sess.SessionDir, strings.TrimSuffix(doc.FileName, ".pdf")+".docx")
		err := h.convertEng.ConvertPDFToWord(destPath, outDocx)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "✅ Word hujjati tayyor!"
		if lang == "ru" {
			succMsg = "✅ Документ Word готов!"
		} else if lang == "en" {
			succMsg = "✅ Word document is ready!"
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outDocx, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StatePDF2JPGAwaitFile:
		procMsg := "⏳ PDF rasmlarga ajratilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Извлечение изображений из PDF..."
		} else if lang == "en" {
			procMsg = "⏳ Extracting images from PDF..."
		}
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, procMsg, nil)

		imgDir := filepath.Join(sess.SessionDir, "images")
		_ = os.MkdirAll(imgDir, 0755)
		imgs, err := h.convertEng.ConvertPDFToImages(destPath, imgDir)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil || len(imgs) == 0 {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, "❌ Error extracting images.", nil)
			return err
		}

		zipPath := filepath.Join(sess.SessionDir, "images.zip")
		_ = createZipArchive(imgs, zipPath)

		succMsg := fmt.Sprintf("✅ %d ta rasm (ZIP arxiv)!", len(imgs))
		if lang == "ru" {
			succMsg = fmt.Sprintf("✅ %d изображений (ZIP архив)!", len(imgs))
		} else if lang == "en" {
			succMsg = fmt.Sprintf("✅ %d images (ZIP archive)!", len(imgs))
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, zipPath, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StateRotateAwaitFile:
		h.sm.SetState(userID, session.StateRotateAwaitAngle)
		kb := keyboards.RotateKeyboard(lang)
		prompt := "🔄 Burish burchagini tanlang:"
		if lang == "ru" {
			prompt = "🔄 Выберите угол поворота:"
		} else if lang == "en" {
			prompt = "🔄 Select rotation angle:"
		}
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateProtectAwaitFile:
		h.sm.SetState(userID, session.StateProtectAwaitPass)
		kb := keyboards.CancelKeyboard(lang)
		prompt := "🔒 PDF uchun parol kiriting:"
		if lang == "ru" {
			prompt = "🔒 Введите пароль для PDF:"
		} else if lang == "en" {
			prompt = "🔒 Enter a password for the PDF:"
		}
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateUnlockAwaitFile:
		h.sm.SetState(userID, session.StateUnlockAwaitPass)
		kb := keyboards.CancelKeyboard(lang)
		prompt := "🔓 PDF ning parolini kiriting:"
		if lang == "ru" {
			prompt = "🔓 Введите пароль от PDF:"
		} else if lang == "en" {
			prompt = "🔓 Enter the password of the PDF:"
		}
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateWatermarkAwaitFile:
		h.sm.SetState(userID, session.StateWatermarkAwaitText)
		kb := keyboards.CancelKeyboard(lang)
		prompt := "🏷 Watermark matnini kiriting (masalan: `CONFIDENTIAL`):"
		if lang == "ru" {
			prompt = "🏷 Введите текст водяного знака (например: `CONFIDENTIAL`):"
		} else if lang == "en" {
			prompt = "🏷 Enter watermark text (e.g.: `CONFIDENTIAL`):"
		}
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StatePagenumAwaitFile:
		procMsg := "⏳ Sahifa raqamlari qo'shilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Добавление нумерации страниц..."
		} else if lang == "en" {
			procMsg = "⏳ Adding page numbers..."
		}
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, procMsg, nil)

		outPDF := filepath.Join(sess.SessionDir, "numbered.pdf")
		err := h.pdfcpuEng.AddPageNumbers(destPath, outPDF)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "✅ Sahifa raqamlari qo'shildi!"
		if lang == "ru" {
			succMsg = "✅ Нумерация страниц добавлена!"
		} else if lang == "en" {
			succMsg = "✅ Page numbers added!"
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outPDF, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StateOrganizeAwaitFile:
		h.sm.SetState(userID, session.StateOrganizeAwaitPages)
		pageCount, _ := h.pdfcpuEng.GetPageCount(destPath)
		prompt := fmt.Sprintf("🗂 **Sahifalar:** %d\n\nYangi tartibni kiriting (masalan: `3, 1, 2`):", pageCount)
		if lang == "ru" {
			prompt = fmt.Sprintf("🗂 **Страниц:** %d\n\nУкажите новый порядок (например: `3, 1, 2`):", pageCount)
		} else if lang == "en" {
			prompt = fmt.Sprintf("🗂 **Pages:** %d\n\nEnter new order (e.g.: `3, 1, 2`):", pageCount)
		}

		kb := keyboards.CancelKeyboard(lang)
		_, err := h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, prompt, &kb)
		return err

	case session.StateOCRAwaitFile:
		procMsg := "🔍 OCR matn tanilmoqda..."
		if lang == "ru" {
			procMsg = "🔍 Распознавание текста OCR..."
		} else if lang == "en" {
			procMsg = "🔍 Processing OCR text recognition..."
		}
		_, _ = h.EditMessage(b, ctx.EffectiveChat.Id, statusMsg.MessageId, procMsg, nil)

		outPDF := filepath.Join(sess.SessionDir, "ocr_searchable.pdf")
		err := h.convertEng.OCRPDF(destPath, outPDF, "eng")
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ OCR error: %v", err), nil)
			return err
		}

		succMsg := "✅ OCR PDF tayyor!"
		if lang == "ru" {
			succMsg = "✅ OCR PDF готов!"
		} else if lang == "en" {
			succMsg = "✅ OCR PDF is ready!"
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outPDF, succMsg, lang)
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
	lang := sess.Language

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
	msgText := fmt.Sprintf("📷 **Qabul qilingan rasmlar (%d):**", len(files))
	if lang == "ru" {
		msgText = fmt.Sprintf("📷 **Полученные изображения (%d):**", len(files))
	} else if lang == "en" {
		msgText = fmt.Sprintf("📷 **Received images (%d):**", len(files))
	}

	kb := keyboards.JPG2PDFKeyboard(len(files), lang)
	msg, err := b.SendMessage(ctx.EffectiveChat.Id, msgText, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	if msg != nil {
		h.sm.AddTempMsg(userID, msg.MessageId)
	}
	return err
}

// HandleTextMessage handles text inputs
func (h *BotHandlers) HandleTextMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	text := ctx.EffectiveMessage.Text
	if text == "" || strings.HasPrefix(text, "/") {
		return nil
	}

	userID := ctx.EffectiveUser.Id
	h.sm.AddTempMsg(userID, ctx.EffectiveMessage.MessageId)
	sess := h.sm.Get(userID)
	lang := sess.Language

	switch sess.State {
	case session.StateSplitAwaitRange:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outDir := filepath.Join(sess.SessionDir, "split_output")
		_ = os.MkdirAll(outDir, 0755)

		procMsg := "⏳ Sahifalar ajratilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Разделение страниц..."
		} else if lang == "en" {
			procMsg = "⏳ Splitting pages..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		pages := engine.ParsePageRange(text)
		outFiles, err := h.pdfcpuEng.SplitPDF(inFile, outDir, pages)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil || len(outFiles) == 0 {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "✅ Ajratilgan PDF tayyor!"
		if lang == "ru" {
			succMsg = "✅ Разделенный PDF готов!"
		} else if lang == "en" {
			succMsg = "✅ Split PDF is ready!"
		}

		if len(outFiles) == 1 {
			_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFiles[0], succMsg, lang)
		} else {
			zipPath := filepath.Join(sess.SessionDir, "split_pages.zip")
			_ = createZipArchive(outFiles, zipPath)
			_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, zipPath, succMsg+" (ZIP)", lang)
		}
		h.sm.Reset(userID)
		return nil

	case session.StateProtectAwaitPass:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "protected.pdf")

		procMsg := "🔒 PDF qulflanmoqda..."
		if lang == "ru" {
			procMsg = "🔒 Защита PDF..."
		} else if lang == "en" {
			procMsg = "🔒 Protecting PDF..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.pdfcpuEng.EncryptPDF(inFile, outFile, text)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "🔒 Protected PDF is ready!"
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StateUnlockAwaitPass:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "unlocked.pdf")

		procMsg := "🔓 PDF ochilmoqda..."
		if lang == "ru" {
			procMsg = "🔓 Снятие пароля PDF..."
		} else if lang == "en" {
			procMsg = "🔓 Unlocking PDF..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.pdfcpuEng.DecryptPDF(inFile, outFile, text)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, "❌ Incorrect password or error.", nil)
			return err
		}

		succMsg := "🔓 Unlocked PDF is ready!"
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StateWatermarkAwaitText:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "watermarked.pdf")

		procMsg := "🏷 Watermark qo'shilmoqda..."
		if lang == "ru" {
			procMsg = "🏷 Добавление водяного знака..."
		} else if lang == "en" {
			procMsg = "🏷 Adding watermark..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.pdfcpuEng.WatermarkPDF(inFile, outFile, text)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "🏷 Watermarked PDF is ready!"
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StateOrganizeAwaitPages:
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "organized.pdf")

		procMsg := "🗂 PDF sahifalari tartiblanmoqda..."
		if lang == "ru" {
			procMsg = "🗂 Организация страниц PDF..."
		} else if lang == "en" {
			procMsg = "🗂 Organizing PDF pages..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		pages := engine.ParsePageRange(text)
		err := h.pdfcpuEng.OrganizePDF(inFile, outFile, pages)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, err = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "🗂 Organized PDF is ready!"
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	case session.StateHTML2PDFAwaitInput:
		outFile := filepath.Join(sess.SessionDir, "webpage.pdf")

		procMsg := "⏳ Web-sahifa PDF ga o'girilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Конвертация Web в PDF..."
		} else if lang == "en" {
			procMsg = "⏳ Converting Web to PDF..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.convertEng.HTMLToPDF(text, outFile)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "🌐 Web PDF is ready!"
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
		h.sm.Reset(userID)
		return nil
	}

	return nil
}

// HandleCallbackActions handles non-tool callback queries
func (h *BotHandlers) HandleCallbackActions(b *gotgbot.Bot, ctx *ext.Context, data string) error {
	userID := ctx.EffectiveUser.Id
	sess := h.sm.Get(userID)
	lang := sess.Language

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

		procMsg := "⏳ PDF fayllar birlashtirilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Объединение PDF файлов..."
		} else if lang == "en" {
			procMsg = "⏳ Merging PDF files..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.pdfcpuEng.MergePDFs(paths, outFile)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := fmt.Sprintf("✅ %d ta PDF birlashtirildi!", len(paths))
		if lang == "ru" {
			succMsg = fmt.Sprintf("✅ %d PDF файлов объединены!", len(paths))
		} else if lang == "en" {
			succMsg = fmt.Sprintf("✅ %d PDF files merged!", len(paths))
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
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

		procMsg := "⏳ Rasmlar PDF ga o'girilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Конвертация изображений в PDF..."
		} else if lang == "en" {
			procMsg = "⏳ Converting images to PDF..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.pdfcpuEng.ImagesToPDF(paths, outFile)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := "✅ PDF ready!"
		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
		h.sm.Reset(userID)
		return nil

	} else if strings.HasPrefix(data, "compress:level:") {
		level := strings.TrimPrefix(data, "compress:level:")
		if len(sess.Files) == 0 {
			return nil
		}
		inFile := sess.Files[0].Path
		outFile := filepath.Join(sess.SessionDir, "compressed.pdf")

		procMsg := "⏳ PDF siqilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Сжатие PDF..."
		} else if lang == "en" {
			procMsg = "⏳ Compressing PDF..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.convertEng.CompressPDF(inFile, outFile, level)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		inStat, _ := os.Stat(inFile)
		outStat, _ := os.Stat(outFile)
		cap := fmt.Sprintf("📉 **PDF Siqildi!**\nBoshlang'ich: `%s` -> Yangi: `%s`", engine.FormatBytes(inStat.Size()), engine.FormatBytes(outStat.Size()))
		if lang == "ru" {
			cap = fmt.Sprintf("📉 **PDF Сжат!**\nБыло: `%s` -> Стало: `%s`", engine.FormatBytes(inStat.Size()), engine.FormatBytes(outStat.Size()))
		} else if lang == "en" {
			cap = fmt.Sprintf("📉 **PDF Compressed!**\nOriginal: `%s` -> New: `%s`", engine.FormatBytes(inStat.Size()), engine.FormatBytes(outStat.Size()))
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, cap, lang)
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

		procMsg := "⏳ PDF burilmoqda..."
		if lang == "ru" {
			procMsg = "⏳ Поворот PDF..."
		} else if lang == "en" {
			procMsg = "⏳ Rotating PDF..."
		}

		statusMsg, _ := b.SendMessage(ctx.EffectiveChat.Id, procMsg, nil)
		if statusMsg != nil {
			h.sm.AddTempMsg(userID, statusMsg.MessageId)
		}

		err := h.pdfcpuEng.RotatePDF(inFile, outFile, angle)
		h.ClearOldMessages(b, ctx.EffectiveChat.Id, userID)

		if err != nil {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, fmt.Sprintf("❌ Error: %v", err), nil)
			return err
		}

		succMsg := fmt.Sprintf("🔄 PDF %d° ga burildi!", angle)
		if lang == "ru" {
			succMsg = fmt.Sprintf("🔄 PDF повернут на %d°!", angle)
		} else if lang == "en" {
			succMsg = fmt.Sprintf("🔄 PDF rotated %d°!", angle)
		}

		_ = h.SendDocumentResponse(b, ctx.EffectiveChat.Id, outFile, succMsg, lang)
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
