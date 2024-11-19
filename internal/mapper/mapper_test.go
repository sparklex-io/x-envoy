package mapper

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func Test_calVotingPower(t *testing.T) {
	type args struct {
		n int
		p float64
		q float64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"0", args{n: 1000000, p: Tau / K, q: 0.017}, 34},
		{"1", args{n: 1000000, p: Tau / K, q: 0.024}, 35},
		{"2", args{n: 1000000, p: Tau / K, q: 0.034}, 36},
		{"3", args{n: 1000000, p: Tau / K, q: 0.048}, 37},
		{"4", args{n: 1000000, p: Tau / K, q: 0.065}, 38},
		{"5", args{n: 1000000, p: Tau / K, q: 0.087}, 39},
		{"6", args{n: 1000000, p: Tau / K, q: 0.12}, 41},
		{"7", args{n: 1000000, p: Tau / K, q: 0.15}, 42},
		{"8", args{n: 1000000, p: Tau / K, q: 0.18}, 43},
		{"9", args{n: 1000000, p: Tau / K, q: 0.23}, 44},
		{"10", args{n: 1000000, p: Tau / K, q: 0.27}, 45},
		{"11", args{n: 1000000, p: Tau / K, q: 0.32}, 46},
		{"12", args{n: 1000000, p: Tau / K, q: 0.37}, 47},
		{"13", args{n: 1000000, p: Tau / K, q: 0.43}, 48},
		{"14", args{n: 1000000, p: Tau / K, q: 0.99}, 65},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VotingPower(tt.args.n, tt.args.p, tt.args.q); got != tt.want {
				t.Errorf("calVotingPower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bApproximatedCDF(t *testing.T) {
	type args struct {
		N int     `json:"n"`
		P float64 `json:"p"`
		K int     `json:"k"`
	}

	file, err := os.ReadFile("../../test/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []args

	err = json.Unmarshal(file, &cases)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Run("case", func(t *testing.T) {
			mean := float64(c.N) * c.P
			sd := math.Sqrt(mean * (1 - c.P))
			got := BApproximatedCDF(c.K, mean, sd)
			t.Log("got", got)
		})
	}
}
