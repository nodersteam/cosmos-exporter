package main

import (
	"context"
	"math/big"
	"net/http"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/cosmos/cosmos-sdk/simapp"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func ValidatorsHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	encCfg := simapp.MakeTestEncodingConfig()
	interfaceRegistry := encCfg.InterfaceRegistry

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
			Help:        "Self declared minimum self delegation shares of the Cosmos-based blockchain validator",
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
			Help:        "1 if the Cosmos-based blockchain validator is in active set, 0 if no",
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

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		validatorsResponse, err := stakingClient.Validators(
			context.Background(),
			&stakingtypes.QueryValidatorsRequest{
				Pagination: &querytypes.PageRequest{
					Limit: Limit,
				},
			},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get validators")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validators")
		validators = validatorsResponse.Validators

		// sorting by delegator shares to display rankings
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
				Pagination: &querytypes.PageRequest{
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
		// Замените преобразование Decimal в Float на использование math/big.Float
		tokensFloat := new(big.Float)
		tokensFloat, _, err := tokensFloat.Parse(validator.Tokens.String(), 10)

		if err != nil {
			log.Error().
				Err(err).
				Str("address", validator.OperatorAddress).
				Msg("Could not parse delegator tokens")
		} else {
			// Ensure that the 'moniker' string contains only valid UTF-8 characters.
			moniker := validator.Description.Moniker
			moniker = sanitizeUTF8(moniker)
			tokensFloatVal, _ := tokensFloat.Float64()
			validatorsTokensGauge.With(prometheus.Labels{
				"address": validator.OperatorAddress,
				"moniker": moniker,
				"denom":   Denom,
			}).Set(tokensFloatVal)

			validatorsStatusGauge.With(prometheus.Labels{
				"address": validator.OperatorAddress,
				"moniker": moniker,
			}).Set(float64(validator.Status))

			// golang doesn't have a ternary operator, so we have to stick with this ugly solution
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

			// Замените преобразование Decimal в Float на использование math/big.Float
			delegatorSharesFloat := new(big.Float)
			delegatorSharesFloat, _, err := delegatorSharesFloat.Parse(validator.DelegatorShares.String(), 10)

			if err != nil {
				log.Error().
					Err(err).
					Str("address", validator.OperatorAddress).
					Msg("Could not parse delegator shares")
			} else {
				delegatorSharesFloatVal, _ := delegatorSharesFloat.Float64()
				validatorsDelegatorSharesGauge.With(prometheus.Labels{
					"address": validator.OperatorAddress,
					"moniker": moniker,
					"denom":   Denom,
				}).Set(delegatorSharesFloatVal)

				// Замените преобразование Decimal в Float на использование math/big.Float
				minSelfDelegationFloat := new(big.Float)
				minSelfDelegationFloat, _, err := minSelfDelegationFloat.Parse(validator.MinSelfDelegation.String(), 10)

				if err != nil {
					log.Error().
						Err(err).
						Str("address", validator.OperatorAddress).
						Msg("Could not parse validator min self delegation")
				} else {
					minSelfDelegationFloatVal, _ := minSelfDelegationFloat.Float64()
					validatorsMinSelfDelegationGauge.With(prometheus.Labels{
						"address": validator.OperatorAddress,
						"moniker": moniker,
						"denom":   Denom,
					}).Set(minSelfDelegationFloatVal)

					err = validator.UnpackInterfaces(interfaceRegistry) // Unpack interfaces, to populate the Anys' cached values
					if err != nil {
						sublogger.Error().
							Str("address", validator.OperatorAddress).
							Err(err).
							Msg("Could not get unpack validator interfaces")
					}

					pubKey, err := validator.GetConsAddr()
					if err != nil {
						sublogger.Error().
							Str("address", validator.OperatorAddress).
							Err(err).
							Msg("Could not get validator pubkey")
					}

					var signingInfo slashingtypes.ValidatorSigningInfo
					found := false

					for _, signingInfoIterated := range signingInfos {
						if pubKey.String() == signingInfoIterated.Address {
							found = true
							signingInfo = signingInfoIterated
							break
						}
					}

					if !found {
						sublogger.Debug().
							Str("address", validator.OperatorAddress).
							Msg("Could not get signing info for validator")
						continue
					}

					if validator.Status == stakingtypes.Bonded {
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
						// golang doesn't have a ternary operator, so we have to stick with this ugly solution
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
			}
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

// Определите функцию для очистки строки от недопустимых символов UTF-8.
func sanitizeUTF8(input string) string {
	result := make([]rune, 0, len(input))
	for _, runeValue := range input {
		if utf8.ValidRune(runeValue) {
			result = append(result, runeValue)
		}
	}
	return string(result)
}
