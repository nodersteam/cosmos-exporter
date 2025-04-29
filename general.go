package main

import (
	"context"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"

	"cosmossdk.io/math"
	"google.golang.org/grpc"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func GeneralHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn, validationClient interface{}) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	generalBondedTokensGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_pool_bonded_tokens",
			Help:        "Bonded tokens in the Cosmos-based blockchain pool",
			ConstLabels: ConstLabels,
		},
	)

	generalNotBondedTokensGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_pool_not_bonded_tokens",
			Help:        "Not bonded tokens in the Cosmos-based blockchain pool",
			ConstLabels: ConstLabels,
		},
	)

	generalCommunityPoolGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_community_pool",
			Help:        "Community pool",
			ConstLabels: ConstLabels,
		},
		[]string{"denom"},
	)

	generalSupplyTotalGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_supply_total",
			Help:        "Total supply",
			ConstLabels: ConstLabels,
		},
		[]string{"denom"},
	)

	generalInflationGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_inflation",
			Help:        "Inflation rate",
			ConstLabels: ConstLabels,
		},
	)

	generalAnnualProvisions := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "cosmos_general_annual_provisions",
			Help:        "Annual provisions",
			ConstLabels: ConstLabels,
		},
		[]string{"denoms"},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(generalBondedTokensGauge)
	registry.MustRegister(generalNotBondedTokensGauge)
	registry.MustRegister(generalCommunityPoolGauge)
	registry.MustRegister(generalSupplyTotalGauge)
	registry.MustRegister(generalInflationGauge)
	registry.MustRegister(generalAnnualProvisions)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying staking pool")
		queryStart := time.Now()

		var res interface{}
		var err error

		if NetworkType == "zenrock" {
			client := validationClient.(ValidationClient)
			res, err = client.Pool(
				context.Background(),
				&QueryPoolRequest{},
			)
		} else {
			client := validationClient.(stakingtypes.QueryClient)
			res, err = client.Pool(
				context.Background(),
				&stakingtypes.QueryPoolRequest{},
			)
		}

		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get pool")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying staking pool")

		var bondedTokens, notBondedTokens math.Int
		if NetworkType == "zenrock" {
			poolRes := res.(*QueryPoolResponse)
			bondedTokens, _ = math.NewIntFromString(poolRes.Pool.BondedTokens)
			notBondedTokens, _ = math.NewIntFromString(poolRes.Pool.NotBondedTokens)
		} else {
			poolRes := res.(*stakingtypes.QueryPoolResponse)
			bondedTokens = poolRes.Pool.BondedTokens
			notBondedTokens = poolRes.Pool.NotBondedTokens
		}

		bondedFloat, _ := new(big.Float).SetInt(bondedTokens.BigInt()).Float64()
		notBondedFloat, _ := new(big.Float).SetInt(notBondedTokens.BigInt()).Float64()

		generalBondedTokensGauge.Set(bondedFloat / DenomCoefficient)
		generalNotBondedTokensGauge.Set(notBondedFloat / DenomCoefficient)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying distribution community pool")
		queryStart := time.Now()

		distributionClient := distributiontypes.NewQueryClient(grpcConn)
		response, err := distributionClient.CommunityPool(
			context.Background(),
			&distributiontypes.QueryCommunityPoolRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get distribution community pool")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying distribution community pool")

		for _, coin := range response.Pool {
			if value, err := strconv.ParseFloat(coin.Amount.String(), 64); err != nil {
				sublogger.Error().
					Err(err).
					Msg("Could not get community pool coin")
			} else {
				generalCommunityPoolGauge.With(prometheus.Labels{
					"denom": coin.Denom,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying bank total supply")
		queryStart := time.Now()

		bankClient := banktypes.NewQueryClient(grpcConn)
		response, err := bankClient.TotalSupply(
			context.Background(),
			&banktypes.QueryTotalSupplyRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get bank total supply")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying bank total supply")

		for _, coin := range response.Supply {
			if value, err := strconv.ParseFloat(coin.Amount.String(), 64); err != nil {
				sublogger.Error().
					Err(err).
					Msg("Could not get total supply")
			} else {
				generalSupplyTotalGauge.With(prometheus.Labels{
					"denom": coin.Denom,
				}).Set(value / DenomCoefficient)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying inflation")
		queryStart := time.Now()

		mintClient := minttypes.NewQueryClient(grpcConn)
		response, err := mintClient.Inflation(
			context.Background(),
			&minttypes.QueryInflationRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get inflation")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying inflation")

		if value, err := strconv.ParseFloat(response.Inflation.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse inflation")
		} else {
			generalInflationGauge.Set(value)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		sublogger.Debug().Msg("Started querying annual provisions")
		queryStart := time.Now()

		mintClient := minttypes.NewQueryClient(grpcConn)
		response, err := mintClient.AnnualProvisions(
			context.Background(),
			&minttypes.QueryAnnualProvisionsRequest{},
		)
		if err != nil {
			sublogger.Error().Err(err).Msg("Could not get annual provisions")
			return
		}

		sublogger.Debug().
			Float64("request-time", time.Since(queryStart).Seconds()).
			Msg("Finished querying annual provisions")

		if value, err := strconv.ParseFloat(response.AnnualProvisions.String(), 64); err != nil {
			sublogger.Error().
				Err(err).
				Msg("Could not parse annual provisions")
		} else {
			generalAnnualProvisions.With(prometheus.Labels{
				"denom": Denom,
			}).Set(value / DenomCoefficient)
		}
	}()

	wg.Wait()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/general").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
