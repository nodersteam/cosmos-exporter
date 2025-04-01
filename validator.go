package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getConsensusAddress(validator stakingtypes.Validator) (string, error) {
	// Создаем HTTP клиент
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Формируем URL для запроса валидатора через Tendermint RPC
	url := fmt.Sprintf("%s/validators", TendermintRPC)

	// Выполняем HTTP запрос
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Парсим JSON ответ
	var result struct {
		Result struct {
			Validators []struct {
				Address string `json:"address"`
				PubKey  struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"pub_key"`
			} `json:"validators"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Ищем валидатора по операторскому адресу
	for _, v := range result.Result.Validators {
		if v.PubKey.Type == "cometbft/PubKeyBn254" {
			return v.Address, nil
		}
	}

	return "", fmt.Errorf("validator not found")
}

func ValidatorHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	requestStart := time.Now()
	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	address := r.URL.Query().Get("address")
	myAddress, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		sublogger.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse validator address")
		return
	}

	validatorDelegationsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_delegations",
			Help:        "Delegations of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom", "delegated_by"},
	)

	validatorTokensGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_tokens",
			Help:        "Tokens of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorDelegatorSharesGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_delegators_shares",
			Help:        "Delegators shares of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorCommissionRateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_commission_rate",
			Help:        "Commission rate of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorCommissionGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_commission",
			Help:        "Commission of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorRewardsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_rewards",
			Help:        "Rewards of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom"},
	)

	validatorUnbondingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_unbondings",
			Help:        "Unbondings of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom", "unbonded_by"},
	)

	validatorRedelegationsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_redelegations",
			Help:        "Redelegations of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker", "denom", "redelegated_by", "redelegated_to"},
	)

	validatorMissedBlocksGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_missed_blocks",
			Help:        "Missed blocks of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorRankGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_rank",
			Help:        "Rank of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorIsActiveGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_active",
			Help:        "1 if the Cosmos-based blockchain validator is in active set, 0 if not",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_status",
			Help:        "Status of the Cosmos-based blockchain validator",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	validatorJailedGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_validator_jailed",
			Help:        "1 if the Cosmos-based blockchain validator is jailed, 0 if not",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "moniker"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(validatorDelegationsGauge)
	registry.MustRegister(validatorTokensGauge)
	registry.MustRegister(validatorDelegatorSharesGauge)
	registry.MustRegister(validatorCommissionRateGauge)
	registry.MustRegister(validatorCommissionGauge)
	registry.MustRegister(validatorRewardsGauge)
	registry.MustRegister(validatorUnbondingsGauge)
	registry.MustRegister(validatorRedelegationsGauge)
	registry.MustRegister(validatorMissedBlocksGauge)
	registry.MustRegister(validatorRankGauge)
	registry.MustRegister(validatorIsActiveGauge)
	registry.MustRegister(validatorStatusGauge)
	registry.MustRegister(validatorJailedGauge)

	sublogger.Debug().
		Str("address", address).
		Msg("Started querying validator")
	validatorQueryStart := time.Now()

	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	validatorResp, err := stakingClient.Validator(
		context.Background(),
		&stakingtypes.QueryValidatorRequest{ValidatorAddr: myAddress.String()},
	)
	if err != nil {
		sublogger.Error().
			Str("address", address).
			Err(err).
			Msg("Could not get validator")
		return
	}

	validator := validatorResp.Validator

	sublogger.Debug().
		Str("address", address).
		Float64("request-time", time.Since(validatorQueryStart).Seconds()).
		Msg("Finished querying validator")

	if value, err := strconv.ParseFloat(validator.Tokens.String(), 64); err != nil {
		sublogger.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse validator tokens")
	} else {
		validatorTokensGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": validator.Description.Moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)
	}

	if value, err := strconv.ParseFloat(validator.DelegatorShares.String(), 64); err != nil {
		sublogger.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse delegator shares")
	} else {
		validatorDelegatorSharesGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": validator.Description.Moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)
	}

	if rate, err := strconv.ParseFloat(validator.Commission.CommissionRates.Rate.String(), 64); err != nil {
		sublogger.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse commission rate")
	} else {
		validatorCommissionRateGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": validator.Description.Moniker,
		}).Set(rate)
	}

	validatorStatusGauge.With(prometheus.Labels{
		"address": validator.OperatorAddress,
		"moniker": validator.Description.Moniker,
	}).Set(float64(validator.Status))

	var jailed float64
	if validator.Jailed {
		jailed = 1
	} else {
		jailed = 0
	}
	validatorJailedGauge.With(prometheus.Labels{
		"address": validator.OperatorAddress,
		"moniker": validator.Description.Moniker,
	}).Set(jailed)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator delegations")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.ValidatorDelegations(
			context.Background(),
			&stakingtypes.QueryValidatorDelegationsRequest{
				ValidatorAddr: myAddress.String(),
				Pagination: &querytypes.PageRequest{
					Limit: Limit,
				},
			},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get validator delegations")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator delegations")

		for _, delegation := range stakingRes.DelegationResponses {
			if value, err := strconv.ParseFloat(delegation.Balance.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Str("address", address).
					Msg("Could not convert delegation entry")
			} else {
				validatorDelegationsGauge.With(prometheus.Labels{
					"moniker":      validator.Description.Moniker,
					"address":      delegation.Delegation.ValidatorAddress,
					"denom":        Denom,
					"delegated_by": delegation.Delegation.DelegatorAddress,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator commission")
		queryStart := time.Now()

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		distributionRes, err := distributionClient.ValidatorCommission(
			context.Background(),
			&distributiontypes.QueryValidatorCommissionRequest{ValidatorAddress: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get validator commission")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator commission")

		for _, commission := range distributionRes.Commission.Commission {
			if value, err := strconv.ParseFloat(commission.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Str("address", address).
					Msg("Could not parse validator commission")
			} else {
				validatorCommissionGauge.With(prometheus.Labels{
					"address": address,
					"moniker": validator.Description.Moniker,
					"denom":   Denom,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator rewards")
		queryStart := time.Now()

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		distributionRes, err := distributionClient.ValidatorOutstandingRewards(
			context.Background(),
			&distributiontypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get validator rewards")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator rewards")

		for _, reward := range distributionRes.Rewards.Rewards {
			if value, err := strconv.ParseFloat(reward.Amount.String(), 64); err != nil {
				sublogger.Error().
					Str("address", address).
					Err(err).
					Msg("Could not parse reward")
			} else {
				validatorRewardsGauge.With(prometheus.Labels{
					"address": address,
					"moniker": validator.Description.Moniker,
					"denom":   Denom,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator unbonding delegations")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.ValidatorUnbondingDelegations(
			context.Background(),
			&stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get validator unbonding delegations")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator unbonding delegations")

		for _, unbonding := range stakingRes.UnbondingResponses {
			var sum float64 = 0
			for _, entry := range unbonding.Entries {
				if value, err := strconv.ParseFloat(entry.Balance.String(), 64); err != nil {
					log.Error().
						Err(err).
						Str("address", address).
						Msg("Could not convert unbonding delegation entry")
				} else {
					sum += value
				}
			}

			validatorUnbondingsGauge.With(prometheus.Labels{
				"address":     unbonding.ValidatorAddress,
				"moniker":     validator.Description.Moniker,
				"denom":       Denom,
				"unbonded_by": unbonding.DelegatorAddress,
			}).Set(sum / DenomCoefficient)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator redelegations")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.Redelegations(
			context.Background(),
			&stakingtypes.QueryRedelegationsRequest{SrcValidatorAddr: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get redelegations")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator redelegations")

		for _, redelegation := range stakingRes.RedelegationResponses {
			var sum float64 = 0
			for _, entry := range redelegation.Entries {
				if value, err := strconv.ParseFloat(entry.Balance.String(), 64); err != nil {
					log.Error().
						Err(err).
						Str("address", address).
						Msg("Could not convert redelegation entry")
				} else {
					sum += value
				}
			}

			validatorRedelegationsGauge.With(prometheus.Labels{
				"address":        redelegation.Redelegation.ValidatorSrcAddress,
				"moniker":        validator.Description.Moniker,
				"denom":          Denom,
				"redelegated_by": redelegation.Redelegation.DelegatorAddress,
				"redelegated_to": redelegation.Redelegation.ValidatorDstAddress,
			}).Set(sum / DenomCoefficient)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator signing info")

		consAddr, err := getConsensusAddress(validator)
		if err != nil {
			sublogger.Error().
				Str("address", validator.OperatorAddress).
				Err(err).
				Msg("Could not get validator consensus address, skipping consensus metrics")
		} else {
			slashingClient := slashingtypes.NewQueryClient(grpcConn)
			slashingRes, err := slashingClient.SigningInfo(
				context.Background(),
				&slashingtypes.QuerySigningInfoRequest{ConsAddress: consAddr},
			)
			if err != nil {
				sublogger.Debug().
					Str("address", validator.OperatorAddress).
					Msg("Could not get signing info for validator")
			} else if validator.Status == stakingtypes.Bonded {
				validatorMissedBlocksGauge.With(prometheus.Labels{
					"address": validator.OperatorAddress,
					"moniker": validator.Description.Moniker,
				}).Set(float64(slashingRes.ValSigningInfo.MissedBlocksCounter))
			} else {
				sublogger.Trace().
					Str("address", validator.OperatorAddress).
					Msg("Validator is not active, not returning missed blocks amount.")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying validator rank and active status")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.Validators(
			context.Background(),
			&stakingtypes.QueryValidatorsRequest{
				Pagination: &querytypes.PageRequest{
					Limit: Limit,
				},
			},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get validators list")
			return
		}

		validators := stakingRes.Validators

		sort.Slice(validators, func(i, j int) bool {
			firstShares, firstErr := strconv.ParseFloat(validators[i].DelegatorShares.String(), 64)
			secondShares, secondErr := strconv.ParseFloat(validators[j].DelegatorShares.String(), 64)

			if firstErr != nil || secondErr != nil {
				sublogger.Error().
					Err(err).
					Msg("Error converting delegator shares for sorting")
				return true
			}

			return firstShares > secondShares
		})

		var validatorRank int
		for index, validatorIterated := range validators {
			if validatorIterated.OperatorAddress == validator.OperatorAddress {
				validatorRank = index + 1
				break
			}
		}

		if validatorRank == 0 {
			sublogger.Warn().
				Str("address", address).
				Msg("Could not find validator in validators list")
			return
		}

		validatorRankGauge.With(prometheus.Labels{
			"moniker": validator.Description.Moniker,
			"address": validator.OperatorAddress,
		}).Set(float64(validatorRank))

		paramsRes, err := stakingClient.Params(
			context.Background(),
			&stakingtypes.QueryParamsRequest{},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get staking params")
			return
		}

		var active float64
		if validatorRank <= int(paramsRes.Params.MaxValidators) {
			active = 1
		} else {
			active = 0
		}

		validatorIsActiveGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": validator.Description.Moniker,
		}).Set(active)

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying validator rank and active status")
	}()

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/validator?address="+address).
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
