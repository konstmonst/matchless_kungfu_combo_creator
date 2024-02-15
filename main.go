package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"slices"
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

type YamlInner struct {
	ID       string `yaml:"id"`
	V        string `yaml:"v"`
	Comment  string `yaml:"comment"`
	ChiType  string `yaml:"chiType"`
	ChiValue int    `yaml:"chiValue"`
}

type Inner struct {
	ID        string
	Comment   string
	ChiType   string
	ChiValue  int
	Bytes     []byte // contets of V only as []byte to speed up processing
	Contained *Inner
}

type MergedInners struct {
	InnerIndices []int  // indices of inners in the global array of inners
	MergePos     []int  // position where a given inner was inserted in the CachedValue
	CachedValue  []byte // the resulting merged string of inners
}

func (m *MergedInners) LastIndex() uint {
	return uint(m.InnerIndices[len(m.InnerIndices)-1])
}

func (m *MergedInners) Merge(inners []*Inner, index int) {
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
		lenA := len(inners[m.LastIndex()].Bytes)
		lenB := len(inners[index].Bytes)
		mergePos := len(m.CachedValue) - lenA + pos
		m.MergePos = append(m.MergePos, mergePos)

		// when we have abc and bcd, we only have to append d
		// when we have abcd and bc, we don't have to append anything
		// if mergePos+lenB > len(m.CachedValue) {
		if lenA-pos < lenB {
			m.CachedValue = append(m.CachedValue, inners[index].Bytes[lenA-pos:]...)
		}

	} else {
		// preallocating slices to avoid reallocations
		m.MergePos = make([]int, 0, len(inners))
		m.MergePos = append(m.MergePos, 0)
		m.InnerIndices = make([]int, 0, len(inners))
		m.CachedValue = make([]byte, 0, 255) // approximate size
		m.CachedValue = append(m.CachedValue, inners[index].Bytes...)
	}
	m.InnerIndices = append(m.InnerIndices, index)
}

func (m *MergedInners) IsBetterThan(other *MergedInners) bool {
	return len(m.CachedValue) < len(other.CachedValue)
}

// converts merged inner to human possible respresentation, separating
// each wordSize letters with space
func (m *MergedInners) String(inners []*Inner, wordSize int) string {
	chiValues := map[string]int{}
	var sb strings.Builder
	sb.WriteString("Inners: ")
	for i, index := range m.InnerIndices {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(inners[index].ID)
		sb.WriteString(fmt.Sprintf(" at %d", m.MergePos[i]))

		if len(inners[index].ChiType) > 0 {
			oldVal := chiValues[inners[index].ChiType]
			chiValues[inners[index].ChiType] = oldVal + inners[index].ChiValue
		}
	}

	sb.WriteByte('\n')
	for i, byte := range m.CachedValue {
		if i > 0 && i%wordSize == 0 {
			sb.WriteByte(' ')
		}
		sb.WriteByte(byte)
	}

	sb.WriteString("\nChi Values: ")
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

	sb.WriteString("\nomitted inners (because they are part of other inners):\n")
	first := true
	for _, index := range m.InnerIndices {
		parent := inners[index]
		child := parent.Contained
		for child != nil {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s is part of %s", child.ID, parent.ID))
			parent = child
			child = child.Contained
			first = false
		}
	}
	return sb.String()
}

// Calculate at witch position in slice b slice a can be inserted
// If slice a doesn't contain the begining of b, the merge position is
// at the end of slice a
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
		// part of b is at the end of a
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

func mergeInners(inners []*Inner, maxResultSize int) MergedInners {
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

// Puts inners, that are part of other inners in parent's Contained field
// If there are multiple, they create a singly linked list with bigger
// matches first
func preprocess(inners []*Inner) []*Inner {
	slices.SortFunc(inners, func(i, j *Inner) int {
		x := len(i.Bytes) - len(j.Bytes)
		if x < 0 {
			return 1
		} else if x > 0 {
			return -1
		}
		return 0
	})
	newLen := len(inners)
	for i := 0; i < len(inners); i++ {
		if inners[i] != nil {
			for j := i + 1; j < len(inners); j++ {
				if inners[j] != nil && bytes.Contains(inners[i].Bytes, inners[j].Bytes) {
					insertAt := inners[i]
					for insertAt.Contained != nil && bytes.Contains(insertAt.Bytes, inners[j].Bytes) {
						insertAt = insertAt.Contained
					}
					inners[j].Contained = insertAt.Contained
					insertAt.Contained = inners[j]
					inners[j] = nil
					newLen--
				}
			}
		}
	}
	result := make([]*Inner, 0, newLen)
	for _, inner := range inners {
		if inner != nil {
			result = append(result, inner)
		}
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
		MaxResultSize int         `yaml:"maxResultSize"`
		KnownInners   []YamlInner `yaml:"knownInners"`
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

	inners := make([]*Inner, 0, len(input.KnownInners))
	for i := 0; i < len(input.KnownInners); i++ {
		if len(input.KnownInners[i].V) > (1<<16)-1 {
			log.Fatalf("inner to long for calculation")
		}
		inners = append(inners, &Inner{
			ID:       input.KnownInners[i].ID,
			Comment:  input.KnownInners[i].Comment,
			Bytes:    []byte(input.KnownInners[i].V),
			ChiType:  input.KnownInners[i].ChiType,
			ChiValue: input.KnownInners[i].ChiValue,
		})
	}

	inners = preprocess(inners)

	log.Printf("input(%d): %v", len(input.KnownInners), input)
	startAt := time.Now()
	result := mergeInners(inners, int(input.MaxResultSize))
	fmt.Printf("calculation took %v\n", time.Since(startAt))
	fmt.Printf("result of size %d:\n%v\n", len(result.CachedValue), result.String(inners, *wordSize))
}
