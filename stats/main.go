package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

// 从目标目录中挑选符合要求的文件, eg: `/2021/01/02/${fileName}.md`
func pickArticlefile(files *[]string, denyList *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		check(err)

		if filepath.Ext(path) != ".md" {
			return nil
		}

		for _, denyItem := range *denyList {
			if strings.Contains(path, denyItem) {
				return nil
			}
		}

		checkArchiveStruct := regexp.MustCompile(`/\d{4}\/\d{2}\/\d{2}\/`)
		matched := checkArchiveStruct.Match([]byte(path))
		if !matched {
			return nil
		}

		*files = append(*files, path)
		return nil
	}
}

// 检查文件关联的元文件，并提取文件夹和元文件
func relativeMetaFile(filePath string, metaFile *string, directory *string) bool {
	dirName, fullName := filepath.Split(filePath)
	fileName := strings.TrimSuffix(fullName, filepath.Ext(fullName))
	*metaFile = filepath.Join(dirName, fileName+".json")
	*directory = dirName

	if _, err := os.Stat(*metaFile); err != nil {
		return false
	}

	return true
}

// meta 数据结构
type Category struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// meta 需要使用的数据结构
type Meta struct {
	Date       string     `json:"date"`
	Tags       []string   `json:"tag"`
	Categories []Category `json:"categories"`
}

// 从文件读取数据
func readFile(filePath string) []byte {
	content, ioError := ioutil.ReadFile(filePath)

	if ioError != nil {
		return []byte{}
	}
	return content
}

// 解析元文件中的相关数据
func extractMeta(fileData []byte) (Status bool, date string, time string, Tags []string, Categories []string) {
	var meta Meta
	parseError := json.Unmarshal(fileData, &meta)
	if parseError != nil {
		fmt.Println(parseError)
		return false, "", "", nil, nil
	}

	reDate := regexp.MustCompile(`\s+`)
	dateParts := reDate.Split(meta.Date, -1)
	date = dateParts[0]
	time = dateParts[1]

	var categories []string
	categoryCount := len(meta.Categories)
	if categoryCount > 0 {
		for i := 0; i < categoryCount; i++ {
			categories = append(categories, meta.Categories[i].Name)
		}
	}

	return true, date, time, meta.Tags, categories
}

// 解析日期中的数据
func extractDate(date string, time string) (year string, month string, day string, hour string, yearAndMonth string) {
	reDate := regexp.MustCompile(`\-`)
	dateParts := reDate.Split(date, -1)

	year = dateParts[0]
	month = dateParts[1]
	day = dateParts[2]

	findYearMonth := regexp.MustCompile(`^\d{4}-\d{2}`)
	yearAndMonth = string(findYearMonth.Find([]byte(date)))

	reTime := regexp.MustCompile(`\:`)
	timeParts := reTime.Split(time, -1)
	hour = timeParts[0]

	return year, month, day, hour, yearAndMonth
}

type JSON struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func toJSON(data []string, limit int) []JSON {
	topN := arrayStats(data, limit)
	var result []JSON
	for _, num := range topN {
		result = append(result, JSON{num.Name, num.Value})
	}
	return result
}

func arrayStats(keys []string, limit int) []JSON {
	m := make(map[string]int)
	for _, key := range keys {
		if _, ok := m[key]; ok {
			m[key]++
		} else {
			m[key] = 1
		}
	}

	keyCounts := make([]JSON, 0, len(m))
	for key, val := range m {
		keyCounts = append(keyCounts, JSON{Name: key, Value: val})
	}

	sort.Slice(keyCounts, func(i, j int) bool {
		return keyCounts[i].Value > keyCounts[j].Value
	})

	retCounts := make([]JSON, 0, limit)

	for i := 0; i < len(keyCounts) && i < limit; i++ {
		retCounts = append(retCounts, keyCounts[i])
	}

	return retCounts
}

