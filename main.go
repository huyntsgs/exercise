package main

import (
	"log"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

const (
	MIN_SEC   int = 1
	MAX_SEC   int = 100
	MIN_VAL   int = 1000
	MAX_VAL   int = 10000
	LIST_SIZE int = 10000
)

type Pair struct {
	Second     int
	Value      int
	CreatedSec int64
}
type PairTest struct {
	Pair
	ID int
}
type List struct {
	pairs []Pair
	sync.RWMutex
}

type SumInfo struct {
	idx       int
	pairInUse *sync.Map
}

type SumList struct {
	sync.RWMutex
	sums []int
}

func (pair Pair) isExpired() bool {
	lived := currentMillisSec() - pair.CreatedSec
	if int(lived) > pair.Second*1000 {
		return true
	}
	return false
}

func randInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}
func currentMillisSec() int64 {
	return time.Now().UTC().Round(time.Millisecond).UnixNano() /
		(int64(time.Millisecond) / int64(time.Nanosecond))
}

func StartJob(list *List, sumInfo *SumInfo) {
	job := cron.New()
	job.AddFunc("* * * * * *", func() {
		list.Lock()
		defer list.Unlock()
		//list is saved as lastest at the end
		pair := Pair{randInt(MIN_SEC, MAX_SEC), randInt(MIN_VAL, MAX_VAL), currentMillisSec()}
		list.pairs = append(list.pairs, pair)
		//log.Println("pairs sz", len(list.pairs))
		log.Println("pairs ", list.pairs)

		//call goroutine to check expired
		go RemoveExpiredPairs(list, sumInfo)

	})
	job.Start()
}

func RemoveExpiredPairs(data *List, sumInfo *SumInfo) {
	data.Lock()
	defer data.Unlock()
	data.pairs = CheckExpiredParallel(data.pairs, sumInfo)
	if len(data.pairs) > LIST_SIZE {
		start := len(data.pairs) - LIST_SIZE
		data.pairs = data.pairs[start:]
		log.Printf("remove %d items in list %d", start, len(data.pairs))
	}
}
func getLastest20(data *List) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		//log.Println("len of data", len(data.pairs))
		i := 0
		result := make([]Pair, 0)
		for {
			if i > len(data.pairs)-1 || i == 20 || len(data.pairs) == 0 {
				break
			}
			result = append(result, data.pairs[len(data.pairs)-i-1])
			i++
		}
		c.JSON(http.StatusOK, gin.H{
			"error":  false,
			"result": result,
		})
	}
	return gin.HandlerFunc(fn)
}

func getSum(data *List, sumInfo *SumInfo, sumList *SumList) gin.HandlerFunc {

	fn := func(c *gin.Context) {
		result := 0
		var newMaps sync.Map
		//check whether there is enough next 10 oldest number for calculate
		if len(data.pairs) >= sumInfo.idx*10+10 {
			data.Lock()
			defer data.Unlock()
			//log.Printf("getSum sumInfo b ", sumInfo)
			for i := sumInfo.idx * 10; i < sumInfo.idx*10+10; i++ {
				result += data.pairs[i].Value
				newMaps.Store(data.pairs[i].CreatedSec, 0)
			}
			sumInfo.pairInUse = &newMaps
			sumList.sums = append(sumList.sums, result)
			if len(sumList.sums) > 0 {
				sumInfo.idx += 1
			}
			//			log.Printf("getSum sumInfo e ", sumInfo.idx)
			//			log.Printf("getSum sumList e ", sumList.sums)

		} else {
			//log.Printf("getSum not enough", sumList.sums)
			if len(sumList.sums) > 0 {
				result = sumList.sums[sumInfo.idx]
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"error":  false,
			"result": result,
		})
	}

	return gin.HandlerFunc(fn)
}
func getMedian(data *List, sumList *SumList) gin.HandlerFunc {

	fn := func(c *gin.Context) {
		var result float32
		n := len(sumList.sums)
		if n > 0 {
			sumList.Lock()
			defer sumList.Unlock()
			//we can sort immediately in getSum for better performance
			sort.Sort(sort.IntSlice(sumList.sums))
			if n%2 > 0 {
				result = float32(sumList.sums[n/2])
			} else {
				median := sumList.sums[(n-1)/2] + sumList.sums[n/2]
				result = float32(median) / 2
			}
		}

		//		log.Println("getMedian sums", sumList.sums)
		//		log.Println("getMedian result %d", result)

		c.JSON(http.StatusOK, gin.H{
			"error":  false,
			"result": result,
		})
	}

	return gin.HandlerFunc(fn)
}

func getHome(data *List, index int) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"home.html",
			nil,
		)
	}
	return gin.HandlerFunc(fn)
}

func StartGin(data *List, sumInfo *SumInfo, sumList *SumList) {
	router := gin.Default()
	router.LoadHTMLGlob("./*.html")

	router.GET("/home", getHome(data, 0))
	router.GET("/getSum/:id", getSum(data, sumInfo, sumList))
	router.GET("/getMedian", getMedian(data, sumList))
	router.GET("/getLastest20", getLastest20(data))
	router.Run(":8080")
}
func main() {

	pairs := make([]Pair, 0)
	//use this init if testing full 10000 items
	for i := 0; i < LIST_SIZE+10000; i++ {
		pair := Pair{randInt(MIN_SEC, MAX_SEC), randInt(MIN_VAL, MAX_VAL), currentMillisSec()}
		pairs = append(pairs, pair)
	}
	data := &List{pairs: pairs}

	//init data for struct
	var pairInUse sync.Map
	sumInfo := &SumInfo{pairInUse: &pairInUse}
	sumList := &SumList{sums: make([]int, 0)}
	rand.Seed(time.Now().UTC().UnixNano())

	StartJob(data, sumInfo)
	StartGin(data, sumInfo, sumList)
}
