package keyboards

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func MonitorMainMenuKeyboard(lang string, isVIP bool) gotgbot.InlineKeyboardMarkup {
	vipBadge := "🆓 Bepul versiya"
	if isVIP {
		vipBadge = "👑 VIP Aktiv"
	}

	if lang == "ru" {
		return gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: "🎯 Мои фильтры", CallbackData: "monitor:my_filters"},
					{Text: "➕ Создать фильтр", CallbackData: "monitor:add_filter"},
				},
				{
					{Text: "⚡️ Горячие объявления (" + vipBadge + ")", CallbackData: "monitor:hot_feed"},
				},
				{
					{Text: "👑 VIP Подписка / Тарифы", CallbackData: "monitor:vip_plans"},
					{Text: "📊 Аналитика цен", CallbackData: "monitor:price_stats"},
				},
				{
					{Text: "⚙️ Настройки & Помощь", CallbackData: "monitor:settings"},
				},
			},
		}
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🎯 Filterlarim", CallbackData: "monitor:my_filters"},
				{Text: "➕ Yangi Filter Yaratish", CallbackData: "monitor:add_filter"},
			},
			{
				{Text: "⚡️ So'nggi Arzon E'lonlar (" + vipBadge + ")", CallbackData: "monitor:hot_feed"},
			},
			{
				{Text: "👑 VIP Obuna / Tariflar", CallbackData: "monitor:vip_plans"},
				{Text: "📊 Bozor Narxlari Statistikasi", CallbackData: "monitor:price_stats"},
			},
			{
				{Text: "⚙️ Sozlamalar & Qo'llanma", CallbackData: "monitor:settings"},
			},
		},
	}
}

func FilterMakeKeyboard() gotgbot.InlineKeyboardMarkup {
	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🚗 Chevrolet", CallbackData: "flt_make:Chevrolet"},
				{Text: "🚘 BYD", CallbackData: "flt_make:BYD"},
			},
			{
				{Text: "🏎 Kia", CallbackData: "flt_make:Kia"},
				{Text: "🚙 Hyundai", CallbackData: "flt_make:Hyundai"},
			},
			{
				{Text: "🌐 Barcha Markalar", CallbackData: "flt_make:All"},
			},
		},
	}
}

func FilterModelKeyboard(make string) gotgbot.InlineKeyboardMarkup {
	if make == "Chevrolet" {
		return gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{Text: "Gentra / Lacetti", CallbackData: "flt_model:Gentra"},
					{Text: "Cobalt", CallbackData: "flt_model:Cobalt"},
				},
				{
					{Text: "Damas / Labo", CallbackData: "flt_model:Damas"},
					{Text: "Spark", CallbackData: "flt_model:Spark"},
				},
				{
					{Text: "Tracker", CallbackData: "flt_model:Tracker"},
					{Text: "Malibu", CallbackData: "flt_model:Malibu"},
				},
				{
					{Text: "🌐 Barcha Modellar", CallbackData: "flt_model:All"},
				},
			},
		}
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "🌐 Barcha Modellar", CallbackData: "flt_model:All"},
			},
		},
	}
}

func SubscriptionKeyboard(userID int64, freeTrialUsed bool) gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{}

	if !freeTrialUsed {
		kb = append(kb, []gotgbot.InlineKeyboardButton{
			{Text: "🎁 24 Soat Bepul VIP Olib Ko'rish", CallbackData: "vip:claim_trial"},
		})
	}

	kb = append(kb, []gotgbot.InlineKeyboardButton{
		{Text: "⚡️ 1 Haftalik Starter (35,000 so'm)", CallbackData: "vip:buy_starter"},
	})
	kb = append(kb, []gotgbot.InlineKeyboardButton{
		{Text: "🚀 1 Oylik Pro (99,000 so'm)", CallbackData: "vip:buy_pro"},
	})
	kb = append(kb, []gotgbot.InlineKeyboardButton{
		{Text: "👑 3 Oylik VIP (249,000 so'm)", CallbackData: "vip:buy_vip"},
	})
	kb = append(kb, []gotgbot.InlineKeyboardButton{
		{Text: "⬅️ Asosiy Menyuga Qaytish", CallbackData: "monitor:main"},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}

func ListingActionKeyboard(listingURL string, phone string) gotgbot.InlineKeyboardMarkup {
	buttons := []gotgbot.InlineKeyboardButton{
		{Text: "🔗 OLX / Avto.uz da Ochish", Url: listingURL},
	}
	if phone != "" {
		buttons = append(buttons, gotgbot.InlineKeyboardButton{
			Text: "📞 Qo'ng'iroq Qilish", Url: fmt.Sprintf("tel:%s", phone),
		})
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{buttons},
	}
}
