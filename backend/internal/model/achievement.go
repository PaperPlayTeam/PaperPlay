package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Achievement represents an achievement definition
type Achievement struct {
	ID          string    `json:"id" gorm:"primaryKey;type:text"`
	Name        string    `json:"name" gorm:"not null;type:text"`
	Description string    `json:"description" gorm:"not null;type:text"`
	Level       int       `json:"level" gorm:"not null;default:1"`      // Achievement level (Bronze=1, Silver=2, Gold=3)
	IconURL     string    `json:"icon_url" gorm:"type:text"`            // Icon image URL
	BadgeType   string    `json:"badge_type" gorm:"not null;type:text"` // Category: "learning", "streak", "speed", etc.
	RuleJSON    string    `json:"rule_json" gorm:"not null;type:text"`  // Achievement trigger rules as JSON
	NFTEnabled  bool      `json:"nft_enabled" gorm:"default:false"`     // Whether this achievement generates NFT
	NFTMetadata string    `json:"nft_metadata" gorm:"type:text"`        // NFT metadata template as JSON
	IsActive    bool      `json:"is_active" gorm:"default:true"`        // Whether this achievement is active
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	UserAchievements []UserAchievement `json:"user_achievements,omitempty" gorm:"foreignKey:AchievementID;constraint:OnDelete:CASCADE"`
	NFTAssets        []NFTAsset        `json:"nft_assets,omitempty" gorm:"foreignKey:AchievementID"`
}

// BeforeCreate generates UUID for new achievement
func (a *Achievement) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// AchievementRule represents the structure of achievement rule JSON
type AchievementRule struct {
	Type       string          `json:"type"`       // "streak", "score", "speed", "total", etc.
	Conditions []ConditionRule `json:"conditions"` // Multiple conditions (AND logic)
	Metadata   map[string]any  `json:"metadata"`   // Additional rule data
}

// ConditionRule represents a single condition within an achievement rule
type ConditionRule struct {
	Field    string `json:"field"`    // Field to check (e.g., "streak_days", "correct_rate")
	Operator string `json:"operator"` // ">=", ">", "<=", "<", "=", "!="
	Value    any    `json:"value"`    // Value to compare against
}

