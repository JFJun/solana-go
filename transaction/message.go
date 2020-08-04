package transaction

/*
func：
author： flynn
date: 2020-08-03
fork: https://github.com/solana-labs/solana-web3.js/src/message.js
*/
import (
	"github.com/btcsuite/btcutil/base58"
)

/*
最大打包限制
*/
const PACK_DATA_SIZE = 1280 - 40 - 8

/*
用于构建交易数据
*/
type Message struct {
	Header          *MessageHeader
	AccountKeys     []string
	RecentBlockHash string
	Instructions    []*CompiledInstruction
}

type CompiledInstruction struct {
	ProgramIdIndex int
	Accounts       []int
	Data           string
}
type MessageHeader struct {
	NumRequiredSignatures       int
	NumReadonlySignedAccounts   int
	NumReadonlyUnsignedAccounts int
}

func (message *Message) Serialize() []byte {
	numKey := len(message.AccountKeys)
	keyCount := encodeLength(numKey)
	var instructionBuffer []byte
	for _, m := range message.Instructions {
		accounts, programIdIndex := m.Accounts, m.ProgramIdIndex
		data := base58.Decode(m.Data)
		keyIndicesCount := encodeLength(len(accounts))

		dataCount := encodeLength(len(data))
		var accs []byte
		for _, a := range accounts {
			accs = append(accs, byte(a))
		}
		instructionBuffer = append(instructionBuffer, byte(programIdIndex))
		instructionBuffer = append(instructionBuffer, keyIndicesCount...)
		instructionBuffer = append(instructionBuffer, accs...)
		instructionBuffer = append(instructionBuffer, dataCount...)
		instructionBuffer = append(instructionBuffer, data...)
	}
	instructionCount := encodeLength(len(message.Instructions))

	instructionBuffer = append(instructionCount, instructionBuffer...)

	var signData []byte
	signData = append(signData, byte(message.Header.NumRequiredSignatures))
	signData = append(signData, byte(message.Header.NumReadonlySignedAccounts))
	signData = append(signData, byte(message.Header.NumReadonlyUnsignedAccounts))
	signData = append(signData, keyCount...)
	for _, key := range message.AccountKeys {
		signData = append(signData, base58.Decode(key)...)
	}
	signData = append(signData, base58.Decode(message.RecentBlockHash)...)
	signData = append(signData, instructionBuffer...)
	return signData
}

func encodeLength(num int) []byte {

	var (
		data []byte
		n    = num
	)
	for true {
		elem := n & 0x7f
		n >>= 7
		if n == 0 {
			data = append(data, byte(elem))
			break
		} else {
			elem = elem | 0x80
			data = append(data, byte(elem))
		}
	}
	return data
}
