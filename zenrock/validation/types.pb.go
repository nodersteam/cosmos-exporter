package validation

import (
	types "github.com/cosmos/cosmos-sdk/types"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

// QueryDelegationRequest is request type for the Query/Delegation RPC method.
type QueryDelegationRequest struct {
	DelegatorAddr string `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	ValidatorAddr string `protobuf:"bytes,2,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
}

// QueryDelegationResponse is response type for the Query/Delegation RPC method.
type QueryDelegationResponse struct {
	DelegationResponse *DelegationResponse `protobuf:"bytes,1,opt,name=delegation_response,json=delegationResponse,proto3" json:"delegation_response,omitempty"`
}

// QueryDelegatorDelegationsRequest is request type for the Query/DelegatorDelegations RPC method.
type QueryDelegatorDelegationsRequest struct {
	DelegatorAddr string             `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	Pagination    *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryDelegatorDelegationsResponse is response type for the Query/DelegatorDelegations RPC method.
type QueryDelegatorDelegationsResponse struct {
	DelegationResponses []*DelegationResponse `protobuf:"bytes,1,rep,name=delegation_responses,json=delegationResponses,proto3" json:"delegation_responses,omitempty"`
	Pagination          *query.PageResponse   `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryDelegatorUnbondingDelegationsRequest is request type for the Query/DelegatorUnbondingDelegations RPC method.
type QueryDelegatorUnbondingDelegationsRequest struct {
	DelegatorAddr string             `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	Pagination    *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryDelegatorUnbondingDelegationsResponse is response type for the Query/DelegatorUnbondingDelegations RPC method.
type QueryDelegatorUnbondingDelegationsResponse struct {
	UnbondingResponses []*UnbondingDelegation `protobuf:"bytes,1,rep,name=unbonding_responses,json=unbondingResponses,proto3" json:"unbonding_responses,omitempty"`
	Pagination         *query.PageResponse    `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryDelegatorValidatorRequest is request type for the Query/DelegatorValidator RPC method.
type QueryDelegatorValidatorRequest struct {
	DelegatorAddr string `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	ValidatorAddr string `protobuf:"bytes,2,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
}

// QueryDelegatorValidatorResponse is response type for the Query/DelegatorValidator RPC method.
type QueryDelegatorValidatorResponse struct {
	Validator *Validator `protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
}

// QueryDelegatorValidatorsRequest is request type for the Query/DelegatorValidators RPC method.
type QueryDelegatorValidatorsRequest struct {
	DelegatorAddr string             `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	Pagination    *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryDelegatorValidatorsResponse is response type for the Query/DelegatorValidators RPC method.
type QueryDelegatorValidatorsResponse struct {
	Validators []*Validator        `protobuf:"bytes,1,rep,name=validators,proto3" json:"validators,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryHistoricalInfoRequest is request type for the Query/HistoricalInfo RPC method.
type QueryHistoricalInfoRequest struct {
	Height int64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
}

// QueryHistoricalInfoResponse is response type for the Query/HistoricalInfo RPC method.
type QueryHistoricalInfoResponse struct {
	Hist *HistoricalInfo `protobuf:"bytes,1,opt,name=hist,proto3" json:"hist,omitempty"`
}

// QueryParamsRequest is request type for the Query/Params RPC method.
type QueryParamsRequest struct{}

// QueryParamsResponse is response type for the Query/Params RPC method.
type QueryParamsResponse struct {
	Params *Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params,omitempty"`
}

// QueryPoolRequest is request type for the Query/Pool RPC method.
type QueryPoolRequest struct{}

// QueryPoolResponse is response type for the Query/Pool RPC method.
type QueryPoolResponse struct {
	Pool *Pool `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool,omitempty"`
}

// QueryRedelegationsRequest is request type for the Query/Redelegations RPC method.
type QueryRedelegationsRequest struct {
	DelegatorAddr    string             `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	SrcValidatorAddr string             `protobuf:"bytes,2,opt,name=src_validator_addr,json=srcValidatorAddr,proto3" json:"src_validator_addr,omitempty"`
	DstValidatorAddr string             `protobuf:"bytes,3,opt,name=dst_validator_addr,json=dstValidatorAddr,proto3" json:"dst_validator_addr,omitempty"`
	Pagination       *query.PageRequest `protobuf:"bytes,4,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryRedelegationsResponse is response type for the Query/Redelegations RPC method.
type QueryRedelegationsResponse struct {
	RedelegationResponses []*RedelegationResponse `protobuf:"bytes,1,rep,name=redelegation_responses,json=redelegationResponses,proto3" json:"redelegation_responses,omitempty"`
	Pagination            *query.PageResponse     `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryUnbondingDelegationRequest is request type for the Query/UnbondingDelegation RPC method.
type QueryUnbondingDelegationRequest struct {
	DelegatorAddr string `protobuf:"bytes,1,opt,name=delegator_addr,json=delegatorAddr,proto3" json:"delegator_addr,omitempty"`
	ValidatorAddr string `protobuf:"bytes,2,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
}

// QueryUnbondingDelegationResponse is response type for the Query/UnbondingDelegation RPC method.
type QueryUnbondingDelegationResponse struct {
	Unbond *UnbondingDelegation `protobuf:"bytes,1,opt,name=unbond,proto3" json:"unbond,omitempty"`
}

// QueryValidatorRequest is request type for the Query/Validator RPC method.
type QueryValidatorRequest struct {
	ValidatorAddr string `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
}

// QueryValidatorResponse is response type for the Query/Validator RPC method.
type QueryValidatorResponse struct {
	Validator *Validator `protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
}

// QueryValidatorDelegationsRequest is request type for the Query/ValidatorDelegations RPC method.
type QueryValidatorDelegationsRequest struct {
	ValidatorAddr string             `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
	Pagination    *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryValidatorDelegationsResponse is response type for the Query/ValidatorDelegations RPC method.
type QueryValidatorDelegationsResponse struct {
	DelegationResponses []*DelegationResponse `protobuf:"bytes,1,rep,name=delegation_responses,json=delegationResponses,proto3" json:"delegation_responses,omitempty"`
	Pagination          *query.PageResponse   `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryPowerRequest is request type for the Query/ValidatorPower RPC method.
type QueryPowerRequest struct {
	ValidatorAddr string `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
}

// QueryPowerResponse is response type for the Query/ValidatorPower RPC method.
type QueryPowerResponse struct {
	Power int64 `protobuf:"varint,1,opt,name=power,proto3" json:"power,omitempty"`
}

// QueryValidatorUnbondingDelegationsRequest is request type for the Query/ValidatorUnbondingDelegations RPC method.
type QueryValidatorUnbondingDelegationsRequest struct {
	ValidatorAddr string             `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
	Pagination    *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryValidatorUnbondingDelegationsResponse is response type for the Query/ValidatorUnbondingDelegations RPC method.
type QueryValidatorUnbondingDelegationsResponse struct {
	UnbondingResponses []*UnbondingDelegation `protobuf:"bytes,1,rep,name=unbonding_responses,json=unbondingResponses,proto3" json:"unbonding_responses,omitempty"`
	Pagination         *query.PageResponse    `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryValidatorsRequest is request type for the Query/Validators RPC method.
type QueryValidatorsRequest struct {
	Status     string             `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Pagination *query.PageRequest `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// QueryValidatorsResponse is response type for the Query/Validators RPC method.
type QueryValidatorsResponse struct {
	Validators []*Validator        `protobuf:"bytes,1,rep,name=validators,proto3" json:"validators,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

// DelegationResponse is equivalent to a delegation except that it contains a balance in addition to shares which is more suitable for client responses.
type DelegationResponse struct {
	Delegation *Delegation `protobuf:"bytes,1,opt,name=delegation,proto3" json:"delegation,omitempty"`
	Balance    *types.Coin `protobuf:"bytes,2,opt,name=balance,proto3" json:"balance,omitempty"`
}

// Delegation represents the bond with tokens held by an account. It is owned by one delegator, and is associated with the voting power of one validator.
type Delegation struct {
	DelegatorAddress string `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	ValidatorAddress string `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Shares           string `protobuf:"bytes,3,opt,name=shares,proto3" json:"shares,omitempty"`
}

// UnbondingDelegation stores all of a single delegator's unbonding bonds for a single validator in an time-ordered list.
type UnbondingDelegation struct {
	DelegatorAddress string            `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	ValidatorAddress string            `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Entries          []*UnbondingEntry `protobuf:"bytes,3,rep,name=entries,proto3" json:"entries,omitempty"`
}

// UnbondingEntry defines an unbonding object with relevant metadata.
type UnbondingEntry struct {
	CreationHeight int64  `protobuf:"varint,1,opt,name=creation_height,json=creationHeight,proto3" json:"creation_height,omitempty"`
	CompletionTime int64  `protobuf:"varint,2,opt,name=completion_time,json=completionTime,proto3" json:"completion_time,omitempty"`
	InitialBalance string `protobuf:"bytes,3,opt,name=initial_balance,json=initialBalance,proto3" json:"initial_balance,omitempty"`
	Balance        string `protobuf:"bytes,4,opt,name=balance,proto3" json:"balance,omitempty"`
}

// Validator defines a validator, together with the total amount of the Validator's bond shares and their exchange rate to coins.
type Validator struct {
	OperatorAddress   string       `protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	ConsensusPubkey   string       `protobuf:"bytes,2,opt,name=consensus_pubkey,json=consensusPubkey,proto3" json:"consensus_pubkey,omitempty"`
	Jailed            bool         `protobuf:"varint,3,opt,name=jailed,proto3" json:"jailed,omitempty"`
	Status            int32        `protobuf:"varint,4,opt,name=status,proto3" json:"status,omitempty"`
	Tokens            string       `protobuf:"bytes,5,opt,name=tokens,proto3" json:"tokens,omitempty"`
	DelegatorShares   string       `protobuf:"bytes,6,opt,name=delegator_shares,json=delegatorShares,proto3" json:"delegator_shares,omitempty"`
	Description       *Description `protobuf:"bytes,7,opt,name=description,proto3" json:"description,omitempty"`
	UnbondingHeight   int64        `protobuf:"varint,8,opt,name=unbonding_height,json=unbondingHeight,proto3" json:"unbonding_height,omitempty"`
	UnbondingTime     int64        `protobuf:"varint,9,opt,name=unbonding_time,json=unbondingTime,proto3" json:"unbonding_time,omitempty"`
	Commission        *Commission  `protobuf:"bytes,10,opt,name=commission,proto3" json:"commission,omitempty"`
	MinSelfDelegation string       `protobuf:"bytes,11,opt,name=min_self_delegation,json=minSelfDelegation,proto3" json:"min_self_delegation,omitempty"`
}

// Description defines a validator description.
type Description struct {
	Moniker         string `protobuf:"bytes,1,opt,name=moniker,proto3" json:"moniker,omitempty"`
	Identity        string `protobuf:"bytes,2,opt,name=identity,proto3" json:"identity,omitempty"`
	Website         string `protobuf:"bytes,3,opt,name=website,proto3" json:"website,omitempty"`
	SecurityContact string `protobuf:"bytes,4,opt,name=security_contact,json=securityContact,proto3" json:"security_contact,omitempty"`
	Details         string `protobuf:"bytes,5,opt,name=details,proto3" json:"details,omitempty"`
}

// Commission defines commission parameters for a given validator.
type Commission struct {
	CommissionRates *CommissionRates `protobuf:"bytes,1,opt,name=commission_rates,json=commissionRates,proto3" json:"commission_rates,omitempty"`
	UpdateTime      int64            `protobuf:"varint,2,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
}

// CommissionRates defines the initial commission rates to be used for creating a validator.
type CommissionRates struct {
	Rate          string `protobuf:"bytes,1,opt,name=rate,proto3" json:"rate,omitempty"`
	MaxRate       string `protobuf:"bytes,2,opt,name=max_rate,json=maxRate,proto3" json:"max_rate,omitempty"`
	MaxChangeRate string `protobuf:"bytes,3,opt,name=max_change_rate,json=maxChangeRate,proto3" json:"max_change_rate,omitempty"`
}

// HistoricalInfo contains header and validator information for a given block.
type HistoricalInfo struct {
	Header *Header      `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Valset []*Validator `protobuf:"bytes,2,rep,name=valset,proto3" json:"valset,omitempty"`
}

// Header defines the structure of a block header.
type Header struct {
	Version            *Version `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	ChainID            string   `protobuf:"bytes,2,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Height             int64    `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Time               int64    `protobuf:"varint,4,opt,name=time,proto3" json:"time,omitempty"`
	LastBlockID        *BlockID `protobuf:"bytes,5,opt,name=last_block_id,json=lastBlockId,proto3" json:"last_block_id,omitempty"`
	LastCommitHash     []byte   `protobuf:"bytes,6,opt,name=last_commit_hash,json=lastCommitHash,proto3" json:"last_commit_hash,omitempty"`
	DataHash           []byte   `protobuf:"bytes,7,opt,name=data_hash,json=dataHash,proto3" json:"data_hash,omitempty"`
	ValidatorsHash     []byte   `protobuf:"bytes,8,opt,name=validators_hash,json=validatorsHash,proto3" json:"validators_hash,omitempty"`
	NextValidatorsHash []byte   `protobuf:"bytes,9,opt,name=next_validators_hash,json=nextValidatorsHash,proto3" json:"next_validators_hash,omitempty"`
	ConsensusHash      []byte   `protobuf:"bytes,10,opt,name=consensus_hash,json=consensusHash,proto3" json:"consensus_hash,omitempty"`
	AppHash            []byte   `protobuf:"bytes,11,opt,name=app_hash,json=appHash,proto3" json:"app_hash,omitempty"`
	LastResultsHash    []byte   `protobuf:"bytes,12,opt,name=last_results_hash,json=lastResultsHash,proto3" json:"last_results_hash,omitempty"`
	EvidenceHash       []byte   `protobuf:"bytes,13,opt,name=evidence_hash,json=evidenceHash,proto3" json:"evidence_hash,omitempty"`
	ProposerAddress    []byte   `protobuf:"bytes,14,opt,name=proposer_address,json=proposerAddress,proto3" json:"proposer_address,omitempty"`
}

// Version captures the consensus rules for processing a block in the blockchain.
type Version struct {
	Block uint64 `protobuf:"varint,1,opt,name=block,proto3" json:"block,omitempty"`
	App   uint64 `protobuf:"varint,2,opt,name=app,proto3" json:"app,omitempty"`
}

// BlockID defines the unique identifier of a block.
type BlockID struct {
	Hash          []byte         `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	PartSetHeader *PartSetHeader `protobuf:"bytes,2,opt,name=part_set_header,json=partSetHeader,proto3" json:"part_set_header,omitempty"`
}

// PartSetHeader defines the structure of a block part set header.
type PartSetHeader struct {
	Total uint32 `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Hash  []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

// Params defines the parameters for the staking module.
type Params struct {
	UnbondingTime     int64  `protobuf:"varint,1,opt,name=unbonding_time,json=unbondingTime,proto3" json:"unbonding_time,omitempty"`
	MaxValidators     uint32 `protobuf:"varint,2,opt,name=max_validators,json=maxValidators,proto3" json:"max_validators,omitempty"`
	MaxEntries        uint32 `protobuf:"varint,3,opt,name=max_entries,json=maxEntries,proto3" json:"max_entries,omitempty"`
	HistoricalEntries uint32 `protobuf:"varint,4,opt,name=historical_entries,json=historicalEntries,proto3" json:"historical_entries,omitempty"`
	BondDenom         string `protobuf:"bytes,5,opt,name=bond_denom,json=bondDenom,proto3" json:"bond_denom,omitempty"`
}

// Pool is used for tracking bonded and not-bonded token supply of the bond denomination.
type Pool struct {
	NotBondedTokens string `protobuf:"bytes,1,opt,name=not_bonded_tokens,json=notBondedTokens,proto3" json:"not_bonded_tokens,omitempty"`
	BondedTokens    string `protobuf:"bytes,2,opt,name=bonded_tokens,json=bondedTokens,proto3" json:"bonded_tokens,omitempty"`
}

// RedelegationResponse is equivalent to a Redelegation except that its entries contain a balance in addition to shares which is more suitable for client responses.
type RedelegationResponse struct {
	Redelegation *Redelegation                `protobuf:"bytes,1,opt,name=redelegation,proto3" json:"redelegation,omitempty"`
	Entries      []*RedelegationEntryResponse `protobuf:"bytes,2,rep,name=entries,proto3" json:"entries,omitempty"`
}

// Redelegation contains the list of a particular delegator's redelegating bonds from a particular source validator to a particular destination validator.
type Redelegation struct {
	DelegatorAddress    string               `protobuf:"bytes,1,opt,name=delegator_address,json=delegatorAddress,proto3" json:"delegator_address,omitempty"`
	ValidatorSrcAddress string               `protobuf:"bytes,2,opt,name=validator_src_address,json=validatorSrcAddress,proto3" json:"validator_src_address,omitempty"`
	ValidatorDstAddress string               `protobuf:"bytes,3,opt,name=validator_dst_address,json=validatorDstAddress,proto3" json:"validator_dst_address,omitempty"`
	Entries             []*RedelegationEntry `protobuf:"bytes,4,rep,name=entries,proto3" json:"entries,omitempty"`
}

// RedelegationEntry defines a redelegation object with relevant metadata.
type RedelegationEntry struct {
	CreationHeight int64  `protobuf:"varint,1,opt,name=creation_height,json=creationHeight,proto3" json:"creation_height,omitempty"`
	CompletionTime int64  `protobuf:"varint,2,opt,name=completion_time,json=completionTime,proto3" json:"completion_time,omitempty"`
	InitialBalance string `protobuf:"bytes,3,opt,name=initial_balance,json=initialBalance,proto3" json:"initial_balance,omitempty"`
	SharesDst      string `protobuf:"bytes,4,opt,name=shares_dst,json=sharesDst,proto3" json:"shares_dst,omitempty"`
}

// RedelegationEntryResponse is equivalent to a RedelegationEntry except that it contains a balance in addition to shares which is more suitable for client responses.
type RedelegationEntryResponse struct {
	RedelegationEntry *RedelegationEntry `protobuf:"bytes,1,opt,name=redelegation_entry,json=redelegationEntry,proto3" json:"redelegation_entry,omitempty"`
	Balance           string             `protobuf:"bytes,4,opt,name=balance,proto3" json:"balance,omitempty"`
}
