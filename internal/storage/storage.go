package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"informant/internal/config"
)

// ReadStatus represents the read status of news items
type ReadStatus struct {
	ReadItems map[string]time.Time `json:"read_items"`
	LastCheck time.Time            `json:"last_check"`
}

// Storage handles persistent storage of read status
type Storage struct {
	filePath string
	status   *ReadStatus
	mutex    sync.RWMutex
}

// New creates a new Storage instance
func New() (*Storage, error) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	filePath := filepath.Join(configPath, ".informant_read_status.json")

	storage := &Storage{
		filePath: filePath,
		status: &ReadStatus{
			ReadItems: make(map[string]time.Time),
			LastCheck: time.Now(),
		},
	}

	// Load existing data if available
	if err := storage.load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load read status: %w", err)
		}
	}

	return storage, nil
}

// IsRead checks if an item has been marked as read
func (s *Storage) IsRead(itemID string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.status.ReadItems[itemID]
	return exists
}

// MarkAsRead marks an item as read
func (s *Storage) MarkAsRead(itemID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.status.ReadItems[itemID] = time.Now()
	return s.save()
}

// MarkAsUnread marks an item as unread
func (s *Storage) MarkAsUnread(itemID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.status.ReadItems, itemID)
	return s.save()
}

// GetReadTime returns the time when an item was marked as read
func (s *Storage) GetReadTime(itemID string) (time.Time, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	readTime, exists := s.status.ReadItems[itemID]
	return readTime, exists
}

// GetReadCount returns the total number of read items
func (s *Storage) GetReadCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.status.ReadItems)
}

// Cleanup removes read status for items older than the specified duration
func (s *Storage) Cleanup(maxAge time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	
	for itemID, readTime := range s.status.ReadItems {
		if readTime.Before(cutoff) {
			delete(s.status.ReadItems, itemID)
		}
	}

	return s.save()
}

// load reads the read status from disk
func (s *Storage) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.status)
}

// save writes the current read status to disk
func (s *Storage) save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	s.status.LastCheck = time.Now()

	data, err := json.MarshalIndent(s.status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal read status: %w", err)
	}

	// Write to temporary file first, then rename for atomicity
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile) // Clean up on failure
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
