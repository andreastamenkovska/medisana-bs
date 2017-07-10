package plugins

import (
	"github.com/gocarina/gocsv"
	"github.com/jovandeginste/medisana-bs/structs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type Csv struct {
	Dir string
}

func (plugin Csv) Initialize() bool {
	log.Println("I am the CSV plugin")
	log.Printf("  - Dir: %s\n", plugin.Dir)
	return true
}
func (plugin Csv) ParseData(person *structs.PersonMetrics) bool {
	log.Println("The csv plugin is parsing new data")
	personId := person.Person
	weights := make(structs.BodyMetrics, len(person.BodyMetrics))
	idx := 0
	for _, value := range person.BodyMetrics {
		weights[idx] = value
		idx++
	}
	sort.Sort(weights)

	csvFile := plugin.Dir + "/" + strconv.Itoa(personId) + ".csv"
	log.Printf("Writing to file '%s'.\n", csvFile)
	CreateCsvDir(csvFile)

	f, err := os.Create(csvFile)
	if err != nil {
		log.Printf("%#v", err)
	}
	defer f.Close()

	err = gocsv.MarshalWithoutHeaders(&weights, f)

	if err != nil {
		log.Printf("%#v", err)
	}
	return true
}

func CreateCsvDir(file string) {
	path := filepath.Dir(file)
	mode := os.FileMode(0700)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, mode)
	}
}
