package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/gonum/stat/combin"
	"gopkg.in/yaml.v3"
)

var (
	mergeCache map[uint]int = map[uint]int{}
)

type Inner struct {
	ID       string `yaml:"id"`
	V        string `yaml:"v"`
	Comment  string `yaml:"comment"`
	ChiType  string `yaml:"chiType"`
	ChiValue int    `yaml:"chiValue"`
}

type MergedInners struct {
	InnerIndices []int
	MergePos     []int
	CachedValue  string
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
		key := (m.LastIndex() << 16) | uint(index)
		pos, found = mergeCache[key]
		if !found {
			pos = calcMergePos(inners[m.LastIndex()].V, inners[index].V)
			mergeCache[key] = pos
		}
		lenA := int(len(inners[m.LastIndex()].V))
		lenB := int(len(inners[index].V))
		if pos+lenB-1 > lenA {
			m.CachedValue += inners[index].V[lenA-pos:]
		}
		m.MergePos = append(m.MergePos, pos)
	} else {
		m.CachedValue = inners[index].V
		m.MergePos = []int{pos}
	}
	m.InnerIndices = append(m.InnerIndices, index)
}

func (m *MergedInners) IsBetterThan(other *MergedInners) bool {
	return len(other.CachedValue) == 0 || len(m.CachedValue) < len(other.CachedValue)
}

func (m *MergedInners) String(inners []Inner) string {
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
	for i, rune := range m.CachedValue {
		sb.WriteRune(rune)
		if i > 0 && i%5 == 0 {
			sb.WriteRune(' ')
		}
	}

	return sb.String()
}

func calcMergePos(a, b string) int {
	lenA := int(len(a))
	lenB := int(len(b))

	for mergeAt := int(0); mergeAt < lenA; mergeAt++ {
		j := int(0)
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
	var initial MergedInners
	result := initial
	numInners := len(inners)

	numIterations := factorial(uint64(numInners))
	iter := uint64(0)
	lastProgres := uint64(0)

	permGen := combin.NewPermutationGenerator(numInners, numInners)
	p := make([]int, numInners)
	for permGen.Next() {
		permGen.Permutation(p)
		var inner MergedInners
		for i := 0; i < numInners; i++ {
			inner.Merge(inners, p[i])
		}
		if inner.IsBetterThan(&result) {
			fmt.Printf("new result: %v\n", inner.String(inners))
			result = inner
		}

		progress := iter * uint64(100) / numIterations
		if progress != lastProgres {
			log.Printf("%d%% done\n", progress)
			lastProgres = progress
		}
		iter++
	}
	if len(result.InnerIndices) == 0 {
		result.CachedValue = ""
	}
	return result
}

func main() {
	filename := "combinations.yaml"
	yamlFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("failed to open %s: %v", filename, err)
	}
	defer yamlFile.Close()

	yamlBytes, _ := io.ReadAll(yamlFile)

	var input struct {
		MaxResultSize int     `yaml:"maxResultSize"`
		KnownInners   []Inner `yaml:"knownInners"`
	}

	err = yaml.Unmarshal(yamlBytes, &input)
	if err != nil {
		log.Fatalf("failed to unmarshal: %s: %v", filename, err)
	}

	if len(input.KnownInners) == 0 {
		log.Fatalf("no known techs provided in %s file", filename)
	}

	if len(input.KnownInners) > 21 {
		log.Fatalf("only max 21 techs at once are supported by now")
	}

	if input.MaxResultSize < 1 || input.MaxResultSize > 127 {
		log.Fatalf("only max string of 127 chars can be computed by now")
	}

	for _, inner := range input.KnownInners {
		if len(inner.V) > (1<<16)-1 {
			log.Fatalf("inner to long for calculation")
		}
	}

	log.Printf("input(%d): %v", len(input.KnownInners), input)
	startAt := time.Now()
	result := findSmallestCommonString(input.KnownInners, int(input.MaxResultSize))
	fmt.Printf("calculation took %v\n", time.Since(startAt))
	fmt.Printf("result of size %d: %v\n", len(result.CachedValue), result.String(input.KnownInners))
}
