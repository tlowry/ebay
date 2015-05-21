// TrainBayesian project main.go
package main

import (
	"bufio"
	"github.com/jbrukh/bayesian"
	"log"
	"os"
	"strings"
)

const (
	GFX      bayesian.Class = "gfx"
	CPU      bayesian.Class = "cpu"
	APU      bayesian.Class = "apu"
	System   bayesian.Class = "system"
	Unwanted bayesian.Class = "unwanted"
)

func LearnFile(classifier *bayesian.Classifier, name string, class bayesian.Class) {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic("could not open file")
	}
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if line == nil || err != nil {
			break
		}
		words := strings.Split(string(line), " ")
		classifier.Learn(words, class)
	}
}

func testClassifier(c *bayesian.Classifier, doc []string) bayesian.Class {

	_, inx, _ := c.ProbScores(doc)
	class := c.Classes[inx]

	return class
}

func TestFile(classifier *bayesian.Classifier, name string, class bayesian.Class) int {
	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		panic("could not open file")
	} else {
		log.Println("Parsing file ", name, "as", class)
	}
	var score, tested int
	reader := bufio.NewReader(file)

	for {
		line, _, err := reader.ReadLine()
		if line == nil || err != nil {
			break
		}
		words := strings.Split(string(line), " ")
		cls := testClassifier(classifier, words)

		if cls == class {
			score++
		} else {
			// wrong, try again
			log.Println("Classifier wrong ", words, "!=", cls, " expected ", class)
			classifier.Learn(words, class)
		}

		tested++
	}
	pc := 100
	if tested > 1 {
		pc = (score * 100) / tested
	}

	return pc
}

func main() {

	c := bayesian.NewClassifier(CPU, GFX, APU, System, Unwanted)

	var err error = nil

	gfxFile := "files/gfx.txt"
	LearnFile(c, gfxFile, GFX)

	cpuFile := "files/cpu.txt"
	LearnFile(c, cpuFile, CPU)

	apuFile := "files/apu.txt"
	LearnFile(c, apuFile, APU)

	systemFile := "files/system.txt"
	LearnFile(c, systemFile, System)

	unWantedFile := "files/unwanted.txt"
	LearnFile(c, unWantedFile, Unwanted)

	log.Printf("classifier is trained: %d documents read\n", c.WordCount())

	accuracy := 0
	for accuracy < 100 {
		accuracy = TestFile(c, gfxFile, GFX)
	}

	accuracy = 0
	for accuracy < 100 {
		accuracy = TestFile(c, cpuFile, CPU)
	}

	accuracy = 0
	for accuracy < 100 {
		accuracy = TestFile(c, apuFile, APU)
	}

	accuracy = 0
	for accuracy < 100 {
		accuracy = TestFile(c, systemFile, System)
	}

	accuracy = 0
	for accuracy < 100 {
		accuracy = TestFile(c, unWantedFile, Unwanted)
	}

	log.Printf("Successfuly trained classifier")

	err = c.WriteToFile("files/item.ebay.classifier")
	if err != nil {
		panic(err)
	}

}
