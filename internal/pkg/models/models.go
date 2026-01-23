package models

import (
	"time"

	"github.com/google/uuid"
)

// User representa un usuario en el sistema
type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Email       string    `json:"email" db:"email"`
	FirebaseUID string    `json:"firebase_uid" db:"firebase_uid"`
	CoinBalance int       `json:"coin_balance" db:"coin_balance"`
	IsPremium   bool      `json:"is_premium" db:"is_premium"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Series representa una serie de micro-dramas
type Series struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	HorizontalPoster string   `json:"horizontal_poster" db:"horizontal_poster"`
	VerticalPoster  string    `json:"vertical_poster" db:"vertical_poster"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Episode representa un episodio de una serie
type Episode struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SeriesID     uuid.UUID `json:"series_id" db:"series_id"`
	EpisodeNumber int      `json:"episode_number" db:"episode_number"`
	Title        string    `json:"title" db:"title"`
	VideoIDBunny string    `json:"video_id_bunny" db:"video_id_bunny"`
	Duration     int       `json:"duration" db:"duration"` // en segundos
	IsFree       bool      `json:"is_free" db:"is_free"`
	PriceCoins   int       `json:"price_coins" db:"price_coins"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Unlock representa el desbloqueo de un episodio por un usuario
type Unlock struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	EpisodeID  uuid.UUID `json:"episode_id" db:"episode_id"`
	Method     string    `json:"method" db:"method"` // COIN, AD, SUB
	UnlockedAt time.Time `json:"unlocked_at" db:"unlocked_at"`
}

// UnlockMethod tipos de desbloqueo
const (
	UnlockMethodCoin = "COIN"
	UnlockMethodAd   = "AD"
	UnlockMethodSub  = "SUB"
)

