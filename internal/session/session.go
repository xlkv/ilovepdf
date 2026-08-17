package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type State string

const (
	StateIdle               State = "IDLE"
	StateMergeUploading     State = "MERGE_UPLOADING"
	StateSplitAwaitFile     State = "SPLIT_AWAIT_FILE"
	StateSplitAwaitRange    State = "SPLIT_AWAIT_RANGE"
	StateCompressAwaitFile  State = "COMPRESS_AWAIT_FILE"
	StateCompressAwaitLevel State = "COMPRESS_AWAIT_LEVEL"
	StateWord2PDFAwaitFile  State = "WORD2PDF_AWAIT_FILE"
	StatePPT2PDFAwaitFile   State = "PPT2PDF_AWAIT_FILE"
	StateExcel2PDFAwaitFile State = "EXCEL2PDF_AWAIT_FILE"
	StatePDF2WordAwaitFile  State = "PDF2WORD_AWAIT_FILE"
	StatePDF2JPGAwaitFile   State = "PDF2JPG_AWAIT_FILE"
	StatePDF2JPGAwaitMode   State = "PDF2JPG_AWAIT_MODE"
	StateJPG2PDFUploading   State = "JPG2PDF_UPLOADING"
	StateRotateAwaitFile    State = "ROTATE_AWAIT_FILE"
	StateRotateAwaitAngle   State = "ROTATE_AWAIT_ANGLE"
	StateProtectAwaitFile   State = "PROTECT_AWAIT_FILE"
	StateProtectAwaitPass   State = "PROTECT_AWAIT_PASS"
	StateUnlockAwaitFile    State = "UNLOCK_AWAIT_FILE"
	StateUnlockAwaitPass    State = "UNLOCK_AWAIT_PASS"
	StateWatermarkAwaitFile State = "WATERMARK_AWAIT_FILE"
	StateWatermarkAwaitText State = "WATERMARK_AWAIT_TEXT"
	StatePagenumAwaitFile   State = "PAGENUM_AWAIT_FILE"
	StatePagenumAwaitPos    State = "PAGENUM_AWAIT_POS"
	StateOrganizeAwaitFile  State = "ORGANIZE_AWAIT_FILE"
	StateOrganizeAwaitPages State = "ORGANIZE_AWAIT_PAGES"
	StateHTML2PDFAwaitInput State = "HTML2PDF_AWAIT_INPUT"
	StateOCRAwaitFile       State = "OCR_AWAIT_FILE"
	StateOCRAwaitLang       State = "OCR_AWAIT_LANG"
)

type FileMeta struct {
	ID       string
	Name     string
	Path     string
	Size     int64
	MimeType string
}

type UserSession struct {
	UserID           int64
	State            State
	Language         string // "uz", "en", "ru"
	LanguageSelected bool   // true if user picked language
	Files            []FileMeta
	Metadata         map[string]string
	TempMessageIDs   []int64
	SessionDir       string
	LastActive       time.Time
}

type UserProfile struct {
	UserID          int64            `json:"user_id"`
	Username        string           `json:"username"`
	FirstName       string           `json:"first_name"`
	LastName        string           `json:"last_name"`
	Language        string           `json:"language"`
	FirstSeen       string           `json:"first_seen"`
	LastActive      string           `json:"last_active"`
	TotalOperations int64            `json:"total_operations"`
	ToolUsage       map[string]int64 `json:"tool_usage"`
}

type AnalyticsData struct {
	UserProfiles        map[int64]*UserProfile `json:"user_profiles"`
	ToolUsage           map[string]int64       `json:"tool_usage"`
	TotalProcessedFiles int64                  `json:"total_processed_files"`
}

type Manager struct {
	mu        sync.RWMutex
	sessions  map[int64]*UserSession
	baseDir   string
	analytics AnalyticsData
	dataFile  string
}

