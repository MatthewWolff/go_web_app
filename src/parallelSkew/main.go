package main

import (
	"fmt"
	"log"
	"path"
	"time"
	//"runtime"
)

type Bacterium struct {
	name      string
	genome    string
	skewArray []int
}

func main() {

	genomeNames := []string{
		"bacillus_anthracis",
		"deinococcus_deserti",
		"escherichia_coli",
		"legionella_pneumophila",
		"porphyromonas_gingivalis",
		"rickettsia_prowazekii",
		"staphylococcus_aureus",
		"streptococcus_pneumoniae",
		"thermotoga_petrophila",
	}

	bacteria := make([]Bacterium, len(genomeNames))

	// add genomes
	for i := range bacteria {
		bacteria[i].name = genomeNames[i]
		bacteria[i].genome, _ = readFasta(path.Join("data", bacteria[i].name+".txt"))
	}

	start := time.Now()

	// fill in serial code here
	for _, bact := range bacteria {
		bact.skewArray = SkewArraySerial(bact.genome)
	}

	elapsed := time.Since(start)
	log.Printf("running serially took %s", elapsed)

	// clear out the skewArrays
	for _, bact := range bacteria {
		bact.skewArray = []int{}
	}

	// fill in parallel code here
	start = time.Now()
	skewArraysParallel := make(map[string][]int)

	c := make(chan Bacterium, len(bacteria))

	for _, bact := range bacteria {
		go SkewArrayParallel(c, bact)
	}

	for i := 0; i < len(bacteria); i++ {
		bact := <-c
		bacteria[i] = bact
		skewArraysParallel[bact.name] = bact.skewArray
	}

	elapsed = time.Since(start)
	log.Printf("running in parallel took %s", elapsed)

	fmt.Println()
	fmt.Println("Generating plots:")
	for _, bact := range bacteria {
		//fmt.Println(bact.skewArray)
		bact.ConvertToPNG()
	}
}
