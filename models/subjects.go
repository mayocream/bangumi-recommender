package models

type Subject struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        int            `json:"type"`
	RatingTotal int64          `json:"ratingTotal"`
	RatingScore float64        `json:"ratingScore"`
	Tags        map[string]int `json:"tags"`

	RelationScore string
}

type Subjects map[string]*Subject
