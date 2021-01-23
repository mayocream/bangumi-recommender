package models

import (
	"github.com/antlabs/strsim"
	"math"
	"sort"
	"strconv"
)

type RelationBuilder struct {
	Subjects Subjects
}

func NewRelationBuilder() *RelationBuilder {
	return &RelationBuilder{}
}

type RelationResult struct {
	ID    string
	Score float64
}

type RelationResults []*RelationResult

func (r RelationResults) Len() int {
	return len(r)
}
func (r RelationResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
func (r RelationResults) Less(i, j int) bool {
	return r[j].Score < r[i].Score
}

func (r *RelationBuilder) Top(rr *RelationResults, n int) []Subject {
	subs := make([]Subject, 0, n)
	for i := 0; i < n; i++ {
		id := (*rr)[i].ID
		sub := r.Subjects[id]
		sub.RelationScore = strconv.FormatFloat((*rr)[i].Score, 'E', -1, 64)
		subs = append(subs, *sub)
	}
	return subs
}

func (r *RelationBuilder) Calculate(target *Subject) *RelationResults {
	relationScores := make(RelationResults, 0)
	for id, sub := range r.Subjects {
		if id == target.ID {
			continue
		}
		rr := &RelationResult{
			ID:    id,
			Score: CalculateSubjectsSimilar(sub, target),
		}
		relationScores = append(relationScores, rr)
	}
	sort.Sort(relationScores)
	return &relationScores
}

func CalculateSubjectsSimilar(a *Subject, b *Subject) float64 {
	var similar float64
	nameSimilar := strsim.Compare(a.Name, b.Name)
	similar += nameSimilar * 10

	var tagsSimilar float64
	var tagsSameCount int
	for bTagName, _ := range b.Tags {
		if _, ok := a.Tags[bTagName]; ok {
			tagsSameCount++
		}
	}
	if tagsSameCount > 0 && len(a.Tags) > 0 && len(b.Tags) > 0 {
		tagsSimilar = float64(tagsSameCount) / math.Sqrt(float64(len(a.Tags)*len(b.Tags)))
		similar += tagsSimilar * 80
	}
	return similar / 100
}
