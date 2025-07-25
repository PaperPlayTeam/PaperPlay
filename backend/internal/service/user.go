package service

import (
	"fmt"
	"paperplay/internal/model"
	"time"

	"gorm.io/gorm"
)

// UserService handles user business logic
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// UserStats represents user statistics
type UserStats struct {
	TotalQuestionsAnswered int           `json:"total_questions_answered"`
	TotalCorrectAnswers    int           `json:"total_correct_answers"`
	OverallCorrectRate     float64       `json:"overall_correct_rate"`
	CurrentStreakDays      int           `json:"current_streak_days"`
	LongestStreakDays      int           `json:"longest_streak_days"`
	TotalStudyTimeMs       int64         `json:"total_study_time_ms"`
	LevelsCompleted        int           `json:"levels_completed"`
	LevelsInProgress       int           `json:"levels_in_progress"`
	AchievementsEarned     int           `json:"achievements_earned"`
	TodayQuestionsAnswered int           `json:"today_questions_answered"`
	TodayCorrectAnswers    int           `json:"today_correct_answers"`
	TodayStudyTimeMs       int           `json:"today_study_time_ms"`
	WeeklyActivityDays     int           `json:"weekly_activity_days"`
	FavoriteSubjects       []SubjectStat `json:"favorite_subjects"`
}

// SubjectStat represents statistics for a subject
type SubjectStat struct {
	SubjectID   string  `json:"subject_id"`
	SubjectName string  `json:"subject_name"`
	Completed   int     `json:"completed"`
	InProgress  int     `json:"in_progress"`
	CorrectRate float64 `json:"correct_rate"`
}

// UserProgressResponse represents user's learning progress
type UserProgressResponse struct {
	Subjects []SubjectProgress `json:"subjects"`
	Recent   []RecentActivity  `json:"recent_activity"`
}

// SubjectProgress represents progress in a subject
type SubjectProgress struct {
	Subject    *model.Subject       `json:"subject"`
	Roadmap    []model.RoadmapNode  `json:"roadmap"`
	Progress   []model.UserProgress `json:"progress"`
	Completion float64              `json:"completion_rate"`
}

// RecentActivity represents recent user activity
type RecentActivity struct {
	Date        string `json:"date"`
	Activity    string `json:"activity"`
	Description string `json:"description"`
	LevelID     string `json:"level_id,omitempty"`
	SubjectID   string `json:"subject_id,omitempty"`
}

