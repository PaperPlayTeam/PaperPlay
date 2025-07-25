package cron

import (
	"fmt"
	"paperplay/config"
	"paperplay/internal/model"
	"paperplay/internal/service"
	"paperplay/internal/websocket"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JobManager handles all scheduled jobs
type JobManager struct {
	cron               *cron.Cron
	db                 *gorm.DB
	logger             *zap.Logger
	config             *config.CronConfig
	achievementService *service.AchievementService
	userService        *service.UserService
	wsHub              *websocket.Hub
}

// NewJobManager creates a new job manager
func NewJobManager(
	db *gorm.DB,
	logger *zap.Logger,
	config *config.CronConfig,
	achievementService *service.AchievementService,
	userService *service.UserService,
	wsHub *websocket.Hub,
) *JobManager {
	c := cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)))

	return &JobManager{
		cron:               c,
		db:                 db,
		logger:             logger,
		config:             config,
		achievementService: achievementService,
		userService:        userService,
		wsHub:              wsHub,
	}
}

// Start starts all scheduled jobs
func (jm *JobManager) Start() error {
	if !jm.config.Enabled {
		jm.logger.Info("Cron jobs disabled")
		return nil
	}

	// Daily stats update job
	if _, err := jm.cron.AddFunc(jm.config.StatsUpdateSpec, jm.dailyStatsUpdate); err != nil {
		return fmt.Errorf("failed to add daily stats update job: %w", err)
	}

	// Weekly report generation job
	if _, err := jm.cron.AddFunc(jm.config.ReportGenerationSpec, jm.weeklyReportGeneration); err != nil {
		return fmt.Errorf("failed to add weekly report generation job: %w", err)
	}

	// Achievement check job
	if _, err := jm.cron.AddFunc(jm.config.AchievementCheckSpec, jm.achievementCheck); err != nil {
		return fmt.Errorf("failed to add achievement check job: %w", err)
	}

	// Start the cron scheduler
	jm.cron.Start()
	jm.logger.Info("Cron jobs started successfully")

	return nil
}

// Stop stops all scheduled jobs
func (jm *JobManager) Stop() {
	if jm.cron != nil {
		ctx := jm.cron.Stop()
		<-ctx.Done()
		jm.logger.Info("Cron jobs stopped")
	}
}

