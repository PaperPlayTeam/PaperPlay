package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID            string    `json:"id" gorm:"primaryKey;type:text"`
	Email         string    `json:"email" gorm:"unique;not null;type:text"`
	PasswordHash  string    `json:"-" gorm:"not null;type:text"` // Never expose password hash in JSON
	DisplayName   string    `json:"display_name" gorm:"type:text"`
	AvatarURL     string    `json:"avatar_url" gorm:"type:text"`
	EthAddress    string    `json:"eth_address" gorm:"type:text"` // Ethereum wallet address
	EthPrivateKey string    `json:"-" gorm:"type:text"`           // Never expose private key in JSON
	CreatedAt     time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	RefreshTokens []RefreshToken    `json:"-" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Progress      []UserProgress    `json:"progress,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Attempts      []UserAttempts    `json:"attempts,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Achievements  []UserAchievement `json:"achievements,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	NFTAssets     []NFTAsset        `json:"nft_assets,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID and sets up user before creation
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// SetPassword hashes and sets the password
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// RefreshToken represents a JWT refresh token
type RefreshToken struct {
	Token     string    `json:"token" gorm:"primaryKey;type:text"`
	UserID    string    `json:"user_id" gorm:"not null;type:text;index"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	// Associations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID for refresh token
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.Token == "" {
		rt.Token = uuid.New().String()
	}
	return nil
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// UserProgress represents user progress on levels
type UserProgress struct {
	UserID        string     `json:"user_id" gorm:"primaryKey;type:text"`
	LevelID       string     `json:"level_id" gorm:"primaryKey;type:text"`
	Status        int        `json:"status" gorm:"not null"` // 0=未开始 1=进行中 2=已通过
	Score         int        `json:"score" gorm:"default:0"` // Latest score
	Stars         int        `json:"stars" gorm:"default:0"` // Star rating 0-3
	LastAttemptAt *time.Time `json:"last_attempt_at" gorm:"type:datetime"`
	CreatedAt     time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"not null"`

	// Associations
	User  *User  `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Level *Level `json:"level,omitempty" gorm:"foreignKey:LevelID;constraint:OnDelete:CASCADE"`
}

// Progress status constants
const (
	ProgressNotStarted = 0
	ProgressInProgress = 1
	ProgressCompleted  = 2
)

// UserAttempts represents daily user statistics and attempts
type UserAttempts struct {
	StatDate                string    `json:"stat_date" gorm:"primaryKey;type:text"` // YYYY-MM-DD format
	UserID                  string    `json:"user_id" gorm:"primaryKey;type:text"`
	AttemptsTotal           int       `json:"attempts_total" gorm:"default:0"`
	AttemptsCorrect         int       `json:"attempts_correct" gorm:"default:0"`
	AttemptsFirstTryCorrect int       `json:"attempts_first_try_correct" gorm:"default:0"`
	CorrectRate             float64   `json:"correct_rate" gorm:"default:0"`
	FirstTryCorrectRate     float64   `json:"first_try_correct_rate" gorm:"default:0"`
	GiveupCount             int       `json:"giveup_count" gorm:"default:0"`
	SkipRate                float64   `json:"skip_rate" gorm:"default:0"`
	AvgDurationMs           int       `json:"avg_duration_ms" gorm:"default:0"`
	TotalTimeMs             int       `json:"total_time_ms" gorm:"default:0"`
	SessionsCount           int       `json:"sessions_count" gorm:"default:0"`
	StreakDays              int       `json:"streak_days" gorm:"default:0"`
	ReviewDueCount          int       `json:"review_due_count" gorm:"default:0"`
	RetentionScore          float64   `json:"retention_score" gorm:"default:0"`
	PeakHour                int       `json:"peak_hour" gorm:"default:0"` // 0-23
	UpdatedAt               time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// UpdateRates calculates and updates the rate fields
func (ua *UserAttempts) UpdateRates() {
	if ua.AttemptsTotal > 0 {
		ua.CorrectRate = float64(ua.AttemptsCorrect) / float64(ua.AttemptsTotal)
		ua.FirstTryCorrectRate = float64(ua.AttemptsFirstTryCorrect) / float64(ua.AttemptsTotal)
		ua.SkipRate = float64(ua.GiveupCount) / float64(ua.AttemptsTotal)
	} else {
		ua.CorrectRate = 0
		ua.FirstTryCorrectRate = 0
		ua.SkipRate = 0
	}
}

// GetTodayStats returns today's stats for a user, creating if not exists
func GetTodayStats(db *gorm.DB, userID string) (*UserAttempts, error) {
	today := time.Now().Format("2006-01-02")

	var stats UserAttempts
	err := db.Where("stat_date = ? AND user_id = ?", today, userID).First(&stats).Error

	if err == gorm.ErrRecordNotFound {
		// Create new stats for today
		stats = UserAttempts{
			StatDate:  today,
			UserID:    userID,
			UpdatedAt: time.Now(),
		}
		if err := db.Create(&stats).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &stats, nil
}

// AddAttempt adds a question attempt to today's stats
func (ua *UserAttempts) AddAttempt(db *gorm.DB, correct bool, firstTry bool, durationMs int) error {
	ua.AttemptsTotal++
	ua.TotalTimeMs += durationMs

	if correct {
		ua.AttemptsCorrect++
		if firstTry {
			ua.AttemptsFirstTryCorrect++
		}
	}

	// Update average duration
	ua.AvgDurationMs = ua.TotalTimeMs / ua.AttemptsTotal

	// Update rates
	ua.UpdateRates()
	ua.UpdatedAt = time.Now()

	return db.Save(ua).Error
}

// AddGiveup adds a giveup/skip to today's stats
func (ua *UserAttempts) AddGiveup(db *gorm.DB) error {
	ua.GiveupCount++
	ua.UpdateRates()
	ua.UpdatedAt = time.Now()

	return db.Save(ua).Error
}