// GetUserStats returns comprehensive user statistics
func (s *UserService) GetUserStats(userID string) (*UserStats, error) {
	stats := &UserStats{}

	// Get today's date
	today := time.Now().Format("2006-01-02")

	// Get today's attempts
	var todayAttempts model.UserAttempts
	if err := s.db.Where("user_id = ? AND stat_date = ?", userID, today).First(&todayAttempts).Error; err == nil {
		stats.TodayQuestionsAnswered = todayAttempts.AttemptsTotal
		stats.TodayCorrectAnswers = todayAttempts.AttemptsCorrect
		stats.TodayStudyTimeMs = todayAttempts.TotalTimeMs
		stats.CurrentStreakDays = todayAttempts.StreakDays
	}

	// Get overall statistics from all attempts
	var allAttempts []model.UserAttempts
	if err := s.db.Where("user_id = ?", userID).Find(&allAttempts).Error; err != nil {
		return nil, fmt.Errorf("failed to get user attempts: %w", err)
	}

	totalQuestions := 0
	totalCorrect := 0
	totalStudyTime := int64(0)
	maxStreak := 0

	for _, attempt := range allAttempts {
		totalQuestions += attempt.AttemptsTotal
		totalCorrect += attempt.AttemptsCorrect
		totalStudyTime += int64(attempt.TotalTimeMs)
		if attempt.StreakDays > maxStreak {
			maxStreak = attempt.StreakDays
		}
	}

	stats.TotalQuestionsAnswered = totalQuestions
	stats.TotalCorrectAnswers = totalCorrect
	stats.TotalStudyTimeMs = totalStudyTime
	stats.LongestStreakDays = maxStreak

	if totalQuestions > 0 {
		stats.OverallCorrectRate = float64(totalCorrect) / float64(totalQuestions)
	}

	// Get progress statistics
	var progressStats []struct {
		Status int
		Count  int
	}
	if err := s.db.Model(&model.UserProgress{}).
		Where("user_id = ?", userID).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&progressStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get progress stats: %w", err)
	}

	for _, stat := range progressStats {
		switch stat.Status {
		case model.ProgressCompleted:
			stats.LevelsCompleted = stat.Count
		case model.ProgressInProgress:
			stats.LevelsInProgress = stat.Count
		}
	}

	// Get achievements count
	var achievementsCount int64
	if err := s.db.Model(&model.UserAchievement{}).
		Where("user_id = ?", userID).
		Count(&achievementsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get achievements count: %w", err)
	}
	stats.AchievementsEarned = int(achievementsCount)

	// Get weekly activity (last 7 days)
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	var weeklyActivityCount int64
	if err := s.db.Model(&model.UserAttempts{}).
		Where("user_id = ? AND stat_date >= ? AND attempts_total > 0", userID, weekAgo).
		Count(&weeklyActivityCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get weekly activity: %w", err)
	}
	stats.WeeklyActivityDays = int(weeklyActivityCount)

	// Get favorite subjects (subjects with most activity)
	var subjectStats []struct {
		SubjectID   string
		SubjectName string
		Completed   int64
		InProgress  int64
		Total       int64
	}

	if err := s.db.Table("user_progresses").
		Select(`
			subjects.id as subject_id,
			subjects.name as subject_name,
			SUM(CASE WHEN user_progresses.status = ? THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN user_progresses.status = ? THEN 1 ELSE 0 END) as in_progress,
			COUNT(*) as total
		`, model.ProgressCompleted, model.ProgressInProgress).
		Joins("JOIN levels ON levels.id = user_progresses.level_id").
		Joins("JOIN papers ON papers.id = levels.paper_id").
		Joins("JOIN subjects ON subjects.id = papers.subject_id").
		Where("user_progresses.user_id = ?", userID).
		Group("subjects.id, subjects.name").
		Having("total > 0").
		Order("completed DESC, total DESC").
		Limit(5).
		Find(&subjectStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get subject stats: %w", err)
	}

	for _, stat := range subjectStats {
		correctRate := 0.0
		if stat.Total > 0 {
			correctRate = float64(stat.Completed) / float64(stat.Total)
		}

		stats.FavoriteSubjects = append(stats.FavoriteSubjects, SubjectStat{
			SubjectID:   stat.SubjectID,
			SubjectName: stat.SubjectName,
			Completed:   int(stat.Completed),
			InProgress:  int(stat.InProgress),
			CorrectRate: correctRate,
		})
	}

	return stats, nil
}

