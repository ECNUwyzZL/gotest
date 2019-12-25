package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"k8s.io/klog"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const WORKER_NUM = 20
const BATCH_SIZE = 20

type HalsteadResult struct {
	N1      int `json:"N1"`
	Ln1     int `json:"n1"`
	N2      int `json:"N2"`
	Ln2     int `json:"n2"`
	LineNum int `json:"line_num"`
}

type HalsteadFullResult struct {
	Size       int     `json:"size"`
	Vocab_size int     `json:"vocab_size"`
	Volume     float64 `json:"volume"`
	Difficulty float64 `json:"difficulty"`
	Level      float64 `json:"level"`
	Effort     float64 `json:"effort"`
}

type RedundancyPair struct {
	f            string
	s            string
	multiplicity int
}

var redundancyPairs []RedundancyPair

var operators = make(map[string]void)
var operatorCounts, operandCounts = make(map[string]int), make(map[string]int)

func popRedundancyPairs() {
	for i := range operators {
		for j := range operators {
			if i != j {
				occurNum := 0
				occurNum = strings.Count(j, i)
				if occurNum > 0 {
					redundancyPairs = append(redundancyPairs, RedundancyPair{
						f:            j,
						s:            i,
						multiplicity: occurNum,
					})
					klog.V(4).Info("re", j, i, occurNum)
				}
			}
		}
	}
}

func popualateOperators() {
	var vmember void
	for _, i := range ops {
		operators[i] = vmember
	}
}

func adjustRedundancy() {
	for _, v := range redundancyPairs {
		klog.V(4).Info("adj", v.f, operatorCounts[v.f], v.s, operatorCounts[v.s], v.multiplicity)
		operatorCounts[v.s] = operatorCounts[v.s] - operatorCounts[v.f]*v.multiplicity
	}
}

func worker(id int, jobs <-chan string, results chan<- HalsteadResult) {
	for j := range jobs {
		klog.V(4).Infof("worker", id, "started  job", j)
		time.Sleep(time.Second)
		klog.V(4).Infof("worker", id, "finished job", j)
		results <- HalsteadResult{1, 2, 3, 4, 1}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func run_metric(filePath string) {
	// fill the operators
	popualateOperators()
	popRedundancyPairs()

	// we now create a regex for identifier
	identifierDef, _ := regexp.Compile("[A-Za-z][A-Za-z0-9_]*")
	numberDef, _ := regexp.Compile("\\b([0-9]+)\\b")

	file, err := os.Open(filePath)
	check(err)
	defer file.Close()

	res := new(HalsteadResult)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err := scanner.Err()
		check(err)
		line := scanner.Text()
		klog.V(4).Info(line)
		res.LineNum += 1

		// now check for operators in the line
		for op := range operators {
			opCount := strings.Count(line, op)
			if opCount == 0 {
				continue
			}

			if _, exist := operatorCounts[op]; exist {
				klog.V(4).Info("before", line, opCount, op, operatorCounts[op])
				operatorCounts[op] += opCount
				klog.V(4).Info(line, opCount, op, operatorCounts[op])
			} else {
				operatorCounts[op] = opCount
				klog.V(4).Info("first", line, opCount, op, operatorCounts[op])
			}
		}

		// now lets check for identifiers
		for _, match := range identifierDef.FindAllString(line, -1) {
			_, exist := operators[match]
			if exist {
				continue
			}
			if _, exist := operandCounts[match]; exist {
				operandCounts[match]++
			} else {
				operandCounts[match] = 1
			}
		}

		// search for numbers
		for _, match := range numberDef.FindAllString(line, -1) {
			_, exist := operators[match]
			if exist {
				continue
			}
			if _, exist := operandCounts[match]; exist {
				operandCounts[match]++
			} else {
				operandCounts[match] = 1
			}
		}

	}
	klog.V(4).Infof("%+v\n", operatorCounts)
	adjustRedundancy()
	klog.V(4).Infof("%+v\n", operatorCounts)

	for idx, value := range operatorCounts {
		klog.V(4).Info(idx, value)
		if value != 0 {
			res.Ln1++
		}
		res.N1 += value
	}

	for _, value := range operandCounts {
		if value != 0 {
			res.Ln2++
		}
		res.N2 += value
	}

	klog.V(4).Infof("n1:%d, n2:%d, N1:%d, N2:%d\n", res.Ln1, res.Ln2, res.N1, res.N2)

	// compute the halstead metrics now

	resFull := new(HalsteadFullResult)
	// program size defined as the sum of
	// all operands and operators
	resFull.Size = res.N1 + res.N2

	// Vocabulary size -- Size of the vocabulary
	// defined as sum of distinct operands and operators
	resFull.Vocab_size = res.Ln1 + res.Ln2

	// Volume - Program Volume , defined as follows:
	// Volume = size x log ( vocab_size )
	resFull.Volume = float64(resFull.Size) * math.Log2(float64(resFull.Vocab_size))

	// Difficulty = ( n1/2 ) x ( N2/n2 ) and level = 1/difficulty
	resFull.Difficulty = (float64(res.Ln1) / 2) * (float64(res.N2) / float64(res.Ln2))
	resFull.Level = 1 / resFull.Difficulty

	// effort = volume x difficulty
	resFull.Effort = resFull.Volume * resFull.Difficulty

	resFullJson, _ := json.Marshal(resFull)
	fmt.Println(string(resFullJson))
}

func main() {
	klog.InitFlags(nil)
	var _ = flag.Set("logtostderr", "true")
	var _ = flag.Set("stderrthreshold", "WARNING")
	var _ = flag.Set("v", "4")
	flag.Parse()

	var targetList = []string{}
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			klog.V(4).Infof("%+v %+v \n", path, info.Name())
			var ignore = false
			for _, v := range pathPrefixIgnore {
				if strings.HasPrefix(path, v) {
					ignore = true
					break
				}
			}
			if ignore {
				return nil
			}
			var ignore2 = true
			for _, v := range cCppExtList {
				if strings.HasSuffix(path, v) {
					ignore2 = false
					break
				}
			}

			if !ignore2 {
				targetList = append(targetList, path)
			}
			return nil
		})
	check(err)
	klog.V(2).Infoln(targetList)

	for _, target := range targetList {
		run_metric(target)
	}

	//jobs := make(chan string, 100)
	//results := make(chan HalsteadResult, 100)
	//
	//for w := 1; w <= WORKER_NUM; w++ {
	//	go worker(w, jobs, results)
	//}
	//for j := 1; j <= BATCH_SIZE; j++ {
	//	jobs <- ops[j]
	//}
	//close(jobs)
	//
	//for a := 1; a <= BATCH_SIZE; a++ {
	//	fmt.Println(<-results)
	//}
}
