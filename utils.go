package main

import (
	"log"
	"reflect"
	"sync"
	"time"
)

func CheckExpiredWg1(s []Pair, i int, maps *sync.Map, wg *sync.WaitGroup, sumInfo *SumInfo) {
	//defer timeTrack(time.Now(), "CheckLoop")

	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := sumInfo.pairInUse.Load(s[i].CreatedSec); !ok {
			if s[i].isExpired() {
				s = append(s[:i], s[i+1:]...)
			}
		}
	}
	//log.Printf("after check loop len %d", len(s))

	maps.Store(i, s)
	wg.Done()

}
func CheckExpiredWg(s []Pair, i int, maps *sync.Map, sumInfo *SumInfo) {
	//defer timeTrack(time.Now(), "CheckLoop")

	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := sumInfo.pairInUse.Load(s[i].CreatedSec); !ok {
			if s[i].isExpired() {
				s = append(s[:i], s[i+1:]...)
			}
		}
	}
	//log.Printf("after check loop len %d", len(s))

	maps.Store(i, s)

}
func CheckExpiredSimple(s []Pair, sumInfo *SumInfo) []Pair {
	//defer timeTrack(time.Now(), "CheckLoop")

	for i := len(s) - 1; i >= 0; i-- {
		if _, ok := sumInfo.pairInUse.Load(s[i].CreatedSec); !ok {
			if s[i].isExpired() {
				s = append(s[:i], s[i+1:]...)
			}
		}
	}
	//log.Printf("after check loop len %d", len(s))
	return s
}
func CheckLoopTest(s []PairTest) []PairTest {
	defer timeTrack(time.Now(), "CheckLoopTest")
	for i := len(s) - 1; i >= 0; i-- {
		if s[i].isExpired() {
			s = append(s[:i], s[i+1:]...)
		}
	}
	log.Printf("after check loop len %d", len(s))
	log.Printf("slice after", s)
	return s
}

//split array of pair smaller arrays to check parallel, in order to increase performance of checking
//performance is much better than linear loop
func CheckExpiredParallel(s []Pair, sumInfo *SumInfo) []Pair {

	defer timeTrack(time.Now(), "CheckExpiredParallel")

	n := len(s) / 1000
	var maps sync.Map

	wg := &sync.WaitGroup{}
	wg.Add(n)

	if len(s) <= 1000 {
		return CheckExpiredSimple(s, sumInfo)
	}

	for i := 0; i < n; i++ {
		startIdx := 1000 * i
		endIdx := 1000 * (i + 1)

		go func(i int) {
			//log.Printf("starti %d, endIdx %d, i %d", startIdx, endIdx, i)
			CheckExpiredWg(s[startIdx:endIdx], i, &maps, sumInfo)
			wg.Done()
		}(i)
	}

	wg.Wait()
	result := make([]Pair, 0)
	for i := 0; i < n; i++ {
		subRet, ok := maps.Load(i)
		if ok && subRet != nil {
			result = append(result, subRet.([]Pair)...)
		} else {
			log.Printf("nil at %d", i)
		}
		log.Printf("type of subRet %s", reflect.TypeOf(subRet))

	}
	//log.Printf("sz end of CheckExpiredParallel %d", len(result))
	return result
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
