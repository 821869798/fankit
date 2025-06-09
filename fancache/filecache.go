package fancache

import (
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	CurrentVersion       = 1
	DefaultMaxItems      = 500
	DefaultEvictPercent  = 0.3  // 淘汰30%的项目
	RandomEvictThreshold = 1000 // 当缓存项超过1000时，使用随机淘汰策略
)

var (
	ErrCacheCorrupted = errors.New("cache file corrupted")
	ErrKeyMismatch    = errors.New("cache key mismatch")
)

// CacheHeader 缓存头部结构
type CacheHeader struct {
	Version    byte   `gob:"v"`
	Expiration int64  `gob:"e"`
	Key        string `gob:"k"`
}

// FileCache 文件缓存结构
type FileCache struct {
	dir          string
	maxItems     int
	evictPercent float64
	mu           sync.RWMutex
	keys         map[string]CacheHeader // key是原始键
}

// NewFileCache 创建新的文件缓存实例
func NewFileCache(cacheDir string, options ...Option) (*FileCache, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	fc := &FileCache{
		dir:          cacheDir,
		maxItems:     DefaultMaxItems,
		evictPercent: DefaultEvictPercent,
		keys:         make(map[string]CacheHeader),
	}

	// 应用配置选项
	for _, opt := range options {
		opt(fc)
	}

	if err := fc.scanCacheDir(); err != nil {
		return nil, fmt.Errorf("failed to scan cache directory: %w", err)
	}

	return fc, nil
}

// Option 配置选项
type Option func(*FileCache)

func WithMaxItems(maxItems int) Option {
	return func(fc *FileCache) {
		fc.maxItems = maxItems
	}
}

func WithEvictPercent(percent float64) Option {
	return func(fc *FileCache) {
		if percent > 0 && percent < 1 {
			fc.evictPercent = percent
		}
	}
}

// scanCacheDir 扫描缓存目录，初始化keys
func (fc *FileCache) scanCacheDir() error {
	entries, err := os.ReadDir(fc.dir)
	if err != nil {
		return err
	}

	tempKeys := make(map[string]CacheHeader, len(entries))
	now := time.Now().UnixNano()
	corruptedFiles := make([]string, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		hashedKeyAsFileName := entry.Name()
		filePath := filepath.Join(fc.dir, hashedKeyAsFileName)

		header, err := fc.readCacheHeader(filePath)
		if err != nil {
			corruptedFiles = append(corruptedFiles, filePath)
			continue
		}

		// 验证原始键的完整性
		if header.Key == "" {
			corruptedFiles = append(corruptedFiles, filePath)
			continue
		}

		// 验证哈希一致性（防止文件名被篡改）
		expectedHash := fc.getHash(header.Key)
		if expectedHash != hashedKeyAsFileName {
			corruptedFiles = append(corruptedFiles, filePath)
			continue
		}

		if now > header.Expiration {
			corruptedFiles = append(corruptedFiles, filePath)
		} else {
			tempKeys[header.Key] = header
		}
	}

	// 清理损坏和过期的文件
	fc.cleanupFiles(corruptedFiles)
	fc.keys = tempKeys

	return nil
}

// readCacheHeader 读取缓存文件头部
func (fc *FileCache) readCacheHeader(filePath string) (CacheHeader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return CacheHeader{}, err
	}
	defer file.Close()

	var header CacheHeader
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&header); err != nil {
		return CacheHeader{}, fmt.Errorf("failed to decode header: %w", err)
	}

	return header, nil
}

// cleanupFiles 批量清理文件
func (fc *FileCache) cleanupFiles(filePaths []string) {
	for _, filePath := range filePaths {
		os.Remove(filePath) // 忽略错误，因为文件可能已经不存在
	}
}

// Set 设置缓存
func (fc *FileCache) Set(key string, value interface{}, duration time.Duration) error {
	if key == "" {
		return errors.New("cache key cannot be empty")
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// 检查是否需要淘汰
	_, keyExists := fc.keys[key]
	if fc.maxItems > 0 && len(fc.keys) >= fc.maxItems && !keyExists {
		if err := fc.evictCache(); err != nil { // 使用evictCache替换evictOldest
			return fmt.Errorf("failed to evict items: %w", err)
		}
	}

	header := CacheHeader{
		Version:    CurrentVersion,
		Expiration: time.Now().Add(duration).UnixNano(),
		Key:        key,
	}

	hashedKey := fc.getHash(key)
	filePath := filepath.Join(fc.dir, hashedKey)

	// 使用临时文件写入，确保原子性
	tempPath := filePath + ".tmp"
	if err := fc.writeToFile(tempPath, header, value); err != nil {
		os.Remove(tempPath)
		return err
	}

	// 原子性重命名
	if err := os.Rename(tempPath, filePath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	fc.keys[key] = header
	return nil
}

// writeToFile 写入数据到文件
func (fc *FileCache) writeToFile(filePath string, header CacheHeader, value interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(header); err != nil {
		return fmt.Errorf("failed to encode header: %w", err)
	}
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	// 确保数据写入磁盘
	return file.Sync()
}

// Get 获取缓存
func (fc *FileCache) Get(key string, value interface{}) (bool, error) {
	if key == "" {
		return false, errors.New("cache key cannot be empty")
	}

	fc.mu.RLock()
	header, exists := fc.keys[key]
	fc.mu.RUnlock()

	if !exists {
		return false, nil
	}

	now := time.Now().UnixNano()
	if now > header.Expiration {
		// 双重检查锁定模式处理过期
		fc.mu.Lock()
		// 再次检查key是否存在以及是否过期
		if currentHeader, stillExists := fc.keys[key]; stillExists && now > currentHeader.Expiration {
			fc.removeUnsafe(key)
		}
		fc.mu.Unlock()
		return false, nil
	}

	err := fc.readFromFile(key, value)
	if err != nil {
		// 如果文件不存在，或文件已损坏，都应清理
		isCorrupted := errors.Is(err, ErrCacheCorrupted) || errors.Is(err, ErrKeyMismatch)
		if errors.Is(err, fs.ErrNotExist) || isCorrupted {
			fc.mu.Lock()
			// 再次检查，防止在获取锁的过程中状态变化
			if h, ok := fc.keys[key]; ok && h.Expiration == header.Expiration { // 确保是同一个缓存项
				fc.removeUnsafe(key)
			}
			fc.mu.Unlock()
		}
		return false, err // 返回interface{}的零值和错误
	}

	return true, nil
}

// readFromFile 从文件读取数据
func (fc *FileCache) readFromFile(key string, value interface{}) error {
	hashedKey := fc.getHash(key)
	filePath := filepath.Join(fc.dir, hashedKey)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)

	var fileHeader CacheHeader
	if err := decoder.Decode(&fileHeader); err != nil {
		return fmt.Errorf("corrupted header: %w", err)
	}

	// 验证头部一致性
	if fileHeader.Key != key {
		return ErrKeyMismatch
	}

	// 解码数据到interface{}变量data的地址
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("corrupted data: %w", err)
	}

	return nil
}

