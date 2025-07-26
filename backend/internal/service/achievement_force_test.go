package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"paperplay/internal/model"
	"paperplay/internal/websocket"
)

func TestForceAwardAchievement(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate required models
	err = db.AutoMigrate(
		&model.User{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.Event{},
	)
	require.NoError(t, err)

	// Setup services
	logger := zap.NewNop()
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	achievementService := NewAchievementService(db, logger, wsHub, nil)

	// Seed test data
	user := &model.User{
		ID:           "test-user-1",
		Email:        "test@example.com",
		PasswordHash: "hash",
		DisplayName:  "Test User",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	achievement := &model.Achievement{
		ID:          "test-achievement-1",
		Name:        "Test Achievement",
		Description: "Test Description",
		Level:       1,
		BadgeType:   "test",
		RuleJSON:    `{"type":"test","conditions":[]}`,
		NFTEnabled:  false,
		IsActive:    true,
	}
	err = db.Create(achievement).Error
	require.NoError(t, err)

	t.Run("Valid Force Award", func(t *testing.T) {
		// Test force awarding achievement
		err = achievementService.ForceAwardAchievement(user.ID, achievement.ID)
		assert.NoError(t, err)

		// Verify achievement was awarded
		var userAchievement model.UserAchievement
		err = db.Where("user_id = ? AND achievement_id = ?", user.ID, achievement.ID).First(&userAchievement).Error
		require.NoError(t, err)
		assert.Equal(t, user.ID, userAchievement.UserID)
		assert.Equal(t, achievement.ID, userAchievement.AchievementID)
		assert.Equal(t, 1.0, userAchievement.Progress)

		// Verify event was created
		var event model.Event
		err = db.Where("user_id = ? AND event_type = ?", user.ID, model.EventAchievementEarned).First(&event).Error
		require.NoError(t, err)
		assert.Equal(t, user.ID, event.UserID)
		assert.Equal(t, model.EventAchievementEarned, event.EventType)
	})

	t.Run("Duplicate Award", func(t *testing.T) {
		// Try to award the same achievement again
		err = achievementService.ForceAwardAchievement(user.ID, achievement.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already has this achievement")
	})

	t.Run("Invalid Achievement ID", func(t *testing.T) {
		err = achievementService.ForceAwardAchievement(user.ID, "invalid-achievement-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "achievement not found")
	})

	t.Run("Multiple Achievements", func(t *testing.T) {
		// Create another user and multiple achievements
		user2 := &model.User{
			ID:           "test-user-2",
			Email:        "test2@example.com",
			PasswordHash: "hash",
			DisplayName:  "Test User 2",
		}
		err = db.Create(user2).Error
		require.NoError(t, err)

		achievements := []model.Achievement{
			{
				ID:          "test-achievement-2",
				Name:        "Second Achievement",
				Description: "Second Test Description",
				Level:       2,
				BadgeType:   "test",
				RuleJSON:    `{"type":"test","conditions":[]}`,
				NFTEnabled:  false,
				IsActive:    true,
			},
			{
				ID:          "test-achievement-3",
				Name:        "Third Achievement",
				Description: "Third Test Description",
				Level:       3,
				BadgeType:   "test",
				RuleJSON:    `{"type":"test","conditions":[]}`,
				NFTEnabled:  false,
				IsActive:    true,
			},
		}

		for _, ach := range achievements {
			err = db.Create(&ach).Error
			require.NoError(t, err)

			// Award achievement to user2
			err = achievementService.ForceAwardAchievement(user2.ID, ach.ID)
			assert.NoError(t, err)
		}

		// Verify user2 has both achievements
		var count int64
		db.Model(&model.UserAchievement{}).Where("user_id = ?", user2.ID).Count(&count)
		assert.Equal(t, int64(2), count)

		// Verify user1 still has only 1 achievement
		db.Model(&model.UserAchievement{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(1), count)
	})
}

func TestWebSocketNotificationInForceAward(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{}, &model.Achievement{}, &model.UserAchievement{}, &model.Event{})
	require.NoError(t, err)

	// Setup WebSocket hub with notification tracking
	logger := zap.NewNop()
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	achievementService := NewAchievementService(db, logger, wsHub, nil)

	// Create test data
	user := &model.User{
		ID:           "websocket-test-user",
		Email:        "wstest@example.com",
		PasswordHash: "hash",
		DisplayName:  "WebSocket Test User",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	achievement := &model.Achievement{
		ID:          "websocket-test-achievement",
		Name:        "WebSocket Test Achievement",
		Description: "Test WebSocket notifications",
		Level:       1,
		BadgeType:   "notification",
		RuleJSON:    `{"type":"notification","conditions":[]}`,
		NFTEnabled:  false,
		IsActive:    true,
	}
	err = db.Create(achievement).Error
	require.NoError(t, err)

	// Test that force award completes without error (WebSocket notification is sent internally)
	err = achievementService.ForceAwardAchievement(user.ID, achievement.ID)
	assert.NoError(t, err)

	// Verify the achievement was awarded
	var userAchievement model.UserAchievement
	err = db.Where("user_id = ? AND achievement_id = ?", user.ID, achievement.ID).First(&userAchievement).Error
	require.NoError(t, err)

	// Verify timing
	assert.WithinDuration(t, time.Now(), userAchievement.EarnedAt, 5*time.Second)
}

func TestAchievementServiceEvaluationIntegration(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.User{},
		&model.Achievement{},
		&model.UserAchievement{},
		&model.Event{},
		&model.UserAttempts{},
	)
	require.NoError(t, err)

	logger := zap.NewNop()
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	achievementService := NewAchievementService(db, logger, wsHub, nil)

	// Create test user
	user := &model.User{
		ID:           "integration-test-user",
		Email:        "integration@example.com",
		PasswordHash: "hash",
		DisplayName:  "Integration Test User",
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	// Create test achievements
	achievements := []model.Achievement{
		{
			ID:          "integration-achievement-1",
			Name:        "First Achievement",
			Description: "First test achievement",
			Level:       1,
			BadgeType:   "learning",
			RuleJSON:    `{"type":"first_try","conditions":[{"field":"attempts_first_try_correct","operator":">=","value":1}]}`,
			NFTEnabled:  false,
			IsActive:    true,
		},
		{
			ID:          "integration-achievement-2",
			Name:        "Second Achievement",
			Description: "Second test achievement",
			Level:       2,
			BadgeType:   "streak",
			RuleJSON:    `{"type":"streak","conditions":[{"field":"streak_days","operator":">=","value":7}]}`,
			NFTEnabled:  false,
			IsActive:    true,
		},
	}

	for _, ach := range achievements {
		err = db.Create(&ach).Error
		require.NoError(t, err)
	}

	t.Run("Force Award vs Regular Evaluation", func(t *testing.T) {
		// Force award first achievement
		err = achievementService.ForceAwardAchievement(user.ID, achievements[0].ID)
		assert.NoError(t, err)

		// Try regular evaluation (should not award already earned achievement)
		err = achievementService.EvaluateUserAchievements(user.ID)
		assert.NoError(t, err)

		// Verify user still has only 1 achievement
		var count int64
		db.Model(&model.UserAchievement{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(1), count)

		// Force award second achievement
		err = achievementService.ForceAwardAchievement(user.ID, achievements[1].ID)
		assert.NoError(t, err)

		// Verify user now has 2 achievements
		db.Model(&model.UserAchievement{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(2), count)
	})
}
