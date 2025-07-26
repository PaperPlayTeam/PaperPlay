package api

import (
	"net/http"
	"paperplay/internal/middleware"
	"paperplay/internal/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// LevelHandler handles level system HTTP requests
type LevelHandler struct {
	db        *gorm.DB
	validator *validator.Validate
}

// NewLevelHandler creates a new level handler
func NewLevelHandler(db *gorm.DB) *LevelHandler {
	return &LevelHandler{
		db:        db,
		validator: validator.New(),
	}
}

// SubmitAnswerRequest represents answer submission request
type SubmitAnswerRequest struct {
	QuestionID string `json:"question_id" validate:"required,uuid"`
	AnswerJSON any    `json:"answer_json" validate:"required"`
	DurationMS int    `json:"duration_ms" validate:"min=0"`
}

// SubmitAnswerResponse represents answer submission response
type SubmitAnswerResponse struct {
	QuestionID  string `json:"question_id"`
	IsCorrect   bool   `json:"is_correct"`
	Score       int    `json:"score"`
	TotalScore  int    `json:"total_score"`
	Explanation string `json:"explanation,omitempty"`
}

// GetAllSubjects handles GET /api/v1/subjects
func (h *LevelHandler) GetAllSubjects(c *gin.Context) {
	var subjects []model.Subject
	if err := h.db.Order("created_at DESC").Find(&subjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve subjects",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Subjects retrieved successfully",
		Data:    subjects,
	})
}

// GetSubject handles GET /api/v1/subjects/{subject_id}
func (h *LevelHandler) GetSubject(c *gin.Context) {
	subjectID := c.Param("subject_id")

	var subject model.Subject
	if err := h.db.First(&subject, "id = ?", subjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "subject_not_found",
				Message: "Subject not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve subject",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Subject retrieved successfully",
		Data:    subject,
	})
}

// GetSubjectPapers handles GET /api/v1/subjects/{subject_id}/papers
func (h *LevelHandler) GetSubjectPapers(c *gin.Context) {
	subjectID := c.Param("subject_id")

	// Check if subject exists
	var subject model.Subject
	if err := h.db.First(&subject, "id = ?", subjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "subject_not_found",
				Message: "Subject not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify subject",
			Details: err.Error(),
		})
		return
	}

	var papers []model.Paper
	if err := h.db.Where("subject_id = ?", subjectID).
		Order("created_at ASC").
		Find(&papers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve papers",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Papers retrieved successfully",
		Data:    papers,
	})
}

// GetPaper handles GET /api/v1/papers/{paper_id}
func (h *LevelHandler) GetPaper(c *gin.Context) {
	paperID := c.Param("paper_id")

	var paper model.Paper
	if err := h.db.Preload("Subject").First(&paper, "id = ?", paperID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "paper_not_found",
				Message: "Paper not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve paper",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Paper retrieved successfully",
		Data:    paper,
	})
}

// GetPaperLevel handles GET /api/v1/papers/{paper_id}/level
func (h *LevelHandler) GetPaperLevel(c *gin.Context) {
	paperID := c.Param("paper_id")

	// Check if paper exists
	var paper model.Paper
	if err := h.db.First(&paper, "id = ?", paperID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "paper_not_found",
				Message: "Paper not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify paper",
			Details: err.Error(),
		})
		return
	}

	var level model.Level
	if err := h.db.Where("paper_id = ?", paperID).First(&level).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "level_not_found",
				Message: "Level not found for this paper",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve level",
			Details: err.Error(),
		})
		return
	}

	// Enhance level data with paper information for API response
	levelResponse := struct {
		model.Level
		PaperAuthor   string `json:"paper_author"`
		PaperPubYM    string `json:"paper_pub_ym"`
		CitationCount string `json:"citation_count"`
	}{
		Level:         level,
		PaperAuthor:   paper.PaperAuthor,
		PaperPubYM:    paper.PaperPubYM,
		CitationCount: paper.PaperCitationCount,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Level retrieved successfully",
		Data:    levelResponse,
	})
}

