package service

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"paperplay/internal/model"
	"paperplay/internal/websocket"
)

// AchievementService handles achievement evaluation and management
type AchievementService struct {
	db              *gorm.DB
	logger          *zap.Logger
	wsHub           *websocket.Hub
	ethereumService *EthereumService
}

// NewAchievementService creates a new achievement service
func NewAchievementService(db *gorm.DB, logger *zap.Logger, wsHub *websocket.Hub, ethService *EthereumService) *AchievementService {
	return &AchievementService{
		db:              db,
		logger:          logger,
		wsHub:           wsHub,
		ethereumService: ethService,
	}
}

// EvaluateUserAchievements checks and awards achievements for a user
func (s *AchievementService) EvaluateUserAchievements(userID string) error {
	// Get all active achievements
	var achievements []model.Achievement
	if err := s.db.Where("is_active = ?", true).Find(&achievements).Error; err != nil {
		return fmt.Errorf("failed to get achievements: %w", err)
	}

	// Get user's current achievements to avoid duplicates
	var userAchievements []model.UserAchievement
	if err := s.db.Where("user_id = ?", userID).Find(&userAchievements).Error; err != nil {
		return fmt.Errorf("failed to get user achievements: %w", err)
	}

	earnedAchievementIDs := make(map[string]bool)
	for _, ua := range userAchievements {
		earnedAchievementIDs[ua.AchievementID] = true
	}

	// Evaluate each achievement
	for _, achievement := range achievements {
		if earnedAchievementIDs[achievement.ID] {
			continue // User already has this achievement
		}

		earned, err := s.evaluateAchievement(userID, &achievement)
		if err != nil {
			s.logger.Error("Failed to evaluate achievement",
				zap.String("achievement_id", achievement.ID),
				zap.String("user_id", userID),
				zap.Error(err),
			)
			continue
		}

		if earned {
			if err := s.awardAchievement(userID, &achievement); err != nil {
				s.logger.Error("Failed to award achievement",
					zap.String("achievement_id", achievement.ID),
					zap.String("user_id", userID),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}

// evaluateAchievement checks if a user meets the criteria for an achievement
func (s *AchievementService) evaluateAchievement(userID string, achievement *model.Achievement) (bool, error) {
	rule, err := achievement.GetRule()
	if err != nil {
		return false, fmt.Errorf("failed to parse achievement rule: %w", err)
	}

	switch rule.Type {
	case "first_try":
		return s.evaluateFirstTryAchievement(userID, rule)
	case "streak":
		return s.evaluateStreakAchievement(userID, rule)
	case "speed_accuracy":
		return s.evaluateSpeedAccuracyAchievement(userID, rule)
	case "endurance":
		return s.evaluateEnduranceAchievement(userID, rule)
	case "memory":
		return s.evaluateMemoryAchievement(userID, rule)
	default:
		s.logger.Warn("Unknown achievement rule type", zap.String("type", rule.Type))
		return false, nil
	}
}

// evaluateFirstTryAchievement evaluates "首战告捷" achievement
func (s *AchievementService) evaluateFirstTryAchievement(userID string, rule *model.AchievementRule) (bool, error) {
	var todayStats model.UserAttempts
	today := time.Now().Format("2006-01-02")

	err := s.db.Where("user_id = ? AND stat_date = ?", userID, today).First(&todayStats).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(&todayStats, condition) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateStreakAchievement evaluates "连击达人" achievement
func (s *AchievementService) evaluateStreakAchievement(userID string, rule *model.AchievementRule) (bool, error) {
	var latestStats model.UserAttempts
	err := s.db.Where("user_id = ?", userID).
		Order("stat_date DESC").
		First(&latestStats).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(&latestStats, condition) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateSpeedAccuracyAchievement evaluates "智速双全" achievement
func (s *AchievementService) evaluateSpeedAccuracyAchievement(userID string, rule *model.AchievementRule) (bool, error) {
	var todayStats model.UserAttempts
	today := time.Now().Format("2006-01-02")

	err := s.db.Where("user_id = ? AND stat_date = ?", userID, today).First(&todayStats).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	// Need at least some attempts to qualify
	if todayStats.AttemptsTotal < 10 {
		return false, nil
	}

	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(&todayStats, condition) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateEnduranceAchievement evaluates "持久战士" achievement
func (s *AchievementService) evaluateEnduranceAchievement(userID string, rule *model.AchievementRule) (bool, error) {
	var todayStats model.UserAttempts
	today := time.Now().Format("2006-01-02")

	err := s.db.Where("user_id = ? AND stat_date = ?", userID, today).First(&todayStats).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(&todayStats, condition) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateMemoryAchievement evaluates "记忆大师" achievement
func (s *AchievementService) evaluateMemoryAchievement(userID string, rule *model.AchievementRule) (bool, error) {
	var todayStats model.UserAttempts
	today := time.Now().Format("2006-01-02")

	err := s.db.Where("user_id = ? AND stat_date = ?", userID, today).First(&todayStats).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(&todayStats, condition) {
			return false, nil
		}
	}

	return true, nil
}

// evaluateCondition evaluates a single condition against user stats
func (s *AchievementService) evaluateCondition(stats *model.UserAttempts, condition model.ConditionRule) bool {
	var value interface{}

	// Get the field value from stats
	switch condition.Field {
	case "attempts_total":
		value = stats.AttemptsTotal
	case "attempts_correct":
		value = stats.AttemptsCorrect
	case "attempts_first_try_correct":
		value = stats.AttemptsFirstTryCorrect
	case "correct_rate":
		value = stats.CorrectRate
	case "first_try_correct_rate":
		value = stats.FirstTryCorrectRate
	case "giveup_count":
		value = stats.GiveupCount
	case "skip_rate":
		value = stats.SkipRate
	case "avg_duration_ms":
		value = stats.AvgDurationMs
	case "total_time_ms":
		value = stats.TotalTimeMs
	case "sessions_count":
		value = stats.SessionsCount
	case "streak_days":
		value = stats.StreakDays
	case "review_due_count":
		value = stats.ReviewDueCount
	case "retention_score":
		value = stats.RetentionScore
	default:
		s.logger.Warn("Unknown condition field", zap.String("field", condition.Field))
		return false
	}

	return s.compareValues(value, condition.Operator, condition.Value)
}

// compareValues compares two values using the specified operator
func (s *AchievementService) compareValues(actual interface{}, operator string, expected interface{}) bool {
	switch operator {
	case ">=":
		return s.numericCompare(actual, expected) >= 0
	case ">":
		return s.numericCompare(actual, expected) > 0
	case "<=":
		return s.numericCompare(actual, expected) <= 0
	case "<":
		return s.numericCompare(actual, expected) < 0
	case "=", "==":
		return s.numericCompare(actual, expected) == 0
	case "!=":
		return s.numericCompare(actual, expected) != 0
	default:
		s.logger.Warn("Unknown operator", zap.String("operator", operator))
		return false
	}
}

// numericCompare compares two numeric values, returns -1, 0, or 1
func (s *AchievementService) numericCompare(a, b interface{}) int {
	af := s.toFloat64(a)
	bf := s.toFloat64(b)

	if af < bf {
		return -1
	} else if af > bf {
		return 1
	}
	return 0
}

// toFloat64 converts various numeric types to float64
func (s *AchievementService) toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

// awardAchievement awards an achievement to a user
func (s *AchievementService) awardAchievement(userID string, achievement *model.Achievement) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Create user achievement record
		userAchievement := &model.UserAchievement{
			UserID:        userID,
			AchievementID: achievement.ID,
			EarnedAt:      time.Now(),
			Progress:      1.0,
		}

		if err := tx.Create(userAchievement).Error; err != nil {
			return fmt.Errorf("failed to create user achievement: %w", err)
		}

		// Create event log
		event := &model.Event{
			UserID:    userID,
			EventType: model.EventAchievementEarned,
		}

		eventData := map[string]interface{}{
			"achievement_id":   achievement.ID,
			"achievement_name": achievement.Name,
			"level":            achievement.Level,
		}

		if err := event.SetData(eventData); err != nil {
			return fmt.Errorf("failed to set event data: %w", err)
		}

		if err := tx.Create(event).Error; err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		s.logger.Info("Achievement awarded",
			zap.String("user_id", userID),
			zap.String("achievement_id", achievement.ID),
			zap.String("achievement_name", achievement.Name),
		)

		// Send real-time notification via WebSocket
		if s.wsHub != nil {
			notification := &websocket.NotificationMessage{
				ID:      userAchievement.ID,
				Type:    "achievement",
				Title:   "成就解锁！",
				Message: fmt.Sprintf("恭喜您获得成就：%s", achievement.Name),
				Achievement: &struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Level       int    `json:"level"`
					IconURL     string `json:"icon_url"`
				}{
					ID:          achievement.ID,
					Name:        achievement.Name,
					Description: achievement.Description,
					Level:       achievement.Level,
					IconURL:     achievement.IconURL,
				},
			}
			s.wsHub.SendNotification(userID, notification)
		}

		// Create NFT if enabled
		if achievement.NFTEnabled && s.ethereumService != nil && s.ethereumService.IsEnabled() {
			if err := s.createAchievementNFT(tx, userID, achievement); err != nil {
				s.logger.Error("Failed to create NFT for achievement",
					zap.String("user_id", userID),
					zap.String("achievement_id", achievement.ID),
					zap.Error(err),
				)
				// Don't fail the transaction for NFT creation errors
			}
		}

		return nil
	})
}

// createAchievementNFT creates an NFT for an achievement
func (s *AchievementService) createAchievementNFT(tx *gorm.DB, userID string, achievement *model.Achievement) error {
	// Get user's ethereum address
	var user model.User
	if err := tx.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.EthAddress == "" {
		return fmt.Errorf("user has no ethereum address")
	}

	// Generate metadata URI
	metadataURI, err := s.ethereumService.GenerateMetadataURI(achievement, userID)
	if err != nil {
		return fmt.Errorf("failed to generate metadata URI: %w", err)
	}

	// Create NFT asset record
	nftAsset, err := s.ethereumService.CreateNFTAsset(userID, &achievement.ID, metadataURI)
	if err != nil {
		return fmt.Errorf("failed to create NFT asset: %w", err)
	}

	// Save NFT asset to database
	if err := tx.Create(nftAsset).Error; err != nil {
		return fmt.Errorf("failed to save NFT asset: %w", err)
	}

	// Queue NFT minting (in a real implementation, this would be done asynchronously)
	go func() {
		mintRequest := NFTMintRequest{
			ToAddress: user.EthAddress,
			TokenURI:  metadataURI,
			Metadata: map[string]interface{}{
				"achievement_id": achievement.ID,
				"user_id":        userID,
			},
		}

		result, err := s.ethereumService.MintNFT(mintRequest)
		if err != nil {
			s.logger.Error("Failed to mint NFT",
				zap.String("nft_asset_id", nftAsset.ID),
				zap.Error(err),
			)

			// Mark NFT as failed
			s.db.Model(nftAsset).Update("status", model.NFTStatusFailed)
			return
		}

		// Update NFT asset with minting results
		s.ethereumService.UpdateNFTAsset(nftAsset, result)
		s.db.Save(nftAsset)

		s.logger.Info("NFT minted successfully",
			zap.String("nft_asset_id", nftAsset.ID),
			zap.String("token_id", result.TokenID),
			zap.String("tx_hash", result.TransactionHash),
		)
	}()

	return nil
}

// EvaluateOnEvent evaluates achievements when specific events occur
func (s *AchievementService) EvaluateOnEvent(userID string, eventType string, eventData map[string]interface{}) error {
	// Update user stats based on event
	if err := s.updateUserStatsFromEvent(userID, eventType, eventData); err != nil {
		s.logger.Error("Failed to update user stats from event",
			zap.String("user_id", userID),
			zap.String("event_type", eventType),
			zap.Error(err),
		)
	}

	// Evaluate achievements
	return s.EvaluateUserAchievements(userID)
}

// updateUserStatsFromEvent updates user statistics based on events
func (s *AchievementService) updateUserStatsFromEvent(userID string, eventType string, eventData map[string]interface{}) error {
	stats, err := model.GetTodayStats(s.db, userID)
	if err != nil {
		return fmt.Errorf("failed to get today's stats: %w", err)
	}

	switch eventType {
	case model.EventQuestionAnswered:
		if correct, ok := eventData["correct"].(bool); ok {
			if firstTry, ok := eventData["first_try"].(bool); ok {
				if duration, ok := eventData["duration_ms"].(float64); ok {
					return stats.AddAttempt(s.db, correct, firstTry, int(duration))
				}
			}
		}

	case model.EventLevelCompleted:
		// Level completion might trigger streak updates
		stats.StreakDays = s.calculateCurrentStreak(userID)
		return s.db.Save(stats).Error

	case "session_started":
		stats.SessionsCount++
		return s.db.Save(stats).Error
	}

	return nil
}

// calculateCurrentStreak calculates the current learning streak for a user
func (s *AchievementService) calculateCurrentStreak(userID string) int {
	var attempts []model.UserAttempts
	err := s.db.Where("user_id = ? AND attempts_total > 0", userID).
		Order("stat_date DESC").
		Limit(30). // Look at last 30 days
		Find(&attempts).Error

	if err != nil {
		s.logger.Error("Failed to get user attempts for streak calculation", zap.Error(err))
		return 0
	}

	if len(attempts) == 0 {
		return 0
	}

	// Calculate consecutive days
	streak := 1

	for i := 1; i < len(attempts); i++ {
		currentDate, _ := time.Parse("2006-01-02", attempts[i-1].StatDate)
		prevDate, _ := time.Parse("2006-01-02", attempts[i].StatDate)

		// Check if dates are consecutive
		if currentDate.AddDate(0, 0, -1).Format("2006-01-02") == prevDate.Format("2006-01-02") {
			streak++
		} else {
			break
		}
	}

	return streak
}

// GetUserAchievements returns all achievements for a user
func (s *AchievementService) GetUserAchievements(userID string) ([]model.UserAchievement, error) {
	var achievements []model.UserAchievement
	err := s.db.Where("user_id = ?", userID).
		Preload("Achievement").
		Order("earned_at DESC").
		Find(&achievements).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
	}

	return achievements, nil
}

// GetAllAchievements returns all available achievements
func (s *AchievementService) GetAllAchievements() ([]model.Achievement, error) {
	var achievements []model.Achievement
	err := s.db.Where("is_active = ?", true).
		Order("level ASC, created_at ASC").
		Find(&achievements).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	return achievements, nil
}
