package main

import (
	"database/sql"
	"log"
)

func RunMigrations(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS movies (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			file_path VARCHAR(500),
			thumbnail_path VARCHAR(500),
			duration INT,
			owner_id INT REFERENCES users(id),
			content_type VARCHAR(20) DEFAULT 'movie',
			parent_id INT REFERENCES movies(id) ON DELETE CASCADE,
			season_number INT,
			episode_number INT,
			cast_members JSONB DEFAULT '[]',
			director VARCHAR(255),
			release_year INT,
			genres JSONB DEFAULT '[]',
			tags JSONB DEFAULT '[]',
			mood VARCHAR(100),
			embedding JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS watch_history (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			movie_id INT REFERENCES movies(id) ON DELETE CASCADE,
			progress INT DEFAULT 0,
			last_watched TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_movies_owner ON movies(owner_id)`,
		`CREATE INDEX IF NOT EXISTS idx_watch_history_user ON watch_history(user_id)`,
		`ALTER TABLE watch_history DROP CONSTRAINT IF EXISTS unique_user_movie`,
		`ALTER TABLE watch_history ADD CONSTRAINT IF NOT EXISTS unique_user_movie UNIQUE(user_id, movie_id)`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN content_type VARCHAR(20) DEFAULT 'movie'; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN parent_id INT REFERENCES movies(id) ON DELETE CASCADE; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN season_number INT; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN episode_number INT; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN cast_members JSONB DEFAULT '[]'; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN director VARCHAR(255); EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`DO $$ BEGIN ALTER TABLE movies ADD COLUMN release_year INT; EXCEPTION WHEN duplicate_column THEN NULL; END $$`,
		`ALTER TABLE movies ALTER COLUMN file_path DROP NOT NULL`,
		// drop the old access table if it still exists (cleanup)
		`DROP TABLE IF EXISTS movie_access`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Migration error: %v", err)
		}
	}
	log.Println("Migrations completed")

}
