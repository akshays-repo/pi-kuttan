// internal/storage/sqlite.go

package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Database struct {
	db *sql.DB
}

type Reminder struct {
	ID            int64
	ChatID        int64
	Type          string
	Interval      int    // in minutes
	Status        string
	Message       string
	CreatedAt     time.Time
	LastTriggered sql.NullTime
	NextTrigger   sql.NullTime
}

type Migration struct {
	Version    int
	Name       string
	SQL        string
}

func NewDatabase(dbPath string) (*Database, error) {
	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	database := &Database{
		db: db,
	}

	// Always run migrate to check for and apply new migrations
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("error running migrations: %w", err)
	}

	return database, nil
}

func (d *Database) initMigrationTable() error {
	// Create migration tracking table
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := d.db.Exec(query)
	return err
}

func (d *Database) loadMigrations() ([]Migration, error) {
	// Read embedded migrations
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("error reading embedded migrations: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			// Parse version from filename (format: 001_name.sql)
			parts := strings.SplitN(entry.Name(), "_", 2)
			if len(parts) != 2 {
				continue
			}

			version, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}

			// Read migration content from embedded file
			content, err := migrationsFS.ReadFile(fmt.Sprintf("migrations/%s", entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("error reading migration %s: %w", entry.Name(), err)
			}

			migrations = append(migrations, Migration{
				Version: version,
				Name:    strings.TrimSuffix(parts[1], ".sql"),
				SQL:     string(content),
			})
		}
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (d *Database) getAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)
	
	rows, err := d.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("error querying applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("error scanning migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

func (d *Database) migrate() error {
	// First ensure migration table exists
	if err := d.initMigrationTable(); err != nil {
		return fmt.Errorf("error initializing migration table: %w", err)
	}

	// Get applied migrations
	applied, err := d.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("error getting applied migrations: %w", err)
	}

	// Load all available migrations
	migrations, err := d.loadMigrations()
	if err != nil {
		return fmt.Errorf("error loading migrations: %w", err)
	}

	// Begin transaction
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Apply each pending migration
	for _, migration := range migrations {
		if !applied[migration.Version] {
			fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Name)

			// Execute migration
			if _, err := tx.Exec(migration.SQL); err != nil {
				return fmt.Errorf("error executing migration %d: %w", migration.Version, err)
			}

			// Record migration as applied
			if err := d.recordMigration(tx, migration); err != nil {
				return fmt.Errorf("error recording migration %d: %w", migration.Version, err)
			}
		}
	}

	// Commit all migrations
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing migrations: %w", err)
	}

	return nil
}

func (d *Database) recordMigration(tx *sql.Tx, migration Migration) error {
	_, err := tx.Exec(
		"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
		migration.Version,
		migration.Name,
	)
	return err
}

// CreateReminder inserts a new reminder into the database
func (d *Database) CreateReminder(chatID int64, interval int, message string) (*Reminder, error) {
	query := `
		INSERT INTO reminders (
			chat_id, type, interval, status, message, created_at, next_trigger
		) VALUES (?, ?, ?, 'active', ?, CURRENT_TIMESTAMP, datetime('now', ?))
	`
	
	// Calculate next trigger time
	nextTriggerMinutes := fmt.Sprintf("+%d minutes", interval)
	
	result, err := d.db.Exec(query, chatID, "custom", interval, message, nextTriggerMinutes)
	if err != nil {
		return nil, fmt.Errorf("error creating reminder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting last insert id: %w", err)
	}

	return d.GetReminder(id)
}

// GetReminder retrieves a reminder by ID
func (d *Database) GetReminder(id int64) (*Reminder, error) {
	query := `
		SELECT id, chat_id, type, interval, status, message, 
			   created_at, last_triggered, next_trigger
		FROM reminders
		WHERE id = ?
	`
	
	reminder := &Reminder{}
	err := d.db.QueryRow(query, id).Scan(
		&reminder.ID,
		&reminder.ChatID,
		&reminder.Type,
		&reminder.Interval,
		&reminder.Status,
		&reminder.Message,
		&reminder.CreatedAt,
		&reminder.LastTriggered,
		&reminder.NextTrigger,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting reminder: %w", err)
	}

	return reminder, nil
}

