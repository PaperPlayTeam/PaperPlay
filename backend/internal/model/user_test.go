package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupModelTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto migrate the schema
	db.AutoMigrate(&User{}, &UserAttempts{})

	return db
}

func TestUser_SetPassword(t *testing.T) {
	user := &User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	password := "testpassword123"
	err := user.SetPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, password, user.PasswordHash)
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	password := "testpassword123"
	user.SetPassword(password)

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "Correct password",
			password: "testpassword123",
			expected: true,
		},
		{
			name:     "Incorrect password",
			password: "wrongpassword",
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.CheckPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_BeforeCreate(t *testing.T) {
	db := setupModelTestDB()

	user := &User{
		Email:       "test@example.com",
		DisplayName: "Test User",
	}

	// ID should be empty before create
	assert.Empty(t, user.ID)

	err := db.Create(user).Error
	assert.NoError(t, err)

	// ID should be generated after create
	assert.NotEmpty(t, user.ID)
	assert.Len(t, user.ID, 36) // UUID length
}

func TestUserAttempts_AddAttempt(t *testing.T) {
	db := setupModelTestDB()

	userID := "test-user-id"
	today := time.Now().Format("2006-01-02")

	// Create initial attempts record
	attempts := &UserAttempts{
		UserID:   userID,
		StatDate: today,
	}
	db.Create(attempts)

	tests := []struct {
		name        string
		correct     bool
		firstTry    bool
		durationMs  int
		expectError bool
	}{
		{
			name:        "Correct first try",
			correct:     true,
			firstTry:    true,
			durationMs:  30000,
			expectError: false,
		},
		{
			name:        "Incorrect attempt",
			correct:     false,
			firstTry:    false,
			durationMs:  45000,
			expectError: false,
		},
		{
			name:        "Correct retry",
			correct:     true,
			firstTry:    false,
			durationMs:  60000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get fresh attempts record
			var attempts UserAttempts
			db.Where("user_id = ? AND stat_date = ?", userID, today).First(&attempts)

			err := attempts.AddAttempt(db, tt.correct, tt.firstTry, tt.durationMs)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the attempt was recorded
				var updated UserAttempts
				db.Where("user_id = ? AND stat_date = ?", userID, today).First(&updated)

				assert.Greater(t, updated.AttemptsTotal, attempts.AttemptsTotal)
				assert.Greater(t, updated.TotalTimeMs, attempts.TotalTimeMs)

				if tt.correct {
					assert.Greater(t, updated.AttemptsCorrect, attempts.AttemptsCorrect)
				}

				if tt.firstTry && tt.correct {
					assert.Greater(t, updated.AttemptsFirstTryCorrect, attempts.AttemptsFirstTryCorrect)
				}
			}
		})
	}
}

func TestGetTodayStats(t *testing.T) {
	db := setupModelTestDB()

	userID := "test-user-id"
	today := time.Now().Format("2006-01-02")

	// Test getting stats when none exist
	stats, err := GetTodayStats(db, userID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, userID, stats.UserID)
	assert.Equal(t, today, stats.StatDate)
	assert.Equal(t, 0, stats.AttemptsTotal)

	// Test getting existing stats
	existingStats := &UserAttempts{
		UserID:          userID,
		StatDate:        today,
		AttemptsTotal:   10,
		AttemptsCorrect: 8,
	}
	db.Create(existingStats)

	stats, err = GetTodayStats(db, userID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 10, stats.AttemptsTotal)
	assert.Equal(t, 8, stats.AttemptsCorrect)
}

func TestUserAttempts_UpdateRates(t *testing.T) {
	attempts := &UserAttempts{
		AttemptsTotal:           10,
		AttemptsCorrect:         8,
		AttemptsFirstTryCorrect: 6,
		GiveupCount:             1,
	}

	attempts.UpdateRates()

	assert.Equal(t, 0.8, attempts.CorrectRate)
	assert.Equal(t, 0.75, attempts.FirstTryCorrectRate) // 6/8 correct answers
	assert.Equal(t, 0.1, attempts.SkipRate)             // 1/10 total attempts
}

// Note: UpdateAverages method tests removed due to method not being available in the model
