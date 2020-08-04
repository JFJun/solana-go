package transaction

/*
func：
author： flynn
date: 2020-08-03
fork: https://github.com/solana-labs/solana-web3.js/src/system-program.js
*/
import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"math/big"
)

type TransferParams struct {
	From   string
	To     string
	Amount *big.Int
}

func (tp *TransferParams) GetFromPublicKey() []byte {
	return base58.Decode(tp.From)
}
func (tp *TransferParams) GetToPublicKey() []byte {
	return base58.Decode(tp.To)
}

func NewTransfer(transfer TransferParams) (ITransactionInstruction, error) {
	transferIndex := uint32(2) //https://github.com/solana-labs/solana-web3.js/src/system-program.js-->p511  version:v0.64.0
	lamports := transfer.Amount.Uint64()
	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)
	err := binary.Write(buf1, binary.LittleEndian, transferIndex)
	if err != nil {
		return nil, fmt.Errorf("encode transfer index error,Err=%v", err)
	}
	err = binary.Write(buf2, binary.LittleEndian, lamports)
	if err != nil {
		return nil, fmt.Errorf("encode lamports error,Err=%v", err)
	}
	var data []byte
	data = append(data, buf1.Bytes()...)
	data = append(data, buf2.Bytes()...)
	if len(data) != 12 {
		return nil, errors.New("transfer data length is not equal 12")
	}
	ti := new(TransactionInstruction)
	ti.SetKeys([]*AccountMeta{
		{transfer.GetFromPublicKey(), true, true},
		{transfer.GetToPublicKey(), false, true},
	})
	err = ti.SetProgramId("11111111111111111111111111111111")
	if err != nil {
		return nil, err
	}
	err = ti.SetData(data)
	if err != nil {
		return nil, err
	}
	return ti, nil
}
