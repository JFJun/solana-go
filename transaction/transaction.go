package transaction

/*
func：
author： flynn
date: 2020-08-03
fork: https://github.com/solana-labs/solana-web3.js/src/transaction.js
*/
import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/JFJun/solana-go/account"
	"github.com/btcsuite/btcutil/base58"
)

type AccountMeta struct {
	PubKey      []byte //
	IsSigner    bool
	IsWriteable bool
}

type SignaturePubkeyPair struct {
	Signature []byte
	PublicKey []byte
}
type TransactionCtorFields struct {
	RecentBlockHash string
	NonceInfo       NonceInformation
	Signatures      []*SignaturePubkeyPair
}

type NonceInformation struct {
	Nonce            string //blockhash
	NonceInstruction ITransactionInstruction
}

type Transaction struct {
	Signatures      []*SignaturePubkeyPair
	Instructions    []ITransactionInstruction
	RecentBlockHash string
	NonceInfo       *NonceInformation
}

func NewTransaction(recentBlockHash string) *Transaction {
	tx := new(Transaction)
	tx.RecentBlockHash = recentBlockHash
	return tx
}
func (tx *Transaction) SetInstructions(itf ITransactionInstruction) {
	tx.Instructions = append(tx.Instructions, itf)
}

func (tx *Transaction) Sign(accounts []*account.Account) error {
	if len(accounts) == 0 {
		return errors.New("do not set account")
	}
	var signatures []*SignaturePubkeyPair
	for _, acc := range accounts {
		spp := new(SignaturePubkeyPair)
		spp.PublicKey = acc.PublicKey
		signatures = append(signatures, spp)
	}
	tx.Signatures = signatures
	signData, err := tx.serializeMessage()

	if err != nil {
		return fmt.Errorf("serial sign message error,Err==%v", err)
	}
	for idx, acc := range accounts {
		priv := ed25519.NewKeyFromSeed(acc.SecretKey)
		sig := ed25519.Sign(priv, signData)
		if len(sig) != 64 {
			return errors.New("sign data length is not equal 64")
		}
		tx.Signatures[idx].Signature = sig
	}
	return nil
}

