package main

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type ValidatorMetricsHandler struct {
	logger *zerolog.Logger
}

// GetValidator получает информацию о валидаторе по его адресу
func GetValidator(address string, grpcConn *grpc.ClientConn) (stakingtypes.Validator, error) {
	valAddr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("could not parse validator address: %v", err)
	}

	stakingClient := stakingtypes.NewQueryClient(grpcConn)
	validatorResp, err := stakingClient.Validator(
		context.Background(),
		&stakingtypes.QueryValidatorRequest{ValidatorAddr: valAddr.String()},
	)
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("could not get validator: %v", err)
	}

	return validatorResp.Validator, nil
}

func ValidatorHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	start := time.Now()
	requestID := uuid.New().String()

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Missing address parameter", http.StatusBadRequest)
		return
	}

	validator, err := GetValidator(address, grpcConn)
	if err != nil {
		log.Error().Err(err).Str("address", address).Str("request-id", requestID).Msg("Could not get validator")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	consAddr := GetValidatorConsAddr(&validator)
	if consAddr == nil {
		log.Printf("Не удалось получить консенсусный адрес валидатора")
		http.Error(w, "Не удалось получить консенсусный адрес валидатора", http.StatusInternalServerError)
		return
	}

	// Получаем информацию о подписях валидатора
	slashingClient := slashingtypes.NewQueryClient(grpcConn)
	signingInfoReq := &slashingtypes.QuerySigningInfoRequest{
		ConsAddress: sdk.ConsAddress(consAddr).String(),
	}
	signingInfoResp, err := slashingClient.SigningInfo(r.Context(), signingInfoReq)
	if err != nil {
		log.Printf("Ошибка при получении информации о подписях валидатора: %v", err)
		http.Error(w, "Ошибка при получении информации о подписях валидатора", http.StatusInternalServerError)
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

	log.Info().
		Str("address", address).
		Str("request-id", requestID).
		Msg("Started querying validator")

	logger := log.With().Str("request-id", requestID).Logger()
	metricsHandler := &ValidatorMetricsHandler{
		logger: &logger,
	}

	if err := metricsHandler.Handle(validator); err != nil {
		log.Error().
			Str("address", address).
			Err(err).
			Str("request-id", requestID).
			Msg("Failed to handle validator metrics")
		return
	}

	if value, err := strconv.ParseFloat(validator.Tokens.String(), 64); err != nil {
		log.Error().
			Err(err).
			Str("address", address).
			Str("request-id", requestID).
			Msg("Could not parse validator tokens")
	} else {
		validatorTokensGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": validator.Description.Moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)
	}

	if value, err := strconv.ParseFloat(validator.DelegatorShares.String(), 64); err != nil {
		log.Error().
			Err(err).
			Str("address", address).
			Str("request-id", requestID).
			Msg("Could not parse delegator shares")
	} else {
		validatorDelegatorSharesGauge.With(prometheus.Labels{
			"address": validator.OperatorAddress,
			"moniker": validator.Description.Moniker,
			"denom":   Denom,
		}).Set(value / DenomCoefficient)
	}

	if rate, err := strconv.ParseFloat(validator.Commission.CommissionRates.Rate.String(), 64); err != nil {
		log.Error().
			Err(err).
			Str("address", address).
			Str("request-id", requestID).
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
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator delegations")

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.ValidatorDelegations(
			context.Background(),
			&stakingtypes.QueryValidatorDelegationsRequest{
				ValidatorAddr: validator.OperatorAddress,
				Pagination: &querytypes.PageRequest{
					Limit: Limit,
				},
			},
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
				Msg("Could not get validator delegations")
			return
		}

		for _, delegation := range stakingRes.DelegationResponses {
			if value, err := strconv.ParseFloat(delegation.Balance.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Str("address", address).
					Str("request-id", requestID).
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
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator commission")

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		distributionRes, err := distributionClient.ValidatorCommission(
			context.Background(),
			&distributiontypes.QueryValidatorCommissionRequest{ValidatorAddress: validator.OperatorAddress},
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
				Msg("Could not get validator commission")
			return
		}

		for _, commission := range distributionRes.Commission.Commission {
			if value, err := strconv.ParseFloat(commission.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Str("address", address).
					Str("request-id", requestID).
					Msg("Could not parse validator commission")
			} else {
				validatorCommissionGauge.With(prometheus.Labels{
					"address": validator.OperatorAddress,
					"moniker": validator.Description.Moniker,
					"denom":   Denom,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator rewards")

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		distributionRes, err := distributionClient.ValidatorOutstandingRewards(
			context.Background(),
			&distributiontypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: validator.OperatorAddress},
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
				Msg("Could not get validator rewards")
			return
		}

		for _, reward := range distributionRes.Rewards.Rewards {
			if value, err := strconv.ParseFloat(reward.Amount.String(), 64); err != nil {
				log.Error().
					Err(err).
					Str("address", address).
					Str("request-id", requestID).
					Msg("Could not parse reward")
			} else {
				validatorRewardsGauge.With(prometheus.Labels{
					"address": validator.OperatorAddress,
					"moniker": validator.Description.Moniker,
					"denom":   Denom,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator unbonding delegations")

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.ValidatorUnbondingDelegations(
			context.Background(),
			&stakingtypes.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: validator.OperatorAddress},
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
				Msg("Could not get validator unbonding delegations")
			return
		}

		for _, unbonding := range stakingRes.UnbondingResponses {
			var sum float64 = 0
			for _, entry := range unbonding.Entries {
				if value, err := strconv.ParseFloat(entry.Balance.String(), 64); err != nil {
					log.Error().
						Err(err).
						Str("address", address).
						Str("request-id", requestID).
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
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator redelegations")

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.Redelegations(
			context.Background(),
			&stakingtypes.QueryRedelegationsRequest{SrcValidatorAddr: validator.OperatorAddress},
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
				Msg("Could not get redelegations")
			return
		}

		for _, redelegation := range stakingRes.RedelegationResponses {
			var sum float64 = 0
			for _, entry := range redelegation.Entries {
				if value, err := strconv.ParseFloat(entry.Balance.String(), 64); err != nil {
					log.Error().
						Err(err).
						Str("address", address).
						Str("request-id", requestID).
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
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator signing info")

		validatorMissedBlocksGauge.With(prometheus.Labels{
			"moniker": validator.Description.Moniker,
			"address": validator.OperatorAddress,
		}).Set(float64(signingInfoResp.ValSigningInfo.MissedBlocksCounter))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debug().
			Str("address", address).
			Str("request-id", requestID).
			Msg("Started querying validator rank and active status")

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		validatorsResp, err := stakingClient.Validators(
			context.Background(),
			&stakingtypes.QueryValidatorsRequest{
				Pagination: &querytypes.PageRequest{
					Limit: Limit,
				},
			},
		)
		if err != nil {
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
				Msg("Could not get validators list")
			return
		}

		validators := validatorsResp.Validators

		sort.Slice(validators, func(i, j int) bool {
			firstShares, firstErr := strconv.ParseFloat(validators[i].DelegatorShares.String(), 64)
			secondShares, secondErr := strconv.ParseFloat(validators[j].DelegatorShares.String(), 64)

			if firstErr != nil || secondErr != nil {
				log.Error().
					Err(err).
					Str("request-id", requestID).
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
			log.Warn().
				Str("address", address).
				Str("request-id", requestID).
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
			log.Error().
				Err(err).
				Str("address", address).
				Str("request-id", requestID).
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
	}()

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	log.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/validator?address="+address).
		Float64("request-time", time.Since(start).Seconds()).
		Str("request-id", requestID).
		Msg("Request processed")
}

// GetValidatorConsAddr возвращает консенсусный адрес валидатора
func GetValidatorConsAddr(validator *stakingtypes.Validator) []byte {
	if validator.ConsensusPubkey == nil {
		return nil
	}

	var pubKey cryptotypes.PubKey
	err := interfaceRegistry.UnpackAny(validator.ConsensusPubkey, &pubKey)
	if err != nil {
		log.Printf("Ошибка при распаковке публичного ключа валидатора: %v", err)
		return nil
	}

	// Для Bn254 ключей возвращаем байты напрямую
	if bn254Key, ok := pubKey.(*PubKeyBn254); ok {
		return bn254Key.Bytes()
	}

	// Для других типов ключей используем стандартный метод
	return pubKey.Address()
}

func (v *ValidatorMetricsHandler) Handle(validator stakingtypes.Validator) error {
	sublogger := v.logger.With().
		Str("validator", validator.OperatorAddress).
		Logger()

	queryStart := time.Now()

	consAddr := GetValidatorConsAddr(&validator)
	if consAddr == nil {
		sublogger.Error().
			Str("validator", validator.OperatorAddress).
			Msg("Failed to get validator consensus address")
		return fmt.Errorf("failed to get validator consensus address")
	}

	// Используем consAddr и queryStart в дальнейшей логике
	_ = consAddr
	_ = queryStart

	return nil
}
