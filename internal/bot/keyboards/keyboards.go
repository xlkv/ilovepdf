package keyboards

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

// MainMenuKeyboard builds the main menu with 20 PDF tools
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
		txtHTML2PDF  = "🌐 HTML to PDF"
		txtOCR       = "🔍 OCR PDF"
		txtLang      = "🌐 Language / Tillar"
	)

	if lang == "ru" {
		txtMerge = "🧩 Объединить PDF"
		txtSplit = "✂️ Разделить PDF"
		txtCompress = "📉 Сжать PDF"
		txtWord2PDF = "📝 Word в PDF"
		txtPPT2PDF = "📊 PPT в PDF"
		txtExcel2PDF = "📈 Excel в PDF"
		txtPDF2Word = "📄 PDF в Word"
		txtPDF2JPG = "🖼 PDF в JPG"
		txtJPG2PDF = "📷 JPG в PDF"
		txtRotate = "🔄 Повернуть PDF"
		txtProtect = "🔒 Защитить PDF"
		txtUnlock = "🔓 Снять пароль"
		txtWatermark = "🏷 Водяной знак"
		txtPageNum = "🔢 Нумерация"
		txtOrganize = "🗂 Организовать"
		txtHTML2PDF = "🌐 HTML в PDF"
		txtOCR = "🔍 OCR Распознавание"
	} else if lang == "uz" {
		txtMerge = "🧩 PDF Birlashtirish"
		txtSplit = "✂️ PDF Ajratish"
		txtCompress = "📉 PDF Siqish"
		txtWord2PDF = "📝 Word -> PDF"
		txtPPT2PDF = "📊 PPT -> PDF"
		txtExcel2PDF = "📈 Excel -> PDF"
		txtPDF2Word = "📄 PDF -> Word"
		txtPDF2JPG = "🖼 PDF -> JPG"
		txtJPG2PDF = "📷 JPG -> PDF"
		txtRotate = "🔄 PDF Burish"
		txtProtect = "🔒 Parol Qo'yish"
		txtUnlock = "🔓 Parolni O'chirish"
		txtWatermark = "🏷 Watermark"
		txtPageNum = "🔢 Raqamlash"
		txtOrganize = "🗂 Tartiblash"
		txtHTML2PDF = "🌐 Web -> PDF"
		txtOCR = "🔍 OCR Matn"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
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

// BackToMenuKeyboard returns button to jump back to main menu
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

// CancelKeyboard returns a single cancel button
func CancelKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "❌ Bekor qilish / Cancel", CallbackData: "action:cancel"},
			},
		},
	}
}

// MergeFilesKeyboard control keyboard during multi-file upload
func MergeFilesKeyboard(count int) gotgbot.InlineKeyboardMarkup {
	btnProcessText := fmt.Sprintf("🚀 Birlashtirish (%d ta fayl)", count)
	btnProcessData := "merge:process"
	if count < 2 {
		btnProcessText = "📥 Kamida 2 ta fayl yuboring"
		btnProcessData = "noop"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: btnProcessText, CallbackData: btnProcessData},
			},
			{
				{Text: "❌ Bekor qilish", CallbackData: "action:cancel"},
			},
		},
	}
}

// JPG2PDFKeyboard control keyboard for image converting
func JPG2PDFKeyboard(count int) gotgbot.InlineKeyboardMarkup {
	btnText := fmt.Sprintf("📷 PDF ga o'girish (%d rasm)", count)
	btnData := "jpg2pdf:process"
	if count < 1 {
		btnText = "📥 Rasm yuboring"
		btnData = "noop"
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: btnText, CallbackData: btnData},
			},
			{
				{Text: "❌ Bekor qilish", CallbackData: "action:cancel"},
			},
		},
	}
}

// CompressLevelKeyboard returns compression options
func CompressLevelKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "⚡ Yuqori siqish (Extreme)", CallbackData: "compress:level:extreme"},
			},
			{
				{Text: "⚖️ Tavsiya etilgan (Recommended)", CallbackData: "compress:level:recommended"},
			},
			{
				{Text: "🔍 Kam siqish (High Quality)", CallbackData: "compress:level:less"},
			},
			{
				{Text: "❌ Bekor qilish", CallbackData: "action:cancel"},
			},
		},
	}
}

// RotateKeyboard returns angle selection
func RotateKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "↪️ 90° O'ngga", CallbackData: "rotate:angle:90"},
				{Text: "↩️ 90° Chapga", CallbackData: "rotate:angle:270"},
				{Text: "🔄 180° Burish", CallbackData: "rotate:angle:180"},
			},
			{
				{Text: "❌ Bekor qilish", CallbackData: "action:cancel"},
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
				{Text: "🇬🇧 English", CallbackData: "lang:en"},
				{Text: "🇷🇺 Русский", CallbackData: "lang:ru"},
			},
			{
				{Text: "🔙 Bosh menyu", CallbackData: "nav:main"},
			},
		},
	}
}
