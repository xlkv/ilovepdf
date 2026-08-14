package session

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type State string

const (
	StateIdle             State = "IDLE"
	StateMergeUploading   State = "MERGE_UPLOADING"
	StateSplitAwaitFile   State = "SPLIT_AWAIT_FILE"
	StateSplitAwaitRange  State = "SPLIT_AWAIT_RANGE"
	StateCompressAwaitFile State = "COMPRESS_AWAIT_FILE"
	StateCompressAwaitLevel State = "COMPRESS_AWAIT_LEVEL"
	StateWord2PDFAwaitFile State = "WORD2PDF_AWAIT_FILE"
	StatePPT2PDFAwaitFile  State = "PPT2PDF_AWAIT_FILE"
	StateExcel2PDFAwaitFile State = "EXCEL2PDF_AWAIT_FILE"
	StatePDF2WordAwaitFile State = "PDF2WORD_AWAIT_FILE"
	StatePDF2JPGAwaitFile  State = "PDF2JPG_AWAIT_FILE"
	StatePDF2JPGAwaitMode  State = "PDF2JPG_AWAIT_MODE"
	StateJPG2PDFUploading State = "JPG2PDF_UPLOADING"
	StateRotateAwaitFile   State = "ROTATE_AWAIT_FILE"
	StateRotateAwaitAngle  State = "ROTATE_AWAIT_ANGLE"
	StateProtectAwaitFile  State = "PROTECT_AWAIT_FILE"
	StateProtectAwaitPass  State = "PROTECT_AWAIT_PASS"
	StateUnlockAwaitFile   State = "UNLOCK_AWAIT_FILE"
	StateUnlockAwaitPass   State = "UNLOCK_AWAIT_PASS"
	StateWatermarkAwaitFile State = "WATERMARK_AWAIT_FILE"
	StateWatermarkAwaitText State = "WATERMARK_AWAIT_TEXT"
	StatePagenumAwaitFile  State = "PAGENUM_AWAIT_FILE"
	StatePagenumAwaitPos   State = "PAGENUM_AWAIT_POS"
	StateOrganizeAwaitFile State = "ORGANIZE_AWAIT_FILE"
	StateOrganizeAwaitPages State = "ORGANIZE_AWAIT_PAGES"
	StateHTML2PDFAwaitInput State = "HTML2PDF_AWAIT_INPUT"
	StateOCRAwaitFile      State = "OCR_AWAIT_FILE"
	StateOCRAwaitLang      State = "OCR_AWAIT_LANG"
)

type FileMeta struct {
	ID       string
	Name     string
	Path     string
	Size     int64
	MimeType string
}

type UserSession struct {
	UserID     int64
	State      State
	Language   string // "uz", "en", "ru"
	Files      []FileMeta
	Metadata   map[string]string
	SessionDir string
	LastActive time.Time
}

type Manager struct {
	mu       sync.RWMutex
	sessions map[int64]*UserSession
	baseDir  string
}

func NewManager(baseDir string) *Manager {
	m := &Manager{
		sessions: make(map[int64]*UserSession),
		baseDir:  baseDir,
	}
	go m.cleanupLoop()
	return m
}

func (m *Manager) Get(userID int64) *UserSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, exists := m.sessions[userID]
	if !exists {
		sessDir := filepath.Join(m.baseDir, fmt.Sprintf("%d_%d", userID, time.Now().UnixNano()))
		_ = os.MkdirAll(sessDir, 0755)

		sess = &UserSession{
			UserID:     userID,
			State:      StateIdle,
			Language:   "uz",
			Files:      make([]FileMeta, 0),
			Metadata:   make(map[string]string),
			SessionDir: sessDir,
			LastActive: time.Now(),
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
		_ = os.RemoveAll(sess.SessionDir)
		delete(m.sessions, userID)
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
		m.mu.Unlock()
	}
}
