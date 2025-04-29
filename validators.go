package main

import (
	"bytes"
	"context"
	"math/big"
	"net/http"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types/query"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ValidatorData struct {
	Moniker           string
	OperatorAddress   string
	ConsensusPubkey   string
	Jailed            bool
	Status            stakingtypes.BondStatus
	Tokens            math.Int
	DelegatorShares   math.LegacyDec
	Description       stakingtypes.Description
	UnbondingHeight   int64
	UnbondingTime     time.Time
	Commission        stakingtypes.Commission
	MinSelfDelegation math.Int
}

type ValidatorSigningData struct {
	Address             string
	StartHeight         int64
	IndexOffset         int64
	JailedUntil         time.Time
	Tombstoned          bool
	MissedBlocksCounter int64
}

type CosmosClient struct {
	grpcConn *grpc.ClientConn
}

func ValidatorsHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn, validationClient interface{}) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	validatorsCommissionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_commission",
			Help:        "Commission of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorsStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_status",
			Help:        "Status of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorsJailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_jailed",
			Help:        "Jailed status of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorsTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_tokens",
			Help:        "Tokens of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorsDelegatorSharesGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_delegator_shares",
			Help:        "Delegator shares of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorsMinSelfDelegationGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_min_self_delegation",
			Help:        "Self-declared minimum self-delegation shares of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorsMissedBlocksGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_missed_blocks",
			Help:        "Missed blocks of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorsRankGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_rank",
			Help:        "Rank of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorsIsActiveGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validators_active",
			Help:        "1 if the Cosmos-based blockchain validator is in active set, 0 if not",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(validatorsCommissionGauge)
	registry.MustRegister(validatorsStatusGauge)
	registry.MustRegister(validatorsJailedGauge)
	registry.MustRegister(validatorsTokensGauge)
	registry.MustRegister(validatorsDelegatorSharesGauge)
	registry.MustRegister(validatorsMinSelfDelegationGauge)
	registry.MustRegister(validatorsMissedBlocksGauge)
	registry.MustRegister(validatorsRankGauge)
	registry.MustRegister(validatorsIsActiveGauge)

	var validators []stakingtypes.Validator
	var signingInfos []slashingtypes.ValidatorSigningInfo
	var validatorSetLength uint32

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying validators")
		queryStart := time.Now()

		var validatorsResponse interface{}
		var err error

		if NetworkType == "zenrock" {
			client := validationClient.(ValidationClient)
			validatorsResponse, err = client.Validators(
				context.Background(),
				&QueryValidatorsRequest{},
			)
		} else {
			client := validationClient.(stakingtypes.QueryClient)
			validatorsResponse, err = client.Validators(
				context.Background(),
				&stakingtypes.QueryValidatorsRequest{
					Pagination: &query.PageRequest{
						Limit: Limit,
					},
				},
			)
		}

		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get validators")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validators")

		if NetworkType == "zenrock" {
			res := validatorsResponse.(*QueryValidatorsResponse)
			validators = make([]stakingtypes.Validator, len(res.Validators))
			for i, v := range res.Validators {
				anyPubKey := &types.Any{
					TypeUrl: "/cosmos.crypto.ed25519.PubKey",
					Value:   []byte(v.ConsensusPubkey),
				}

				tokens, ok := math.NewIntFromString(v.Tokens)
				if !ok {
					sublogger.Error().Msg("Failed to parse tokens")
					continue
				}

				delegatorShares, err := math.LegacyNewDecFromStr(v.DelegatorShares)
				if err != nil {
					sublogger.Error().Err(err).Msg("Failed to parse delegator shares")
					continue
				}

				validators[i] = stakingtypes.Validator{
					OperatorAddress: v.OperatorAddress,
					ConsensusPubkey: anyPubKey,
					Jailed:          v.Jailed,
					Status:          stakingtypes.BondStatus(stakingtypes.BondStatus_value[v.Status]),
					Tokens:          tokens,
					DelegatorShares: delegatorShares,
					Description: stakingtypes.Description{
						Moniker: v.Description.Moniker,
					},
				}
			}
		} else {
			res := validatorsResponse.(*stakingtypes.QueryValidatorsResponse)
			validators = res.Validators
		}

		sort.Slice(validators, func(i, j int) bool {
			return validators[i].DelegatorShares.GT(validators[j].DelegatorShares)
		})
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying validators signing infos")
		queryStart := time.Now()

		slashingClient := slashingtypes.NewQueryClient(grpcConn)
		signingInfosResponse, err := slashingClient.SigningInfos(
			context.Background(),
			&slashingtypes.QuerySigningInfosRequest{
				Pagination: &query.PageRequest{
					Limit: Limit,
				},
			},
		)
		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get validators signing infos")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator signing infos")
		signingInfos = signingInfosResponse.Info
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying staking params")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		paramsResponse, err := stakingClient.Params(
			context.Background(),
			&stakingtypes.QueryParamsRequest{},
		)
		if err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not get staking params")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying staking params")
		validatorSetLength = paramsResponse.Params.MaxValidators
	}()

	wg.Wait()

	sublogger.Debug().
		Int("signingLength", len(signingInfos)).
		Int("validatorsLength", len(validators)).
		Msg("Validators info")

	for index, validator := range validators {
		moniker := validator.Description.Moniker
		moniker = sanitizeUTF8(moniker)

		// Исправление для validator.Tokens
		value, _ := new(big.Float).SetInt(validator.Tokens.BigInt()).Float64()
		validatorsTokensGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)

		validatorsStatusGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": moniker,
		}).Set(float64(validator.Status))

		var jailed float64
		if validator.Jailed {
			jailed = 1
		} else {
			jailed = 0
		}
		validatorsJailedGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": moniker,
		}).Set(jailed)

		// Исправление для validator.DelegatorShares
		value, _ = new(big.Float).SetInt(validator.DelegatorShares.BigInt()).Float64()
		validatorsDelegatorSharesGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)

		// Исправление для validator.MinSelfDelegation
		value, _ = new(big.Float).SetInt(validator.MinSelfDelegation.BigInt()).Float64()
		validatorsMinSelfDelegationGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)

		consAddr, err := validator.GetConsAddr()
		if err != nil {
			sublogger.Error().
				Str("address", validator.OperatorAddress).
				Err(err).
				Msg("Could not get validator consensus address")
			continue
		}

		var signingInfo slashingtypes.ValidatorSigningInfo
		found := false
		for _, signingInfoIterated := range signingInfos {
			if bytes.Equal(consAddr, []byte(signingInfoIterated.Address)) {
				found = true
				signingInfo = signingInfoIterated
				break
			}
		}

		if !found {
			sublogger.Debug().
				Str("address", validator.OperatorAddress).
				Msg("Could not get signing info for validator")
		} else if validator.Status == stakingtypes.Bonded {
			validatorsMissedBlocksGauge.With(prometheus.Labels{
				"address": validator.OperatorAddress,
				"moniker": moniker,
			}).Set(float64(signingInfo.MissedBlocksCounter))
		} else {
			sublogger.Trace().
				Str("address", validator.OperatorAddress).
				Msg("Validator is not active, not returning missed blocks amount.")
		}

		validatorsRankGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": moniker,
		}).Set(float64(index + 1))

		if validatorSetLength != 0 {
			var active float64
			if index+1 <= int(validatorSetLength) {
				active = 1
			} else {
				active = 0
			}
			validatorsIsActiveGauge.With(prometheus.Labels{
				"address": validator.OperatorAddress,
				"moniker": moniker,
			}).Set(active)
		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/validators").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}

func sanitizeUTF8(input string) string {
	buf := &bytes.Buffer{}
	for _, runeValue := range input {
		if utf8.ValidRune(runeValue) {
			buf.WriteRune(runeValue)
		}
	}
	return buf.String()
}

func (c *CosmosClient) GetValidators(ctx context.Context) ([]ValidatorData, error) {
	stakingClient := stakingtypes.NewQueryClient(c.grpcConn)

	validatorsResponse, err := stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Pagination: &query.PageRequest{
			Limit: 1000,
		},
	})
	if err != nil {
		return nil, err
	}

	var validators []ValidatorData
	for _, v := range validatorsResponse.Validators {
		validator := ValidatorData{
			Moniker:           v.Description.Moniker,
			OperatorAddress:   v.OperatorAddress,
			ConsensusPubkey:   v.ConsensusPubkey.String(),
			Jailed:            v.Jailed,
			Status:            v.Status,
			Tokens:            v.Tokens,
			DelegatorShares:   v.DelegatorShares,
			Description:       v.Description,
			UnbondingHeight:   v.UnbondingHeight,
			UnbondingTime:     v.UnbondingTime,
			Commission:        v.Commission,
			MinSelfDelegation: v.MinSelfDelegation,
		}
		validators = append(validators, validator)
	}

	return validators, nil
}
