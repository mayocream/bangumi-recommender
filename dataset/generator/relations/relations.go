package main

import (
	"bangumi-recommender/models"
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/afero"
	"golang.org/x/sync/semaphore"
	"os"
	"strconv"
	"strings"
	"sync"
)

type pass struct {
	Relations []models.Subject
	ID        string
}

func main() {
	subjects := readPoints("./dataset/points.txt")
	fmt.Printf("数量: %v", len(subjects))

	builder := models.NewRelationBuilder()
	builder.Subjects = subjects

	outCh := make(chan *pass, 100)

	go writeRelations(outCh)

	wg := sync.WaitGroup{}
	weight := semaphore.NewWeighted(100)
	for _, target := range subjects {
		wg.Add(1)
		weight.Acquire(context.Background(), 1)
		go func(target *models.Subject) {
			defer wg.Done()
			defer weight.Release(1)
			result := builder.Calculate(target)
			if result.Len() == 0 {
				panic(errors.New("查询关系失败"))
			}
			topn := builder.Top(result, 10)
			if topn == nil {
				panic(errors.New("没有关联项"))
			}
			outCh <- &pass{
				Relations: topn,
				ID:        target.ID,
			}
		}(target)
	}
	wg.Wait()
	close(outCh)
	select {}
}

func writeRelations(recvCh chan *pass) {
	fs := afero.NewOsFs()
	fs.RemoveAll("./relations")
	for p := range recvCh {
		go func(p *pass) {
			if p == nil || p.Relations == nil {
				fmt.Println("没有 p")
				return
			}
			bytes, err := jsoniter.Marshal(p.Relations)
			if err != nil {
				panic(err)
			}
			parent := p.ID[:len(p.ID)-1]
			if parent == "" {
				parent = "0"
			}
			if err := fs.MkdirAll("relations/"+parent, os.ModePerm); err != nil {
				panic(err)
			}
			if err := afero.WriteFile(fs, "./relations/"+parent+"/"+p.ID+".json", bytes, os.ModePerm); err != nil {
				panic(err)
			}
			fmt.Printf("写 %s \n", p.ID)
		}(p)
	}
	os.Exit(0)
}

func readPoints(path string) models.Subjects {
	subjects := make(models.Subjects)
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		strArr := strings.Split(line, ",")
		typ, err := strconv.Atoi(strArr[2])
		if err != nil {
			panic(err)
		}
		ratingScore, err := strconv.ParseFloat(strArr[3], 64)
		if err != nil {
			fmt.Println("没有评分: ", line)
			ratingScore = 0
		}
		ratingTotal, err := strconv.Atoi(strArr[4])
		if err != nil {
			fmt.Println("没有评分人数: ", line)
			ratingTotal = 0
		}
		name, err := base64.StdEncoding.DecodeString(strArr[1])
		if err != nil {
			panic(err)
		}
		subject := &models.Subject{
			ID:          strArr[0],
			Name:        string(name),
			Type:        typ,
			RatingScore: ratingScore,
			RatingTotal: int64(ratingTotal),
			Tags:        make(map[string]int, 0),
		}
		tags := strings.Split(strArr[5], "|")
		if len(tags) > 0 {
			for _, tag := range tags {
				if tag != "" {
					subject.Tags[tag] = 1
				}
			}
		}
		if len(subject.Tags) == 0 {
			fmt.Println("没有 tags: ", subject.ID)
		}
		subjects[subject.ID] = subject
	}
	return subjects
}
