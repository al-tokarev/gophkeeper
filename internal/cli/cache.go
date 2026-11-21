package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/al-tokarev/gophkeeper/internal/client"
)

// Cache — локальный кеш записей для оффлайн-режима.
type Cache struct {
	mu   sync.Mutex
	path string
	recs map[string]client.Record
}

// NewCache загружает кеш с диска.
func NewCache() (*Cache, error) {
	p, err := cachePath()
	if err != nil {
		return nil, err
	}
	c := &Cache{path: p, recs: make(map[string]client.Record)}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil // первый запуск — пустой кеш
		}
		return nil, err
	}
	if err := json.Unmarshal(b, &c.recs); err != nil {
		return nil, err
	}
	return c, nil
}

// cachePath возвращает путь к файлу кеша.
func cachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gophkeeper", "cache.json"), nil
}

// Put сохраняет запись в кеше и пишет на диск.
func (c *Cache) Put(rec client.Record) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.recs[rec.ID] = rec
	c.persist()
}

// PutAll сохраняет список записей.
func (c *Cache) PutAll(recs []client.Record) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, r := range recs {
		if existing, ok := c.recs[r.ID]; ok && len(r.Data) == 0 {
			r.Data = existing.Data
		}
		c.recs[r.ID] = r
	}
	c.persist()
}

// Get возвращает запись из кеша.
func (c *Cache) Get(id string) (client.Record, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.recs[id]
	return r, ok
}

// List возвращает все записи из кеша.
func (c *Cache) List() []client.Record {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]client.Record, 0, len(c.recs))
	for _, r := range c.recs {
		out = append(out, r)
	}
	return out
}

// Remove удаляет запись из кеша.
func (c *Cache) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.recs, id)
	c.persist()
}

// persist пишет кеш на диск
func (c *Cache) persist() {
	b, err := json.Marshal(c.recs)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(c.path), 0o700)
	_ = os.WriteFile(c.path, b, 0o600)
}
