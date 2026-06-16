package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MD-Mushfiqur123/lychee/api"
	_ "modernc.org/sqlite"
)

// ConversationSummary provides high-level info about a saved conversation.
type ConversationSummary struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	Title     string    `json:"title"`
	Messages  int       `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Conversation holds the full metadata and list of messages for a chat.
type Conversation struct {
	ID        string        `json:"id"`
	Model     string        `json:"model"`
	Messages  []api.Message `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// MemoryStore manages conversation persistence using JSON and SQLite.
type MemoryStore struct {
	dir    string
	dbPath string
	db     *sql.DB
	mu     sync.RWMutex
}

// NewMemoryStore initializes a new MemoryStore at the specified directory.
func NewMemoryStore(dir string) *MemoryStore {
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create conversation directory: %v\n", err)
	}

	dbPath := filepath.Join(dir, "conversations.db")
	db, err := sql.Open("sqlite", dbPath)
	if err == nil {
		schema := `
		CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			model TEXT,
			title TEXT,
			created_at DATETIME,
			updated_at DATETIME
		);
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id TEXT,
			role TEXT,
			content TEXT,
			images TEXT,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		);`
		if _, err := db.Exec(schema); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize sqlite schema: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "failed to open sqlite db: %v\n", err)
	}

	ms := &MemoryStore{
		dir:    dir,
		dbPath: dbPath,
		db:     db,
	}

	if db != nil {
		if err := ms.runMigrations(db); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		}
	}

	return ms
}

// Save writes a conversation to both SQLite and JSON file.
func (ms *MemoryStore) Save(conv *Conversation) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// 1. Write to JSON
	jsonPath := filepath.Join(ms.dir, conv.ID+".json")
	jsonData, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	// 2. Write to SQLite
	if ms.db != nil {
		title := "New Conversation"
		for _, m := range conv.Messages {
			if m.Role == "user" && m.Content != "" {
				title = m.Content
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				break
			}
		}

		// Update title if conversation has enough context (mature conversations)
		if len(conv.Messages) >= 4 {
			for _, m := range conv.Messages {
				if m.Role == "user" && m.Content != "" {
					title = m.Content
					if len(title) > 50 {
						title = title[:47] + "..."
					}
				}
			}
		}

		_, err = ms.db.Exec(`
			INSERT OR REPLACE INTO conversations (id, model, title, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`,
			conv.ID, conv.Model, title, conv.CreatedAt, conv.UpdatedAt,
		)
		if err != nil {
			return err
		}

		if _, err := ms.db.Exec("DELETE FROM messages WHERE conversation_id = ?", conv.ID); err != nil {
			slog.Warn("memory: failed to delete old messages before re-insert", "error", err, "conversation_id", conv.ID)
		}

		for _, msg := range conv.Messages {
			var imagesJSON []byte
			if len(msg.Images) > 0 {
				imagesJSON, _ = json.Marshal(msg.Images)
			}
			_, err = ms.db.Exec(`
				INSERT INTO messages (conversation_id, role, content, images)
				VALUES (?, ?, ?, ?)`,
				conv.ID, msg.Role, msg.Content, string(imagesJSON),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Load retrieves a conversation by ID, checking SQLite and falling back to JSON.
func (ms *MemoryStore) Load(id string) (*Conversation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.db != nil {
		var conv Conversation
		err := ms.db.QueryRow(`
			SELECT id, model, created_at, updated_at FROM conversations WHERE id = ?`,
			id,
		).Scan(&conv.ID, &conv.Model, &conv.CreatedAt, &conv.UpdatedAt)
		if err == nil {
			rows, err := ms.db.Query(`
				SELECT role, content, images FROM messages WHERE conversation_id = ? ORDER BY id ASC`,
				id,
			)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var msg api.Message
					var imagesStr sql.NullString
					if err := rows.Scan(&msg.Role, &msg.Content, &imagesStr); err == nil {
						if imagesStr.Valid && imagesStr.String != "" {
							_ = json.Unmarshal([]byte(imagesStr.String), &msg.Images)
						}
						conv.Messages = append(conv.Messages, msg)
					}
				}
				return &conv, nil
			}
		}
	}

	jsonPath := filepath.Join(ms.dir, id+".json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}
	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

// List returns a list of summaries for all saved conversations with pagination.
func (ms *MemoryStore) List(limit, offset int) ([]ConversationSummary, int, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.db != nil {
		var total int
		_ = ms.db.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&total)

		rows, err := ms.db.Query(`
			SELECT c.id, c.model, c.title, COUNT(m.id), c.created_at, c.updated_at
			FROM conversations c
			LEFT JOIN messages m ON c.id = m.conversation_id
			GROUP BY c.id
			ORDER BY c.updated_at DESC
			LIMIT ? OFFSET ?`, limit, offset,
		)
		if err == nil {
			defer rows.Close()
			var summaries []ConversationSummary
			for rows.Next() {
				var s ConversationSummary
				if err := rows.Scan(&s.ID, &s.Model, &s.Title, &s.Messages, &s.CreatedAt, &s.UpdatedAt); err == nil {
					summaries = append(summaries, s)
				}
			}
			return summaries, total, nil
		}
	}

	files, err := os.ReadDir(ms.dir)
	if err != nil {
		return nil, 0, err
	}

	var summaries []ConversationSummary
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ms.dir, f.Name()))
		if err != nil {
			continue
		}
		var conv Conversation
		if err := json.Unmarshal(data, &conv); err != nil {
			continue
		}
		title := "New Conversation"
		for _, m := range conv.Messages {
			if m.Role == "user" && m.Content != "" {
				title = m.Content
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				break
			}
		}
		summaries = append(summaries, ConversationSummary{
			ID:        conv.ID,
			Model:     conv.Model,
			Title:     title,
			Messages:  len(conv.Messages),
			CreatedAt: conv.CreatedAt,
			UpdatedAt: conv.UpdatedAt,
		})
	}

	total := len(summaries)
	end := offset + limit
	if offset > len(summaries) {
		summaries = []ConversationSummary{}
	} else if end > len(summaries) {
		summaries = summaries[offset:]
	} else {
		summaries = summaries[offset:end]
	}
	return summaries, total, nil
}

