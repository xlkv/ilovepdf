package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

type AnalyticsData struct {
	Users               map[int64]string `json:"users"` // userID -> lastActive ISO string
	ToolUsage           map[string]int64 `json:"tool_usage"`
	TotalProcessedFiles int64            `json:"total_processed_files"`
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
			Users:               make(map[int64]string),
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
	if m.analytics.Users == nil {
		m.analytics.Users = make(map[int64]string)
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

func (m *Manager) TrackUser(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.analytics.Users[userID] = time.Now().Format(time.RFC3339)
	m.saveAnalytics()
}

func (m *Manager) TrackToolUsage(toolID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.analytics.ToolUsage[toolID]++
	m.analytics.TotalProcessedFiles++
	m.saveAnalytics()
}

func (m *Manager) GetStats() (totalUsers int, activeToday int, totalFiles int64, toolStats map[string]int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	todayStr := time.Now().Format("2006-01-02")
	active := 0
	for _, lastActive := range m.analytics.Users {
		if len(lastActive) >= 10 && lastActive[:10] == todayStr {
			active++
		}
	}

	toolStatsCopy := make(map[string]int64)
	for k, v := range m.analytics.ToolUsage {
		toolStatsCopy[k] = v
	}

	return len(m.analytics.Users), active, m.analytics.TotalProcessedFiles, toolStatsCopy
}

func (m *Manager) GetAllUsers() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]int64, 0, len(m.analytics.Users))
	for userID := range m.analytics.Users {
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

	m.analytics.Users[userID] = time.Now().Format(time.RFC3339)
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
