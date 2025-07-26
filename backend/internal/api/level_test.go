package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"paperplay/internal/model"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLevelTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Auto migrate the schema
	db.AutoMigrate(
		&model.Subject{},
		&model.Paper{},
		&model.Level{},
		&model.Question{},
		&model.RoadmapNode{},
		&model.UserProgress{},
		&model.User{},
	)

	return db
}

func setupLevelTestRouter(handler *LevelHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000") // UUID format user ID
		c.Next()
	})

	v1 := router.Group("/api/v1")
	{
		subjects := v1.Group("/subjects")
		{
			subjects.GET("", handler.GetAllSubjects)
			subjects.GET("/:subject_id", handler.GetSubject)
			subjects.GET("/:subject_id/papers", handler.GetSubjectPapers)
			subjects.GET("/:subject_id/roadmap", handler.GetSubjectRoadmap)
		}

		papers := v1.Group("/papers")
		{
			papers.GET("/:paper_id", handler.GetPaper)
			papers.GET("/:paper_id/level", handler.GetPaperLevel)
		}

		levels := v1.Group("/levels")
		{
			levels.GET("/:level_id", handler.GetLevel)
			levels.GET("/:level_id/questions", handler.GetLevelQuestions)
			levels.POST("/:level_id/start", handler.StartLevel)
			levels.POST("/:level_id/submit", handler.SubmitAnswer)
			levels.POST("/:level_id/complete", handler.CompleteLevel)
		}

		questions := v1.Group("/questions")
		{
			questions.GET("/:question_id", handler.GetQuestion)
		}
	}

	return router
}

func seedLevelTestData(db *gorm.DB) {
	// Create test subject
	subject := &model.Subject{
		ID:          "550e8400-e29b-41d4-a716-446655440001",
		Name:        "Computer Vision",
		Description: "AI and computer vision research",
	}
	db.Create(subject)

	// Create test paper
	paper := &model.Paper{
		ID:                 "550e8400-e29b-41d4-a716-446655440002",
		SubjectID:          "550e8400-e29b-41d4-a716-446655440001",
		Title:              "Deep Learning for Image Recognition",
		PaperAuthor:        "Goodfellow; Bengio; Courville",
		PaperPubYM:         "2016-11",
		PaperCitationCount: "5240",
	}
	db.Create(paper)

	// Create test level
	level := &model.Level{
		ID:            "550e8400-e29b-41d4-a716-446655440003",
		PaperID:       "550e8400-e29b-41d4-a716-446655440002",
		Name:          "Introduction to Deep Learning",
		PassCondition: `{"min_score":80}`,
		MetaJSON:      `{}`,
		X:             100,
		Y:             200,
	}
	db.Create(level)

	// Create test question
	question := &model.Question{
		ID:          "550e8400-e29b-41d4-a716-446655440004",
		LevelID:     "550e8400-e29b-41d4-a716-446655440003",
		Subtitle:    "About backpropagation",
		Stem:        "What is backpropagation?",
		ContentJSON: `{"type":"mcq","options":["Option A","Option B","Option C","Option D"]}`,
		AnswerJSON:  `{"type":"single","correct_options":["Option A"],"explanation":"Backpropagation is..."}`,
		Score:       10,
		CreatedBy:   "test-creator",
	}
	db.Create(question)

	// Create test roadmap node
	roadmapNode := &model.RoadmapNode{
		ID:        "550e8400-e29b-41d4-a716-446655440005",
		SubjectID: "550e8400-e29b-41d4-a716-446655440001",
		LevelID:   "550e8400-e29b-41d4-a716-446655440003",
		ParentID:  nil,
		SortOrder: 1,
		Path:      "001",
	}
	db.Create(roadmapNode)
}

func TestLevelHandler_GetAllSubjects(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subjects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Subjects retrieved successfully", response.Message)

	// Check if data contains our test subject
	data, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
}

