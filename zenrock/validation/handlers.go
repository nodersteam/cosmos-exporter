package validation

import (
	"context"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

// ValidatorHandler handles requests for validator metrics
func ValidatorHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	validatorAddr := r.URL.Query().Get("address")
	if validatorAddr == "" {
		http.Error(w, "validator address is required", http.StatusBadRequest)
		return
	}

	client := NewClient(grpcConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get validator info
	validator, err := client.Validator(ctx, validatorAddr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get validator info")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get validator delegations
	delegations, err := client.ValidatorDelegations(ctx, validatorAddr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get validator delegations")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get validator unbonding delegations
	unbondingDelegations, err := client.ValidatorUnbondingDelegations(ctx, validatorAddr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get validator unbonding delegations")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Export metrics
	exportValidatorMetrics(w, validator, delegations, unbondingDelegations)
}

// ValidatorsHandler handles requests for all validators metrics
func ValidatorsHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	client := NewClient(grpcConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all validators
	validators, err := client.Validators(ctx, "BOND_STATUS_BONDED")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get validators")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Export metrics
	exportValidatorsMetrics(w, validators)
}

// ParamsHandler handles requests for staking parameters
func ParamsHandler(w http.ResponseWriter, r *http.Request, grpcConn *grpc.ClientConn) {
	client := NewClient(grpcConn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get staking parameters
	params, err := client.Params(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get staking parameters")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Export metrics
	exportParamsMetrics(w, params)
}

// Helper functions for exporting metrics
func exportValidatorMetrics(w http.ResponseWriter, validator *QueryValidatorResponse, delegations *QueryValidatorDelegationsResponse, unbondingDelegations *QueryValidatorUnbondingDelegationsResponse) {
	// Create Prometheus metrics
	validatorPower := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_power",
		Help: "Validator voting power",
	})
	validatorCommission := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_commission",
		Help: "Validator commission rate",
	})
	validatorDelegations := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_delegations",
		Help: "Total delegations to validator",
	})
	validatorUnbondingDelegations := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_unbonding_delegations",
		Help: "Total unbonding delegations from validator",
	})

	// Set metric values
	tokens, _ := validator.Validator.Tokens.BigInt().Float64()
	commission := validator.Validator.Commission.CommissionRates.Rate.MustFloat64()
	validatorPower.Set(tokens)
	validatorCommission.Set(commission)
	validatorDelegations.Set(float64(len(delegations.DelegationResponses)))
	validatorUnbondingDelegations.Set(float64(len(unbondingDelegations.UnbondingResponses)))

	// Export metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(validatorPower, validatorCommission, validatorDelegations, validatorUnbondingDelegations)
	prometheus.WriteToTextfile("/tmp/metrics", registry)
}

func exportValidatorsMetrics(w http.ResponseWriter, validators *QueryValidatorsResponse) {
	// Create Prometheus metrics
	validatorsCount := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validators_count",
		Help: "Total number of validators",
	})
	validatorsTotalPower := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validators_total_power",
		Help: "Total voting power of all validators",
	})

	// Calculate total power
	var totalPower float64
	for _, v := range validators.Validators {
		power, _ := v.Tokens.BigInt().Float64()
		totalPower += power
	}

	// Set metric values
	validatorsCount.Set(float64(len(validators.Validators)))
	validatorsTotalPower.Set(totalPower)

	// Export metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(validatorsCount, validatorsTotalPower)
	prometheus.WriteToTextfile("/tmp/metrics", registry)
}

func exportParamsMetrics(w http.ResponseWriter, params *QueryParamsResponse) {
	// Create Prometheus metrics
	maxValidators := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "staking_max_validators",
		Help: "Maximum number of validators",
	})
	unbondingTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "staking_unbonding_time",
		Help: "Unbonding time in seconds",
	})
	maxEntries := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "staking_max_entries",
		Help: "Maximum number of entries in unbonding/delegation queue",
	})
	historicalEntries := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "staking_historical_entries",
		Help: "Number of historical entries",
	})
	bondDenom := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "staking_bond_denom",
		Help: "Bond denomination",
	})

	// Set metric values
	maxValidators.Set(float64(params.Params.MaxValidators))
	unbondingTime.Set(float64(params.Params.UnbondingTime))
	maxEntries.Set(float64(params.Params.MaxEntries))
	historicalEntries.Set(float64(params.Params.HistoricalEntries))
	bondDenom.Set(1) // We'll use 1 as a placeholder since we can't set string values

	// Export metrics
	registry := prometheus.NewRegistry()
	registry.MustRegister(maxValidators, unbondingTime, maxEntries, historicalEntries, bondDenom)
	prometheus.WriteToTextfile("/tmp/metrics", registry)
}
