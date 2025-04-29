package main

import (
	"context"

	"google.golang.org/grpc"
)

// ValidationClient представляет клиент для работы с модулем validation в Zenrock
type ValidationClient interface {
	DelegatorDelegations(ctx context.Context, in *QueryDelegatorDelegationsRequest, opts ...grpc.CallOption) (*QueryDelegatorDelegationsResponse, error)
	ValidatorDelegations(ctx context.Context, in *QueryValidatorDelegationsRequest, opts ...grpc.CallOption) (*QueryValidatorDelegationsResponse, error)
	Validators(ctx context.Context, in *QueryValidatorsRequest, opts ...grpc.CallOption) (*QueryValidatorsResponse, error)
	Validator(ctx context.Context, in *QueryValidatorRequest, opts ...grpc.CallOption) (*QueryValidatorResponse, error)
	Pool(ctx context.Context, in *QueryPoolRequest, opts ...grpc.CallOption) (*QueryPoolResponse, error)
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
}

// QueryDelegatorDelegationsRequest и другие структуры запросов
type QueryDelegatorDelegationsRequest struct {
	DelegatorAddr string
}

type QueryDelegatorDelegationsResponse struct {
	DelegationResponses []DelegationResponse
}

type DelegationResponse struct {
	Delegation Delegation
	Balance    Coin
}

type Delegation struct {
	DelegatorAddress string
	ValidatorAddress string
	Shares           string
}

type Coin struct {
	Denom  string
	Amount string
}

// QueryValidatorDelegationsRequest и другие структуры запросов
type QueryValidatorDelegationsRequest struct {
	ValidatorAddr string
}

type QueryValidatorDelegationsResponse struct {
	DelegationResponses []DelegationResponse
}

type QueryValidatorsRequest struct {
	Status string
}

type QueryValidatorsResponse struct {
	Validators []Validator
}

type Validator struct {
	OperatorAddress string
	ConsensusPubkey string
	Jailed          bool
	Status          string
	Tokens          string
	DelegatorShares string
	Description     Description
}

type Description struct {
	Moniker string
}

type QueryValidatorRequest struct {
	ValidatorAddr string
}

type QueryValidatorResponse struct {
	Validator Validator
}

type QueryPoolRequest struct{}

type QueryPoolResponse struct {
	Pool Pool
}

type Pool struct {
	NotBondedTokens string
	BondedTokens    string
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params
}

type Params struct {
	UnbondingTime     string
	MaxValidators     uint32
	MaxEntries        uint32
	HistoricalEntries uint32
	BondDenom         string
}

// NewValidationClient создает новый клиент для работы с модулем validation
func NewValidationClient(conn *grpc.ClientConn) ValidationClient {
	return &validationClient{conn}
}

type validationClient struct {
	cc *grpc.ClientConn
}

func (c *validationClient) DelegatorDelegations(ctx context.Context, in *QueryDelegatorDelegationsRequest, opts ...grpc.CallOption) (*QueryDelegatorDelegationsResponse, error) {
	out := new(QueryDelegatorDelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/DelegatorDelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *validationClient) ValidatorDelegations(ctx context.Context, in *QueryValidatorDelegationsRequest, opts ...grpc.CallOption) (*QueryValidatorDelegationsResponse, error) {
	out := new(QueryValidatorDelegationsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/ValidatorDelegations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *validationClient) Validators(ctx context.Context, in *QueryValidatorsRequest, opts ...grpc.CallOption) (*QueryValidatorsResponse, error) {
	out := new(QueryValidatorsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Validators", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *validationClient) Validator(ctx context.Context, in *QueryValidatorRequest, opts ...grpc.CallOption) (*QueryValidatorResponse, error) {
	out := new(QueryValidatorResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Validator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *validationClient) Pool(ctx context.Context, in *QueryPoolRequest, opts ...grpc.CallOption) (*QueryPoolResponse, error) {
	out := new(QueryPoolResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Pool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *validationClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/zrchain.validation.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
