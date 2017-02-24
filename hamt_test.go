package hamt_test

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt-functional/hamt64"
	"github.com/lleo/go-hamt/key"
	"github.com/lleo/go-hamt/stringkey"
	"github.com/pkg/errors"

	"github.com/lleo/stringutil"
)

type StrVal struct {
	Str string
	Val int
}

var numKvs int = 2 * 1024 * 1024

var KVS []key.KeyVal
var SVS []StrVal

var LookupMap = make(map[string]int, numKvs)
var DeleteMap = make(map[string]int, numKvs)

var LookupHamt32 hamt32.Hamt
var DeleteHamt32 hamt32.Hamt

var LookupHamt64 hamt64.Hamt
var DeleteHamt64 hamt64.Hamt

var Inc = stringutil.Lower.Inc

var StartTime = make(map[string]time.Time)
var RunTime = make(map[string]time.Duration)

func TestMain(m *testing.M) {
	var fullonly, componly, hybrid, all bool
	flag.BoolVar(&fullonly, "F", false, "Use full tables only and exclude C and H Options.")
	flag.BoolVar(&componly, "C", false, "Use compressed tables only and exclude F and Hpppppppp Options.")
	flag.BoolVar(&hybrid, "H", false, "Use compressed tables initially and exclude F and C Options.")
	flag.BoolVar(&all, "A", false, "Run all Tests w/ Options set to hamt32.FullTablesOnly, hamt32.CompTablesOnly, and hamt32.HybridTables; in that order.")

	flag.Parse()

	// If all flag set, ignore fullonly, componly, and hybrid.
	if !all {

		// only one flag may be set between fullonly, componly, and hybrid
		if (fullonly && (componly || hybrid)) ||
			(componly && (fullonly || hybrid)) ||
			(hybrid && (componly || fullonly)) {
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// If no flags given, run all tests.
	if !(all || fullonly || componly || hybrid) {
		all = true
	}

	log.SetFlags(log.Lshortfile)

	var logfile, err = os.Create("test.log")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to os.Create(\"test.log\")"))
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	// SETUP
	log.Println("TestMain: and so it begins...")

	KVS, SVS = buildKeyVals(numKvs)

	//Build LookupMap & DeleteMap
	for i, sv := range SVS {
		var s = sv.Str
		LookupMap[s] = i
		DeleteMap[s] = i
	}

	// execute
	var xit int

	if all {
		//Full Tables Only
		hamt32.GradeTables = false
		hamt32.FullTableInit = true
		hamt64.GradeTables = false
		hamt64.FullTableInit = true
		log.Println("TestMain: Full Tables Only")
		log.Printf("TestMain: GradeTables=%t; FullTableInit=%t\n", hamt32.GradeTables, hamt32.FullTableInit)
		initialize()
		xit = m.Run()
		if xit != 0 {
			os.Exit(xit)
		}

		//Compressed Tables Only
		hamt32.GradeTables = false
		hamt32.FullTableInit = false
		hamt64.GradeTables = false
		hamt64.FullTableInit = false
		log.Println("TestMain: Compressed Tables Only")
		log.Printf("TestMain: GradeTables=%t; FullTableInit=%t\n", hamt32.GradeTables, hamt32.FullTableInit)
		initialize()
		xit = m.Run()
		if xit != 0 {
			os.Exit(xit)
		}

		//Hybrid Tables
		hamt32.GradeTables = true
		hamt32.FullTableInit = false
		hamt64.GradeTables = true
		hamt64.FullTableInit = false
		log.Println("TestMain: Hybrid Tables")
		log.Printf("TestMain: GradeTables=%t; FullTableInit=%t\n", hamt32.GradeTables, hamt32.FullTableInit)
		initialize()
		xit = m.Run()
	} else {
		if hybrid {
			hamt32.GradeTables = true
			hamt32.FullTableInit = false
			hamt64.GradeTables = true
			hamt64.FullTableInit = false
			log.Println("TestMain: Hybrid Tables")
		} else if fullonly {
			hamt32.GradeTables = false
			hamt32.FullTableInit = true
			hamt64.GradeTables = false
			hamt64.FullTableInit = true
			log.Println("TestMain: Full Tables Only")
		} else /* if componly */ {
			hamt32.GradeTables = false
			hamt32.FullTableInit = false
			hamt64.GradeTables = false
			hamt64.FullTableInit = false
			log.Println("TestMain: Compressed Tables Only")
		}

		log.Printf("TestMain: GradeTables=%t; FullTableInit=%t\n", hamt32.GradeTables, hamt32.FullTableInit)
		initialize()
		xit = m.Run()
	}

	log.Println("\n", RunTimes())
	log.Println("TestMain: the end.")

	// TEARDOWN

	os.Exit(xit)
}

func RunTimes() string {
	var s = ""

	s += "Key                                                               Val\n"
	s += "=================================================================+==========\n"

	for key, val := range RunTime {
		s += fmt.Sprintf("%-65s %s\n", key, val)
	}
	return s
}

var initializeNum int

func initialize() {
	var name = fmt.Sprintf("initialize-%d", initializeNum)
	initializeNum++
	StartTime[name] = time.Now()

	log.Printf("%s: GradeTables=%t; FullTableInit=%t\n", name, hamt32.GradeTables, hamt32.FullTableInit)

	LookupHamt32 = hamt32.Hamt{}
	DeleteHamt32 = hamt32.Hamt{}

	LookupHamt64 = hamt64.Hamt{}
	DeleteHamt64 = hamt64.Hamt{}

	for _, kv := range KVS {
		var inserted bool

		LookupHamt32, inserted = LookupHamt32.Put(kv.Key, kv.Val)
		if !inserted {
			log.Fatalf("failed to LookupHamt32.Put(%s, %v)", kv.Key, kv.Val)
		}

		DeleteHamt32, inserted = DeleteHamt32.Put(kv.Key, kv.Val)
		if !inserted {
			log.Fatalf("failed to DeleteHamt32.Put(%s, %v)", kv.Key, kv.Val)
		}

		LookupHamt64, inserted = LookupHamt64.Put(kv.Key, kv.Val)
		if !inserted {
			log.Fatalf("failed to LookupHamt64.Put(%s, %v)", kv.Key, kv.Val)
		}

		DeleteHamt64, inserted = DeleteHamt64.Put(kv.Key, kv.Val)
		if !inserted {
			log.Fatalf("failed to DeleteHamt64.Put(%s, %v)", kv.Key, kv.Val)
		}
	}

	RunTime[name] = time.Since(StartTime[name])
}

func buildKeyVals(num int) ([]key.KeyVal, []StrVal) {
	var name = "buildKeyVals"
	StartTime[name] = time.Now()

	var kvs = make([]key.KeyVal, num, num)
	var svs = make([]StrVal, num, num)

	s := "aaa"
	for i := 0; i < num; i++ {
		kvs[i].Key = stringkey.New(s)
		kvs[i].Val = i

		svs[i].Str = s
		svs[i].Val = i

		s = Inc(s)
	}

	RunTime[name] = time.Since(StartTime[name])
	return kvs, svs
}

//First genRandomizedSvs() copies []KeyVal passed in. Then it randomizes that
//copy in-place. Finnally, it returns the randomized copy.
func genRandomizedKvs(kvs []key.KeyVal) []key.KeyVal {
	var name = "genRandomizedKvs"
	StartTime[name] = time.Now()

	var randKvs = make([]key.KeyVal, len(kvs))
	copy(randKvs, kvs)

	//From: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
	var n = len(randKvs) // n is the number of elements
	var limit = n - 1
	for i := 0; i < limit; /* aka i_max = n-2 */ i++ {
		j := n - rand.Intn(i+1) - 1 // i <= j < n
		// j_min = 0   => n - (i_max + 1) - 1 = n - (n-2 + 1) - 1 = n-n+2-1-1 = 0
		// j_max = n-1 => n - 0 - 1 = n - 1
		randKvs[i], randKvs[j] = randKvs[j], randKvs[i]
	}

	RunTime[name] = time.Since(StartTime[name])
	return randKvs
}

//First genRandomizedSvs() copies []StrVal passed in. Then it randomizes that
//copy in-place. Finnally, it returns the randomized copy.
func genRandomizedSvs(svs []StrVal) []StrVal {
	var name = "genRandomizedSvs"
	StartTime[name] = time.Now()

	var randSvs = make([]StrVal, len(svs))
	copy(randSvs, svs)

	//From: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
	var limit = len(randSvs) //n-1
	for i := 0; i < limit; /* aka i_max = n-2 */ i++ {
		j := rand.Intn(i+1) - 1 // i <= j < n; j_min=n-(n-2+1)-1=0; j_max=n-0-1=n-1
		randSvs[i], randSvs[j] = randSvs[j], randSvs[i]
	}

	RunTime[name] = time.Since(StartTime[name])
	return randSvs
}

func BenchmarkMapGet(b *testing.B) {
	var name = "BenchmarkMapGet"
	log.Printf("%s b.N=%d\n", name, b.N)
	StartTime[name] = time.Now()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var j = i % numKvs
		var s = SVS[j].Str
		var v = SVS[j].Val
		var val, ok = LookupMap[s]
		if !ok {
			b.Fatalf("LookupMap[%s] does not exist", s)
		}
		if val != v {
			b.Fatalf("LookupMap[%s] != %d", s, KVS[j].Val)
		}
	}

	RunTime[name] = time.Since(StartTime[name])
}