// UpdateReminderStatus updates the status of a reminder
func (d *Database) UpdateReminderStatus(id int64, status string) error {
	query := `UPDATE reminders SET status = ? WHERE id = ?`
	
	result, err := d.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("error updating reminder status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reminder not found")
	}

	return nil
}

// UpdateReminderTrigger updates the last_triggered and next_trigger times
func (d *Database) UpdateReminderTrigger(id int64) error {
	query := `
		UPDATE reminders 
		SET last_triggered = CURRENT_TIMESTAMP,
			next_trigger = datetime('now', ? || ' minutes')
		WHERE id = ?
	`
	
	// Get the reminder's interval
	reminder, err := d.GetReminder(id)
	if err != nil {
		return fmt.Errorf("error getting reminder: %w", err)
	}
	
	result, err := d.db.Exec(query, reminder.Interval, id)
	if err != nil {
		return fmt.Errorf("error updating reminder trigger: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reminder not found")
	}

	return nil
}

// ListActiveReminders returns all active reminders for a chat
func (d *Database) ListActiveReminders(chatID int64) ([]*Reminder, error) {
	query := `
		SELECT id, chat_id, type, interval, status, message, 
			   created_at, last_triggered, next_trigger
		FROM reminders
		WHERE chat_id = ? AND status = 'active'
		ORDER BY next_trigger ASC
	`
	
	rows, err := d.db.Query(query, chatID)
	if err != nil {
		return nil, fmt.Errorf("error querying reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*Reminder
	for rows.Next() {
		reminder := &Reminder{}
		err := rows.Scan(
			&reminder.ID,
			&reminder.ChatID,
			&reminder.Type,
			&reminder.Interval,
			&reminder.Status,
			&reminder.Message,
			&reminder.CreatedAt,
			&reminder.LastTriggered,
			&reminder.NextTrigger,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning reminder: %w", err)
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

// DeleteReminder deletes a reminder and its history
func (d *Database) DeleteReminder(id int64) error {
	query := `DELETE FROM reminders WHERE id = ?`
	
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting reminder: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("reminder not found")
	}

	return nil
}

// AddReminderHistory adds a history entry for a reminder
func (d *Database) AddReminderHistory(reminderID int64, status string) error {
	query := `
		INSERT INTO reminder_history (reminder_id, status)
		VALUES (?, ?)
	`
	
	_, err := d.db.Exec(query, reminderID, status)
	if err != nil {
		return fmt.Errorf("error adding reminder history: %w", err)
	}

	return nil
}

// UpdateReminderInterval updates the interval of a reminder
func (d *Database) UpdateReminderInterval(id int64, interval int) error {
	query := `UPDATE reminders SET interval = ? WHERE id = ?`
	_, err := d.db.Exec(query, interval, id)
	return err
}

// GetAllActiveReminders returns all active reminders in the database
func (d *Database) GetAllActiveReminders() ([]*Reminder, error) {
	query := `
		SELECT id, chat_id, type, interval, status, message, 
			   created_at, last_triggered, next_trigger
		FROM reminders
		WHERE status = 'active'
		ORDER BY next_trigger ASC
	`
	
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying active reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*Reminder
	for rows.Next() {
		reminder := &Reminder{}
		err := rows.Scan(
			&reminder.ID,
			&reminder.ChatID,
			&reminder.Type,
			&reminder.Interval,
			&reminder.Status,
			&reminder.Message,
			&reminder.CreatedAt,
			&reminder.LastTriggered,
			&reminder.NextTrigger,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning reminder: %w", err)
		}
		reminders = append(reminders, reminder)
	}

	return reminders, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}