package mapper

import (
	"crypto/ecdsa"
	"errors"
	"github.com/furoxr/go-ecvrf"
	"github.com/rs/zerolog/log"
	"github.com/sparklex-io/envoy/generated/reducer"
	"gonum.org/v1/gonum/stat/distuv"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
)

const (
	Tau float64 = 5_000
	K   float64 = 100_000_000
)

type Mapper struct{}

type Vote struct {
	PublicKey   [2]*big.Int
	VRFHash     []byte
	Proof       []byte
	UPoint      [2]*big.Int
	VComponents [4]*big.Int
	VotingPower *big.Int
	Stake       *big.Int
}

func (m *Mapper) Prove(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, []byte, error) {
	beta, pi, err := ecvrf.Secp256k1Sha256Tai.Prove(privateKey, message)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate VRF")
		return nil, nil, err
	}
	return beta, pi, nil
}

func (m *Mapper) RandomNumber(vrfOutput []byte) (float64, error) {
	t := &big.Int{}
	t.SetBytes(vrfOutput[:])

	precision := uint(8 * (len(vrfOutput) + 1))
	maxFloat, b, err := big.ParseFloat("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0, precision, big.ToNearestEven)
	if b != 16 || err != nil {
		log.Error().Err(err).Msg("failed to generate random number")
		return 0, err
	}

	h := big.Float{}
	h.SetPrec(precision)
	h.SetInt(t)

	ratio := big.Float{}
	rNumber, _ := ratio.Quo(&h, maxFloat).Float64()
	return rNumber, nil
}

func (m *Mapper) QueryStake() (int64, error) {
	envName := "TEMP_STAKE_NUMBER"
	value := os.Getenv(envName)
	if value == "" {
		return 1000000, nil
	}
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (m *Mapper) BuildVotePayload(privateKey *ecdsa.PrivateKey, p float64, message []byte, stake int64) (reducer.VoteFastPayload, error) {
	vrfHash, pi, err := m.Prove(privateKey, message)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate VRF")
		return reducer.VoteFastPayload{}, err
	}
	u, sH, cGamma, err := ecvrf.Secp256k1Sha256Tai.ComputeFastVerifyParams(
		&privateKey.PublicKey,
		message,
		pi)
	if err != nil {
		log.Error().Err(err).Msg("failed to build parameters for fast verify")
		return reducer.VoteFastPayload{}, err
	}

	randomNum, err := m.RandomNumber(vrfHash)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate random number")
		return reducer.VoteFastPayload{}, err
	}
	power := VotingPower(int(stake), p, randomNum)

	gamma, C, S, err := ecvrf.Secp256k1Sha256Tai.DecodeProof(pi)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode VRF proof")
		return reducer.VoteFastPayload{}, err
	}

	stakeOnChain, err := toFixedNumber(stake)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert to fixed number")
		return reducer.VoteFastPayload{}, err
	}
	powerOnChain, err := toFixedNumber(int64(power))
	if err != nil {
		log.Error().Err(err).Msg("failed to convert to fixed number")
		return reducer.VoteFastPayload{}, err
	}
	vote := reducer.VoteFastPayload{
		PublicKey:   [2]*big.Int{privateKey.PublicKey.X, privateKey.PublicKey.Y},
		Proof:       [4]*big.Int{gamma.X, gamma.Y, C, S},
		UPoint:      [2]*big.Int{u.X, u.Y},
		VComponents: [4]*big.Int{sH.X, sH.Y, cGamma.X, cGamma.Y},
		VotingPower: powerOnChain,
		Stake:       stakeOnChain,
		Message:     reducer.MessagePayload{},
	}
	return vote, nil
}

func VotingPower(n int, p, q float64) int {
	mean := float64(n) * p
	sd := math.Sqrt(mean * (1 - p))

	low, high := 0, n
	for low < high {
		mid := (low + high) / 2
		cdf := BApproximatedCDF(mid, mean, sd)
		if math.Abs(cdf-q) <= 0.0000001 {
			return mid
		} else if cdf < q {
			low = mid + 1
		} else {
			high = mid
		}
	}

	if low == 0 {
		return 0
	} else {
		return low - 1
	}
}

func BApproximatedCDF(x int, mean float64, sd float64) float64 {
	standardX := (float64(x) + 0.5 - mean) / sd
	return distuv.UnitNormal.CDF(standardX)
}

func toFixedNumber(value int64) (*big.Int, error) {
	r := new(big.Int)
	strValue := strconv.FormatInt(value, 10)
	paddingStr := strValue + strings.Repeat("0", 18)
	r, ok := r.SetString(paddingStr, 10)
	if !ok {
		return nil, errors.New("failed to convert to fixed number")
	}
	return r, nil
}
