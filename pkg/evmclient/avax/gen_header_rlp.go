package avax

import (
	"io"

	"github.com/ethereum/go-ethereum/rlp"
)

//nolint:funlen,gocognit,gocyclo,cyclop,stylecheck
func (obj *Header) EncodeRLP(_w io.Writer) error {
	w := rlp.NewEncoderBuffer(_w)
	_tmp0 := w.List()
	w.WriteBytes(obj.ParentHash[:])
	w.WriteBytes(obj.UncleHash[:])
	w.WriteBytes(obj.Coinbase[:])
	w.WriteBytes(obj.Root[:])
	w.WriteBytes(obj.TxHash[:])
	w.WriteBytes(obj.ReceiptHash[:])
	w.WriteBytes(obj.Bloom[:])
	if obj.Difficulty == nil {
		_, _ = w.Write(rlp.EmptyString)
	} else {
		if obj.Difficulty.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.Difficulty)
	}
	if obj.Number == nil {
		_, _ = w.Write(rlp.EmptyString)
	} else {
		if obj.Number.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.Number)
	}
	w.WriteUint64(obj.GasLimit)
	w.WriteUint64(obj.GasUsed)
	w.WriteUint64(obj.Time)
	w.WriteBytes(obj.Extra)
	w.WriteBytes(obj.MixDigest[:])
	w.WriteBytes(obj.Nonce[:])
	w.WriteBytes(obj.ExtDataHash[:])
	_tmp1 := obj.BaseFee != nil
	_tmp2 := obj.ExtDataGasUsed != nil
	_tmp3 := obj.BlockGasCost != nil
	_tmp4 := obj.ExcessDataGas != nil
	if _tmp1 || _tmp2 || _tmp3 || _tmp4 {
		if obj.BaseFee == nil {
			_, _ = w.Write(rlp.EmptyString)
		} else {
			if obj.BaseFee.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.BaseFee)
		}
	}
	if _tmp2 || _tmp3 || _tmp4 {
		if obj.ExtDataGasUsed == nil {
			_, _ = w.Write(rlp.EmptyString)
		} else {
			if obj.ExtDataGasUsed.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.ExtDataGasUsed)
		}
	}
	if _tmp3 || _tmp4 {
		if obj.BlockGasCost == nil {
			_, _ = w.Write(rlp.EmptyString)
		} else {
			if obj.BlockGasCost.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.BlockGasCost)
		}
	}
	if _tmp4 {
		if obj.ExcessDataGas == nil {
			_, _ = w.Write(rlp.EmptyString)
		} else {
			if obj.ExcessDataGas.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.ExcessDataGas)
		}
	}
	w.ListEnd(_tmp0)

	return w.Flush()
}
