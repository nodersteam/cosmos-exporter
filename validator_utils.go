package main

import (
	"fmt"

	cometbftcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetConsAddr возвращает консенсусный адрес валидатора
func (v *types.Validator) GetConsAddr() ([]byte, error) {
	if v.ConsensusPubkey == nil {
		return nil, fmt.Errorf("validator consensus pubkey is nil")
	}

	var pubKey cryptotypes.PubKey
	if err := v.ConsensusPubkey.UnpackInterfaces(interfaceRegistry); err != nil {
		return nil, fmt.Errorf("failed to unpack consensus pubkey: %w", err)
	}

	pubKey, ok := v.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		// Пробуем получить ключ как cometbft/PubKeyBn254
		if pubKeyBn254, ok := v.ConsensusPubkey.GetCachedValue().(*cometbftcrypto.PubKeyBn254); ok {
			return pubKeyBn254.Address(), nil
		}
		return nil, fmt.Errorf("expecting cryptotypes.PubKey, got %T: invalid type", v.ConsensusPubkey.GetCachedValue())
	}

	return pubKey.Address(), nil
}