// GetLevel handles GET /api/v1/levels/{level_id}
func (h *LevelHandler) GetLevel(c *gin.Context) {
	levelID := c.Param("level_id")

	var level model.Level
	if err := h.db.Preload("Paper").First(&level, "id = ?", levelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "level_not_found",
				Message: "Level not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve level",
			Details: err.Error(),
		})
		return
	}

	// Enhance level data with paper information for API response
	levelResponse := struct {
		model.Level
		PaperAuthor   string `json:"paper_author"`
		PaperPubYM    string `json:"paper_pub_ym"`
		CitationCount string `json:"citation_count"`
	}{
		Level: level,
	}

	if level.Paper != nil {
		levelResponse.PaperAuthor = level.Paper.PaperAuthor
		levelResponse.PaperPubYM = level.Paper.PaperPubYM
		levelResponse.CitationCount = level.Paper.PaperCitationCount
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Level retrieved successfully",
		Data:    levelResponse,
	})
}

// GetLevelQuestions handles GET /api/v1/levels/{level_id}/questions
func (h *LevelHandler) GetLevelQuestions(c *gin.Context) {
	levelID := c.Param("level_id")

	// Check if level exists
	var level model.Level
	if err := h.db.First(&level, "id = ?", levelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "level_not_found",
				Message: "Level not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify level",
			Details: err.Error(),
		})
		return
	}

	var questions []model.Question
	if err := h.db.Where("level_id = ?", levelID).
		Order("created_at ASC").
		Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve questions",
			Details: err.Error(),
		})
		return
	}

	// For security, don't expose answer_json to students
	questionsResponse := make([]map[string]any, len(questions))
	for i, q := range questions {
		questionsResponse[i] = map[string]any{
			"id":           q.ID,
			"level_id":     q.LevelID,
			"stem":         q.Stem,
			"content_json": q.ContentJSON,
			"score":        q.Score,
			"created_by":   q.CreatedBy,
			"created_at":   q.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Questions retrieved successfully",
		Data:    questionsResponse,
	})
}

// GetQuestion handles GET /api/v1/questions/{question_id}
func (h *LevelHandler) GetQuestion(c *gin.Context) {
	questionID := c.Param("question_id")

	var question model.Question
	if err := h.db.First(&question, "id = ?", questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "question_not_found",
				Message: "Question not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve question",
			Details: err.Error(),
		})
		return
	}

	// For security, don't expose answer_json to students
	questionResponse := map[string]any{
		"id":           question.ID,
		"level_id":     question.LevelID,
		"stem":         question.Stem,
		"content_json": question.ContentJSON,
		"score":        question.Score,
		"created_by":   question.CreatedBy,
		"created_at":   question.CreatedAt,
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Question retrieved successfully",
		Data:    questionResponse,
	})
}

// GetSubjectRoadmap handles GET /api/v1/subjects/{subject_id}/roadmap
func (h *LevelHandler) GetSubjectRoadmap(c *gin.Context) {
	subjectID := c.Param("subject_id")

	// Check if subject exists
	var subject model.Subject
	if err := h.db.First(&subject, "id = ?", subjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "subject_not_found",
				Message: "Subject not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify subject",
			Details: err.Error(),
		})
		return
	}

	// Get roadmap using the service
	roadmapService := model.NewRoadmapNodeService(h.db)
	roadmap, err := roadmapService.GetSubjectRoadmap(subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "roadmap_error",
			Message: "Failed to retrieve roadmap",
			Details: err.Error(),
		})
		return
	}

	// Simplify response structure for API
	roadmapResponse := make([]map[string]any, len(roadmap))
	for i, node := range roadmap {
		roadmapResponse[i] = map[string]any{
			"id":         node.ID,
			"subject_id": node.SubjectID,
			"level_id":   node.LevelID,
			"parent_id":  node.ParentID,
			"sort":       node.SortOrder,
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Roadmap retrieved successfully",
		Data:    roadmapResponse,
	})
}

