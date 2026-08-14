package keyboards

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// MainMenuKeyboard builds the clean minimalist main menu in 3 languages
func MainMenuKeyboard(lang string) gotgbot.InlineKeyboardMarkup {
	var (
		txtMerge     = "🧩 Merge PDF"
		txtSplit     = "✂️ Split PDF"
		txtCompress  = "📉 Compress PDF"
		txtWord2PDF  = "📝 Word to PDF"
		txtPPT2PDF   = "📊 PPT to PDF"
		txtExcel2PDF = "📈 Excel to PDF"
		txtPDF2Word  = "📄 PDF to Word"
		txtPDF2JPG   = "🖼 PDF to JPG"
		txtJPG2PDF   = "📷 JPG to PDF"
		txtRotate    = "🔄 Rotate PDF"
		txtProtect   = "🔒 Protect PDF"
		txtUnlock    = "🔓 Unlock PDF"
		txtWatermark = "🏷 Watermark"
		txtPageNum   = "🔢 Page Numbers"
		txtOrganize  = "🗂 Organize PDF"
		txtHTML2PDF  = "🌐 Web to PDF"
		txtOCR       = "🔍 OCR PDF"
		txtLang      = "🌐 Language"
		txtMiniApp   = "📱 Visual Editor"
	)

	if lang == "ru" {
		txtMerge = "🧩 Объединить"
		txtSplit = "✂️ Разделить"
		txtCompress = "📉 Сжать PDF"
		txtWord2PDF = "📝 Word в PDF"
		txtPPT2PDF = "📊 PPT в PDF"
		txtExcel2PDF = "📈 Excel в PDF"
		txtPDF2Word = "📄 PDF в Word"
		txtPDF2JPG = "🖼 PDF в JPG"
		txtJPG2PDF = "📷 JPG в PDF"
		txtRotate = "🔄 Повернуть"
		txtProtect = "🔒 Защитить"
		txtUnlock = "🔓 Снять пароль"
		txtWatermark = "🏷 Водяной знак"
		txtPageNum = "🔢 Нумерация"
		txtOrganize = "🗂 Организовать"
		txtHTML2PDF = "🌐 Web в PDF"
		txtOCR = "🔍 OCR Текст"
		txtLang = "🌐 Сменить язык"
		txtMiniApp = "📱 Визуальный редактор"
	} else if lang == "uz" {
		txtMerge = "🧩 Birlashtirish"
		txtSplit = "✂️ Ajratish"
		txtCompress = "📉 Siqish"
		txtWord2PDF = "📝 Word -> PDF"
		txtPPT2PDF = "📊 PPT -> PDF"
		txtExcel2PDF = "📈 Excel -> PDF"
		txtPDF2Word = "📄 PDF -> Word"
		txtPDF2JPG = "🖼 PDF -> JPG"
		txtJPG2PDF = "📷 JPG -> PDF"
		txtRotate = "🔄 Burish"
		txtProtect = "🔒 Parol Qo'yish"
		txtUnlock = "🔓 Parol O'chirish"
		txtWatermark = "🏷 Watermark"
		txtPageNum = "🔢 Raqamlash"
		txtOrganize = "🗂 Tartiblash"
		txtHTML2PDF = "🌐 Web -> PDF"
		txtOCR = "🔍 OCR Matn"
		txtLang = "🌐 Tilni O'zgartirish"
		txtMiniApp = "📱 Visual Editor"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: txtMiniApp, WebApp: &gotgbot.WebAppInfo{Url: "https://ilovepdf.xlkv.uz"}},
			},
			{
				{Text: txtMerge, CallbackData: "tool:merge"},
				{Text: txtSplit, CallbackData: "tool:split"},
			},
			{
				{Text: txtCompress, CallbackData: "tool:compress"},
				{Text: txtWord2PDF, CallbackData: "tool:word2pdf"},
			},
			{
				{Text: txtPPT2PDF, CallbackData: "tool:ppt2pdf"},
				{Text: txtExcel2PDF, CallbackData: "tool:excel2pdf"},
			},
			{
				{Text: txtPDF2Word, CallbackData: "tool:pdf2word"},
				{Text: txtPDF2JPG, CallbackData: "tool:pdf2jpg"},
			},
			{
				{Text: txtJPG2PDF, CallbackData: "tool:jpg2pdf"},
				{Text: txtRotate, CallbackData: "tool:rotate"},
			},
			{
				{Text: txtProtect, CallbackData: "tool:protect"},
				{Text: txtUnlock, CallbackData: "tool:unlock"},
			},
			{
				{Text: txtWatermark, CallbackData: "tool:watermark"},
				{Text: txtPageNum, CallbackData: "tool:pagenum"},
			},
			{
				{Text: txtOrganize, CallbackData: "tool:organize"},
				{Text: txtHTML2PDF, CallbackData: "tool:html2pdf"},
			},
			{
				{Text: txtOCR, CallbackData: "tool:ocr"},
				{Text: txtLang, CallbackData: "nav:lang"},
			},
		},
	}
}

// BackToMenuKeyboard returns clean button to jump back to main menu
func BackToMenuKeyboard(lang string) gotgbot.InlineKeyboardMarkup {
	text := "🔙 Asosiy Menyu"
	if lang == "ru" {
		text = "🔙 Главное меню"
	} else if lang == "en" {
		text = "🔙 Main Menu"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: text, CallbackData: "nav:main"},
			},
		},
	}
}

