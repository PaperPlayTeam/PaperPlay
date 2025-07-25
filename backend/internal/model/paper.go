package model

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Paper represents a research paper in the system
type Paper struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:text"`
	SubjectID          string    `json:"subject_id" gorm:"not null;type:text;index"`
	Title              string    `json:"title" gorm:"not null;type:text"`
	PaperAuthor        string    `json:"paper_author" gorm:"not null;type:text"` // Multiple authors separated by ';'
	PaperPubYM         string    `json:"paper_pub_ym" gorm:"not null;type:text"` // Publication year/month
	PaperCitationCount string    `json:"paper_citation_count" gorm:"not null;type:text"`
	CreatedAt          time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	Subject *Subject `json:"subject,omitempty" gorm:"foreignKey:SubjectID;constraint:OnDelete:CASCADE"`
	Level   *Level   `json:"level,omitempty" gorm:"foreignKey:PaperID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID for new paper
func (p *Paper) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// GetAuthors returns authors as a slice
func (p *Paper) GetAuthors() []string {
	if p.PaperAuthor == "" {
		return []string{}
	}
	authors := strings.Split(p.PaperAuthor, ";")
	// Trim whitespace
	for i := range authors {
		authors[i] = strings.TrimSpace(authors[i])
	}
	return authors
}

// SetAuthors sets authors from a slice
func (p *Paper) SetAuthors(authors []string) {
	p.PaperAuthor = strings.Join(authors, "; ")
}

// GetCitationCountInt attempts to parse citation count as integer
func (p *Paper) GetCitationCountInt() int {
	count, err := strconv.Atoi(p.PaperCitationCount)
	if err != nil {
		return 0
	}
	return count
}