func (tx *Transaction) serializeMessage() ([]byte, error) {
	message, err := tx.CompileMessage()
	if err != nil {
		return nil, err
	}
	return message.Serialize(), nil
}
func (tx *Transaction) CompileMessage() (*Message, error) {
	if tx.RecentBlockHash == "" {
		return nil, errors.New("tx recent block hash is null")
	}
	if len(tx.Instructions) < 1 {
		return nil, errors.New("tx instruction length is less than 1")
	}

	if tx.NonceInfo != nil && tx.Instructions[0] != tx.NonceInfo.NonceInstruction {
		tx.RecentBlockHash = tx.NonceInfo.Nonce
		var ins []ITransactionInstruction
		ins = append(ins, tx.NonceInfo.NonceInstruction)
		ins = append(ins, tx.Instructions...)
		tx.Instructions = nil
		tx.Instructions = ins
	}
	var (
		numReadonlySignedAccounts, numReadonlyUnsignedAccounts int
		programIds                                             []string
		accountMetas                                           []*AccountMeta
	)
	for _, in := range tx.Instructions {
		accountMetas = append(accountMetas, in.GetKeys()...)
		if len(programIds) == 0 {
			programIds = append(programIds, in.GetProgramId())
		} else {
			if !isIncludes(in.GetProgramId(), programIds) {
				programIds = append(programIds, in.GetProgramId())
			}
		}
	}
	for _, p := range programIds {
		accountMetas = append(accountMetas, &AccountMeta{
			base58.Decode(p),
			false,
			false,
		})
	}
	// 排序accountMeta
	accountMetas = sortAccountMetas(accountMetas)
	// 删除重复的accountMeta
	var uniqueAccountMeta []*AccountMeta
	if len(accountMetas) <= 1 {
		uniqueAccountMeta = accountMetas
	} else {
		for _, accM := range accountMetas {
			if len(uniqueAccountMeta) == 0 {
				uniqueAccountMeta = append(uniqueAccountMeta, accM)
				continue
			}
			isHave := false
			for _, uAccM := range uniqueAccountMeta {
				if bytes.Compare(uAccM.PubKey, accM.PubKey) == 0 {
					isHave = true
					break
				}
			}
			if !isHave {
				uniqueAccountMeta = append(uniqueAccountMeta, accM)
			}

		}

		////冒泡排序
		//for i:=0;i<len(accountMetas)-1;i++{
		//	pubkeyString:=accountMetas[i].PubKey
		//	isHave :=false
		//	for j:=i+1;j<len(accountMetas);j++{
		//		if bytes.Compare(pubkeyString,accountMetas[j].PubKey)==0 {
		//			isHave = true
		//			break
		//		}
		//	}
		//	if !isHave {
		//		uniqueAccountMeta = append(uniqueAccountMeta,accountMetas[i])
		//	}
		//}
	}

	if len(tx.Signatures) > 0 {
		for _, s := range tx.Signatures {
			sigPubkeyString := s.PublicKey
			isHave := false
			for _, u := range uniqueAccountMeta {
				if bytes.Compare(sigPubkeyString, u.PubKey) == 0 {
					isHave = true
					u.IsSigner = true
					break
				}
			}
			if !isHave {
				uniqueAccountMeta = append([]*AccountMeta{
					{
						PubKey:      sigPubkeyString,
						IsSigner:    true,
						IsWriteable: true,
					},
				}, uniqueAccountMeta...)
			}
		}
	}
	var signedKeys, unsignedKeys []string
	for _, u := range uniqueAccountMeta {
		if u.IsSigner {
			// Promote the first signer to writable as it is the fee payer
			length := len(signedKeys)
			signedKeys = append(signedKeys, base58.Encode(u.PubKey))
			if length > 0 && !u.IsWriteable {
				numReadonlySignedAccounts++
			}
		} else {
			unsignedKeys = append(unsignedKeys, base58.Encode(u.PubKey))
			if !u.IsWriteable {
				numReadonlyUnsignedAccounts++
			}
		}
	}

	// Initialize signature array, if needed
	if len(tx.Signatures) == 0 {
		if len(signedKeys) == 0 {
			tx.Signatures = nil
		}
		var signatures []*SignaturePubkeyPair
		for _, s := range signedKeys {
			signatures = append(signatures, &SignaturePubkeyPair{
				nil,
				base58.Decode(s),
			})
		}
		tx.Signatures = signatures
	}
	var accountKeys []string
	accountKeys = append(accountKeys, signedKeys...)
	accountKeys = append(accountKeys, unsignedKeys...)
	var instructions []*CompiledInstruction
	for _, ins := range tx.Instructions {
		data := base58.Encode(ins.GetData())
		program := ins.GetProgramId()
		var (
			programIdIndex int
			accounts       []int
		)
		//for i,a:=range accountKeys{
		//	if a==program {
		//		programIdIndex = i
		//	}
		//	for _,k:=range ins.GetKeys(){
		//
		//		if base58.Encode(k.PubKey)==a {
		//			accounts = append(accounts,i)
		//		}
		//	}
		//}
		for _, k := range ins.GetKeys() {
			for i, a := range accountKeys {
				if a == program {
					programIdIndex = i
				}
				if base58.Encode(k.PubKey) == a {
					accounts = append(accounts, i)
				}
			}
		}

		instructions = append(instructions, &CompiledInstruction{
			ProgramIdIndex: programIdIndex,
			Accounts:       accounts,
			Data:           data,
		})
	}

	messageHeader := new(MessageHeader)
	messageHeader.NumRequiredSignatures = len(tx.Signatures)
	messageHeader.NumReadonlySignedAccounts = numReadonlySignedAccounts
	messageHeader.NumReadonlyUnsignedAccounts = numReadonlyUnsignedAccounts

	message := new(Message)
	message.Header = messageHeader
	message.AccountKeys = accountKeys
	message.RecentBlockHash = tx.RecentBlockHash
	message.Instructions = instructions
	return message, nil
}
func isIncludes(p string, ProgramIds []string) bool {
	for _, pp := range ProgramIds {
		if p == pp {
			return true
		}
	}
	return false
}

func sortAccountMetas(accountMetas []*AccountMeta) []*AccountMeta {
	if len(accountMetas) <= 1 {
		return accountMetas
	}
	cmp := func(a, b *AccountMeta) bool {
		//先判断签名
		if !a.IsSigner && b.IsSigner {
			return true
		}
		//在判断write
		if !a.IsWriteable && b.IsWriteable {
			return true
		}
		return false
	}
	for i := 0; i < len(accountMetas)-1; i++ {
		for j := i + 1; j < len(accountMetas); j++ {
			if cmp(accountMetas[i], accountMetas[j]) {
				accountMetas[i], accountMetas[j] = accountMetas[j], accountMetas[i]
			}
		}
	}
	return accountMetas
}

func (tx *Transaction) Serialize() ([]byte, error) {
	if tx.Signatures == nil || len(tx.Signatures) == 0 {
		return nil, errors.New("transaction has not been signedl")
	}
	signData, err := tx.serializeMessage()
	if err != nil {
		return nil, fmt.Errorf("tx serialize message error,err=%v", err)
	}
	signatureCount := encodeLength(len(tx.Signatures))
	var wireTransaction []byte
	wireTransaction = append(wireTransaction, signatureCount...)
	for _, sig := range tx.Signatures {
		wireTransaction = append(wireTransaction, sig.Signature...)
	}
	wireTransaction = append(wireTransaction, signData...)
	if len(wireTransaction) > PACK_DATA_SIZE {
		return nil, fmt.Errorf("tx is too large,tx length=[%d] big than PACK_DATA_SIZE=[%d]", len(wireTransaction), PACK_DATA_SIZE)
	}
	return wireTransaction, nil
}
