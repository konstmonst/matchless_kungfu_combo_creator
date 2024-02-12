package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Tech struct {
	ID       string `yaml:"id"`
	V        string `yaml:"v"`
	Comment  string `yaml:"comment"`
	ChiType  string `yaml:"chiType"`
	ChiValue int    `yaml:"chiValue"`
}

type Input struct {
	MaxResultSize int    `yaml:"maxResultSize"`
	KnownTechs    []Tech `yaml:"knownTechs"`
}

type MergedTechs struct {
	Techs []*Tech
	Value string
}

func (m *MergedTechs) Merge(tech *Tech) {
	m.Techs = append(m.Techs, tech)
	m.Value = mergeStrings(m.Value, tech.V)
}

func (m *MergedTechs) IsBetterThan(other *MergedTechs) bool {
	return len(m.Value) < len(other.Value)
}

func (m *MergedTechs) String() string {
	chiValues := map[string]int{}
	var sb strings.Builder
	for i, tech := range m.Techs {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(tech.ID)
		if len(tech.ChiType) > 0 {
			oldVal := chiValues[tech.ChiType]
			chiValues[tech.ChiType] = oldVal + tech.ChiValue
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
	for i, rune := range m.Value {
		sb.WriteRune(rune)
		if i > 0 && i%5 == 0 {
			sb.WriteRune(' ')
		}
	}

	return sb.String()
}

// https://stackoverflow.com/questions/30226438/generate-all-permutations-in-go
// The slice p keeps the intermediate state as offsets in a Fisher-Yates shuffle algorithm.
// This has the nice property that the zero value for p describes the identity permutation.
func nextPerm(p []int8) {
	for i := int8(len(p)) - 1; i >= 0; i-- {
		if i == 0 || p[i] < int8(len(p))-i-1 {
			p[i]++
			return
		}
		p[i] = 0
	}
}

func getPerm(orig, p []int8) []int8 {
	result := append([]int8{}, orig...)
	for z, v := range p {
		i := int8(z)
		result[i], result[i+v] = result[i+v], result[i]
	}
	return result
}

/*
func main() {
    orig := []int{11, 22, 33}
    for p := make([]int, len(orig)); p[0] < len(p); nextPerm(p) {
        fmt.Println(getPerm(orig, p))
    }
}*/

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mergeStrings(a, b string) string {
	for i := 0; i < len(a); i++ {
		j := 0
		for j < len(b) && i+j < len(a) && a[i+j] == b[j] {
			j++
		}
		// b i in a
		if j == len(b) {
			return a
		}
		if i+j == len(a) {
			return a + b[j:]
		}
	}
	return a + b
}

func mergeTechs(a, b Tech) Tech {
	return Tech{
		ID: "A",
		V:  mergeStrings(a.V, b.V),
	}
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

func findSmallestCommonString(techs []Tech, maxResultSize int) MergedTechs {
	initial := MergedTechs{Value: strings.Repeat("A", maxResultSize+1)}
	result := initial

	// generate original permutation
	orig := make([]int8, len(techs))
	for i := int8(0); i < int8(len(orig)); i++ {
		orig[i] = i
	}

	numIterations := factorial(uint64(len(orig)))
	iter := uint64(0)
	lastProgres := uint64(0)

	for p := make([]int8, len(orig)); p[0] < int8(len(p)); nextPerm(p) {
		v := getPerm(orig, p)
		var tech MergedTechs
		for i := 0; i < len(v); i++ {
			tech.Merge(&techs[v[i]])
		}
		if tech.IsBetterThan(&result) {
			fmt.Printf("new result: %v\n", tech.String())
			result = tech
		}

		progress := iter * uint64(100) / numIterations
		if progress != lastProgres {
			log.Printf("%d%% done\n", progress)
			lastProgres = progress
		}
		iter++
	}
	if len(result.Techs) == 0 {
		result.Value = ""
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

	var input Input
	err = yaml.Unmarshal(yamlBytes, &input)
	if err != nil {
		log.Fatalf("failed to unmarshal: %s: %v", filename, err)
	}

	if len(input.KnownTechs) == 0 {
		log.Fatalf("no known techs provided in %s file", filename)
	}

	if len(input.KnownTechs) > 21 {
		log.Fatalf("only max 21 techs at once are supported by now")
	}

	if input.MaxResultSize < 1 || input.MaxResultSize > 127 {
		log.Fatalf("only max string of 127 chars can be computed by now")
	}

	log.Printf("input(%d): %v", len(input.KnownTechs), input)

	result := findSmallestCommonString(input.KnownTechs, int(input.MaxResultSize))
	fmt.Printf("result of size %d: %v\n", len(result.Value), result.String())
}
