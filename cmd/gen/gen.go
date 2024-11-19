package main

import (
	"encoding/json"
	"envoy/internal/mapper"
	"math"
	"os"
	"strconv"
)

type Case struct {
	N   int     `json:"n"`
	P   float64 `json:"p"`
	K   int     `json:"k"`
	CDF float64 `json:"cdf"`
}

type FixedNumberCase struct {
	N   string `json:"n"`
	P   string `json:"p"`
	K   string `json:"k"`
	CDF string `json:"cdf"`
}

func main() {
	file, err := os.ReadFile("test/cases.json")
	if err != nil {
		panic(err)
	}
	var cases []Case

	err = json.Unmarshal(file, &cases)
	if err != nil {
		panic(err)
	}

	var nCases []FixedNumberCase
	for index := range cases {
		c := cases[index]
		mean := float64(c.N) * c.P
		sd := math.Sqrt(mean * (1 - c.P))
		cases[index].CDF = mapper.BApproximatedCDF(c.K, mean, sd)

		if cases[index].CDF == 0 || cases[index].CDF == 1 {
			continue
		}

		nCases = append(nCases, FixedNumberCase{
			N:   strconv.Itoa(c.N) + "000000000000000000",
			P:   strconv.Itoa(int(c.P * 1e18)),
			K:   strconv.Itoa(c.K) + "000000000000000000",
			CDF: strconv.Itoa(int(cases[index].CDF * 1e18)),
		})
	}

	data, err := json.Marshal(nCases)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("output.json", data, 0644)
	if err != nil {
		panic(err)
	}
}