// CancelKeyboard returns a clean cancel button in 3 languages
func CancelKeyboard(lang string) gotgbot.InlineKeyboardMarkup {
	text := "❌ Bekor qilish"
	if lang == "ru" {
		text = "❌ Отмена"
	} else if lang == "en" {
		text = "❌ Cancel"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: text, CallbackData: "action:cancel"},
			},
		},
	}
}

// MergeFilesKeyboard control keyboard during multi-file upload in 3 languages
func MergeFilesKeyboard(count int, lang string) gotgbot.InlineKeyboardMarkup {
	btnProcessText := fmt.Sprintf("🚀 Birlashtirish (%d)", count)
	if lang == "ru" {
		btnProcessText = fmt.Sprintf("🚀 Объединить (%d)", count)
	} else if lang == "en" {
		btnProcessText = fmt.Sprintf("🚀 Merge (%d)", count)
	}

	btnProcessData := "merge:process"
	if count < 2 {
		btnProcessText = "📥 Kamida 2 ta fayl yuboring"
		if lang == "ru" {
			btnProcessText = "📥 Отправьте минимум 2 файла"
		} else if lang == "en" {
			btnProcessText = "📥 Upload at least 2 files"
		}
		btnProcessData = "noop"
	}

	btnCancelText := "❌ Bekor qilish"
	if lang == "ru" {
		btnCancelText = "❌ Отмена"
	} else if lang == "en" {
		btnCancelText = "❌ Cancel"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: btnProcessText, CallbackData: btnProcessData},
			},
			{
				{Text: btnCancelText, CallbackData: "action:cancel"},
			},
		},
	}
}

// JPG2PDFKeyboard control keyboard for image converting in 3 languages
func JPG2PDFKeyboard(count int, lang string) gotgbot.InlineKeyboardMarkup {
	btnText := fmt.Sprintf("📷 PDF ga o'girish (%d)", count)
	if lang == "ru" {
		btnText = fmt.Sprintf("📷 Конвертировать в PDF (%d)", count)
	} else if lang == "en" {
		btnText = fmt.Sprintf("📷 Convert to PDF (%d)", count)
	}

	btnData := "jpg2pdf:process"
	if count < 1 {
		btnText = "📥 Rasm yuboring"
		if lang == "ru" {
			btnText = "📥 Отправьте изображение"
		} else if lang == "en" {
			btnText = "📥 Upload an image"
		}
		btnData = "noop"
	}

	btnCancelText := "❌ Bekor qilish"
	if lang == "ru" {
		btnCancelText = "❌ Отмена"
	} else if lang == "en" {
		btnCancelText = "❌ Cancel"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: btnText, CallbackData: btnData},
			},
			{
				{Text: btnCancelText, CallbackData: "action:cancel"},
			},
		},
	}
}

// CompressLevelKeyboard returns clean compression options in 3 languages
func CompressLevelKeyboard(lang string) gotgbot.InlineKeyboardMarkup {
	tExtreme := "⚡ Yuqori siqish (Extreme)"
	tRec := "⚖️ Tavsiya etilgan (Recommended)"
	tLess := "🔍 Kam siqish (High Quality)"
	tCancel := "❌ Bekor qilish"

	if lang == "ru" {
		tExtreme = "⚡ Экстремальное сжатие"
		tRec = "⚖️ Рекомендуемое сжатие"
		tLess = "🔍 Высокое качество"
		tCancel = "❌ Отмена"
	} else if lang == "en" {
		tExtreme = "⚡ Extreme Compression"
		tRec = "⚖️ Recommended Compression"
		tLess = "🔍 High Quality (Less)"
		tCancel = "❌ Cancel"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: tExtreme, CallbackData: "compress:level:extreme"},
			},
			{
				{Text: tRec, CallbackData: "compress:level:recommended"},
			},
			{
				{Text: tLess, CallbackData: "compress:level:less"},
			},
			{
				{Text: tCancel, CallbackData: "action:cancel"},
			},
		},
	}
}

// RotateKeyboard returns angle selection in 3 languages
func RotateKeyboard(lang string) gotgbot.InlineKeyboardMarkup {
	tRight := "↪️ 90° O'ngga"
	tLeft := "↩️ 90° Chapga"
	t180 := "🔄 180° Burish"
	tCancel := "❌ Bekor qilish"

	if lang == "ru" {
		tRight = "↪️ 90° Вправо"
		tLeft = "↩️ 90° Влево"
		t180 = "🔄 180° Повернуть"
		tCancel = "❌ Отмена"
	} else if lang == "en" {
		tRight = "↪️ 90° Right"
		tLeft = "↩️ 90° Left"
		t180 = "🔄 180° Rotate"
		tCancel = "❌ Cancel"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: tRight, CallbackData: "rotate:angle:90"},
				{Text: tLeft, CallbackData: "rotate:angle:270"},
				{Text: t180, CallbackData: "rotate:angle:180"},
			},
			{
				{Text: tCancel, CallbackData: "action:cancel"},
			},
		},
	}
}

// LanguageKeyboard selection menu
func LanguageKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🇺🇿 O'zbekcha", CallbackData: "lang:uz"},
				{Text: "🇷🇺 Русский", CallbackData: "lang:ru"},
				{Text: "🇬🇧 English", CallbackData: "lang:en"},
			},
		},
	}
}
