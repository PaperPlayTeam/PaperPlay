package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Level represents a learning level/stage in the system
type Level struct {
	ID            string    `json:"id" gorm:"primaryKey;type:text"`
	PaperID       string    `json:"paper_id" gorm:"unique;not null;type:text;index"`
	Name          string    `json:"name" gorm:"not null;type:text"`
	PassCondition string    `json:"pass_condition" gorm:"not null;type:text"` // JSON string
	MetaJSON      string    `json:"meta_json" gorm:"type:text"`               // Additional metadata
	X             int       `json:"x" gorm:"not null"`                        // UI X coordinate
	Y             int       `json:"y" gorm:"not null"`                        // UI Y coordinate
	CreatedAt     time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	Paper     *Paper     `json:"paper,omitempty" gorm:"foreignKey:PaperID;constraint:OnDelete:CASCADE"`
	Questions []Question `json:"questions,omitempty" gorm:"foreignKey:LevelID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID for new level
func (l *Level) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = uuid.New().String()
	}
	return nil
}

// PassConditionData represents the structure of pass condition JSON
type PassConditionData struct {
	MinScore      int     `json:"min_score"`       // Minimum score required
	MinCorrectPct float64 `json:"min_correct_pct"` // Minimum correct percentage (0-1)
	MaxAttempts   int     `json:"max_attempts"`    // Maximum attempts allowed
	TimeLimit     int     `json:"time_limit"`      // Time limit in seconds (0 = no limit)
}

// GetPassCondition parses and returns the pass condition
func (l *Level) GetPassCondition() (*PassConditionData, error) {
	var condition PassConditionData
	if err := json.Unmarshal([]byte(l.PassCondition), &condition); err != nil {
		return nil, err
	}
	return &condition, nil
}

// SetPassCondition sets the pass condition from struct
func (l *Level) SetPassCondition(condition *PassConditionData) error {
	data, err := json.Marshal(condition)
	if err != nil {
		return err
	}
	l.PassCondition = string(data)
	return nil
}

// MetaData represents additional metadata for the level
type MetaData struct {
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Difficulty  int                    `json:"difficulty"` // 1-5 scale
	Resources   []string               `json:"resources"`  // URLs or file paths
	Custom      map[string]interface{} `json:"custom"`     // Custom fields
}

// GetMetaData parses and returns the metadata
func (l *Level) GetMetaData() (*MetaData, error) {
	if l.MetaJSON == "" {
		return &MetaData{}, nil
	}
	var meta MetaData
	if err := json.Unmarshal([]byte(l.MetaJSON), &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// SetMetaData sets the metadata from struct
func (l *Level) SetMetaData(meta *MetaData) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	l.MetaJSON = string(data)
	return nil
}

// Question represents a question within a level
type Question struct {
	ID          string    `json:"id" gorm:"primaryKey;type:text"`
	LevelID     string    `json:"level_id" gorm:"not null;type:text;index"`
	Stem        string    `json:"stem" gorm:"not null;type:text"`         // Question stem/context
	ContentJSON string    `json:"content_json" gorm:"not null;type:text"` // Question content as JSON
	AnswerJSON  string    `json:"answer_json" gorm:"not null;type:text"`  // Standard answer as JSON
	Score       int       `json:"score" gorm:"not null"`                  // Points for this question
	CreatedBy   string    `json:"created_by" gorm:"type:text"`            // Creator/generator
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`

	// Associations
	Level *Level `json:"level,omitempty" gorm:"foreignKey:LevelID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID for new question
func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	return nil
}

// QuestionContent represents the structure of question content JSON
type QuestionContent struct {
	Type        string                 `json:"type"`        // "multiple_choice", "text", "code", etc.
	Text        string                 `json:"text"`        // Question text
	Options     []string               `json:"options"`     // For multiple choice
	Code        string                 `json:"code"`        // For code questions
	Images      []string               `json:"images"`      // Image URLs
	Attachments []string               `json:"attachments"` // File URLs
	Metadata    map[string]interface{} `json:"metadata"`    // Additional data
}

// GetContent parses and returns the question content
func (q *Question) GetContent() (*QuestionContent, error) {
	var content QuestionContent
	if err := json.Unmarshal([]byte(q.ContentJSON), &content); err != nil {
		return nil, err
	}
	return &content, nil
}

// SetContent sets the question content from struct
func (q *Question) SetContent(content *QuestionContent) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	q.ContentJSON = string(data)
	return nil
}

// QuestionAnswer represents the structure of answer JSON
type QuestionAnswer struct {
	Type           string                 `json:"type"`            // "single", "multiple", "text", "code"
	CorrectOptions []string               `json:"correct_options"` // For multiple choice
	CorrectText    string                 `json:"correct_text"`    // For text answers
	CorrectCode    string                 `json:"correct_code"`    // For code answers
	Explanation    string                 `json:"explanation"`     // Answer explanation
	Keywords       []string               `json:"keywords"`        // For text matching
	CaseSensitive  bool                   `json:"case_sensitive"`  // For text answers
	Metadata       map[string]interface{} `json:"metadata"`        // Additional data
}

// GetAnswer parses and returns the question answer
func (q *Question) GetAnswer() (*QuestionAnswer, error) {
	var answer QuestionAnswer
	if err := json.Unmarshal([]byte(q.AnswerJSON), &answer); err != nil {
		return nil, err
	}
	return &answer, nil
}

// SetAnswer sets the question answer from struct
func (q *Question) SetAnswer(answer *QuestionAnswer) error {
	data, err := json.Marshal(answer)
	if err != nil {
		return err
	}
	q.AnswerJSON = string(data)
	return nil
}

// IsCorrectAnswer checks if the provided answer is correct
func (q *Question) IsCorrectAnswer(userAnswer string) (bool, error) {
	answer, err := q.GetAnswer()
	if err != nil {
		return false, err
	}

	switch answer.Type {
	case "single":
		if len(answer.CorrectOptions) == 0 {
			return false, nil
		}
		return userAnswer == answer.CorrectOptions[0], nil

	case "multiple":
		// For multiple choice, userAnswer should be comma-separated
		userOptions := strings.Split(userAnswer, ",")
		for i := range userOptions {
			userOptions[i] = strings.TrimSpace(userOptions[i])
		}

		// Check if all correct options are present and no extra ones
		if len(userOptions) != len(answer.CorrectOptions) {
			return false, nil
		}

		userMap := make(map[string]bool)
		for _, opt := range userOptions {
			userMap[opt] = true
		}

		for _, correct := range answer.CorrectOptions {
			if !userMap[correct] {
				return false, nil
			}
		}
		return true, nil

	case "text":
		compareText := userAnswer
		correctText := answer.CorrectText

		if !answer.CaseSensitive {
			compareText = strings.ToLower(compareText)
			correctText = strings.ToLower(correctText)
		}

		// Exact match or keyword matching
		if compareText == correctText {
			return true, nil
		}

		// Check keywords if available
		for _, keyword := range answer.Keywords {
			checkKeyword := keyword
			if !answer.CaseSensitive {
				checkKeyword = strings.ToLower(checkKeyword)
			}
			if strings.Contains(compareText, checkKeyword) {
				return true, nil
			}
		}
		return false, nil

	case "code":
		// Simple string comparison for code (can be enhanced)
		return strings.TrimSpace(userAnswer) == strings.TrimSpace(answer.CorrectCode), nil

	default:
		return false, fmt.Errorf("unsupported answer type: %s", answer.Type)
	}
}
