package hamt_test

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt-functional/hamt64"
	"github.com/lleo/go-hamt-key"
	"github.com/lleo/go-hamt-key/stringkey"
	"github.com/pkg/errors"

	"github.com/lleo/stringutil"
)

// In a Hamt32 with 3149824 entries, there is 32 collisionLeafs. The creation
// of collisionLeafs happends after 3m+2k but before 3m+4k entries given my
// string "aaa" stringutil.Lower.Inc() string incrementer test harness.
var numKvs int = (3 * 1024 * 1024) + (4 * 1024) // between 3m+2k & 3m+4k

var KVS []key.KeyVal

//var SVS []StrVal

// Global variables to be initialized & shared between tests.
// These used to be initialized in main. This slowed down everything, even
// if they weren't used. This was changed that they are initialized only
// where they are needed and only if they are not already initialized.
var TestHamt32 hamt32.Hamt
var TestHamt64 hamt64.Hamt

var Inc = stringutil.Lower.Inc

var StartTime = make(map[string]time.Time)
var RunTime = make(map[string]time.Duration)

const (
	hybrid   = 0
	fullonly = 1
	componly = 2
)

var cfgStr = []string{"hybrid", "fullonly", "componly"}
var cfgMap = map[string]int{"hybrid": hybrid, "fullonly": fullonly, "componly": componly}

var TYP int
var CFG string