func BenchmarkMapPut(b *testing.B) {
	var name = "BenchmarkMapPut"
	log.Printf("%s b.N=%d\n", name, b.N)
	StartTime[name] = time.Now()
	var m = make(map[string]int)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var j = i % numKvs
		var s = SVS[j].Str
		var v = SVS[j].Val
		m[s] = v
	}

	RunTime[name] = time.Since(StartTime[name])
}

var rebuildDeleteMapNum int

func rebuildDeleteMap(svs []StrVal) {
	var name = fmt.Sprintf("BenchmarkMapPut-%d", rebuildDeleteMapNum)
	rebuildDeleteMapNum++
	StartTime[name] = time.Now()

	for _, sv := range svs {
		var _, ok = DeleteMap[sv.Str]
		if ok {
			break
		}
		//else
		delete(DeleteMap, sv.Str)

		DeleteMap[sv.Str] = sv.Val
	}

	RunTime[name] = time.Since(StartTime[name])
}

func BenchmarkMapDel(b *testing.B) {
	var name = "BenchmarkMapDel"
	log.Printf("%s b.N=%d\n", name, b.N)
	StartTime[name] = time.Now()
	rebuildDeleteMap(SVS)
	RunTime[name] = time.Since(StartTime[name])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var j = i % numKvs
		var k = SVS[j].Str
		var v = SVS[j].Val

		var val, ok = DeleteMap[k]
		if ok {
			delete(DeleteMap, k)
		} else if val != v {
			b.Fatalf("DeleteMap[%s],%d != %d", k, v, val)
		}

		//b.StopTimer()
		DeleteMap[k] = v
		//b.StartTimer()
	}
}
