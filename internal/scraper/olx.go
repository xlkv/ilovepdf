package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xlkv/ilovepdf/internal/analytics"
	"github.com/xlkv/ilovepdf/internal/db"
)

type OLXScraper struct {
	client    *http.Client
	evaluator *analytics.PriceEvaluator
}

func NewOLXScraper(evaluator *analytics.PriceEvaluator) *OLXScraper {
	return &OLXScraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		evaluator: evaluator,
	}
}

type OLXAPIResponse struct {
	Data []struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		CreatedTime string `json:"created_time"`
		Params      []struct {
			Key   string `json:"key"`
			Name  string `json:"name"`
			Value struct {
				Value interface{} `json:"value"`
				Label string      `json:"label"`
			} `json:"value"`
		} `json:"params"`
		ParamsList struct {
			Price struct {
				Value        float64 `json:"value"`
				Currency     string  `json:"currency"`
				DisplayValue string  `json:"display_value"`
			} `json:"price"`
		} `json:"params_list"`
		Location struct {
			City struct {
				Name string `json:"name"`
			} `json:"city"`
			Region struct {
				Name string `json:"name"`
			} `json:"region"`
		} `json:"location"`
		Photos []struct {
			Link string `json:"link"`
		} `json:"photos"`
	} `json:"data"`
}

func (s *OLXScraper) FetchLatestCars() ([]*db.Listing, error) {
	// OLX.uz public API endpoint for Transport -> Cars
	url := "https://www.olx.uz/api/v1/offers/?category_id=85&limit=20"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("olx returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp OLXAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	var listings []*db.Listing
	for _, item := range apiResp.Data {
		l := s.parseItem(&item)
		if l != nil {
			listings = append(listings, l)
		}
	}

	return listings, nil
}

func (s *OLXScraper) parseItem(item *OLXAPIResponseData) *db.Listing {
	title := item.Title
	if title == "" {
		return nil
	}

	priceUSD := 0.0
	priceUZS := 0.0

	// Parse price
	val := item.ParamsList.Price.Value
	curr := strings.ToUpper(item.ParamsList.Price.Currency)
	if curr == "USD" || curr == "$" {
		priceUSD = val
		priceUZS = val * 12600.0 // Approx USD to UZS rate
	} else {
		priceUZS = val
		priceUSD = val / 12600.0
	}

	if priceUSD <= 0 {
		return nil
	}

	makeName, modelName := parseMakeAndModel(title)
	year := parseYearFromTitle(title)

	photoURL := ""
	if len(item.Photos) > 0 {
		photoURL = strings.Replace(item.Photos[0].Link, "{width}", "600", 1)
		photoURL = strings.Replace(photoURL, "{height}", "450", 1)
	}

	location := item.Location.City.Name
	if location == "" {
		location = item.Location.Region.Name
	}
	if location == "" {
		location = "Toshkent"
	}

	marketAvgUSD := s.evaluator.GetMarketAvgUSD(makeName, modelName, year)
	discountPct := s.evaluator.CalculateBelowMarketPct(priceUSD, marketAvgUSD)

	return &db.Listing{
		ID:             fmt.Sprintf("olx_%d", item.ID),
		Source:         "olx",
		ExternalID:     fmt.Sprintf("olx_%d", item.ID),
		Title:          title,
		Category:       db.CategoryCar,
		Make:           makeName,
		Model:          modelName,
		PriceUSD:       priceUSD,
		PriceUZS:       priceUZS,
		MarketAvgUSD:   marketAvgUSD,
		BelowMarketPct: discountPct,
		Year:           year,
		Region:         location,
		URL:            item.URL,
		PhotoURL:       photoURL,
		PublishedAt:    time.Now(),
		CreatedAt:      time.Now(),
	}
}

type OLXAPIResponseData = struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	CreatedTime string `json:"created_time"`
	Params      []struct {
		Key   string `json:"key"`
		Name  string `json:"name"`
		Value struct {
			Value interface{} `json:"value"`
			Label string      `json:"label"`
		} `json:"value"`
	} `json:"params"`
	ParamsList struct {
		Price struct {
			Value        float64 `json:"value"`
			Currency     string  `json:"currency"`
			DisplayValue string  `json:"display_value"`
		} `json:"price"`
	} `json:"params_list"`
	Location struct {
		City struct {
			Name string `json:"name"`
		} `json:"city"`
		Region struct {
			Name string `json:"name"`
		} `json:"region"`
	} `json:"location"`
	Photos []struct {
		Link string `json:"link"`
	} `json:"photos"`
}

func parseMakeAndModel(title string) (string, string) {
	tLower := strings.ToLower(title)
	makeName := "Chevrolet"

	models := []string{"gentra", "lacetti", "cobalt", "damas", "spark", "nexia", "matiz", "tracker", "malibu", "equinox", "tahoe", "monza", "byd", "kia"}

	for _, m := range models {
		if strings.Contains(tLower, m) {
			return makeName, strings.Title(m)
		}
	}

	return makeName, "Avto"
}

func parseYearFromTitle(title string) int {
	re := regexp.MustCompile(`\b(20[0-2][0-9])\b`)
	match := re.FindString(title)
	if match != "" {
		yr, _ := strconv.Atoi(match)
		return yr
	}
	return 2022
}