func TestLevelHandler_GetSubject(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		subjectID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid subject ID",
			subjectID:      "550e8400-e29b-41d4-a716-446655440001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid subject ID",
			subjectID:      "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "subject_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/subjects/"+tt.subjectID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_GetSubjectPapers(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		subjectID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid subject ID",
			subjectID:      "550e8400-e29b-41d4-a716-446655440001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid subject ID",
			subjectID:      "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "subject_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/subjects/"+tt.subjectID+"/papers", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_GetPaper(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		paperID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid paper ID",
			paperID:        "550e8400-e29b-41d4-a716-446655440002",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid paper ID",
			paperID:        "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "paper_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/"+tt.paperID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_GetPaperLevel(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		paperID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid paper ID with level",
			paperID:        "550e8400-e29b-41d4-a716-446655440002",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid paper ID",
			paperID:        "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "paper_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/"+tt.paperID+"/level", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_StartLevel(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		levelID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid level ID",
			levelID:        "550e8400-e29b-41d4-a716-446655440003",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid level ID",
			levelID:        "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "level_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/levels/"+tt.levelID+"/start", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_SubmitAnswer(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	// Create user progress for the level
	progress := &model.UserProgress{
		UserID:  "550e8400-e29b-41d4-a716-446655440000",
		LevelID: "550e8400-e29b-41d4-a716-446655440003",
		Status:  model.ProgressInProgress,
		Score:   0,
		Stars:   0,
	}
	db.Create(progress)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		levelID        string
		payload        SubmitAnswerRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Valid answer submission",
			levelID: "550e8400-e29b-41d4-a716-446655440003",
			payload: SubmitAnswerRequest{
				QuestionID: "550e8400-e29b-41d4-a716-446655440004",
				AnswerJSON: "Option A",
				DurationMS: 30000,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Invalid question ID",
			levelID: "550e8400-e29b-41d4-a716-446655440003",
			payload: SubmitAnswerRequest{
				QuestionID: "550e8400-e29b-41d4-a716-446655440999",
				AnswerJSON: "Option A",
				DurationMS: 30000,
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "question_not_found",
		},
		{
			name:    "Missing question ID",
			levelID: "550e8400-e29b-41d4-a716-446655440003",
			payload: SubmitAnswerRequest{
				AnswerJSON: "Option A",
				DurationMS: 30000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/levels/"+tt.levelID+"/submit", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_CompleteLevel(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	// Create user progress for the level
	progress := &model.UserProgress{
		UserID:  "550e8400-e29b-41d4-a716-446655440000",
		LevelID: "550e8400-e29b-41d4-a716-446655440003",
		Status:  model.ProgressInProgress,
		Score:   8, // 80% of 10 points
		Stars:   0,
	}
	db.Create(progress)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		levelID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid level completion",
			levelID:        "550e8400-e29b-41d4-a716-446655440003",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid level ID",
			levelID:        "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "level_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/levels/"+tt.levelID+"/complete", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_GetLevelQuestions(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/levels/550e8400-e29b-41d4-a716-446655440003/questions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
}

func TestLevelHandler_GetQuestion(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		questionID     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid question ID",
			questionID:     "550e8400-e29b-41d4-a716-446655440004",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid question ID",
			questionID:     "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "question_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/questions/"+tt.questionID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestLevelHandler_GetSubjectRoadmap(t *testing.T) {
	db := setupLevelTestDB()
	seedLevelTestData(db)

	handler := NewLevelHandler(db)
	router := setupLevelTestRouter(handler)

	tests := []struct {
		name           string
		subjectID      string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid subject ID",
			subjectID:      "550e8400-e29b-41d4-a716-446655440001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid subject ID",
			subjectID:      "550e8400-e29b-41d4-a716-446655440999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "subject_not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/subjects/"+tt.subjectID+"/roadmap", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response.Error)
			} else {
				var response SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}
