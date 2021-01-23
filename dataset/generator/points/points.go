package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/semaphore"
	"os"
	"regexp"
	"strings"
	"sync"
)

func main() {
	strCh := make(chan string, 1024)
	writePoints(strCh, "./dataset/points.txt")
	readFiles(strCh, "./dataset/subjects/data")
	select {}
}

func writePoints(recvCh chan string, path string) {
	go func() {
		var file *os.File
		os.Remove(path)
		file, err := os.Open(path)
		if err != nil {
			file, err = os.Create(path)
			if err != nil {
				panic(err)
			}
		}
		defer file.Close()
		for str := range recvCh {
			file.WriteString(str + "\n")
		}
		fmt.Println("完成")
		os.Exit(0)
	}()
}

func readFiles(outCh chan string, path string) {
	wg := sync.WaitGroup{}
	weight := semaphore.NewWeighted(1000)
	fs := afero.NewOsFs()
	afero.Walk(fs, path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !regexp.MustCompile(`\.json$`).MatchString(info.Name()) {
			return nil
		}
		if err := weight.Acquire(context.Background(), 1); err != nil {
			panic(err)
		}
		go func() {
			defer weight.Release(1)
			defer wg.Done()
			wg.Add(1)
			readFile(path, outCh)
		}()
		return nil
	})
	wg.Wait()
	close(outCh)
}

func readFile(path string, outCh chan string) {
	outCh <- parseSubject(path)
}

// 解析单个 Subject (JSON 格式)
// 格式:
//   id title type score total tags
//   1,标题,4,7.9,185,tag1|tag2|tag3
func parseSubject(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 提前开辟内存
	buf := make([]byte, 1024)
	reader := bufio.NewReader(file)
	_, err = reader.Read(buf)
	if err != nil {
		panic(err)
	}

	str := string(buf)
	id := gjson.Get(str, "id").String()
	typ := gjson.Get(str, "type").String()
	name := gjson.Get(str, "name").String()
	rateTotal := gjson.Get(str, "rating.total").String()
	rateScore := gjson.Get(str, "rating.score").String()

	tags := gjson.Get(str, "tags.#.name").Array()
	tagsArr := make([]string, 0)
	for _, tag := range tags {
		tagStr := tag.String()
		if tagStr == "" {
			continue
		}
		tagsArr = append(tagsArr, tag.String())
	}
	tagsStr := strings.Join(tagsArr, "|")

	pt := strings.Join([]string{id, base64.StdEncoding.EncodeToString([]byte(name)), typ, rateScore, rateTotal, tagsStr}, ",")
	return pt
}
