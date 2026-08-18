package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/xlkv/ilovepdf/internal/analytics"
	"github.com/xlkv/ilovepdf/internal/bot/keyboards"
	"github.com/xlkv/ilovepdf/internal/db"
	"github.com/xlkv/ilovepdf/internal/matcher"
	"github.com/xlkv/ilovepdf/internal/payments"
	"github.com/xlkv/ilovepdf/internal/scraper"
)

type MonitorHandlers struct {
	storage    *db.Storage
	evaluator  *analytics.PriceEvaluator
	olxScraper *scraper.OLXScraper
	matcher    *matcher.FilterMatcher
	payMgr     *payments.PaymentManager
}

func NewMonitorHandlers(storage *db.Storage, evaluator *analytics.PriceEvaluator, olx *scraper.OLXScraper, m *matcher.FilterMatcher, pm *payments.PaymentManager) *MonitorHandlers {
	return &MonitorHandlers{
		storage:    storage,
		evaluator:  evaluator,
		olxScraper: olx,
		matcher:    m,
		payMgr:     pm,
	}
}

func (mh *MonitorHandlers) HandleMonitorStart(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveUser
	user := mh.storage.GetUser(u.Id)
	if user == nil {
		user = &db.User{
			UserID:    u.Id,
			Username:  u.Username,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Language:  "uz",
		}
		mh.storage.UpsertUser(user)
	}

	welcomeText := fmt.Sprintf(`🚀 **Avto & E'lon Monitor Botiga Xush Kelibsiz, %s!**

⚡️ **Ushbu bot nimasi bilan afzal?**
• OLX va Avto.uz e'lonlarini har **5 soniyada** skanerlaydi.
• Bozor narxidan **10%% - 25%% arzon** tushgan e'lonlarni **1 soniya** ichida birinchi bo'lib sizga xabar qiladi!
• "Perepuk" va "Rieltor"lar uchun eng arzon va tekor takliflarni ilib olish vositasi.

🎁 **Bonus:** Yangi foydalanuvchilar uchun **24 Soatlik Bepul VIP** imkoniyat mavjud!`, u.FirstName)

	kb := keyboards.MonitorMainMenuKeyboard(user.Language, user.IsVIP())

	if ctx.CallbackQuery != nil {
		_, _ = ctx.CallbackQuery.Answer(b, nil)
		_, _, _ = b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        welcomeText,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return nil
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, welcomeText, &gotgbot.SendMessageOpts{
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return err
}

func (mh *MonitorHandlers) HandleMyFilters(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveUser
	filters := mh.storage.GetUserFilters(u.Id)

	if len(filters) == 0 {
		text := "🎯 **Sizda hali saqlangan filterlar yo'q.**\n\nYangi e'lonlar haqida birinchi bo'lib xabar olish uchun **➕ Yangi Filter Yaratish** tugmasini bosing!"
		kb := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "➕ Yangi Filter Yaratish", CallbackData: "monitor:add_filter"}},
				{{Text: "⬅️ Asosiy Menyu", CallbackData: "monitor:main"}},
			},
		}
		_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        text,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	text := fmt.Sprintf("🎯 **Sizning Faol Filterlaringiz (%d ta):**\n\n", len(filters))
	for i, f := range filters {
		text += fmt.Sprintf("%d. 🚗 **%s %s** (%d - %d yillar)\n   💰 Narx: $%d - $%d | 🔥 Min. arzonlik: %0.0f%%\n\n",
			i+1, f.Make, f.Model, f.MinYear, f.MaxYear, int(f.MinPrice), int(f.MaxPrice), f.BelowMarketPct)
	}

	kb := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{Text: "➕ Yangi Filter Qo'shish", CallbackData: "monitor:add_filter"}},
			{{Text: "🗑 Barcha Filterlarni O'chirish", CallbackData: "monitor:clear_filters"}},
			{{Text: "⬅️ Asosiy Menyu", CallbackData: "monitor:main"}},
		},
	}

	_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:      ctx.EffectiveChat.Id,
		MessageId:   ctx.EffectiveMessage.MessageId,
		Text:        text,
		ParseMode:   "Markdown",
		ReplyMarkup: kb,
	})
	return err
}

func (mh *MonitorHandlers) HandleAddFilterWizard(b *gotgbot.Bot, ctx *ext.Context, step string) error {
	if step == "make" {
		text := "🚗 **1-Bosqich:** Qaysi avtomobil markasi bo'yicha e'lonlarni kuzatmoqchisiz?"
		kb := keyboards.FilterMakeKeyboard()
		_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        text,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}
	return nil
}