// StartLevel handles POST /api/v1/levels/{level_id}/start
func (h *LevelHandler) StartLevel(c *gin.Context) {
	levelID := c.Param("level_id")
	userID := middleware.MustGetCurrentUserID(c)

	// Check if level exists
	var level model.Level
	if err := h.db.First(&level, "id = ?", levelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "level_not_found",
				Message: "Level not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify level",
			Details: err.Error(),
		})
		return
	}

	// Check existing progress
	var progress model.UserProgress
	if err := h.db.Where("user_id = ? AND level_id = ?", userID, levelID).First(&progress).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new progress record
			now := time.Now()
			progress = model.UserProgress{
				UserID:        userID,
				LevelID:       levelID,
				Status:        1, // 1 = in progress
				Score:         0,
				Stars:         0,
				LastAttemptAt: &now,
			}
			if err := h.db.Create(&progress).Error; err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error:   "database_error",
					Message: "Failed to create progress record",
					Details: err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to check progress",
				Details: err.Error(),
			})
			return
		}
	} else {
		// Update existing progress to in progress
		now := time.Now()
		progress.Status = 1
		progress.LastAttemptAt = &now
		if err := h.db.Save(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to update progress",
				Details: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Level started",
		Data: map[string]any{
			"user_id":    progress.UserID,
			"level_id":   progress.LevelID,
			"status":     progress.Status,
			"started_at": progress.LastAttemptAt,
		},
	})
}

// SubmitAnswer handles POST /api/v1/levels/{level_id}/submit
func (h *LevelHandler) SubmitAnswer(c *gin.Context) {
	levelID := c.Param("level_id")
	userID := middleware.MustGetCurrentUserID(c)

	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
		return
	}

	// Check if question exists and belongs to the level
	var question model.Question
	if err := h.db.Where("id = ? AND level_id = ?", req.QuestionID, levelID).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "question_not_found",
				Message: "Question not found in this level",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify question",
			Details: err.Error(),
		})
		return
	}

	// Convert answer to string for validation
	userAnswer := ""
	switch v := req.AnswerJSON.(type) {
	case string:
		userAnswer = v
	case float64:
		userAnswer = string(rune(int(v)))
	case int:
		userAnswer = string(rune(v))
	default:
		// For complex answers, convert to JSON string
		userAnswer = c.GetString("answer_json")
	}

	// Check if answer is correct
	isCorrect, err := question.IsCorrectAnswer(userAnswer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "answer_validation_error",
			Message: "Failed to validate answer",
			Details: err.Error(),
		})
		return
	}

	// Calculate score
	score := 0
	if isCorrect {
		score = question.Score
	}

	// Get current total score for the level
	var currentProgress model.UserProgress
	totalScore := score
	if err := h.db.Where("user_id = ? AND level_id = ?", userID, levelID).First(&currentProgress).Error; err == nil {
		totalScore = currentProgress.Score + score
	}

	// Get answer explanation if available
	answer, _ := question.GetAnswer()
	explanation := ""
	if answer != nil {
		explanation = answer.Explanation
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Answer submitted",
		Data: SubmitAnswerResponse{
			QuestionID:  req.QuestionID,
			IsCorrect:   isCorrect,
			Score:       score,
			TotalScore:  totalScore,
			Explanation: explanation,
		},
	})
}

// CompleteLevel handles POST /api/v1/levels/{level_id}/complete
func (h *LevelHandler) CompleteLevel(c *gin.Context) {
	levelID := c.Param("level_id")
	userID := middleware.MustGetCurrentUserID(c)

	// Check if level exists
	var level model.Level
	if err := h.db.First(&level, "id = ?", levelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "level_not_found",
				Message: "Level not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to verify level",
			Details: err.Error(),
		})
		return
	}

	// Get user progress
	var progress model.UserProgress
	if err := h.db.Where("user_id = ? AND level_id = ?", userID, levelID).First(&progress).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "progress_not_found",
			Message: "No progress found for this level",
		})
		return
	}

	// Calculate total possible score for the level
	var totalPossibleScore int
	h.db.Model(&model.Question{}).Where("level_id = ?", levelID).Select("COALESCE(SUM(score), 0)").Scan(&totalPossibleScore)

	// Calculate stars based on score percentage
	stars := 0
	if totalPossibleScore > 0 {
		percentage := float64(progress.Score) / float64(totalPossibleScore)
		if percentage >= 0.9 {
			stars = 3
		} else if percentage >= 0.7 {
			stars = 2
		} else if percentage >= 0.5 {
			stars = 1
		}
	}

	// Update progress
	now := time.Now()
	progress.Status = 2 // 2 = completed
	progress.Stars = stars
	progress.LastAttemptAt = &now

	if err := h.db.Save(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update progress",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Level completed",
		Data: map[string]any{
			"user_id":      progress.UserID,
			"level_id":     progress.LevelID,
			"score":        progress.Score,
			"stars":        stars,
			"completed_at": progress.LastAttemptAt,
		},
	})
}
