package db

import (
	"time"
)

type Category string

const (
	CategoryCar    Category = "car"
	CategoryRealty Category = "realty"
)

type User struct {
	UserID             int64     `json:"user_id"`
	Username           string    `json:"username"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	Language           string    `json:"language"` // "uz", "ru", "en"
	FreeTrialUsed      bool      `json:"free_trial_used"`
	SubscriptionExpire time.Time `json:"subscription_expire"`
	IsAdmin            bool      `json:"is_admin"`
	CreatedAt          time.Time `json:"created_at"`
	LastActiveAt       time.Time `json:"last_active_at"`
}

func (u *User) IsVIP() bool {
	return u.IsAdmin || time.Now().Before(u.SubscriptionExpire)
}

type UserFilter struct {
	ID             string   `json:"id"`
	UserID         int64    `json:"user_id"`
	Name           string   `json:"name"`
	Category       Category `json:"category"`
	Make           string   `json:"make"`             // e.g. "Chevrolet", "Hyundai"
	Model          string   `json:"model"`            // e.g. "Gentra", "Cobalt", "3-x xona"
	MinYear        int      `json:"min_year"`         // e.g. 2018
	MaxYear        int      `json:"max_year"`         // e.g. 2024
	MinPrice       float64  `json:"min_price"`        // e.g. 5000
	MaxPrice       float64  `json:"max_price"`        // e.g. 15000
	Region         string   `json:"region"`           // e.g. "Toshkent", "All"
	BelowMarketPct float64  `json:"below_market_pct"` // e.g. 10.0 (minimum 10% below market price)
	Active         bool     `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
}

type Listing struct {
	ID             string    `json:"id"`
	Source         string    `json:"source"` // "olx", "avto"
	ExternalID     string    `json:"external_id"`
	Title          string    `json:"title"`
	Category       Category  `json:"category"`
	Make           string    `json:"make"`
	Model          string    `json:"model"`
	PriceUSD       float64   `json:"price_usd"`
	PriceUZS       float64   `json:"price_uzs"`
	MarketAvgUSD   float64   `json:"market_avg_usd"`
	BelowMarketPct float64   `json:"below_market_pct"` // % discount
	Year           int       `json:"year"`
	MileageKM      int       `json:"mileage_km"`
	Region         string    `json:"region"`
	URL            string    `json:"url"`
	PhotoURL       string    `json:"photo_url"`
	Phone          string    `json:"phone"`
	Description    string    `json:"description"`
	PublishedAt    time.Time `json:"published_at"`
	CreatedAt      time.Time `json:"created_at"`
}

type SubscriptionPlan string

const (
	PlanTrial    SubscriptionPlan = "trial"    // 24 hours free
	PlanWeekly   SubscriptionPlan = "starter"  // 7 days
	PlanMonthly  SubscriptionPlan = "pro"      // 30 days
	PlanQuarterly SubscriptionPlan = "vip"     // 90 days
)

type Subscription struct {
	ID              string           `json:"id"`
	UserID          int64            `json:"user_id"`
	Plan            SubscriptionPlan `json:"plan"`
	AmountUZS       float64          `json:"amount_uzs"`
	PaymentProvider string           `json:"payment_provider"` // "click", "payme"
	TransactionID   string           `json:"transaction_id"`
	Status          string           `json:"status"` // "pending", "paid", "cancelled"
	ExpireAt        time.Time        `json:"expire_at"`
	CreatedAt       time.Time        `json:"created_at"`
}
