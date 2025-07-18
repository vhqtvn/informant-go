package storage

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"informant/internal/config"
)

// ReadStatus represents the read status of news items
type ReadStatus struct {
	ReadItems map[string]time.Time `json:"read_items"`
	LastCheck time.Time            `json:"last_check"`
}

// CacheEntry represents a cached RSS feed
type CacheEntry struct {
	Data      []byte    `json:"data"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
}

// Storage handles persistent storage of read status
type Storage struct {
	filePath     string
	status       *ReadStatus
	mutex        sync.RWMutex
	isSystemWide bool
	cacheDir     string
}

// showStorageFallbackWarning displays a warning about falling back to per-user storage
func showStorageFallbackWarning() {
	fmt.Println("Warning: Cannot write to system-wide storage (/var/lib/informant-go.dat)")
	fmt.Println("Falling back to per-user storage. This means read status won't be shared between users.")
}

// New creates a new Storage instance
func New() (*Storage, error) {
	return NewWithConfirmation(true)
}

// NewWithConfirmation creates a new Storage instance with optional confirmation prompts
func NewWithConfirmation(requireConfirmation bool) (*Storage, error) {
	// Try system-wide storage first
	systemFilePath := "/var/lib/informant-go.dat"
	systemCacheDir := "/var/cache/informant"

	// Check if we're running as root
	isRoot := os.Geteuid() == 0

	var filePath, cacheDir string
	var isSystemWide bool

	if isRoot {
		// Running as root - create system directories with proper permissions
		if err := createSystemDirectories(systemFilePath, systemCacheDir); err != nil {
			return nil, fmt.Errorf("failed to create system directories: %w", err)
		}
		filePath = systemFilePath
		cacheDir = systemCacheDir
		isSystemWide = true
	} else {
		// Try to use system-wide storage
		if canUseSystemStorage(systemFilePath, systemCacheDir) {
			filePath = systemFilePath
			cacheDir = systemCacheDir
			isSystemWide = true
		} else {
			// Fall back to per-user storage
			if requireConfirmation {
				if !confirmFallback() {
					return nil, fmt.Errorf("user declined to use per-user storage")
				}
			} else {
				// Show warning but don't require confirmation
				showStorageFallbackWarning()
			}

			var err error
			filePath, cacheDir, err = getUserStoragePaths()
			if err != nil {
				return nil, fmt.Errorf("failed to get user storage paths: %w", err)
			}
			isSystemWide = false
		}
	}

	storage := &Storage{
		filePath:     filePath,
		cacheDir:     cacheDir,
		isSystemWide: isSystemWide,
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

// createSystemDirectories creates system directories with proper permissions
func createSystemDirectories(filePath, cacheDir string) error {
	// Create /var/lib directory if it doesn't exist
	libDir := filepath.Dir(filePath)
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %w", err)
	}

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Set permissions on cache directory so all users can write (only if needed)
	if info, err := os.Stat(cacheDir); err == nil {
		if info.Mode().Perm() != 0777 {
			if err := os.Chmod(cacheDir, 0777); err != nil {
				return fmt.Errorf("failed to set cache directory permissions: %w", err)
			}
		}
	}

	return nil
}

// canUseSystemStorage checks if the current user can use system-wide storage
func canUseSystemStorage(filePath, cacheDir string) bool {
	// Check if we can write to the storage file
	if !canWriteToFile(filePath) {
		return false
	}

	// Check if we can write to the cache directory
	if !canWriteToDirectory(cacheDir) {
		return false
	}

	return true
}

// canWriteToFile checks if we can write to a specific file
func canWriteToFile(filePath string) bool {
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, check if we can write to the directory
			return canWriteToDirectory(filepath.Dir(filePath))
		}
		return false
	}

	// File exists, try to open it for writing
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// canWriteToDirectory checks if we can write to a directory
func canWriteToDirectory(dir string) bool {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// Try to create a temporary file in the directory
	tempFile := filepath.Join(dir, ".informant_test_write")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		return false
	}
	os.Remove(tempFile)
	return true
}

// getUserStoragePaths returns per-user storage paths
func getUserStoragePaths() (string, string, error) {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return "", "", fmt.Errorf("failed to get config path: %w", err)
	}

	filePath := filepath.Join(configPath, ".informant_read_status.json")
	cacheDir := filepath.Join(configPath, ".informant_cache")

	// Create cache directory
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create user cache directory: %w", err)
	}

	return filePath, cacheDir, nil
}

// confirmFallback asks user for confirmation to use per-user storage
func confirmFallback() bool {
	showStorageFallbackWarning()
	fmt.Print("Continue with per-user storage? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// GetCacheFile returns cached RSS data if available and not expired
func (s *Storage) GetCacheFile(url string, maxAge time.Duration) ([]byte, bool) {
	cacheFile := s.getCacheFilePath(url)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Check if cache is still valid
	if time.Since(entry.Timestamp) > maxAge {
		return nil, false
	}

	return entry.Data, true
}

// SetCacheFile saves RSS data to cache
func (s *Storage) SetCacheFile(url string, data []byte) error {
	cacheFile := s.getCacheFilePath(url)

	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		URL:       url,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Ensure cache directory exists (only if we have permission)
	if _, err := os.Stat(s.cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.cacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	// Write cache file directly
	if err := os.WriteFile(cacheFile, jsonData, 0666); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getCacheFilePath generates a cache file path for a URL
func (s *Storage) getCacheFilePath(url string) string {
	// Use MD5 hash of URL as filename to avoid filesystem issues
	hash := md5.Sum([]byte(url))
	filename := fmt.Sprintf("%x.json", hash)
	return filepath.Join(s.cacheDir, filename)
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

// IsSystemWide returns whether storage is system-wide or per-user
func (s *Storage) IsSystemWide() bool {
	return s.isSystemWide
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
	// Ensure directory exists (only if we have permission)
	dir := filepath.Dir(s.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	s.status.LastCheck = time.Now()

	data, err := json.MarshalIndent(s.status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal read status: %w", err)
	}

	// Set appropriate permissions based on whether we're using system-wide storage
	var perm os.FileMode = 0644
	if s.isSystemWide {
		perm = 0666
	}

	// Write directly to the file in-place
	if err := os.WriteFile(s.filePath, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Set final permissions on the file if system-wide and we're root
	if s.isSystemWide && os.Geteuid() == 0 {
		// Check if file already has correct permissions
		if info, err := os.Stat(s.filePath); err == nil {
			if info.Mode().Perm() != 0666 {
				if err := os.Chmod(s.filePath, 0666); err != nil {
					return fmt.Errorf("failed to set file permissions: %w", err)
				}
			}
		}
	}

	return nil
}
