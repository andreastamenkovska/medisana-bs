package main

import (
	"github.com/jovandeginste/medisana-bs/plugins"
	"github.com/jovandeginste/medisana-bs/structs"
	"log"
	"math"
	"time"
)

var allPersons = make([]*structs.PersonMetrics, 8)

// MetricParser will initialize the Persons from csv and parse incoming metrics
func MetricParser() {
	for i := range allPersons {
		allPersons[i] = &structs.PersonMetrics{Person: i + 1, BodyMetrics: make(map[int]structs.BodyMetric)}
		allPersons[i].ImportBodyMetrics(structs.ImportCsv(i + 1))
	}
	syncChan := make(chan bool)
	go debounce(3*time.Second, syncChan)
	for {
		partialMetric := <-metricChan
		updatePerson(partialMetric.Person)
		updateBody(partialMetric.Body)
		updateWeight(partialMetric.Weight)
		syncChan <- true
	}
}

func getPersonMetrics(personID int) *structs.PersonMetrics {
	return allPersons[personID-1]
}

func updatePerson(update structs.Person) {
	if !update.Valid {
		return
	}
	log.Printf("[METRIC PARSER] Received person metrics: %+v", update)
	person := getPersonMetrics(update.Person)
	person.Gender = update.Gender
	person.Age = update.Age
	person.Size = update.Size
	person.Activity = update.Activity
	printPerson(person)
}

func updateBody(update structs.Body) {
	if !update.Valid {
		return
	}
	log.Printf("[METRIC PARSER] Received body metrics: %+v", update)
	person := getPersonMetrics(update.Person)
	person.Updated = true
	if _, ok := person.BodyMetrics[update.Timestamp]; !ok {
		log.Printf("[METRIC PARSER] No body metric - creating")
		person.BodyMetrics[update.Timestamp] = structs.BodyMetric{}
	}
	bodyMetric := person.BodyMetrics[update.Timestamp]
	bodyMetric.Timestamp = update.Timestamp
	bodyMetric.Kcal = update.Kcal
	bodyMetric.Fat = update.Fat
	bodyMetric.Tbw = update.Tbw
	bodyMetric.Muscle = update.Muscle
	bodyMetric.Bone = update.Bone
	person.BodyMetrics[update.Timestamp] = bodyMetric
	printPerson(person)
}
func updateWeight(update structs.Weight) {
	if !update.Valid {
		return
	}
	log.Printf("[METRIC PARSER] Received weight metrics: %+v", update)
	person := getPersonMetrics(update.Person)
	person.Updated = true
	if _, ok := person.BodyMetrics[update.Timestamp]; !ok {
		log.Printf("[METRIC PARSER] No body metric - creating")
		person.BodyMetrics[update.Timestamp] = structs.BodyMetric{}
	}
	bodyMetric := person.BodyMetrics[update.Timestamp]
	bodyMetric.Weight = update.Weight
	bodyMetric.Timestamp = update.Timestamp
	if bodyMetric.Weight > 0 && person.Size > 0 {
		bodyMetric.Bmi = bodyMetric.Weight / float32(math.Pow(float64(person.Size)/100, 2))
	}

	person.BodyMetrics[update.Timestamp] = bodyMetric
	printPerson(person)
}
func printPerson(person *structs.PersonMetrics) {
	log.Printf("[METRIC PARSER] Person %d now has %d metrics.\n", person.Person, len(person.BodyMetrics))
}

func debounce(lull time.Duration, in chan bool) {
	for {
		select {
		case <-in:
		case <-time.Tick(lull):
			for _, person := range allPersons {
				if person.Updated {
					log.Printf("[METRIC PARSER] Person %d was updated -- calling all plugins.\n", person.Person)
					plugins.ParseData(person)
					person.Updated = false
				}
			}
		}
	}
}
