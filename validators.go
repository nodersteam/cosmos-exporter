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

	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/slashing/types/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking/types/staking"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

	var validators []staking.Validator
	var signingInfos []slashing.ValidatorSigningInfo
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
			client := validationClient.(staking.QueryClient)
			validatorsResponse, err = client.Validators(
				context.Background(),
				&staking.QueryValidatorsRequest{
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
			validators = make([]staking.Validator, len(res.Validators))
			for i, v := range res.Validators {
				validators[i] = staking.Validator{
					OperatorAddress: v.OperatorAddress,
					ConsensusPubkey: &types.Any{
						TypeUrl: "/cosmos.crypto.ed25519.PubKey",
						Value:   []byte(v.ConsensusPubkey),
					},
					Jailed:          v.Jailed,
					Status:          staking.BondStatus(staking.BondStatus_value[v.Status]),
					Tokens:          sdk.NewIntFromString(v.Tokens),
					DelegatorShares: sdk.NewDecFromStr(v.DelegatorShares),
					Description: staking.Description{
						Moniker: v.Description.Moniker,
					},
				}
			}
		} else {
			res := validatorsResponse.(*staking.QueryValidatorsResponse)
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

		slashingClient := slashing.NewQueryClient(grpcConn)
		signingInfosResponse, err := slashingClient.SigningInfos(
			context.Background(),
			&slashing.QuerySigningInfosRequest{
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

		stakingClient := staking.NewQueryClient(grpcConn)
		paramsResponse, err := stakingClient.Params(
			context.Background(),
			&staking.QueryParamsRequest{},
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

		var signingInfo slashing.ValidatorSigningInfo
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
		} else if validator.Status == staking.Bonded {
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
