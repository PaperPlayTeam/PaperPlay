package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subject represents a subject/discipline in the learning system
type Subject struct {
	ID          string    `json:"id" gorm:"primaryKey;type:text"`
	Name        string    `json:"name" gorm:"not null;type:text"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"not null"`

	// Associations
	Papers       []Paper       `json:"papers,omitempty" gorm:"foreignKey:SubjectID;constraint:OnDelete:CASCADE"`
	RoadmapNodes []RoadmapNode `json:"roadmap_nodes,omitempty" gorm:"foreignKey:SubjectID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate generates UUID for new subject
func (s *Subject) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// RoadmapNode represents a node in the subject roadmap tree structure
// Uses materialized path pattern for efficient tree operations
type RoadmapNode struct {
	ID        string  `json:"id" gorm:"primaryKey;type:text"`
	SubjectID string  `json:"subject_id" gorm:"not null;type:text;index"`
	LevelID   string  `json:"level_id" gorm:"not null;type:text;index"`
	ParentID  *string `json:"parent_id" gorm:"type:text;index"` // NULL for root nodes
	SortOrder int     `json:"sort_order" gorm:"not null;default:1"`
	Path      string  `json:"path" gorm:"not null;type:text;index"` // e.g., "001.002.005"
	Depth     int     `json:"depth" gorm:"not null"`                // starts from 1 for root

	// Associations
	Subject  *Subject      `json:"subject,omitempty" gorm:"foreignKey:SubjectID;constraint:OnDelete:CASCADE"`
	Level    *Level        `json:"level,omitempty" gorm:"foreignKey:LevelID;constraint:OnDelete:CASCADE"`
	Parent   *RoadmapNode  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []RoadmapNode `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}

// BeforeCreate generates UUID for new roadmap node
func (rn *RoadmapNode) BeforeCreate(tx *gorm.DB) error {
	if rn.ID == "" {
		rn.ID = uuid.New().String()
	}
	return nil
}

// GetPathArray returns the path as an array of segments
func (rn *RoadmapNode) GetPathArray() []string {
	if rn.Path == "" {
		return []string{}
	}
	return strings.Split(rn.Path, ".")
}

// IsRoot checks if this node is a root node
func (rn *RoadmapNode) IsRoot() bool {
	return rn.ParentID == nil
}

// IsLeaf checks if this node is a leaf node (has no children)
func (rn *RoadmapNode) IsLeaf() bool {
	return len(rn.Children) == 0
}

// GetNextSortOrder calculates the next sort order for a child node
func (rn *RoadmapNode) GetNextSortOrder(db *gorm.DB) (int, error) {
	var maxOrder int
	err := db.Model(&RoadmapNode{}).
		Where("parent_id = ?", rn.ID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error

	if err != nil {
		return 0, err
	}

	return maxOrder + 1, nil
}

// BuildPath constructs the materialized path for this node
func (rn *RoadmapNode) BuildPath(db *gorm.DB) error {
	if rn.ParentID == nil {
		// Root node
		rn.Path = fmt.Sprintf("%03d", rn.SortOrder)
		rn.Depth = 1
		return nil
	}

	// Get parent node
	var parent RoadmapNode
	if err := db.First(&parent, "id = ?", *rn.ParentID).Error; err != nil {
		return err
	}

	// Construct path
	rn.Path = fmt.Sprintf("%s.%03d", parent.Path, rn.SortOrder)
	rn.Depth = parent.Depth + 1

	return nil
}

// RoadmapNodeService provides business logic for roadmap operations
type RoadmapNodeService struct {
	db *gorm.DB
}

// NewRoadmapNodeService creates a new roadmap node service
func NewRoadmapNodeService(db *gorm.DB) *RoadmapNodeService {
	return &RoadmapNodeService{db: db}
}

// GetSubjectRoadmap returns the complete roadmap tree for a subject
func (s *RoadmapNodeService) GetSubjectRoadmap(subjectID string) ([]RoadmapNode, error) {
	var nodes []RoadmapNode
	err := s.db.Where("subject_id = ?", subjectID).
		Order("path ASC").
		Preload("Level").
		Find(&nodes).Error

	if err != nil {
		return nil, err
	}

	// Build tree structure
	return s.buildTree(nodes), nil
}

// buildTree converts flat list to hierarchical tree
func (s *RoadmapNodeService) buildTree(nodes []RoadmapNode) []RoadmapNode {
	nodeMap := make(map[string]*RoadmapNode)
	var roots []RoadmapNode

	// Create map of all nodes
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
		nodes[i].Children = []RoadmapNode{} // Initialize children slice
	}

	// Build parent-child relationships
	for i := range nodes {
		node := &nodes[i]
		if node.ParentID != nil {
			if parent, exists := nodeMap[*node.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		} else {
			roots = append(roots, *node)
		}
	}

	return roots
}

// CreateNode creates a new roadmap node
func (s *RoadmapNodeService) CreateNode(node *RoadmapNode) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Generate next sort order
		if node.ParentID != nil {
			var maxOrder int
			err := tx.Model(&RoadmapNode{}).
				Where("parent_id = ?", *node.ParentID).
				Select("COALESCE(MAX(sort_order), 0)").
				Scan(&maxOrder).Error
			if err != nil {
				return err
			}
			node.SortOrder = maxOrder + 1
		} else {
			var maxOrder int
			err := tx.Model(&RoadmapNode{}).
				Where("subject_id = ? AND parent_id IS NULL", node.SubjectID).
				Select("COALESCE(MAX(sort_order), 0)").
				Scan(&maxOrder).Error
			if err != nil {
				return err
			}
			node.SortOrder = maxOrder + 1
		}

		// Build path
		if err := node.BuildPath(tx); err != nil {
			return err
		}

		// Create node
		return tx.Create(node).Error
	})
}

// MoveNode moves a node to a new parent or position
func (s *RoadmapNodeService) MoveNode(nodeID string, newParentID *string, newSortOrder int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var node RoadmapNode
		if err := tx.First(&node, "id = ?", nodeID).Error; err != nil {
			return err
		}

		// Update parent and sort order
		node.ParentID = newParentID
		node.SortOrder = newSortOrder

		// Rebuild path
		if err := node.BuildPath(tx); err != nil {
			return err
		}

		// Update node
		if err := tx.Save(&node).Error; err != nil {
			return err
		}

		// Update all descendant paths
		return s.updateDescendantPaths(tx, &node)
	})
}

// updateDescendantPaths recursively updates paths for all descendants
func (s *RoadmapNodeService) updateDescendantPaths(tx *gorm.DB, node *RoadmapNode) error {
	var children []RoadmapNode
	if err := tx.Where("parent_id = ?", node.ID).Find(&children).Error; err != nil {
		return err
	}

	for _, child := range children {
		if err := child.BuildPath(tx); err != nil {
			return err
		}
		if err := tx.Save(&child).Error; err != nil {
			return err
		}
		if err := s.updateDescendantPaths(tx, &child); err != nil {
			return err
		}
	}

	return nil
}