var WeekMap = map[string]string{
	"Monday":    "周一",
	"Tuesday":   "周二",
	"Wednesday": "周三",
	"Thursday":  "周四",
	"Friday":    "周五",
	"Saturday":  "周六",
	"Sunday":    "周日",
}

// 读取所有的元文件
func readMetaFiles(root string) (metaFiles []string) {
	denyList := []string{
		".DS_Store",
		".git",
		".gitignore",
		".gitea",
		"README.json",
		"README.md",
	}

	var files []string
	err := filepath.Walk(root, pickArticlefile(&files, &denyList))
	check(err)

	for _, file := range files {
		metaFile := ""
		directory := ""
		exists := relativeMetaFile(file, &metaFile, &directory)
		if exists {
			metaFiles = append(metaFiles, metaFile)
		} else {
			fmt.Printf("存在缺少 Meta 文件的内容: %s", file)
			fmt.Println()
		}
	}

	return metaFiles
}

// 合并数据
func concat(metaFiles *[]string) (tags []string, categories []string, years []string, months []string, days []string, hours []string, yearMonths []string, weeks []string) {
	for _, meta := range *metaFiles {
		metaData := readFile(meta)
		isOk, dateStr, timeStr, tags, categories := extractMeta(metaData)
		if isOk {
			year, month, day, hour, yearAndMonth := extractDate(dateStr, timeStr)

			t, _ := time.Parse("2006-01-02", dateStr)
			week := WeekMap[t.Weekday().String()]
			weeks = append(weeks, week)

			years = append(years, year)
			months = append(months, month)
			days = append(days, day)
			hours = append(hours, hour)
			yearMonths = append(yearMonths, yearAndMonth)

			tagCount := len(tags)
			if tagCount > 0 {
				for i := 0; i < tagCount; i++ {
					tags = append(tags, tags[i])
				}
			}

			cateCount := len(categories)
			if cateCount > 0 {
				for i := 0; i < cateCount; i++ {
					categories = append(categories, categories[i])
				}
			}
		}
	}

	return tags, categories, years, months, days, hours, yearMonths, weeks
}

// 计算各个维度 TOP N
func calcTopN(tags *[]string, categories *[]string, years *[]string, months *[]string, days *[]string, hours *[]string, yearMonths *[]string, weeks *[]string) (result []byte) {
	curYear := time.Now().Year()
	repeatYear := curYear - 2007

	type Response struct {
		Timpstamp    time.Time `json:"timestamp,omitempty"`
		Tag          []JSON    `json:"byTag,omitmepty"`
		Category     []JSON    `json:"byCategory,omitmepty"`
		Year         []JSON    `json:"byYear,omitmepty"`
		Month        []JSON    `json:"byMonth,omitmepty"`
		Day          []JSON    `json:"byDay,omitmepty"`
		Hour         []JSON    `json:"byHour,omitmepty"`
		Week         []JSON    `json:"byWeek,omitmepty"`
		YearAndMonth []JSON    `json:"byYearAndMonth,omitmepty"`
	}

	result, err := json.Marshal(Response{
		Timpstamp:    time.Now().UTC(),
		Tag:          toJSON(*tags, 30),
		Category:     toJSON(*categories, 10),
		Year:         toJSON(*years, repeatYear),
		Month:        toJSON(*months, 12),
		Day:          toJSON(*days, 31),
		Hour:         toJSON(*hours, 24),
		Week:         toJSON(*weeks, 7),
		YearAndMonth: toJSON(*yearMonths, repeatYear*12),
	})
	check(err)

	return result
}

func main() {
	flag.Parse()
	src := flag.Arg(0)

	metaFiles := readMetaFiles(src)
	tags, categories, years, months, days, hours, yearMonths, weeks := concat(&metaFiles)
	data := calcTopN(&tags, &categories, &years, &months, &days, &hours, &yearMonths, &weeks)

	os.MkdirAll("./report", os.ModePerm)
	err := ioutil.WriteFile("./report/stats.json", data, 0644)

	fmt.Println(string(data))

	check(err)
}
