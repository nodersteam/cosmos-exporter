package validation

import (
	context "context"

	grpc "google.golang.org/grpc"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	// Delegation queries validator info for given validator address.
	Delegation(ctx context.Context, in *QueryDelegationRequest, opts ...grpc.CallOption) (*QueryDelegationResponse, error)
	// DelegatorDelegations queries all delegations of a given delegator address.
	DelegatorDelegations(ctx context.Context, in *QueryDelegatorDelegationsRequest, opts ...grpc.CallOption) (*QueryDelegatorDelegationsResponse, error)
	// DelegatorUnbondingDelegations queries all unbonding delegations of a given delegator address.
	DelegatorUnbondingDelegations(ctx context.Context, in *QueryDelegatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*QueryDelegatorUnbondingDelegationsResponse, error)
	// DelegatorValidator queries validator info for given delegator validator pair.
	DelegatorValidator(ctx context.Context, in *QueryDelegatorValidatorRequest, opts ...grpc.CallOption) (*QueryDelegatorValidatorResponse, error)
	// DelegatorValidators queries all validators info for given delegator address.
	DelegatorValidators(ctx context.Context, in *QueryDelegatorValidatorsRequest, opts ...grpc.CallOption) (*QueryDelegatorValidatorsResponse, error)
	// HistoricalInfo queries the historical info for given height.
	HistoricalInfo(ctx context.Context, in *QueryHistoricalInfoRequest, opts ...grpc.CallOption) (*QueryHistoricalInfoResponse, error)
	// Params queries all parameters.
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// Pool queries the pool info.
	Pool(ctx context.Context, in *QueryPoolRequest, opts ...grpc.CallOption) (*QueryPoolResponse, error)
	// Redelegations queries redelegations of given address.
	Redelegations(ctx context.Context, in *QueryRedelegationsRequest, opts ...grpc.CallOption) (*QueryRedelegationsResponse, error)
	// UnbondingDelegation queries unbonding info for given validator delegator pair.
	UnbondingDelegation(ctx context.Context, in *QueryUnbondingDelegationRequest, opts ...grpc.CallOption) (*QueryUnbondingDelegationResponse, error)
	// Validator queries validator info for given validator address.
	Validator(ctx context.Context, in *QueryValidatorRequest, opts ...grpc.CallOption) (*QueryValidatorResponse, error)
	// ValidatorDelegations queries all delegations of a given validator address.
	ValidatorDelegations(ctx context.Context, in *QueryValidatorDelegationsRequest, opts ...grpc.CallOption) (*QueryValidatorDelegationsResponse, error)
	// ValidatorPower queries validator power for given validator address.
	ValidatorPower(ctx context.Context, in *QueryPowerRequest, opts ...grpc.CallOption) (*QueryPowerResponse, error)
	// ValidatorUnbondingDelegations queries all unbonding delegations of a given validator address.
	ValidatorUnbondingDelegations(ctx context.Context, in *QueryValidatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*QueryValidatorUnbondingDelegationsResponse, error)
	// Validators queries all validators that match the given status.
	Validators(ctx context.Context, in *QueryValidatorsRequest, opts ...grpc.CallOption) (*QueryValidatorsResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Delegation(ctx context.Context, in *QueryDelegationRequest, opts ...grpc.CallOption) (*QueryDelegationResponse, error) {
	out := new(QueryDelegationResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Delegation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DelegatorDelegations(ctx context.Context, in *QueryDelegatorDelegationsRequest, opts ...grpc.CallOption) (*QueryDelegatorDelegationsResponse, error) {
	out := new(QueryDelegatorDelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/DelegatorDelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DelegatorUnbondingDelegations(ctx context.Context, in *QueryDelegatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*QueryDelegatorUnbondingDelegationsResponse, error) {
	out := new(QueryDelegatorUnbondingDelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/DelegatorUnbondingDelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DelegatorValidator(ctx context.Context, in *QueryDelegatorValidatorRequest, opts ...grpc.CallOption) (*QueryDelegatorValidatorResponse, error) {
	out := new(QueryDelegatorValidatorResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/DelegatorValidator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DelegatorValidators(ctx context.Context, in *QueryDelegatorValidatorsRequest, opts ...grpc.CallOption) (*QueryDelegatorValidatorsResponse, error) {
	out := new(QueryDelegatorValidatorsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/DelegatorValidators", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) HistoricalInfo(ctx context.Context, in *QueryHistoricalInfoRequest, opts ...grpc.CallOption) (*QueryHistoricalInfoResponse, error) {
	out := new(QueryHistoricalInfoResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/HistoricalInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Pool(ctx context.Context, in *QueryPoolRequest, opts ...grpc.CallOption) (*QueryPoolResponse, error) {
	out := new(QueryPoolResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Pool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Redelegations(ctx context.Context, in *QueryRedelegationsRequest, opts ...grpc.CallOption) (*QueryRedelegationsResponse, error) {
	out := new(QueryRedelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Redelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) UnbondingDelegation(ctx context.Context, in *QueryUnbondingDelegationRequest, opts ...grpc.CallOption) (*QueryUnbondingDelegationResponse, error) {
	out := new(QueryUnbondingDelegationResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/UnbondingDelegation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Validator(ctx context.Context, in *QueryValidatorRequest, opts ...grpc.CallOption) (*QueryValidatorResponse, error) {
	out := new(QueryValidatorResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Validator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ValidatorDelegations(ctx context.Context, in *QueryValidatorDelegationsRequest, opts ...grpc.CallOption) (*QueryValidatorDelegationsResponse, error) {
	out := new(QueryValidatorDelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/ValidatorDelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ValidatorPower(ctx context.Context, in *QueryPowerRequest, opts ...grpc.CallOption) (*QueryPowerResponse, error) {
	out := new(QueryPowerResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/ValidatorPower", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ValidatorUnbondingDelegations(ctx context.Context, in *QueryValidatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*QueryValidatorUnbondingDelegationsResponse, error) {
	out := new(QueryValidatorUnbondingDelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/ValidatorUnbondingDelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Validators(ctx context.Context, in *QueryValidatorsRequest, opts ...grpc.CallOption) (*QueryValidatorsResponse, error) {
	out := new(QueryValidatorsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Validators", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
