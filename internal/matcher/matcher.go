package matcher

import (
	"strings"

	"github.com/xlkv/ilovepdf/internal/db"
)

type FilterMatcher struct {
	storage *db.Storage
}

func NewFilterMatcher(storage *db.Storage) *FilterMatcher {
	return &FilterMatcher{
		storage: storage,
	}
}

// FindMatchingUsers returns list of user filters and users that match a newly scraped listing
func (m *FilterMatcher) FindMatchingUsers(listing *db.Listing) []int64 {
	activeFilters := m.storage.GetAllActiveFilters()
	matchedUsers := make(map[int64]bool)

	for _, filter := range activeFilters {
		if !filter.Active {
			continue
		}

		// Category check
		if filter.Category != "" && filter.Category != listing.Category {
			continue
		}

		// Make check
		if filter.Make != "" && filter.Make != "All" {
			if !strings.EqualFold(filter.Make, listing.Make) {
				continue
			}
		}

		// Model check
		if filter.Model != "" && filter.Model != "All" {
			if !strings.Contains(strings.ToLower(listing.Model), strings.ToLower(filter.Model)) &&
				!strings.Contains(strings.ToLower(listing.Title), strings.ToLower(filter.Model)) {
				continue
			}
		}

		// Year range check
		if filter.MinYear > 0 && listing.Year < filter.MinYear {
			continue
		}
		if filter.MaxYear > 0 && listing.Year > filter.MaxYear {
			continue
		}

		// Price range check
		if filter.MinPrice > 0 && listing.PriceUSD < filter.MinPrice {
			continue
		}
		if filter.MaxPrice > 0 && listing.PriceUSD > filter.MaxPrice {
			continue
		}

		// Region check
		if filter.Region != "" && filter.Region != "All" {
			if !strings.Contains(strings.ToLower(listing.Region), strings.ToLower(filter.Region)) {
				continue
			}
		}

		// Below Market Percentage check
		if filter.BelowMarketPct > 0 && listing.BelowMarketPct < filter.BelowMarketPct {
			continue
		}

		// Check if user's VIP subscription is active
		user := m.storage.GetUser(filter.UserID)
		if user != nil && user.IsVIP() {
			matchedUsers[filter.UserID] = true
		}
	}

	var userIDs []int64
	for userID := range matchedUsers {
		userIDs = append(userIDs, userID)
	}

	return userIDs
}
