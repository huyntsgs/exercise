package main

import (
	"sync"
	"testing"
	"time"
)

var data []Pair

func init() {
	data = make([]Pair, 0)
	for i := 0; i < 100000; i++ {
		pair := Pair{randInt(MIN_SEC, MAX_SEC), randInt(MIN_VAL, MAX_VAL), currentMillisSec()}
		data = append(data, pair)
	}
}

//func BenchmarkCheckLoop(b *testing.B) {
//	b.ReportAllocs()
//	for i := 0; i < b.N; i++ {
//		CheckLoop(data)
//	}
//	//log.Printf("len data %d", len(data))

//}

func TestCheckExpiredParallel(t *testing.T) {
	data := make([]Pair, 0)
	for i := 0; i < 1000000; i++ {
		pair := Pair{randInt(MIN_SEC, MAX_SEC), randInt(MIN_VAL, MAX_VAL), currentMillisSec()}
		data = append(data, pair)
	}
	//sleep to create huge number of expired pairs
	time.Sleep(time.Second * 80)
	var pairInUse sync.Map
	sumInfo := &SumInfo{pairInUse: &pairInUse}
	CheckExpiredParallel(data, sumInfo)

}
