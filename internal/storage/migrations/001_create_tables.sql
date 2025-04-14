-- migrations/001_create_tables.sql

-- First create the reminders table
CREATE TABLE IF NOT EXISTS reminders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('custom')),
    interval INTEGER NOT NULL, -- in minutes
    status TEXT NOT NULL CHECK(status IN ('active', 'paused', 'stopped')) DEFAULT 'active',
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_triggered TIMESTAMP,
    next_trigger TIMESTAMP
);

-- Create index separately (correct SQLite syntax)
CREATE INDEX IF NOT EXISTS idx_chat_status ON reminders(chat_id, status);

-- Create the history table
CREATE TABLE IF NOT EXISTS reminder_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reminder_id INTEGER NOT NULL,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL,
    FOREIGN KEY(reminder_id) REFERENCES reminders(id) ON DELETE CASCADE
);