// GetUserProgress returns user's learning progress across all subjects
func (s *UserService) GetUserProgress(userID string) (*UserProgressResponse, error) {
	var response UserProgressResponse

	// Get all subjects
	var subjects []model.Subject
	if err := s.db.Find(&subjects).Error; err != nil {
		return nil, fmt.Errorf("failed to get subjects: %w", err)
	}

	// Get user progress for each subject
	for _, subject := range subjects {
		// Get roadmap for this subject
		roadmapService := model.NewRoadmapNodeService(s.db)
		roadmap, err := roadmapService.GetSubjectRoadmap(subject.ID)
		if err != nil {
			continue // Skip subjects with roadmap errors
		}

		if len(roadmap) == 0 {
			continue // Skip subjects without roadmap
		}

		// Get user progress for levels in this subject
		var progress []model.UserProgress
		if err := s.db.Where(`
			user_id = ? AND level_id IN (
				SELECT levels.id FROM levels 
				JOIN papers ON papers.id = levels.paper_id 
				WHERE papers.subject_id = ?
			)
		`, userID, subject.ID).
			Preload("Level").
			Find(&progress).Error; err != nil {
			continue // Skip on error
		}

		// Calculate completion rate
		totalLevels := s.countLevelsInRoadmap(roadmap)
		completedLevels := 0
		for _, p := range progress {
			if p.Status == model.ProgressCompleted {
				completedLevels++
			}
		}

		completionRate := 0.0
		if totalLevels > 0 {
			completionRate = float64(completedLevels) / float64(totalLevels)
		}

		response.Subjects = append(response.Subjects, SubjectProgress{
			Subject:    &subject,
			Roadmap:    roadmap,
			Progress:   progress,
			Completion: completionRate,
		})
	}

	// Get recent activity
	recentActivity, err := s.getRecentActivity(userID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	response.Recent = recentActivity

	return &response, nil
}

// GetUserAchievements returns user's achievements
func (s *UserService) GetUserAchievements(userID string) ([]model.UserAchievement, error) {
	var achievements []model.UserAchievement
	if err := s.db.Where("user_id = ?", userID).
		Preload("Achievement").
		Order("earned_at DESC").
		Find(&achievements).Error; err != nil {
		return nil, fmt.Errorf("failed to get user achievements: %w", err)
	}

	return achievements, nil
}

// countLevelsInRoadmap recursively counts all levels in roadmap
func (s *UserService) countLevelsInRoadmap(nodes []model.RoadmapNode) int {
	count := 0
	for _, node := range nodes {
		count++                                        // Count this node
		count += s.countLevelsInRoadmap(node.Children) // Count children
	}
	return count
}

// getRecentActivity gets user's recent learning activity
func (s *UserService) getRecentActivity(userID string, limit int) ([]RecentActivity, error) {
	var events []model.Event
	if err := s.db.Where("user_id = ?", userID).
		Preload("Level").
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent events: %w", err)
	}

	var activities []RecentActivity
	for _, event := range events {
		activity := RecentActivity{
			Date: event.CreatedAt.Format("2006-01-02"),
		}

		switch event.EventType {
		case model.EventLevelCompleted:
			activity.Activity = "level_completed"
			activity.Description = "Completed a level"
			if event.Level != nil {
				activity.Description = fmt.Sprintf("Completed level: %s", event.Level.Name)
				activity.LevelID = event.Level.ID
			}
		case model.EventLevelStarted:
			activity.Activity = "level_started"
			activity.Description = "Started a level"
			if event.Level != nil {
				activity.Description = fmt.Sprintf("Started level: %s", event.Level.Name)
				activity.LevelID = event.Level.ID
			}
		case model.EventQuestionAnswered:
			activity.Activity = "question_answered"
			activity.Description = "Answered questions"
		case model.EventAchievementEarned:
			activity.Activity = "achievement_earned"
			activity.Description = "Earned an achievement"
		default:
			continue // Skip unknown event types
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// CreateOrUpdateProgress creates or updates user progress for a level
func (s *UserService) CreateOrUpdateProgress(userID, levelID string, status int, score int, stars int) error {
	now := time.Now()

	// Check if progress already exists
	var existing model.UserProgress
	err := s.db.Where("user_id = ? AND level_id = ?", userID, levelID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new progress
		progress := model.UserProgress{
			UserID:        userID,
			LevelID:       levelID,
			Status:        status,
			Score:         score,
			Stars:         stars,
			LastAttemptAt: &now,
		}
		return s.db.Create(&progress).Error
	} else if err != nil {
		return fmt.Errorf("failed to check existing progress: %w", err)
	}

	// Update existing progress
	updates := map[string]any{
		"status":          status,
		"score":           score,
		"stars":           stars,
		"last_attempt_at": now,
		"updated_at":      now,
	}

	return s.db.Model(&existing).Updates(updates).Error
}

// GetUserProgressByLevel returns user's progress for a specific level
func (s *UserService) GetUserProgressByLevel(userID, levelID string) (*model.UserProgress, error) {
	var progress model.UserProgress
	err := s.db.Where("user_id = ? AND level_id = ?", userID, levelID).
		Preload("Level").
		First(&progress).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // No progress found
	} else if err != nil {
		return nil, fmt.Errorf("failed to get user progress: %w", err)
	}

	return &progress, nil
}