// Delete removes a conversation.
func (ms *MemoryStore) Delete(id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.db != nil {
		if _, err := ms.db.Exec("DELETE FROM conversations WHERE id = ?", id); err != nil {
			slog.Warn("memory: failed to delete conversation from sqlite", "error", err, "id", id)
		}
		if _, err := ms.db.Exec("DELETE FROM messages WHERE conversation_id = ?", id); err != nil {
			slog.Warn("memory: failed to delete messages from sqlite", "error", err, "id", id)
		}
	}

	jsonPath := filepath.Join(ms.dir, id+".json")
	return os.Remove(jsonPath)
}

// AppendMessage appends a single message to a conversation.
func (ms *MemoryStore) AppendMessage(id string, msg api.Message) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	now := time.Now()

	// Primary path: SQLite (O(1) insert)
	if ms.db != nil {
		if _, err := ms.db.Exec("UPDATE conversations SET updated_at = ? WHERE id = ?", now, id); err != nil {
			slog.Warn("memory: update timestamp failed", "error", err, "id", id)
		}
		var imagesJSON []byte
		if len(msg.Images) > 0 {
			imagesJSON, _ = json.Marshal(msg.Images)
		}
		if _, err := ms.db.Exec(`
			INSERT INTO messages (conversation_id, role, content, images)
			VALUES (?, ?, ?, ?)`,
			id, msg.Role, msg.Content, string(imagesJSON),
		); err != nil {
			slog.Warn("memory: insert message failed", "error", err, "id", id)
			// Fall through to JSON path
		} else {
			return nil // SQLite succeeded, skip JSON rewrite
		}
	}

	// Fallback: JSON read-modify-write (only when no SQLite)
	jsonPath := filepath.Join(ms.dir, id+".json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("conversation not found: %s", id)
	}
	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return err
	}
	conv.Messages = append(conv.Messages, msg)
	conv.UpdatedAt = now
	jsonData, err := json.MarshalIndent(conv, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		slog.Warn("memory: failed to write json backup", "error", err, "path", jsonPath)
		return err
	}
	return nil
}