func NewManager(baseDir string) *Manager {
	dataDir := filepath.Join(baseDir, "data")
	_ = os.MkdirAll(dataDir, 0755)
	dataFile := filepath.Join(dataDir, "analytics.json")

	m := &Manager{
		sessions: make(map[int64]*UserSession),
		baseDir:  baseDir,
		dataFile: dataFile,
		analytics: AnalyticsData{
			UserProfiles:        make(map[int64]*UserProfile),
			ToolUsage:           make(map[string]int64),
			TotalProcessedFiles: 0,
		},
	}

	m.loadAnalytics()
	go m.cleanupLoop()
	return m
}

func (m *Manager) loadAnalytics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.dataFile)
	if err == nil {
		_ = json.Unmarshal(data, &m.analytics)
	}
	if m.analytics.UserProfiles == nil {
		m.analytics.UserProfiles = make(map[int64]*UserProfile)
	}
	if m.analytics.ToolUsage == nil {
		m.analytics.ToolUsage = make(map[string]int64)
	}
}

func (m *Manager) saveAnalytics() {
	data, err := json.MarshalIndent(m.analytics, "", "  ")
	if err == nil {
		_ = os.WriteFile(m.dataFile, data, 0644)
	}
}

func (m *Manager) TrackUserFull(userID int64, username, firstName, lastName, lang string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nowStr := time.Now().Format(time.RFC3339)
	profile, exists := m.analytics.UserProfiles[userID]
	if !exists {
		profile = &UserProfile{
			UserID:          userID,
			Username:        username,
			FirstName:       firstName,
			LastName:        lastName,
			Language:        lang,
			FirstSeen:       nowStr,
			LastActive:      nowStr,
			TotalOperations: 0,
			ToolUsage:       make(map[string]int64),
		}
		m.analytics.UserProfiles[userID] = profile
	} else {
		if username != "" {
			profile.Username = username
		}
		if firstName != "" {
			profile.FirstName = firstName
		}
		if lastName != "" {
			profile.LastName = lastName
		}
		if lang != "" {
			profile.Language = lang
		}
		profile.LastActive = nowStr
	}

	m.saveAnalytics()
}

func (m *Manager) TrackToolUsageFull(userID int64, toolID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.analytics.ToolUsage[toolID]++
	m.analytics.TotalProcessedFiles++

	if profile, exists := m.analytics.UserProfiles[userID]; exists {
		profile.TotalOperations++
		if profile.ToolUsage == nil {
			profile.ToolUsage = make(map[string]int64)
		}
		profile.ToolUsage[toolID]++
		profile.LastActive = time.Now().Format(time.RFC3339)
	}

	m.saveAnalytics()
}

type DetailedStats struct {
	TotalUsers    int
	ActiveToday   int
	NewToday      int
	TotalFiles    int64
	LangCounts    map[string]int
	TopTools      []ToolStat
	RecentUsers   []*UserProfile
}

type ToolStat struct {
	ToolID string
	Count  int64
}

func (m *Manager) GetDetailedStats() DetailedStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	todayStr := time.Now().Format("2006-01-02")
	activeToday := 0
	newToday := 0
	langCounts := map[string]int{"uz": 0, "ru": 0, "en": 0, "other": 0}

	profiles := make([]*UserProfile, 0, len(m.analytics.UserProfiles))

	for _, p := range m.analytics.UserProfiles {
		profiles = append(profiles, p)

		if len(p.LastActive) >= 10 && p.LastActive[:10] == todayStr {
			activeToday++
		}
		if len(p.FirstSeen) >= 10 && p.FirstSeen[:10] == todayStr {
			newToday++
		}

		lang := p.Language
		if lang == "uz" || lang == "ru" || lang == "en" {
			langCounts[lang]++
		} else {
			langCounts["other"]++
		}
	}

	// Sort recent users by LastActive descending
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].LastActive > profiles[j].LastActive
	})

	recentLimit := 10
	if len(profiles) < recentLimit {
		recentLimit = len(profiles)
	}
	recentUsers := profiles[:recentLimit]

	// Sort top tools
	toolList := make([]ToolStat, 0, len(m.analytics.ToolUsage))
	for toolID, count := range m.analytics.ToolUsage {
		toolList = append(toolList, ToolStat{ToolID: toolID, Count: count})
	}
	sort.Slice(toolList, func(i, j int) bool {
		return toolList[i].Count > toolList[j].Count
	})

	return DetailedStats{
		TotalUsers:  len(m.analytics.UserProfiles),
		ActiveToday: activeToday,
		NewToday:    newToday,
		TotalFiles:  m.analytics.TotalProcessedFiles,
		LangCounts:  langCounts,
		TopTools:    toolList,
		RecentUsers: recentUsers,
	}
}