func TestMain(m *testing.M) {
	var fullonlyOpt, componlyOpt, hybridOpt, allOpt bool
	flag.BoolVar(&fullonlyOpt, "F", false, "Use full tables only and exclude C and H Options.")
	flag.BoolVar(&componlyOpt, "C", false, "Use compressed tables only and exclude F and H Options.")
	flag.BoolVar(&hybridOpt, "H", false, "Use compressed tables initially and exclude F and C Options.")
	flag.BoolVar(&allOpt, "A", false, "Run all Tests w/ Options set to hamt32.FullTablesOnly, hamt32.CompTablesOnly, and hamt32.HybridTables; in that order.")

	flag.Parse()

	// If allOpt flag set, ignore fullonlyOpt, componlyOpt, and hybridOpt.
	if !allOpt {

		// only one flag may be set between fullonlyOpt, componlyOpt, and hybridOpt
		if (fullonlyOpt && (componlyOpt || hybridOpt)) ||
			(componlyOpt && (fullonlyOpt || hybridOpt)) ||
			(hybridOpt && (componlyOpt || fullonlyOpt)) {
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	// If no flags given, run all tests.
	if !(allOpt || fullonlyOpt || componlyOpt || hybridOpt) {
		allOpt = true
		//fullonlyOpt = true
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

	KVS = buildKeyVals("global", numKvs, "aaa", 0)

	// execute
	var xit int

	if allOpt {
		//for _, TYP = range []int{fullonly, componly, hybrid} {
		for _, TYP = range []int{hybrid, componly, fullonly} {
			CFG = cfgStr[TYP]
			var name = "all tests: " + CFG
			StartTime[name] = time.Now()

			log.Printf("allOpt: for type = %s\n", CFG)

			fmt.Println("Running all tests:", CFG)
			xit = m.Run()
			if xit != 0 {
				os.Exit(xit)
			}

			RunTime[name] = time.Since(StartTime[name])
			fmt.Printf("RunTime[%q] = %v\n", name, RunTime[name])
		}
	} else {
		var msg string
		var name string
		if hybridOpt {
			TYP = hybrid
			CFG = cfgStr[TYP]
			msg = fmt.Sprintf("hybridOpt: for type = %s", CFG)
			name = "one test: " + CFG
		} else if fullonlyOpt {
			TYP = fullonly
			CFG = cfgStr[TYP]
			msg = fmt.Sprintf("fullonlyOpt: for type = %s", CFG)
			name = "one test: " + CFG
		} else /* if componlyOpt */ {
			TYP = componly
			CFG = cfgStr[TYP]
			msg = fmt.Sprintf("componlyOpt: for type = %s", CFG)
			name = "one test: " + CFG
		}

		StartTime[name] = time.Now()

		setLibrary(TYP)

		log.Println(msg)
		fmt.Println(msg)

		log.Printf("TestMain: GradeTables=%t; FullTableInit=%t\n", hamt32.GradeTables, hamt32.FullTableInit)

		xit = m.Run()

		RunTime[name] = time.Since(StartTime[name])
		fmt.Printf("RunTime[%q] = %v\n", name, RunTime[name])
	}

	log.Println("\n", RunTimes())
	log.Println("TestMain: the end.")

	// TEARDOWN

	os.Exit(xit)
}

func keysMapStringDuration(m map[string]time.Duration) []string {
	var ks = make([]string, len(m))
	i := 0
	for k := range m {
		ks[i] = k
		i++
	}
	return ks
}

func RunTimes() string {
	var ks = keysMapStringDuration(RunTime)
	sort.Strings(ks)

	var s = ""

	s += "Key                                                               Val\n"
	s += "=================================================================+==========\n"

	for _, k := range ks {
		v := RunTime[k]
		s += fmt.Sprintf("%-65s %s\n", k, v)
	}
	return s
}

var initializeNum int

func setLibrary(typ int) {
	switch typ {
	case hybrid:
		hamt32.GradeTables = true
		hamt32.FullTableInit = false
		//hamt64.GradeTables = true
		//hamt64.FullTableInit = false
	case fullonly:
		hamt32.GradeTables = false
		hamt32.FullTableInit = true
		//hamt64.GradeTables = false
		//hamt64.FullTableInit = true
	case componly:
		hamt32.GradeTables = false
		hamt32.FullTableInit = false
		//hamt64.GradeTables = false
		//hamt64.FullTableInit = false
	default:
		panic(fmt.Sprintf("unknown type %d", typ))
	}
}

func createHamt32(prefix string, kvs []key.KeyVal, typ int) hamt32.Hamt {
	var name = fmt.Sprintf("%s++createHamt32:%s", prefix, cfgStr[typ])
	setLibrary(typ)
	StartTime[name] = time.Now()

	var h = hamt32.Hamt{}

	for _, kv := range kvs {
		var inserted bool
		h, inserted = h.Put(kv.Key, kv.Val)
		if !inserted {
			log.Fatalf("%s: failed to h.Put(%s, %v)", prefix, kv.Key, kv.Val)
		}
	}

	RunTime[name] = time.Since(StartTime[name])

	return h
}

func createHamt64(prefix string, kvs []key.KeyVal, typ int) hamt64.Hamt {
	var name = fmt.Sprintf("%s++createHamt64:%s", prefix, cfgStr[typ])
	setLibrary(typ)
	StartTime[name] = time.Now()

	var h = hamt64.Hamt{}

	for _, kv := range kvs {
		var inserted bool
		h, inserted = h.Put(kv.Key, kv.Val)
		if !inserted {
			log.Fatalf("%s: failed to Hamt64.Put(%s, %v)", prefix, kv.Key, kv.Val)
		}
	}

	RunTime[name] = time.Since(StartTime[name])

	return h
}

func buildMap(prefix string, num int) (map[string]int, []string) {
	var name = fmt.Sprintf("%s-buildMap-%d", prefix, num)
	StartTime[name] = time.Now()

	var m = make(map[string]int, num)
	var k = make([]string, num)

	var s = "aaa"
	var v int = 0

	for i := 0; i < num; i++ {
		m[s] = v
		k[i] = s

		s = Inc(s)
		v++
	}

	RunTime[name] = time.Since(StartTime[name])

	return m, k
}

//type StrVal struct {
//	Str string
//	Val interface{}
//}
//
//func buildStrVals(prefix string, num int) []StrVal {
//	var name = fmt.Sprintf("%s-buildMap-%d", prefix, num)
//	StartTime[name] = time.Now()
//
//	var svs = make([]StrVal, num)
//	var s = "aaa"
//
//	for i := 0; i < num; i++ {
//		svs[s] = i
//
//		s = Inc(s)
//	}
//
//	RunTime[name] = time.Since(StartTime[name])
//	return svs
//}

func buildKeyVals(prefix string, num int, initStr string, initVal int) []key.KeyVal {
	var name = fmt.Sprintf("%s++buildKeyVals#%d", prefix, num)
	StartTime[name] = time.Now()

	var kvs = make([]key.KeyVal, num)
	var s = initStr

	var limit = initVal + num
	for i := initVal; i < limit; i++ {
		var k = stringkey.New(s)

		kvs[i] = key.KeyVal{k, i}
		s = Inc(s)
	}

	RunTime[name] = time.Since(StartTime[name])
	return kvs
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

////First genRandomizedSvs() copies []StrVal passed in. Then it randomizes that
////copy in-place. Finnally, it returns the randomized copy.
//func genRandomizedSvs(svs []StrVal) []StrVal {
//	var name = "genRandomizedSvs"
//	StartTime[name] = time.Now()
//
//	var randSvs = make([]StrVal, len(svs))
//	copy(randSvs, svs)
//
//	//From: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
//	var limit = len(randSvs) //n-1
//	for i := 0; i < limit; /* aka i_max = n-2 */ i++ {
//		j := rand.Intn(i+1) - 1 // i <= j < n; j_min=n-(n-2+1)-1=0; j_max=n-0-1=n-1
//		randSvs[i], randSvs[j] = randSvs[j], randSvs[i]
//	}
//
//	RunTime[name] = time.Since(StartTime[name])
//	return randSvs
//}

//func BenchmarkMapGet(b *testing.B) {
//	var name = "BenchmarkMapGet"
//	log.Printf("%s b.N=%d\n", name, b.N)
//
//	var _, ok = LookupMap["aaa"]
//	if !ok {
//		LookupMap, DeleteMap = buildMaps(numKvs)
//	}
//
//	StartTime[name] = time.Now()
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		var j = i % numKvs
//		var s = SVS[j].Str
//		var v = SVS[j].Val
//		var val, ok = LookupMap[s]
//		if !ok {
//			b.Fatalf("LookupMap[%q] does not exist", s)
//		}
//		if val != v {
//			b.Fatalf("LookupMap[%q] != %d", s, KVS[j].Val)
//		}
//	}
//
//	RunTime[name] = time.Since(StartTime[name])
//}
//
//func BenchmarkMapPut(b *testing.B) {
//	var name = "BenchmarkMapPut"
//	log.Printf("%s b.N=%d\n", name, b.N)
//	StartTime[name] = time.Now()
//	var m = make(map[string]int)
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		var j = i % numKvs
//		var s = SVS[j].Str
//		var v = SVS[j].Val
//		m[s] = v
//	}
//
//	RunTime[name] = time.Since(StartTime[name])
//}
//
//var rebuildDeleteMapNum int
//
//func rebuildDeleteMap(svs []StrVal) {
//	var name = fmt.Sprintf("BenchmarkMapPut-%d", rebuildDeleteMapNum)
//	rebuildDeleteMapNum++
//
//	StartTime[name] = time.Now()
//
//	for _, sv := range svs {
//		var _, ok = DeleteMap[sv.Str]
//		if ok {
//			break
//		}
//		//else
//		delete(DeleteMap, sv.Str)
//
//		DeleteMap[sv.Str] = sv.Val
//	}
//
//	RunTime[name] = time.Since(StartTime[name])
//}
//
//func BenchmarkMapDel(b *testing.B) {
//	var name = "BenchmarkMapDel"
//	log.Printf("%s b.N=%d\n", name, b.N)
//	StartTime[name] = time.Now()
//	rebuildDeleteMap(SVS)
//	RunTime[name] = time.Since(StartTime[name])
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		var j = i % numKvs
//		var k = SVS[j].Str
//		var v = SVS[j].Val
//
//		var val, ok = DeleteMap[k]
//		if ok {
//			delete(DeleteMap, k)
//		} else if val != v {
//			b.Fatalf("DeleteMap[%s],%d != %d", k, v, val)
//		}
//
//		//b.StopTimer()
//		DeleteMap[k] = v
//		//b.StartTimer()
//	}
//}
