package main

import (
	"context"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func WalletHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn, validationClient interface{}) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	address := r.URL.Query().Get("address")
	myAddress, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		sublogger.Error().
			Str("address", address).
			Err(err).
			Msg("Could not parse address")
		return
	}

	walletBalanceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_wallet_balance",
			Help:        "Balance of the Cosmos-based blockchain wallet",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "denom"},
	)

	walletDelegationGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_wallet_delegations",
			Help:        "Delegations of the Cosmos-based blockchain wallet",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "denom", "delegated_to"},
	)

	walletRedelegationGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_wallet_redelegations",
			Help:        "Redelegations of the Cosmos-based blockchain wallet",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "denom", "redelegated_from", "redelegated_to"},
	)

	walletUnbondingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_wallet_unbondings",
			Help:        "Unbondings of the Cosmos-based blockchain wallet",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "denom", "unbonded_from"},
	)

	walletRewardsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_wallet_rewards",
			Help:        "Rewards of the Cosmos-based blockchain wallet",
			ConstLabels: ConstLabels,
		},
		[]string{"address", "denom", "validator_address"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(walletBalanceGauge)
	registry.MustRegister(walletDelegationGauge)
	registry.MustRegister(walletUnbondingsGauge)
	registry.MustRegister(walletRedelegationGauge)
	registry.MustRegister(walletRewardsGauge)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying balance")
		queryStart := time.Now()

		bankClient := banktypes.NewQueryClient(grpcConn)
		bankRes, err := bankClient.AllBalances(
			context.Background(),
			&banktypes.QueryAllBalancesRequest{Address: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get balance")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying balance")

		for _, balance := range bankRes.Balances {
			value, _ := new(big.Float).SetInt(balance.Amount.BigInt()).Float64()
			walletBalanceGauge.With(prometheus.Labels{
				"address": address,
				"denom":   balance.Denom, // Используем реальный denom из ответа
			}).Set(value / DenomCoefficient)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying delegations")
		queryStart := time.Now()

		var stakingRes interface{}
		var err error

		if NetworkType == "zenrock" {
			client := validationClient.(ValidationClient)
			stakingRes, err = client.DelegatorDelegations(
				context.Background(),
				&QueryDelegatorDelegationsRequest{DelegatorAddr: myAddress.String()},
			)
		} else {
			client := validationClient.(stakingtypes.QueryClient)
			stakingRes, err = client.DelegatorDelegations(
				context.Background(),
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: myAddress.String(),
					Pagination: &query.PageRequest{
						Limit: Limit,
					},
				},
			)
		}

		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get delegations")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying delegations")

		if NetworkType == "zenrock" {
			res := stakingRes.(*QueryDelegatorDelegationsResponse)
			for _, delegation := range res.DelegationResponses {
				value, _ := strconv.ParseFloat(delegation.Balance.Amount, 64)
				walletDelegationGauge.With(prometheus.Labels{
					"address":      address,
					"denom":        delegation.Balance.Denom,
					"delegated_to": delegation.Delegation.ValidatorAddress,
				}).Set(value / DenomCoefficient)
			}
		} else {
			res := stakingRes.(*stakingtypes.QueryDelegatorDelegationsResponse)
			for _, delegation := range res.DelegationResponses {
				value, _ := strconv.ParseFloat(delegation.Balance.Amount.String(), 64)
				walletDelegationGauge.With(prometheus.Labels{
					"address":      address,
					"denom":        delegation.Balance.Denom,
					"delegated_to": delegation.Delegation.ValidatorAddress,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying unbonding delegations")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.DelegatorUnbondingDelegations(
			context.Background(),
			&stakingtypes.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get unbonding delegations")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying unbonding delegations")

		for _, unbonding := range stakingRes.UnbondingResponses {
			var sum float64 = 0
			for _, entry := range unbonding.Entries {
				value, _ := new(big.Float).SetInt(entry.Balance.BigInt()).Float64()
				sum += value
			}

			walletUnbondingsGauge.With(prometheus.Labels{
				"address":       unbonding.DelegatorAddress,
				"denom":         Denom, // Нет denoma в ответе, используем глобальный
				"unbonded_from": unbonding.ValidatorAddress,
			}).Set(sum / DenomCoefficient)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying redelegations")
		queryStart := time.Now()

		stakingClient := stakingtypes.NewQueryClient(grpcConn)
		stakingRes, err := stakingClient.Redelegations(
			context.Background(),
			&stakingtypes.QueryRedelegationsRequest{DelegatorAddr: myAddress.String()},
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
			Msg("Finished querying redelegations")

		for _, redelegation := range stakingRes.RedelegationResponses {
			var sum float64 = 0
			for _, entry := range redelegation.Entries {
				value, _ := new(big.Float).SetInt(entry.Balance.BigInt()).Float64()
				sum += value
			}

			walletRedelegationGauge.With(prometheus.Labels{
				"address":          redelegation.Redelegation.DelegatorAddress,
				"denom":            Denom, // Нет denoma в ответе, используем глобальный
				"redelegated_from": redelegation.Redelegation.ValidatorSrcAddress,
				"redelegated_to":   redelegation.Redelegation.ValidatorDstAddress,
			}).Set(sum / DenomCoefficient)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().
			Str("address", address).
			Msg("Started querying rewards")
		queryStart := time.Now()

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		distributionRes, err := distributionClient.DelegationTotalRewards(
			context.Background(),
			&distributiontypes.QueryDelegationTotalRewardsRequest{DelegatorAddress: myAddress.String()},
		)
		if err != nil {
			sublogger.Error().
				Str("address", address).
				Err(err).
				Msg("Could not get rewards")
			return
		}

		sublogger.Debug().
			Str("address", address).
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying rewards")

		for _, reward := range distributionRes.Rewards {
			for _, entry := range reward.Reward {
				value, _ := new(big.Float).SetInt(entry.Amount.BigInt()).Float64()
				walletRewardsGauge.With(prometheus.Labels{
					"address":           address,
					"denom":             entry.Denom, // Используем реальный denom
					"validator_address": reward.ValidatorAddress,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/wallet?address="+address).
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