// GetRule parses and returns the achievement rule
func (a *Achievement) GetRule() (*AchievementRule, error) {
	var rule AchievementRule
	if err := json.Unmarshal([]byte(a.RuleJSON), &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

// SetRule sets the achievement rule from struct
func (a *Achievement) SetRule(rule *AchievementRule) error {
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	a.RuleJSON = string(data)
	return nil
}

// NFTMetadataTemplate represents NFT metadata structure
type NFTMetadataTemplate struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Image       string         `json:"image"`
	Attributes  []NFTAttribute `json:"attributes"`
	Properties  map[string]any `json:"properties"`
}

// NFTAttribute represents an NFT attribute
type NFTAttribute struct {
	TraitType string `json:"trait_type"`
	Value     any    `json:"value"`
}

// GetNFTMetadata parses and returns the NFT metadata template
func (a *Achievement) GetNFTMetadata() (*NFTMetadataTemplate, error) {
	if a.NFTMetadata == "" {
		return &NFTMetadataTemplate{}, nil
	}
	var metadata NFTMetadataTemplate
	if err := json.Unmarshal([]byte(a.NFTMetadata), &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

// SetNFTMetadata sets the NFT metadata from struct
func (a *Achievement) SetNFTMetadata(metadata *NFTMetadataTemplate) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	a.NFTMetadata = string(data)
	return nil
}

// UserAchievement represents a user's earned achievement
type UserAchievement struct {
	ID            string     `json:"id" gorm:"primaryKey;type:text"`
	UserID        string     `json:"user_id" gorm:"not null;type:text;index"`
	AchievementID string     `json:"achievement_id" gorm:"not null;type:text;index"`
	EarnedAt      time.Time  `json:"earned_at" gorm:"not null"`
	Progress      float64    `json:"progress" gorm:"default:1.0"`      // Progress towards achievement (0.0-1.0)
	MetaJSON      string     `json:"meta_json" gorm:"type:text"`       // Additional metadata about earning
	NotifiedAt    *time.Time `json:"notified_at" gorm:"type:datetime"` // When user was notified
	ViewedAt      *time.Time `json:"viewed_at" gorm:"type:datetime"`   // When user viewed the achievement

	// Associations
	User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Achievement *Achievement `json:"achievement,omitempty" gorm:"foreignKey:AchievementID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID for new user achievement
func (ua *UserAchievement) BeforeCreate(tx *gorm.DB) error {
	if ua.ID == "" {
		ua.ID = uuid.New().String()
	}
	return nil
}

// IsNew checks if this achievement is newly earned (not notified yet)
func (ua *UserAchievement) IsNew() bool {
	return ua.NotifiedAt == nil
}

// MarkNotified marks the achievement as notified
func (ua *UserAchievement) MarkNotified() {
	now := time.Now()
	ua.NotifiedAt = &now
}

// MarkViewed marks the achievement as viewed by user
func (ua *UserAchievement) MarkViewed() {
	now := time.Now()
	ua.ViewedAt = &now
}

// Event represents user events for achievement tracking
type Event struct {
	ID         string    `json:"id" gorm:"primaryKey;type:text"`
	UserID     string    `json:"user_id" gorm:"not null;type:text;index"`
	EventType  string    `json:"event_type" gorm:"not null;type:text;index"` // "question_answered", "level_completed", etc.
	LevelID    *string   `json:"level_id" gorm:"type:text;index"`            // Related level if applicable
	QuestionID *string   `json:"question_id" gorm:"type:text;index"`         // Related question if applicable
	DataJSON   string    `json:"data_json" gorm:"type:text"`                 // Event-specific data as JSON
	CreatedAt  time.Time `json:"created_at" gorm:"not null;index"`

	// Associations
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Level    *Level    `json:"level,omitempty" gorm:"foreignKey:LevelID"`
	Question *Question `json:"question,omitempty" gorm:"foreignKey:QuestionID"`
}

// BeforeCreate generates UUID for new event
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// EventData represents generic event data structure
type EventData map[string]any

// GetData parses and returns event data
func (e *Event) GetData() (EventData, error) {
	if e.DataJSON == "" {
		return EventData{}, nil
	}
	var data EventData
	if err := json.Unmarshal([]byte(e.DataJSON), &data); err != nil {
		return nil, err
	}
	return data, nil
}

// SetData sets event data from map
func (e *Event) SetData(data EventData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	e.DataJSON = string(jsonData)
	return nil
}

// Event type constants
const (
	EventQuestionAnswered  = "question_answered"
	EventLevelCompleted    = "level_completed"
	EventLevelStarted      = "level_started"
	EventLevelFailed       = "level_failed"
	EventSessionStarted    = "session_started"
	EventSessionEnded      = "session_ended"
	EventStreakUpdated     = "streak_updated"
	EventAchievementEarned = "achievement_earned"
)

// NFTAsset represents an NFT asset owned by a user
type NFTAsset struct {
	ID              string    `json:"id" gorm:"primaryKey;type:text"`
	UserID          string    `json:"user_id" gorm:"not null;type:text;index"`
	AchievementID   *string   `json:"achievement_id" gorm:"type:text;index"` // Related achievement if any
	ContractAddress string    `json:"contract_address" gorm:"not null;type:text"`
	TokenID         string    `json:"token_id" gorm:"not null;type:text"`
	MetadataURI     string    `json:"metadata_uri" gorm:"not null;type:text"`
	MintTxHash      string    `json:"mint_tx_hash" gorm:"type:text"`
	Status          string    `json:"status" gorm:"not null;default:'pending'"` // pending, minted, failed
	CreatedAt       time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	User        *User        `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Achievement *Achievement `json:"achievement,omitempty" gorm:"foreignKey:AchievementID"`
}

// BeforeCreate generates UUID for new NFT asset
func (nft *NFTAsset) BeforeCreate(tx *gorm.DB) error {
	if nft.ID == "" {
		nft.ID = uuid.New().String()
	}
	return nil
}

// NFT status constants
const (
	NFTStatusPending = "pending"
	NFTStatusMinted  = "minted"
	NFTStatusFailed  = "failed"
)

// IsPending checks if the NFT is still pending
func (nft *NFTAsset) IsPending() bool {
	return nft.Status == NFTStatusPending
}

// IsMinted checks if the NFT has been successfully minted
func (nft *NFTAsset) IsMinted() bool {
	return nft.Status == NFTStatusMinted
}

// MarkMinted marks the NFT as successfully minted
func (nft *NFTAsset) MarkMinted(txHash string) {
	nft.Status = NFTStatusMinted
	nft.MintTxHash = txHash
	nft.UpdatedAt = time.Now()
}

// MarkFailed marks the NFT minting as failed
func (nft *NFTAsset) MarkFailed() {
	nft.Status = NFTStatusFailed
	nft.UpdatedAt = time.Now()
}