func (m *Manager) GetAllUserIDs() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]int64, 0, len(m.analytics.UserProfiles))
	for userID := range m.analytics.UserProfiles {
		users = append(users, userID)
	}
	return users
}

func (m *Manager) Get(userID int64) *UserSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[userID]
	if !exists {
		sessDir := filepath.Join(m.baseDir, fmt.Sprintf("%d_%d", userID, time.Now().UnixNano()))
		_ = os.MkdirAll(sessDir, 0755)

		sess = &UserSession{
			UserID:           userID,
			State:            StateIdle,
			Language:         "uz",
			LanguageSelected: false,
			Files:            make([]FileMeta, 0),
			Metadata:         make(map[string]string),
			TempMessageIDs:   make([]int64, 0),
			SessionDir:       sessDir,
			LastActive:       time.Now(),
		}
		m.sessions[userID] = sess
	} else {
		sess.LastActive = time.Now()
	}

	return sess
}

func (m *Manager) Reset(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[userID]
	if exists {
		lang := sess.Language
		langSelected := sess.LanguageSelected
		_ = os.RemoveAll(sess.SessionDir)

		sessDir := filepath.Join(m.baseDir, fmt.Sprintf("%d_%d", userID, time.Now().UnixNano()))
		_ = os.MkdirAll(sessDir, 0755)

		m.sessions[userID] = &UserSession{
			UserID:           userID,
			State:            StateIdle,
			Language:         lang,
			LanguageSelected: langSelected,
			Files:            make([]FileMeta, 0),
			Metadata:         make(map[string]string),
			TempMessageIDs:   make([]int64, 0),
			SessionDir:       sessDir,
			LastActive:       time.Now(),
		}
	}
}

func (m *Manager) SetState(userID int64, state State) {
	sess := m.Get(userID)
	m.mu.Lock()
	defer m.mu.Unlock()
	sess.State = state
}

func (m *Manager) SetLanguage(userID int64, lang string) {
	sess := m.Get(userID)
	m.mu.Lock()
	defer m.mu.Unlock()
	sess.Language = lang
	sess.LanguageSelected = true
}

func (m *Manager) AddFile(userID int64, file FileMeta) {
	sess := m.Get(userID)
	m.mu.Lock()
	defer m.mu.Unlock()
	sess.Files = append(sess.Files, file)
}

func (m *Manager) SetMeta(userID int64, key, val string) {
	sess := m.Get(userID)
	m.mu.Lock()
	defer m.mu.Unlock()
	sess.Metadata[key] = val
}

func (m *Manager) AddTempMsg(userID int64, msgID int64) {
	if msgID <= 0 {
		return
	}
	sess := m.Get(userID)
	m.mu.Lock()
	defer m.mu.Unlock()
	sess.TempMessageIDs = append(sess.TempMessageIDs, msgID)
}

func (m *Manager) PopTempMsgs(userID int64) []int64 {
	sess := m.Get(userID)
	m.mu.Lock()
	defer m.mu.Unlock()
	msgs := sess.TempMessageIDs
	sess.TempMessageIDs = make([]int64, 0)
	return msgs
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for userID, sess := range m.sessions {
			if now.Sub(sess.LastActive) > 30*time.Minute {
				_ = os.RemoveAll(sess.SessionDir)
				delete(m.sessions, userID)
			}
		}
		m.saveAnalytics()
		m.mu.Unlock()
	}
}