func (mh *MonitorHandlers) HandleClaimFreeTrial(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveUser
	success := mh.storage.GrantFreeTrial(u.Id)

	var text string
	if success {
		text = "🎉 **Tabriklaymiz! Sizga 24 Soatlik Bepul VIP Tarif faollashtirildi!**\n\nBarcha tezkor arzon e'lonlar xabarnomalari 24 soat davomida avtomatik kelib turadi."
	} else {
		text = "⚠️ **Siz ilgari Bepul Trial VIP tarifidan foydalangansiz.**\n\nVIP imkoniyatni uzaytirish uchun tariflardan birini tanlang."
	}

	kb := keyboards.MonitorMainMenuKeyboard("uz", true)
	_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:      ctx.EffectiveChat.Id,
		MessageId:   ctx.EffectiveMessage.MessageId,
		Text:        text,
		ParseMode:   "Markdown",
		ReplyMarkup: &kb,
	})
	return err
}

func (mh *MonitorHandlers) HandleVIPPlans(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveUser
	user := mh.storage.GetUser(u.Id)
	freeTrialUsed := false
	if user != nil {
		freeTrialUsed = user.FreeTrialUsed
	}

	text := `👑 **iLovePDF Avto-E'lon VIP Tariflari**

⚡️ **VIP Tarifning Afzalliklari:**
• OLX va Avto.uz dagi arzon e'lonlarni **1 soniyada** olish.
• Bozor narxidan **10%-25% arzon** e'lonlar haqida aniq chegirma hisob-kitobi.
• Cheksiz filterlar o'rnatish imkoniyati.

Tarifni tanlang:`

	kb := keyboards.SubscriptionKeyboard(u.Id, freeTrialUsed)
	_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:      ctx.EffectiveChat.Id,
		MessageId:   ctx.EffectiveMessage.MessageId,
		Text:        text,
		ParseMode:   "Markdown",
		ReplyMarkup: &kb,
	})
	return err
}

