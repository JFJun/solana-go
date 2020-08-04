package transaction

/*
func：
author： flynn
date: 2020-08-03
fork: https://github.com/solana-labs/solana-web3.js/src/transaction.js
*/
import "errors"

type ITransactionInstruction interface {
	GetKeys() []*AccountMeta
	SetKeys(keys []*AccountMeta) error
	GetProgramId() string
	SetProgramId(programId string) error
	GetData() []byte
	SetData(data []byte) error
}
type TransactionInstruction struct {
	keys      []*AccountMeta
	programId string
	data      []byte
}

func (ti *TransactionInstruction) GetKeys() []*AccountMeta {
	return ti.keys
}

func (ti *TransactionInstruction) SetKeys(keys []*AccountMeta) error {
	if keys == nil {
		return errors.New("keys is null")
	}
	ti.keys = keys
	return nil
}
func (ti *TransactionInstruction) GetProgramId() string {
	return ti.programId
}
func (ti *TransactionInstruction) SetProgramId(programId string) error {
	if programId == "" {
		return errors.New("program id is null")
	}
	ti.programId = programId
	return nil
}
func (ti *TransactionInstruction) GetData() []byte {
	return ti.data
}
func (ti *TransactionInstruction) SetData(data []byte) error {
	if data == nil {
		return errors.New("data is null")
	}
	ti.data = data
	return nil
}
