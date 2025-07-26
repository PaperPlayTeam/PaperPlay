package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"paperplay/internal/model"
)

func setupUserServiceTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto migrate the schema
	db.AutoMigrate(
		&model.User{},
		&model.Subject{},
		&model.Paper{},
		&model.Level{},
		&model.UserProgress{},
		&model.UserAttempts{},
		&model.UserAchievement{},
		&model.Achievement{},
		&model.Event{},
		&model.RoadmapNode{},
	)

	return db
}

func seedUserServiceTestData(db *gorm.DB) {
	// Create test user
	user := &model.User{
		ID:          "test-user-id",
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	db.Create(user)

	// Create test subject
	subject := &model.Subject{
		ID:          "test-subject-1",
		Name:        "Computer Science",
		Description: "Computer Science courses",
	}
	db.Create(subject)

	// Create test paper
	paper := &model.Paper{
		ID:        "test-paper-1",
		SubjectID: "test-subject-1",
		Title:     "Introduction to Algorithms",
	}
	db.Create(paper)

	// Create test level
	level := &model.Level{
		ID:      "test-level-1",
		PaperID: "test-paper-1",
		Name:    "Basic Sorting",
	}
	db.Create(level)

	// Create user progress
	progress := &model.UserProgress{
		UserID:  "test-user-id",
		LevelID: "test-level-1",
		Status:  model.ProgressCompleted,
		Score:   85,
		Stars:   3,
	}
	db.Create(progress)

	// Create test achievement
	achievement := &model.Achievement{
		ID:          "test-achievement-1",
		Name:        "First Steps",
		Description: "Complete your first level",
		Level:       1,
		IsActive:    true,
	}
	db.Create(achievement)

	// Create user achievement
	userAchievement := &model.UserAchievement{
		UserID:        "test-user-id",
		AchievementID: "test-achievement-1",
		EarnedAt:      time.Now(),
		Progress:      1.0,
	}
	db.Create(userAchievement)

	// Create test attempts data
	today := time.Now().Format("2006-01-02")
	attempts := &model.UserAttempts{
		UserID:                  "test-user-id",
		StatDate:                today,
		AttemptsTotal:           20,
		AttemptsCorrect:         15,
		AttemptsFirstTryCorrect: 10,
		CorrectRate:             0.75,
		FirstTryCorrectRate:     0.50,
		GiveupCount:             2,
		SkipRate:                0.10,
		AvgDurationMs:           45000,
		TotalTimeMs:             900000,
		SessionsCount:           3,
		StreakDays:              5,
		ReviewDueCount:          8,
		RetentionScore:          0.80,
	}
	db.Create(attempts)

	// Create roadmap node
	roadmapNode := &model.RoadmapNode{
		ID:        "test-roadmap-1",
		SubjectID: "test-subject-1",
		LevelID:   "test-level-1",
		SortOrder: 1,
		Path:      "001",
		Depth:     1,
	}
	db.Create(roadmapNode)
}

func TestUserService_GetUserStats(t *testing.T) {
	db := setupUserServiceTestDB()
	seedUserServiceTestData(db)

	service := NewUserService(db)

	stats, err := service.GetUserStats("test-user-id")

	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Check basic stats
	assert.Equal(t, 20, stats.TotalQuestionsAnswered)
	assert.Equal(t, 15, stats.TotalCorrectAnswers)
	assert.Equal(t, 0.75, stats.OverallCorrectRate)
	assert.Equal(t, 5, stats.CurrentStreakDays)
	assert.Equal(t, 5, stats.LongestStreakDays)
	assert.Equal(t, int64(900000), stats.TotalStudyTimeMs)
	assert.Equal(t, 1, stats.LevelsCompleted)
	assert.Equal(t, 0, stats.LevelsInProgress)
	assert.Equal(t, 1, stats.AchievementsEarned)

	// Check today's stats
	assert.Equal(t, 20, stats.TodayQuestionsAnswered)
	assert.Equal(t, 15, stats.TodayCorrectAnswers)
	assert.Equal(t, 900000, stats.TodayStudyTimeMs)
	assert.Equal(t, 7, stats.WeeklyActivityDays) // 1 day from mock data

	// Check favorite subjects
	assert.Len(t, stats.FavoriteSubjects, 1)
	assert.Equal(t, "test-subject-1", stats.FavoriteSubjects[0].SubjectID)
	assert.Equal(t, "Computer Science", stats.FavoriteSubjects[0].SubjectName)
	assert.Equal(t, 1, stats.FavoriteSubjects[0].Completed)
	assert.Equal(t, 0, stats.FavoriteSubjects[0].InProgress)
}

func TestUserService_GetUserStats_NoData(t *testing.T) {
	db := setupUserServiceTestDB()

	service := NewUserService(db)

	stats, err := service.GetUserStats("nonexistent-user")

	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// All stats should be zero
	assert.Equal(t, 0, stats.TotalQuestionsAnswered)
	assert.Equal(t, 0, stats.TotalCorrectAnswers)
	assert.Equal(t, 0.0, stats.OverallCorrectRate)
	assert.Equal(t, 0, stats.CurrentStreakDays)
	assert.Equal(t, 0, stats.LongestStreakDays)
	assert.Equal(t, int64(0), stats.TotalStudyTimeMs)
	assert.Equal(t, 0, stats.LevelsCompleted)
	assert.Equal(t, 0, stats.LevelsInProgress)
	assert.Equal(t, 0, stats.AchievementsEarned)
}

func TestUserService_GetUserProgress(t *testing.T) {
	db := setupUserServiceTestDB()
	seedUserServiceTestData(db)

	service := NewUserService(db)

	progress, err := service.GetUserProgress("test-user-id")

	assert.NoError(t, err)
	assert.NotNil(t, progress)

	// Check subjects
	assert.Len(t, progress.Subjects, 1)
	assert.Equal(t, "Computer Science", progress.Subjects[0].Subject.Name)
	assert.Len(t, progress.Subjects[0].Roadmap, 1)
	assert.Len(t, progress.Subjects[0].Progress, 1)
	assert.Equal(t, 1.0, progress.Subjects[0].Completion)

	// Check recent activity
	assert.NotNil(t, progress.Recent)
}

func TestUserService_GetUserAchievements(t *testing.T) {
	db := setupUserServiceTestDB()
	seedUserServiceTestData(db)

	service := NewUserService(db)

	achievements, err := service.GetUserAchievements("test-user-id")

	assert.NoError(t, err)
	assert.Len(t, achievements, 1)
	assert.Equal(t, "test-user-id", achievements[0].UserID)
	assert.Equal(t, "test-achievement-1", achievements[0].AchievementID)
	assert.Equal(t, "First Steps", achievements[0].Achievement.Name)
}

func TestUserService_CreateOrUpdateProgress_NewProgress(t *testing.T) {
	db := setupUserServiceTestDB()
	seedUserServiceTestData(db)

	service := NewUserService(db)

	err := service.CreateOrUpdateProgress("test-user-id", "new-level-id", model.ProgressInProgress, 50, 2)

	assert.NoError(t, err)

	// Verify the progress was created
	var progress model.UserProgress
	err = db.Where("user_id = ? AND level_id = ?", "test-user-id", "new-level-id").First(&progress).Error
	assert.NoError(t, err)
	assert.Equal(t, model.ProgressInProgress, progress.Status)
	assert.Equal(t, 50, progress.Score)
	assert.Equal(t, 2, progress.Stars)
}

func TestUserService_CreateOrUpdateProgress_UpdateExisting(t *testing.T) {
	db := setupUserServiceTestDB()
	seedUserServiceTestData(db)

	service := NewUserService(db)

	err := service.CreateOrUpdateProgress("test-user-id", "test-level-1", model.ProgressCompleted, 95, 3)

	assert.NoError(t, err)

	// Verify the progress was updated
	var progress model.UserProgress
	err = db.Where("user_id = ? AND level_id = ?", "test-user-id", "test-level-1").First(&progress).Error
	assert.NoError(t, err)
	assert.Equal(t, model.ProgressCompleted, progress.Status)
	assert.Equal(t, 95, progress.Score)
	assert.Equal(t, 3, progress.Stars)
}

func TestUserService_GetUserProgressByLevel(t *testing.T) {
	db := setupUserServiceTestDB()
	seedUserServiceTestData(db)

	service := NewUserService(db)

	tests := []struct {
		name     string
		userID   string
		levelID  string
		expected bool
	}{
		{
			name:     "Existing progress",
			userID:   "test-user-id",
			levelID:  "test-level-1",
			expected: true,
		},
		{
			name:     "Non-existing progress",
			userID:   "test-user-id",
			levelID:  "nonexistent-level",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress, err := service.GetUserProgressByLevel(tt.userID, tt.levelID)

			assert.NoError(t, err)

			if tt.expected {
				assert.NotNil(t, progress)
				assert.Equal(t, tt.userID, progress.UserID)
				assert.Equal(t, tt.levelID, progress.LevelID)
			} else {
				assert.Nil(t, progress)
			}
		})
	}
}

func TestUserService_countLevelsInRoadmap(t *testing.T) {
	db := setupUserServiceTestDB()
	service := NewUserService(db)

	tests := []struct {
		name     string
		nodes    []model.RoadmapNode
		expected int
	}{
		{
			name:     "Empty roadmap",
			nodes:    []model.RoadmapNode{},
			expected: 0,
		},
		{
			name: "Single level",
			nodes: []model.RoadmapNode{
				{ID: "node-1", LevelID: "level-1"},
			},
			expected: 1,
		},
		{
			name: "Multiple levels",
			nodes: []model.RoadmapNode{
				{ID: "node-1", LevelID: "level-1"},
				{ID: "node-2", LevelID: "level-2"},
			},
			expected: 2,
		},
		{
			name: "Nested levels",
			nodes: []model.RoadmapNode{
				{
					ID:      "node-1",
					LevelID: "level-1",
					Children: []model.RoadmapNode{
						{ID: "node-2", LevelID: "level-2"},
						{ID: "node-3", LevelID: "level-3"},
					},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := service.countLevelsInRoadmap(tt.nodes)
			assert.Equal(t, tt.expected, count)
		})
	}
}
