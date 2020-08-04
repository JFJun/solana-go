package test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/JFJun/solana-go/account"
	"github.com/JFJun/solana-go/rpc"
	"github.com/JFJun/solana-go/transaction"
	"github.com/btcsuite/btcutil/base58"
	"math/big"
	"testing"
)

func Test_TransferTest(t *testing.T) {
	a1 := []byte{
		131, 158, 144, 122, 59, 1, 90, 107, 206, 5, 55,
		58, 64, 222, 94, 76, 173, 0, 9, 240, 27, 122,
		37, 146, 137, 94, 111, 197, 158, 179, 28, 222, 155,
		24, 28, 215, 82, 94, 70, 57, 195, 24, 207, 42,
		22, 93, 107, 124, 67, 37, 251, 95, 152, 92, 241,
		141, 135, 143, 21, 2, 155, 184, 49, 76,
	}
	a2 := []byte{190, 250, 118, 57, 122, 206, 17, 161, 253, 9, 177,
		162, 79, 20, 93, 163, 121, 39, 77, 196, 160, 227,
		126, 135, 49, 231, 170, 6, 55, 16, 217, 153, 105,
		153, 158, 74, 251, 70, 251, 193, 155, 253, 45, 156,
		122, 77, 2, 163, 227, 135, 126, 58, 208, 50, 75,
		77, 172, 4, 100, 44, 36, 4, 21, 20}
	account1 := account.NewAccountBySecret(a1[:32])
	account2 := account.NewAccountBySecret(a2[:32])
	tp := transaction.TransferParams{
		From:   account1.ToBase58(),
		To:     account2.ToBase58(),
		Amount: big.NewInt(100000000),
	}
	transfer, err := transaction.NewTransfer(tp)
	if err != nil {
		panic(err)
	}
	tp2 := transaction.TransferParams{
		From:   account2.ToBase58(),
		To:     account1.ToBase58(),
		Amount: big.NewInt(123),
	}
	transfer2, err := transaction.NewTransfer(tp2)
	if err != nil {
		panic(err)
	}
	recentBlockHash := account1.ToBase58()
	tx := transaction.NewTransaction(recentBlockHash)
	tx.SetInstructions(transfer)
	tx.SetInstructions(transfer2)
	var accounts []*account.Account
	accounts = append(accounts, account1)
	accounts = append(accounts, account2)
	err = tx.Sign(accounts)
	if err != nil {
		panic(err)
	}
	fmt.Println(tx.Signatures[0].Signature)
	fmt.Println(hex.EncodeToString(tx.Signatures[0].PublicKey))
	fmt.Println(tx.Signatures[1].Signature)
	fmt.Println(hex.EncodeToString(tx.Signatures[1].PublicKey))
}

func Test_Transfer(t *testing.T) {
	url := "http://sol.rylink.io:28899"
	client := rpc.New(url, "", "")
	from := "9SvsEyncSPjZaqjEsGjfvgaQowxq1BTNTJo6imGxseyx"
	to := "BHUNqtk5Vv6vfQTxpPjqWo2v8GPZJbqBonCaqhhK1Hub"
	// 构建交易参数
	tp := transaction.TransferParams{
		From:   from,
		To:     to,
		Amount: big.NewInt(123),
	}
	transfer, err := transaction.NewTransfer(tp)
	if err != nil {
		t.Fatal(err)
	}
	//创建交易
	data, err := client.SendRequest("getRecentBlockhash", nil)
	if err != nil {
		t.Fatal(err)
	}
	var recentBlockHashResult map[string]interface{}
	err = json.Unmarshal(data, &recentBlockHashResult)
	if err != nil {
		panic(err)
	}
	rcbh := recentBlockHashResult["value"].(map[string]interface{})
	recentBlockHash := rcbh["blockhash"].(string)

	//构建 tx
	tx := transaction.NewTransaction(recentBlockHash)
	tx.SetInstructions(transfer)

	//todo 根据私钥找公钥
	privateKey := ""
	p, _ := hex.DecodeString(privateKey)
	//priv:=ed25519.NewKeyFromSeed(p)
	acc := account.NewAccountBySecret(p)
	var accounts []*account.Account
	accounts = append(accounts, acc)

	//签名
	err = tx.Sign(accounts)
	if err != nil {
		panic(err)
	}
	wireTx, err := tx.Serialize()
	if err != nil {
		panic(err)
	}
	b58Tx := base58.Encode(wireTx)

	//发送交易
	sendData, err := client.SendRequest("sendTransaction", []interface{}{b58Tx})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(sendData))
}