// dailyStatsUpdate updates daily statistics and streaks
func (jm *JobManager) dailyStatsUpdate() {
	jm.logger.Info("Starting daily stats update job")
	startTime := time.Now()

	// Get all users who had activity yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var userIDs []string
	if err := jm.db.Model(&model.UserAttempts{}).
		Where("stat_date = ? AND attempts_total > 0", yesterday).
		Pluck("user_id", &userIDs).Error; err != nil {
		jm.logger.Error("Failed to get users with activity", zap.Error(err))
		return
	}

	successCount := 0
	errorCount := 0

	// Update streak information for each user
	for _, userID := range userIDs {
		if err := jm.updateUserStreak(userID); err != nil {
			jm.logger.Error("Failed to update user streak",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	// Update review recommendations
	if err := jm.updateReviewRecommendations(); err != nil {
		jm.logger.Error("Failed to update review recommendations", zap.Error(err))
	}

	duration := time.Since(startTime)
	jm.logger.Info("Daily stats update job completed",
		zap.Duration("duration", duration),
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)
}

// weeklyReportGeneration generates weekly learning reports
func (jm *JobManager) weeklyReportGeneration() {
	jm.logger.Info("Starting weekly report generation job")
	startTime := time.Now()

	// Get all active users (users with activity in the last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	var userIDs []string
	if err := jm.db.Model(&model.UserAttempts{}).
		Where("stat_date >= ? AND attempts_total > 0", weekAgo).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		jm.logger.Error("Failed to get active users", zap.Error(err))
		return
	}

	successCount := 0
	errorCount := 0

	// Generate report for each user
	for _, userID := range userIDs {
		if err := jm.generateWeeklyReport(userID); err != nil {
			jm.logger.Error("Failed to generate weekly report",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)
	jm.logger.Info("Weekly report generation job completed",
		zap.Duration("duration", duration),
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)
}

// achievementCheck checks and awards achievements for active users
func (jm *JobManager) achievementCheck() {
	jm.logger.Info("Starting achievement check job")
	startTime := time.Now()

	// Get users with recent activity (last 24 hours)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	var userIDs []string
	if err := jm.db.Model(&model.UserAttempts{}).
		Where("(stat_date = ? OR stat_date = ?) AND attempts_total > 0", yesterday, today).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		jm.logger.Error("Failed to get users with recent activity", zap.Error(err))
		return
	}

	successCount := 0
	errorCount := 0

	// Check achievements for each user
	for _, userID := range userIDs {
		if err := jm.achievementService.EvaluateUserAchievements(userID); err != nil {
			jm.logger.Error("Failed to evaluate achievements",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			errorCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)
	jm.logger.Info("Achievement check job completed",
		zap.Duration("duration", duration),
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)
}

// updateUserStreak updates a user's learning streak
func (jm *JobManager) updateUserStreak(userID string) error {
	// Get user's recent attempts to calculate streak
	var attempts []model.UserAttempts
	if err := jm.db.Where("user_id = ? AND attempts_total > 0", userID).
		Order("stat_date DESC").
		Limit(30).
		Find(&attempts).Error; err != nil {
		return fmt.Errorf("failed to get user attempts: %w", err)
	}

	if len(attempts) == 0 {
		return nil
	}

	// Calculate current streak
	streak := jm.calculateStreak(attempts)

	// Update the latest stats record with new streak
	latestStats := &attempts[0]
	latestStats.StreakDays = streak

	if err := jm.db.Save(latestStats).Error; err != nil {
		return fmt.Errorf("failed to update streak: %w", err)
	}

	jm.logger.Debug("Updated user streak",
		zap.String("user_id", userID),
		zap.Int("streak_days", streak),
	)

	return nil
}

// calculateStreak calculates consecutive learning days
func (jm *JobManager) calculateStreak(attempts []model.UserAttempts) int {
	if len(attempts) == 0 {
		return 0
	}

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

// updateReviewRecommendations updates review recommendations based on forgetting curve
func (jm *JobManager) updateReviewRecommendations() error {
	// Simple spaced repetition algorithm
	// Items should be reviewed at: 1 day, 3 days, 7 days, 14 days, 30 days

	reviewIntervals := []int{1, 3, 7, 14, 30}
	today := time.Now()

	for _, interval := range reviewIntervals {
		reviewDate := today.AddDate(0, 0, -interval).Format("2006-01-02")

		// Find levels completed on this date that need review
		var progressRecords []model.UserProgress
		if err := jm.db.Where("status = ? AND DATE(last_attempt_at) = ?",
			model.ProgressCompleted, reviewDate).
			Find(&progressRecords).Error; err != nil {
			continue
		}

		// Update review due count for affected users
		for _, progress := range progressRecords {
			var stats model.UserAttempts
			todayStr := today.Format("2006-01-02")

			if err := jm.db.Where("user_id = ? AND stat_date = ?",
				progress.UserID, todayStr).First(&stats).Error; err == gorm.ErrRecordNotFound {
				// Create today's stats if not exists
				stats = model.UserAttempts{
					StatDate:  todayStr,
					UserID:    progress.UserID,
					UpdatedAt: time.Now(),
				}
				jm.db.Create(&stats)
			}

			stats.ReviewDueCount++
			jm.db.Save(&stats)
		}
	}

	return nil
}

// generateWeeklyReport generates a weekly learning report for a user
func (jm *JobManager) generateWeeklyReport(userID string) error {
	// Get user's weekly stats
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	var weeklyStats []model.UserAttempts
	if err := jm.db.Where("user_id = ? AND stat_date >= ? AND stat_date <= ?",
		userID, weekAgo, today).Find(&weeklyStats).Error; err != nil {
		return fmt.Errorf("failed to get weekly stats: %w", err)
	}

	if len(weeklyStats) == 0 {
		return nil // No activity this week
	}

	// Calculate weekly summary
	report := jm.calculateWeeklyReport(weeklyStats)

	// Send notification to user if they're connected
	if jm.wsHub != nil {
		notification := &websocket.NotificationMessage{
			Type:  "weekly_report",
			Title: "本周学习报告",
			Message: fmt.Sprintf("本周您学习了%d天，答对了%d道题，保持得很好！",
				report.ActiveDays, report.TotalCorrect),
		}
		jm.wsHub.SendNotification(userID, notification)
	}

	jm.logger.Debug("Generated weekly report",
		zap.String("user_id", userID),
		zap.Int("active_days", report.ActiveDays),
		zap.Int("total_questions", report.TotalQuestions),
		zap.Int("total_correct", report.TotalCorrect),
	)

	return nil
}

// WeeklyReport represents a weekly learning report
type WeeklyReport struct {
	UserID          string    `json:"user_id"`
	WeekStartDate   string    `json:"week_start_date"`
	WeekEndDate     string    `json:"week_end_date"`
	ActiveDays      int       `json:"active_days"`
	TotalQuestions  int       `json:"total_questions"`
	TotalCorrect    int       `json:"total_correct"`
	TotalStudyTime  int       `json:"total_study_time_ms"`
	AverageAccuracy float64   `json:"average_accuracy"`
	LongestStreak   int       `json:"longest_streak"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// calculateWeeklyReport calculates weekly statistics
func (jm *JobManager) calculateWeeklyReport(stats []model.UserAttempts) *WeeklyReport {
	if len(stats) == 0 {
		return &WeeklyReport{}
	}

	report := &WeeklyReport{
		UserID:        stats[0].UserID,
		WeekStartDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
		WeekEndDate:   time.Now().Format("2006-01-02"),
		GeneratedAt:   time.Now(),
	}

	totalQuestions := 0
	totalCorrect := 0
	totalStudyTime := 0
	maxStreak := 0

	for _, stat := range stats {
		if stat.AttemptsTotal > 0 {
			report.ActiveDays++
		}
		totalQuestions += stat.AttemptsTotal
		totalCorrect += stat.AttemptsCorrect
		totalStudyTime += stat.TotalTimeMs

		if stat.StreakDays > maxStreak {
			maxStreak = stat.StreakDays
		}
	}

	report.TotalQuestions = totalQuestions
	report.TotalCorrect = totalCorrect
	report.TotalStudyTime = totalStudyTime
	report.LongestStreak = maxStreak

	if totalQuestions > 0 {
		report.AverageAccuracy = float64(totalCorrect) / float64(totalQuestions)
	}

	return report
}

// GetJobStats returns statistics about job execution
func (jm *JobManager) GetJobStats() map[string]any {
	entries := jm.cron.Entries()

	jobs := make([]map[string]any, len(entries))
	for i, entry := range entries {
		jobs[i] = map[string]any{
			"id":       entry.ID,
			"next_run": entry.Next,
			"prev_run": entry.Prev,
		}
	}

	return map[string]any{
		"enabled":    jm.config.Enabled,
		"total_jobs": len(entries),
		"jobs":       jobs,
	}
}