// evictRandom 随机淘汰缓存项，适用于大数量缓存
func (fc *FileCache) evictRandom() error {
	numToEvict := int(float64(len(fc.keys)) * fc.evictPercent)
	if numToEvict == 0 && len(fc.keys) > 0 { // 确保至少淘汰一个（如果缓存不为空）
		numToEvict = 1
	}
	if numToEvict == 0 { // 如果缓存为空，则不执行任何操作
		return nil
	}

	keysToEvictSource := make([]string, 0, len(fc.keys))
	for k := range fc.keys {
		keysToEvictSource = append(keysToEvictSource, k)
	}

	finalKeysToEvict := make(map[string]struct{})

	// 使用更可靠的随机源，但为了简单，这里保持原有逻辑的意图
	// 在实际应用中，考虑使用 crypto/rand 或 math/rand.Shuffle
	for i := 0; i < numToEvict && len(keysToEvictSource) > 0; i++ {
		// 简单的伪随机，注意并发环境下time.Now().UnixNano()可能不够随机
		// 更好的做法是使用 math/rand.Intn，并确保正确播种
		randIndex := int(uint32(time.Now().UnixNano()+int64(i)) % uint32(len(keysToEvictSource))) // 增加i以尝试避免快速调用下的重复
		keyToEvict := keysToEvictSource[randIndex]
		finalKeysToEvict[keyToEvict] = struct{}{}

		keysToEvictSource[randIndex] = keysToEvictSource[len(keysToEvictSource)-1]
		keysToEvictSource = keysToEvictSource[:len(keysToEvictSource)-1]
	}

	for keyToEvict := range finalKeysToEvict {
		if err := fc.removeUnsafe(keyToEvict); err != nil {
			// log error but continue
		}
	}

	return nil
}

// evictOldest 淘汰最旧的缓存项
func (fc *FileCache) evictOldest() error {
	if len(fc.keys) == 0 {
		return nil
	}

	numToEvict := int(float64(len(fc.keys)) * fc.evictPercent)
	if numToEvict == 0 {
		numToEvict = 1
	}

	type expiringItem struct {
		key        string
		expiration int64
	}

	items := make([]expiringItem, 0, len(fc.keys))
	for k, header := range fc.keys {
		items = append(items, expiringItem{key: k, expiration: header.Expiration})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].expiration < items[j].expiration
	})

	for i := 0; i < numToEvict && i < len(items); i++ {
		if err := fc.removeUnsafe(items[i].key); err != nil {
			// 考虑记录错误，但继续淘汰其他项
		}
	}

	return nil
}

func (fc *FileCache) evictCache() error {
	if len(fc.keys) < fc.maxItems {
		return nil
	}
	if fc.maxItems > RandomEvictThreshold && len(fc.keys) > RandomEvictThreshold {
		return fc.evictRandom()
	} else {
		return fc.evictOldest()
	}
}

// removeUnsafe 不加锁的删除方法（内部使用）
func (fc *FileCache) removeUnsafe(key string) error {
	hashedKey := fc.getHash(key)
	filePath := filepath.Join(fc.dir, hashedKey)
	delete(fc.keys, key)

	if err := os.Remove(filePath); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// CleanExpired 清理所有过期缓存
func (fc *FileCache) CleanExpired() error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	now := time.Now().UnixNano()
	expiredKeys := make([]string, 0)

	for key, header := range fc.keys {
		if now > header.Expiration {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		fc.removeUnsafe(key) // 忽略错误，继续清理
	}

	return nil
}

// Remove 删除指定缓存
func (fc *FileCache) Remove(key string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if _, exists := fc.keys[key]; !exists {
		return nil
	}

	return fc.removeUnsafe(key)
}

// Size 获取当前缓存项数量
func (fc *FileCache) Size() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return len(fc.keys)
}

// Clear 清空所有缓存
func (fc *FileCache) Clear() error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	keysToRemove := make([]string, 0, len(fc.keys))
	for key := range fc.keys {
		keysToRemove = append(keysToRemove, key)
	}

	for _, key := range keysToRemove {
		fc.removeUnsafe(key) // 忽略错误，继续清理
	}
	// fc.keys = make(map[string]CacheHeader) // 确保map被清空, delete(fc.keys, key) 已经处理
	return nil
}

// getHash 生成键的哈希值
func (fc *FileCache) getHash(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}
