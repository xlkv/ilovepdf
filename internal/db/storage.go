package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StorageData struct {
	Users         map[int64]*User         `json:"users"`
	Filters       map[string]*UserFilter  `json:"filters"`
	Listings      map[string]*Listing     `json:"listings"`
	Subscriptions map[string]*Subscription `json:"subscriptions"`
}

type Storage struct {
	mu       sync.RWMutex
	data     StorageData
	filePath string
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	filePath := filepath.Join(dataDir, "monitor_data.json")
	s := &Storage{
		filePath: filePath,
		data: StorageData{
			Users:         make(map[int64]*User),
			Filters:       make(map[string]*UserFilter),
			Listings:      make(map[string]*Listing),
			Subscriptions: make(map[string]*Subscription),
		},
	}

	s.load()
	return s, nil
}

func (s *Storage) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err == nil {
		_ = json.Unmarshal(data, &s.data)
	}

	if s.data.Users == nil {
		s.data.Users = make(map[int64]*User)
	}
	if s.data.Filters == nil {
		s.data.Filters = make(map[string]*UserFilter)
	}
	if s.data.Listings == nil {
		s.data.Listings = make(map[string]*Listing)
	}
	if s.data.Subscriptions == nil {
		s.data.Subscriptions = make(map[string]*Subscription)
	}
}

func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bytes, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, bytes, 0644)
}

// User methods
func (s *Storage) GetUser(userID int64) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if u, exists := s.data.Users[userID]; exists {
		return u
	}
	return nil
}

func (s *Storage) UpsertUser(u *User) {
	s.mu.Lock()
	if existing, exists := s.data.Users[u.UserID]; exists {
		if u.Username != "" {
			existing.Username = u.Username
		}
		if u.FirstName != "" {
			existing.FirstName = u.FirstName
		}
		if u.LastName != "" {
			existing.LastName = u.LastName
		}
		if u.Language != "" {
			existing.Language = u.Language
		}
		existing.LastActiveAt = time.Now()
	} else {
		u.CreatedAt = time.Now()
		u.LastActiveAt = time.Now()
		s.data.Users[u.UserID] = u
	}
	s.mu.Unlock()
	_ = s.Save()
}

func (s *Storage) GrantFreeTrial(userID int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, exists := s.data.Users[userID]
	if !exists || u.FreeTrialUsed {
		return false
	}

	u.FreeTrialUsed = true
	u.SubscriptionExpire = time.Now().Add(24 * time.Hour)
	_ = s.Save()
	return true
}

func (s *Storage) ActivateSubscription(userID int64, duration time.Duration) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, exists := s.data.Users[userID]
	if !exists {
		u = &User{
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		s.data.Users[userID] = u
	}

	now := time.Now()
	if u.SubscriptionExpire.After(now) {
		u.SubscriptionExpire = u.SubscriptionExpire.Add(duration)
	} else {
		u.SubscriptionExpire = now.Add(duration)
	}

	_ = s.Save()
	return u.SubscriptionExpire
}

// Filter methods
func (s *Storage) SaveFilter(f *UserFilter) {
	s.mu.Lock()
	if f.ID == "" {
		f.ID = fmt.Sprintf("flt_%d_%d", f.UserID, time.Now().UnixNano())
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now()
	}
	s.data.Filters[f.ID] = f
	s.mu.Unlock()
	_ = s.Save()
}

func (s *Storage) GetUserFilters(userID int64) []*UserFilter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*UserFilter
	for _, f := range s.data.Filters {
		if f.UserID == userID {
			list = append(list, f)
		}
	}
	return list
}

func (s *Storage) DeleteFilter(filterID string) {
	s.mu.Lock()
	delete(s.data.Filters, filterID)
	s.mu.Unlock()
	_ = s.Save()
}

func (s *Storage) GetAllActiveFilters() []*UserFilter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var active []*UserFilter
	for _, f := range s.data.Filters {
		if f.Active {
			active = append(active, f)
		}
	}
	return active
}

// Listing methods
func (s *Storage) IsListingSeen(externalID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.data.Listings[externalID]
	return exists
}

func (s *Storage) SaveListing(l *Listing) {
	s.mu.Lock()
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	s.data.Listings[l.ExternalID] = l
	s.mu.Unlock()
	_ = s.Save()
}

func (s *Storage) GetRecentListings(limit int) []*Listing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var list []*Listing
	for _, l := range s.data.Listings {
		list = append(list, l)
	}

	// Sort recent first
	if len(list) > limit {
		return list[len(list)-limit:]
	}
	return list
}
