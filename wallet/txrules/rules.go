// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Package txrules provides transaction rules that should be followed by
// transaction authors for wide mempool acceptance and quick mining.
package txrules

import (
	"errors"
	"github.com/Qitmeer/qitmeer/core/serialization"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
)

// DefaultRelayFeePerKb is the default minimum relay fee policy for a mempool.
var DefaultRelayFeePerKb = types.Amount{Value: 1e3, Id: types.MEERID}

// GetDustThreshold is used to define the amount below which output will be
// determined as dust. Threshold is determined as 3 times the relay fee.
func GetDustThreshold(scriptSize int, relayFeePerKb types.Amount) types.Amount {
	// Calculate the total (estimated) cost to the network.  This is
	// calculated using the serialize size of the output plus the serial
	// size of a transaction input which redeems it.  The output is assumed
	// to be compressed P2PKH as this is the most common script type.  Use
	// the average size of a compressed P2PKH redeem input (148) rather than
	// the largest possible (txsizes.RedeemP2PKHInputSize).
	totalSize := 8 + serialization.VarIntSerializeSize(uint64(scriptSize)) +
		scriptSize + 148

	byteFee := relayFeePerKb.MulF64(1.0/1000)
	replayFee := byteFee.MulF64(float64(totalSize))
	return *replayFee.MulF64(3.0)
}

// IsDustAmount determines whether a transaction output value and script length would
// cause the output to be considered dust.  Transactions with dust outputs are
// not standard and are rejected by mempools with default policies.
func IsDustAmount(amount types.Amount, scriptSize int, relayFeePerKb types.Amount) bool {
	return amount.Value < GetDustThreshold(scriptSize, relayFeePerKb).Value
}

// IsDustOutput determines whether a transaction output is considered dust.
// Transactions with dust outputs are not standard and are rejected by mempools
// with default policies.
func IsDustOutput(output *types.TxOutput, relayFeePerKb types.Amount) bool {
	// Unspendable outputs which solely carry data are not checked for dust.
	if txscript.GetScriptClass(0,output.PkScript) == txscript.NullDataTy {
		return false
	}

	// All other unspendable outputs are considered dust.
	if txscript.IsUnspendable(output.PkScript) {
		return true
	}

	return IsDustAmount(types.Amount(output.Amount), len(output.PkScript),
		relayFeePerKb)
}

// Transaction rule violations
var (
	ErrAmountNegative   = errors.New("transaction output amount is negative")
	ErrAmountExceedsMax = errors.New("transaction output amount exceeds maximum value")
	ErrOutputIsDust     = errors.New("transaction output is dust")
)
const (
	// SatoshiPerBitcoin is the number of satoshi in one bitcoin (1 BTC).
	SatoshiPerBitcoin = 1e8

	// MaxSatoshi is the maximum transaction amount allowed in satoshi.
	MaxSatoshi = 21e6 * SatoshiPerBitcoin
)


// CheckOutput performs simple consensus and policy tests on a transaction
// output.
func CheckOutput(output *types.TxOutput, relayFeePerKb types.Amount) error {
	if output.Amount.Value < 0 {
		return ErrAmountNegative
	}
	if output.Amount.Value > MaxSatoshi {
		return ErrAmountExceedsMax
	}
	if IsDustOutput(output, relayFeePerKb) {
		return ErrOutputIsDust
	}
	return nil
}

// FeeForSerializeSize calculates the required fee for a transaction of some
// arbitrary size given a mempool's relay fee policy.
func FeeForSerializeSize(relayFeePerKb types.Amount, txSerializeSize int) types.Amount {
	fee := relayFeePerKb.MulF64(float64(txSerializeSize)/1000)

	if fee.Value == 0 && relayFeePerKb.Value > 0 {
		fee = &relayFeePerKb
	}

	if fee.Value < 0 || fee.Value > MaxSatoshi {
		fee.Value = MaxSatoshi
	}

	return *fee
}
