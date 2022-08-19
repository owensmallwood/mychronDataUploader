package parsers

import (
	"encoding/csv"
	"log"
	"os"
	"strconv"
)

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func strToFloat(str string) float64 {
	res, _ := strconv.ParseFloat(str, 64)

	return res
}

func ReadCsv(filename string) [][]string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln("Failed to open file", err)
	}
	defer f.Close()

	// Read the whole csv
	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	return data
}
