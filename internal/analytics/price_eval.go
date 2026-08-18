package analytics

import (
	"math"
	"strings"
	"sync"
)

// MarketPriceData stores benchmark prices for popular car models and real estate in Uzbekistan
type PriceEvaluator struct {
	mu            sync.RWMutex
	baseCarPrices map[string]float64 // "chevrolet_cobalt_2022" -> 11500 USD
}

func NewPriceEvaluator() *PriceEvaluator {
	pe := &PriceEvaluator{
		baseCarPrices: map[string]float64{
			"chevrolet_gentra_2023": 13200,
			"chevrolet_gentra_2022": 12400,
			"chevrolet_gentra_2021": 11800,
			"chevrolet_gentra_2020": 11000,

			"chevrolet_cobalt_2024": 12800,
			"chevrolet_cobalt_2023": 11900,
			"chevrolet_cobalt_2022": 11200,
			"chevrolet_cobalt_2021": 10500,

			"chevrolet_damas_2023":  8800,
			"chevrolet_damas_2022":  8300,

			"chevrolet_tracker_2023": 21000,
			"chevrolet_tracker_2022": 19500,

			"kia_sonet_2023":        18500,
			"byd_chazor_2023":       20500,
			"byd_song_plus_2023":    28500,
		},
	}
	return pe
}

func (pe *PriceEvaluator) GetMarketAvgUSD(make, model string, year int) float64 {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	key := strings.ToLower(strings.TrimSpace(make) + "_" + strings.TrimSpace(model) + "_" + string(rune(year)))
	if price, exists := pe.baseCarPrices[key]; exists {
		return price
	}

	// Dynamic fallback algorithm based on model category
	mLower := strings.ToLower(model)
	if strings.Contains(mLower, "gentra") || strings.Contains(mLower, "lacetti") {
		return 12000.0
	}
	if strings.Contains(mLower, "cobalt") {
		return 11000.0
	}
	if strings.Contains(mLower, "spark") || strings.Contains(mLower, "nexia") {
		return 8500.0
	}
	if strings.Contains(mLower, "tracker") || strings.Contains(mLower, "malibu") {
		return 22000.0
	}

	return 0.0
}

// CalculateBelowMarketPct returns the percentage discount below market average (e.g. 15.5%)
func (pe *PriceEvaluator) CalculateBelowMarketPct(actualPriceUSD, marketAvgUSD float64) float64 {
	if marketAvgUSD <= 0 || actualPriceUSD <= 0 || actualPriceUSD >= marketAvgUSD {
		return 0.0
	}

	discount := ((marketAvgUSD - actualPriceUSD) / marketAvgUSD) * 100.0
	return math.Round(discount*10.0) / 10.0
}