func (ms *MemoryStore) runMigrations(db *sql.DB) error {
	_, _ = db.Exec("CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL)")

	var version int
	row := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version")
	if err := row.Scan(&version); err != nil {
		version = 0
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, ""}, // v1 = initial schema, already created
		{2, "ALTER TABLE messages ADD COLUMN thinking TEXT DEFAULT ''"},
		{3, "ALTER TABLE messages ADD COLUMN tool_calls TEXT DEFAULT ''"},
		{4, "ALTER TABLE conversations ADD COLUMN tags TEXT DEFAULT ''"},
	}

	for _, m := range migrations {
		if m.version > version && m.sql != "" {
			if _, err := db.Exec(m.sql); err != nil {
				if !strings.Contains(err.Error(), "duplicate column") {
					return fmt.Errorf("migration v%d failed: %w", m.version, err)
				}
			}
			_, _ = db.Exec("INSERT INTO schema_version (version) VALUES (?)", m.version)
		} else if m.version > version {
			_, _ = db.Exec("INSERT INTO schema_version (version) VALUES (?)", m.version)
		}
	}
	return nil
}

// Search queries conversations matching text in messages.
func (ms *MemoryStore) Search(query string, limit, offset int) ([]ConversationSummary, int, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.db != nil {
		var total int
		_ = ms.db.QueryRow(`
			SELECT COUNT(DISTINCT c.id) FROM conversations c
			JOIN messages m ON c.id = m.conversation_id
			WHERE m.content LIKE '%' || ? || '%'`, query).Scan(&total)

		rows, err := ms.db.Query(`
			SELECT c.id, c.model, c.title, COUNT(m.id), c.created_at, c.updated_at
			FROM conversations c
			JOIN messages m ON c.id = m.conversation_id
			WHERE m.content LIKE '%' || ? || '%'
			GROUP BY c.id
			ORDER BY c.updated_at DESC
			LIMIT ? OFFSET ?`, query, limit, offset)
		if err == nil {
			defer rows.Close()
			var summaries []ConversationSummary
			for rows.Next() {
				var s ConversationSummary
				if err := rows.Scan(&s.ID, &s.Model, &s.Title, &s.Messages, &s.CreatedAt, &s.UpdatedAt); err == nil {
					summaries = append(summaries, s)
				}
			}
			return summaries, total, nil
		}
	}

	// JSON fallback: linear scan (slow but functional)
	files, err := os.ReadDir(ms.dir)
	if err != nil {
		return nil, 0, err
	}

	var summaries []ConversationSummary
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(ms.dir, f.Name()))
		if err != nil {
			continue
		}
		var conv Conversation
		if err := json.Unmarshal(data, &conv); err != nil {
			continue
		}

		matched := false
		lowerQuery := strings.ToLower(query)
		for _, m := range conv.Messages {
			if strings.Contains(strings.ToLower(m.Content), lowerQuery) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		title := "New Conversation"
		for _, m := range conv.Messages {
			if m.Role == "user" && m.Content != "" {
				title = m.Content
				if len(title) > 50 {
					title = title[:47] + "..."
				}
				break
			}
		}
		summaries = append(summaries, ConversationSummary{
			ID:        conv.ID,
			Model:     conv.Model,
			Title:     title,
			Messages:  len(conv.Messages),
			CreatedAt: conv.CreatedAt,
			UpdatedAt: conv.UpdatedAt,
		})
	}

	total := len(summaries)
	end := offset + limit
	if offset > len(summaries) {
		summaries = []ConversationSummary{}
	} else if end > len(summaries) {
		summaries = summaries[offset:]
	} else {
		summaries = summaries[offset:end]
	}
	return summaries, total, nil
}

// Export marshals a conversation to JSON bytes.
func (ms *MemoryStore) Export(id string) ([]byte, error) {
	conv, err := ms.Load(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(conv)
}

// Import unmarshals a conversation from JSON bytes and saves it.
func (ms *MemoryStore) Import(data []byte) (*Conversation, error) {
	var conv Conversation
	if err := json.Unmarshal(data, &conv); err != nil {
		return nil, err
	}
	if conv.ID == "" {
		return nil, errors.New("conversation ID is required")
	}
	if conv.Model == "" {
		return nil, errors.New("conversation model is required")
	}
	if conv.CreatedAt.IsZero() {
		conv.CreatedAt = time.Now()
	}
	if conv.UpdatedAt.IsZero() {
		conv.UpdatedAt = time.Now()
	}
	if err := ms.Save(&conv); err != nil {
		return nil, err
	}
	return &conv, nil
}
