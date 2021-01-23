package models

import (
	"fmt"
	"testing"
)

func TestCalculateSubjectsSimilar(t *testing.T) {
	a := &Subject{
		ID:          "297969",
		Name:        "ひぐらしのなく頃に",
		Type:        1,
		RatingTotal: 100,
		RatingScore: 7.1,
		Tags: map[string]int{
			"2020年10月": 10,
			"悬疑":       1,
		},
	}
	b := &Subject{
		ID:          "297969",
		Name:        "ひぐらしのなく頃に 業",
		Type:        1,
		RatingTotal: 100,
		RatingScore: 7.1,
		Tags: map[string]int{
			"2020年10月": 10,
			"悬":       1,
		},
	}
	resp := CalculateSubjectsSimilar(a, b)
	fmt.Println(resp)
}
