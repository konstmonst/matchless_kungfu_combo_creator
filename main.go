package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/gonum/stat/combin"
	"gopkg.in/yaml.v3"
)

var (
	mergeCache       map[uint]int
	enableMergeCache = flag.Bool("enableMergeCache", true, "enables merge cache")
	cpuprofile       = flag.String("cpuprofile", "", "write cpu profile to file")
	filename         = flag.String("filename", "combinations.yaml", "yaml file with combinations")
	wordSize         = flag.Int("wordSize", 5, "max characters in a word wwhen splitting result string")
)

type Inner struct {
	ID       string `yaml:"id"`
	V        string `yaml:"v"`
	Comment  string `yaml:"comment"`
	ChiType  string `yaml:"chiType"`
	ChiValue int    `yaml:"chiValue"`
	Bytes    []byte
}

type MergedInners struct {
	InnerIndices []int
	MergePos     []int
	CachedValue  []byte
}

func (m *MergedInners) LastIndex() uint {
	return uint(m.InnerIndices[len(m.InnerIndices)-1])
}

func (m *MergedInners) Merge(inners []Inner, index int) {
	var (
		pos   int
		found bool
	)

	if len(m.InnerIndices) > 0 {
		if mergeCache != nil {
			key := (m.LastIndex() << 16) | uint(index)
			pos, found = mergeCache[key]
			if !found {
				pos = calcMergePos(inners[m.LastIndex()].Bytes, inners[index].Bytes)
				mergeCache[key] = pos
			}
		} else {
			pos = calcMergePos(inners[m.LastIndex()].Bytes, inners[index].Bytes)
		}
		lenA := int(len(inners[m.LastIndex()].Bytes))
		lenB := int(len(inners[index].Bytes))
		if pos+lenB-1 > lenA {
			m.CachedValue = append(m.CachedValue, inners[index].Bytes[lenA-pos:]...)
		}
	} else {
		m.MergePos = make([]int, 0, len(inners))
		m.InnerIndices = make([]int, 0, len(inners))
		m.CachedValue = make([]byte, 0, 255)
		m.CachedValue = append(m.CachedValue, inners[index].Bytes...)
	}
	m.MergePos = append(m.MergePos, pos)
	m.InnerIndices = append(m.InnerIndices, index)
}

func (m *MergedInners) IsBetterThan(other *MergedInners) bool {
	return len(m.CachedValue) < len(other.CachedValue)
}

func (m *MergedInners) String(inners []Inner, wordSize int) string {
	chiValues := map[string]int{}
	var sb strings.Builder
	for i, index := range m.InnerIndices {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(inners[index].ID)
		if len(inners[index].ChiType) > 0 {
			oldVal := chiValues[inners[index].ChiType]
			chiValues[inners[index].ChiType] = oldVal + inners[index].ChiValue
		}
	}

	sb.WriteString(" Chi Values: ")
	j := 0
	for chiType, chiValue := range chiValues {
		if j > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteByte('+')
		sb.WriteString(strconv.FormatInt(int64(chiValue), 10))
		sb.WriteString(chiType)
		j++
	}

	sb.WriteByte('\n')
	for i, byte := range m.CachedValue {
		sb.WriteByte(byte)
		if i > 0 && i%wordSize == 0 {
			sb.WriteRune(' ')
		}
	}

	return sb.String()
}

func calcMergePos(a, b []byte) int {
	lenA := len(a)
	lenB := len(b)

	for mergeAt := 0; mergeAt < lenA; mergeAt++ {
		j := 0
		for j < lenB && mergeAt+j < lenA && a[mergeAt+j] == b[j] {
			j++
		}
		// b is completely in a
		if j == lenB {
			return mergeAt
		}
		// part of b at the end of a
		if mergeAt+j == lenA {
			return mergeAt
		}
	}
	return lenA
}

func factorial(number uint64) uint64 {
	if number <= 1 {
		return 1
	}

	fact := number

	for number > 1 {
		number--
		fact *= number
	}
	return fact
}

func findSmallestCommonString(inners []Inner, maxResultSize int) MergedInners {
	var result MergedInners
	numInners := len(inners)

	numIterations := factorial(uint64(numInners))
	lastProgres := uint64(0)

	permGen := combin.NewPermutationGenerator(numInners, numInners)
	p := make([]int, numInners)
	for iter := uint64(0); permGen.Next(); iter++ {
		permGen.Permutation(p)
		var inner MergedInners
		for i := 0; i < numInners; i++ {
			inner.Merge(inners, p[i])
		}
		if len(inner.CachedValue) <= maxResultSize && (result.CachedValue == nil || inner.IsBetterThan(&result)) {
			fmt.Printf("new result: %v\n", inner.String(inners, *wordSize))
			result = inner
		}

		progress := iter * uint64(100) / numIterations
		if progress != lastProgres {
			log.Printf("%d%% done\n", progress)
			lastProgres = progress
		}
	}
	if len(result.InnerIndices) == 0 {
		result.CachedValue = nil
	}
	return result
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *enableMergeCache {
		mergeCache = make(map[uint]int)
	}
	yamlFile, err := os.Open(*filename)
	if err != nil {
		log.Fatalf("failed to open %s: %v", *filename, err)
	}
	defer yamlFile.Close()

	yamlBytes, _ := io.ReadAll(yamlFile)

	var input struct {
		MaxResultSize int     `yaml:"maxResultSize"`
		KnownInners   []Inner `yaml:"knownInners"`
	}

	err = yaml.Unmarshal(yamlBytes, &input)
	if err != nil {
		log.Fatalf("failed to unmarshal: %s: %v", *filename, err)
	}

	if len(input.KnownInners) == 0 {
		log.Fatalf("no known techs provided in %s file", *filename)
	}

	if len(input.KnownInners) > 21 {
		log.Fatalf("only max 21 techs at once are supported by now")
	}

	if input.MaxResultSize < 1 || input.MaxResultSize > 127 {
		log.Fatalf("only max string of 127 chars can be computed by now")
	}

	for i := 0; i < len(input.KnownInners); i++ {
		if len(input.KnownInners[i].V) > (1<<16)-1 {
			log.Fatalf("inner to long for calculation")
		}
		input.KnownInners[i].Bytes = []byte(input.KnownInners[i].V)
	}

	log.Printf("input(%d): %v", len(input.KnownInners), input)
	startAt := time.Now()
	result := findSmallestCommonString(input.KnownInners, int(input.MaxResultSize))
	fmt.Printf("calculation took %v\n", time.Since(startAt))
	fmt.Printf("result of size %d: %v\n", len(result.CachedValue), result.String(input.KnownInners, *wordSize))
}
