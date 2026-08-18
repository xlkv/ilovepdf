package payments

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xlkv/ilovepdf/internal/db"
)

type PaymentManager struct {
	storage     *db.Storage
	clickSecret string
	clickServID string
	clickMerchID string
}

func NewPaymentManager(storage *db.Storage, secret, serviceID, merchantID string) *PaymentManager {
	return &PaymentManager{
		storage:      storage,
		clickSecret:  secret,
		clickServID:  serviceID,
		clickMerchID: merchantID,
	}
}

// GenerateClickURL generates payment URL for Click Uzbek Payment Gateway
func (pm *PaymentManager) GenerateClickURL(userID int64, plan db.SubscriptionPlan, amountUZS float64) string {
	return fmt.Sprintf("https://my.click.uz/services/pay?service_id=%s&merchant_id=%s&amount=%.2f&transaction_param=%d&additional_param=%s",
		pm.clickServID, pm.clickMerchID, amountUZS, userID, plan)
}

// HandleClickWebhook handles incoming Click webhook requests
func (pm *PaymentManager) HandleClickWebhook(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	clickTransID := r.FormValue("click_trans_id")
	serviceID := r.FormValue("service_id")
	merchantTransID := r.FormValue("merchant_trans_id")
	amount := r.FormValue("amount")
	action := r.FormValue("action")
	errorVal := r.FormValue("error")
	signTime := r.FormValue("sign_time")
	signString := r.FormValue("sign_string")

	userID, _ := strconv.ParseInt(merchantTransID, 10, 64)

	// Validate Click Sign MD5: md5(click_trans_id + service_id + secret_key + merchant_trans_id + amount + action + sign_time)
	checkStr := clickTransID + serviceID + pm.clickSecret + merchantTransID + amount + action + signTime
	hasher := md5.New()
	hasher.Write([]byte(checkStr))
	expectedSign := hex.EncodeToString(hasher.Sum(nil))

	if signString != expectedSign {
		jsonResp(w, map[string]interface{}{"error": -1, "error_note": "SIGN CHECK FAILED"})
		return
	}

	if errorVal != "0" {
		jsonResp(w, map[string]interface{}{"error": -2, "error_note": "Transaction error"})
		return
	}

	// Action 0: Prepare, Action 1: Complete
	if action == "0" {
		jsonResp(w, map[string]interface{}{
			"error":                0,
			"error_note":           "Success",
			"click_trans_id":       clickTransID,
			"merchant_trans_id":    merchantTransID,
			"merchant_prepare_id":  time.Now().Unix(),
		})
		return
	} else if action == "1" {
		// Activate 30 days VIP subscription
		pm.storage.ActivateSubscription(userID, 30*24*time.Hour)

		jsonResp(w, map[string]interface{}{
			"error":                0,
			"error_note":           "Success",
			"click_trans_id":       clickTransID,
			"merchant_trans_id":    merchantTransID,
			"merchant_confirm_id":  time.Now().Unix(),
		})
		return
	}

	jsonResp(w, map[string]interface{}{"error": -3, "error_note": "Action not found"})
}

func jsonResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = fmt.Fprintf(w, `{"error": %v, "error_note": "%v"}`, data.(map[string]interface{})["error"], data.(map[string]interface{})["error_note"])
}
