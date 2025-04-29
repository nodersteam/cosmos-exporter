package validation

import (
	"context"

	"google.golang.org/grpc"
)

// Client wraps the QueryClient
type Client struct {
	queryClient QueryClient
}

// NewClient creates a new validation client
func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		queryClient: NewQueryClient(conn),
	}
}

// DelegatorDelegations returns all delegations of a given delegator address
func (c *Client) DelegatorDelegations(ctx context.Context, delegatorAddr string) (*QueryDelegatorDelegationsResponse, error) {
	return c.queryClient.DelegatorDelegations(ctx, &QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegatorAddr,
	})
}

// DelegatorUnbondingDelegations returns all unbonding delegations of a given delegator address
func (c *Client) DelegatorUnbondingDelegations(ctx context.Context, delegatorAddr string) (*QueryDelegatorUnbondingDelegationsResponse, error) {
	return c.queryClient.DelegatorUnbondingDelegations(ctx, &QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: delegatorAddr,
	})
}

// Validator returns validator info for given validator address
func (c *Client) Validator(ctx context.Context, validatorAddr string) (*QueryValidatorResponse, error) {
	return c.queryClient.Validator(ctx, &QueryValidatorRequest{
		ValidatorAddr: validatorAddr,
	})
}

// Validators returns all validators that match the given status
func (c *Client) Validators(ctx context.Context, status string) (*QueryValidatorsResponse, error) {
	return c.queryClient.Validators(ctx, &QueryValidatorsRequest{
		Status: status,
	})
}

// ValidatorDelegations returns all delegations of a given validator address
func (c *Client) ValidatorDelegations(ctx context.Context, validatorAddr string) (*QueryValidatorDelegationsResponse, error) {
	return c.queryClient.ValidatorDelegations(ctx, &QueryValidatorDelegationsRequest{
		ValidatorAddr: validatorAddr,
	})
}

// ValidatorUnbondingDelegations returns all unbonding delegations of a given validator address
func (c *Client) ValidatorUnbondingDelegations(ctx context.Context, validatorAddr string) (*QueryValidatorUnbondingDelegationsResponse, error) {
	return c.queryClient.ValidatorUnbondingDelegations(ctx, &QueryValidatorUnbondingDelegationsRequest{
		ValidatorAddr: validatorAddr,
	})
}

// Pool returns the pool info
func (c *Client) Pool(ctx context.Context) (*QueryPoolResponse, error) {
	return c.queryClient.Pool(ctx, &QueryPoolRequest{})
}

// Params returns all parameters
func (c *Client) Params(ctx context.Context) (*QueryParamsResponse, error) {
	return c.queryClient.Params(ctx, &QueryParamsRequest{})
}

// Redelegations returns redelegations of given address
func (c *Client) Redelegations(ctx context.Context, delegatorAddr string) (*QueryRedelegationsResponse, error) {
	return c.queryClient.Redelegations(ctx, &QueryRedelegationsRequest{
		DelegatorAddr: delegatorAddr,
	})
}