func (mh *MonitorHandlers) HandleHotFeed(b *gotgbot.Bot, ctx *ext.Context) error {
	recent := mh.storage.GetRecentListings(5)

	if len(recent) == 0 {
		text := "⚡️ **Hozircha tizimda yangi arzon e'lonlar keshlanmagan.**\n\nSkreper har 5 soniyada yangi e'lonlarni tekshirib turadi."
		kb := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "⬅️ Asosiy Menyu", CallbackData: "monitor:main"}},
			},
		}
		_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        text,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	for _, l := range recent {
		discountBadge := ""
		if l.BelowMarketPct > 0 {
			discountBadge = fmt.Sprintf("🔥 **Bozor narxidan %0.1f%% arzon!**\n", l.BelowMarketPct)
		}

		msgText := fmt.Sprintf(`🚗 **%s**
%s
💰 **Narxi:** $%0.0f *(Bozor o'rtacha narxi: $%0.0f)*
📍 **Hudud:** %s
⏱ **Topildi:** Hozirgina

🔗 [OLX.uz da Ko'rish](%s)`, l.Title, discountBadge, l.PriceUSD, l.MarketAvgUSD, l.Region, l.URL)

		kb := keyboards.ListingActionKeyboard(l.URL, l.Phone)
		_, _ = b.SendMessage(ctx.EffectiveChat.Id, msgText, &gotgbot.SendMessageOpts{
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
	}

	return nil
}

// StartBackgroundScraper Worker periodically scrapes OLX/Avto and sends instant alerts
func (mh *MonitorHandlers) StartBackgroundScraper(b *gotgbot.Bot) {
	go func() {
		ticker := time.NewTicker(8 * time.Second)
		log.Println("🚀 Real-time Marketplace Scraper Engine Started (Interval: 8s)...")

		for range ticker.C {
			listings, err := mh.olxScraper.FetchLatestCars()
			if err != nil {
				log.Printf("[SCRAPER WARN] Error fetching listings: %v", err)
				continue
			}

			for _, listing := range listings {
				if mh.storage.IsListingSeen(listing.ExternalID) {
					continue
				}

				// Save listing to persistent storage
				mh.storage.SaveListing(listing)

				// Match listing against active user filters
				matchedUsers := mh.matcher.FindMatchingUsers(listing)

				for _, userID := range matchedUsers {
					discountBadge := ""
					if listing.BelowMarketPct > 0 {
						discountBadge = fmt.Sprintf("🔥 **Bozor narxidan %0.1f%% ARZON!**\n", listing.BelowMarketPct)
					}

					alertMsg := fmt.Sprintf(`🚨 **YANGI ARZON E'LON TOPILDI!**
%s
🚗 **%s**
💰 **Narxi:** $%0.0f *(O'rtacha bozor narxi: $%0.0f)*
📍 **Joylashuv:** %s
⏱ **Sana:** Hozirgina (Real-time)`, discountBadge, listing.Title, listing.PriceUSD, listing.MarketAvgUSD, listing.Region)

					kb := keyboards.ListingActionKeyboard(listing.URL, listing.Phone)
					_, sendErr := b.SendMessage(userID, alertMsg, &gotgbot.SendMessageOpts{
						ParseMode:   "Markdown",
						ReplyMarkup: kb,
					})
					if sendErr != nil {
						log.Printf("[ALERT ERROR] Failed to send alert to user %d: %v", userID, sendErr)
					}
				}
			}
		}
	}()
}

func (mh *MonitorHandlers) HandleCallbackQuery(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.CallbackQuery
	if cb == nil {
		return nil
	}

	data := cb.Data
	if data == "monitor:main" {
		return mh.HandleMonitorStart(b, ctx)
	} else if data == "monitor:my_filters" {
		return mh.HandleMyFilters(b, ctx)
	} else if data == "monitor:add_filter" {
		return mh.HandleAddFilterWizard(b, ctx, "make")
	} else if data == "vip:claim_trial" {
		return mh.HandleClaimFreeTrial(b, ctx)
	} else if data == "monitor:vip_plans" {
		return mh.HandleVIPPlans(b, ctx)
	} else if data == "monitor:hot_feed" {
		return mh.HandleHotFeed(b, ctx)
	} else if strings.HasPrefix(data, "flt_make:") {
		makeName := strings.TrimPrefix(data, "flt_make:")
		text := fmt.Sprintf("🚗 Marka: **%s**\n\n**2-Bosqich:** Qaysi modelni qidirmoqchisiz?", makeName)
		kb := keyboards.FilterModelKeyboard(makeName)
		_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        text,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	} else if strings.HasPrefix(data, "flt_model:") {
		modelName := strings.TrimPrefix(data, "flt_model:")
		u := ctx.EffectiveUser

		// Save new filter
		f := &db.UserFilter{
			UserID:         u.Id,
			Name:           "Chevrolet " + modelName,
			Category:       db.CategoryCar,
			Make:           "Chevrolet",
			Model:          modelName,
			MinYear:        2018,
			MaxYear:        2024,
			MinPrice:       3000,
			MaxPrice:       25000,
			BelowMarketPct: 5.0,
			Active:         true,
		}
		mh.storage.SaveFilter(f)

		text := fmt.Sprintf("✅ **Filter Saqlandi!**\n\n🚗 **%s %s** bo'yicha bozor narxidan arzon e'lon tushishi bilan birinchi bo'lib sizga xabar yuboriladi!", f.Make, f.Model)
		kb := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "🎯 Filterlarimni Ko'rish", CallbackData: "monitor:my_filters"}},
				{{Text: "⬅️ Asosiy Menyu", CallbackData: "monitor:main"}},
			},
		}
		_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        text,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	} else if data == "monitor:clear_filters" {
		u := ctx.EffectiveUser
		filters := mh.storage.GetUserFilters(u.Id)
		for _, f := range filters {
			mh.storage.DeleteFilter(f.ID)
		}
		return mh.HandleMyFilters(b, ctx)
	} else if strings.HasPrefix(data, "vip:buy_") {
		planType := strings.TrimPrefix(data, "vip:buy_")
		u := ctx.EffectiveUser
		amount := 99000.0
		if planType == "starter" {
			amount = 35000.0
		} else if planType == "vip" {
			amount = 249000.0
		}

		clickURL := mh.payMgr.GenerateClickURL(u.Id, db.SubscriptionPlan(planType), amount)

		text := fmt.Sprintf("💳 **VIP Obuna Uchun To'lov**\n\nTarif: **%s**\nSumma: **%s so'm**\n\nQuyidagi tugma orqali Click tizimida xavfsiz to'lovni amalga oshiring:", strings.ToUpper(planType), formatCurrency(amount))

		kb := gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{{Text: "💳 Click Orqali To'lash", Url: clickURL}},
				{{Text: "⬅️ Ortga", CallbackData: "monitor:vip_plans"}},
			},
		}
		_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveChat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			Text:        text,
			ParseMode:   "Markdown",
			ReplyMarkup: kb,
		})
		return err
	}

	return nil
}

func formatCurrency(val float64) string {
	str := strconv.FormatFloat(val, 'f', 0, 64)
	if len(str) <= 3 {
		return str
	}
	var res []string
	for i := len(str); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		res = append([]string{str[start:i]}, res...)
	}
	return strings.Join(res, " ")
}
