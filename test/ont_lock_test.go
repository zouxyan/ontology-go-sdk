package test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	mcutils "github.com/ontio/multi-chain/native/service/utils"
	"github.com/ontio/ontology-crypto/keypair"
	. "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	olp "github.com/ontio/ontology/smartcontract/service/native/cross_chain/ont_lock_proxy"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"testing"
	"time"
)

var (
	testOntSdk   *OntologySdk
	testWallet   *Wallet
	testPasswd   = []byte("passwordtest")
	testDefAcc   *Account
	testGasPrice = uint64(0)
	testGasLimit = uint64(20000)
	testNetUrl   = "http://172.168.3.78:20336"
	walletPath   = "./wallet/wallet.dat"
)

func Init() {
	testOntSdk = NewOntologySdk()
	testOntSdk.NewRpcClient().SetAddress(testNetUrl)

	var err error
	var wallet *Wallet
	if !common.FileExisted(walletPath) {
		wallet, err = testOntSdk.CreateWallet(walletPath)
		if err != nil {
			fmt.Println("[CreateWallet] error:", err)
			return
		}
	} else {
		wallet, err = testOntSdk.OpenWallet(walletPath)
		if err != nil {
			fmt.Println("[CreateWallet] error:", err)
			return
		}
	}
	_, err = wallet.NewDefaultSettingAccount(testPasswd)
	if err != nil {
		fmt.Println("")
		return
	}
	//wallet.Save()
	testWallet, err = testOntSdk.OpenWallet(walletPath)
	if err != nil {
		fmt.Printf("account.Open error:%s\n", err)
		return
	}
	testDefAcc, err = testWallet.GetDefaultAccount(testPasswd)
	if err != nil {
		fmt.Printf("GetDefaultAccount error:%s\n", err)
		return
	}

	return

}

func Test_Ont_Transfer(t *testing.T) {
	Init()
	toAddr, _ := common.AddressFromBase58("AQf4Mzu1YJrhz9f3aRkkwSm9n3qhXGSh4p")
	txHash, err := testOntSdk.Native.Ont.Transfer(testGasPrice, testGasLimit, nil, testDefAcc, toAddr, 10000)

	if err != nil {
		t.Errorf("NewTransferTransaction error:%s", err)
		return
	}
	testOntSdk.WaitForGenerateBlock(30*time.Second, 1)
	evts, err := testOntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		t.Errorf("GetSmartContractEvent error:%s", err)
		return
	}
	fmt.Printf("TxHash:%s\n", txHash.ToHexString())
	fmt.Printf("State:%d\n", evts.State)
	fmt.Printf("GasConsume:%d\n", evts.GasConsumed)
	for _, notify := range evts.Notify {
		fmt.Printf("ContractAddress:%s\n", notify.ContractAddress)
		fmt.Printf("States:%+v\n", notify.States)
	}
	secondAccount, err := testWallet.GetAccountByIndex(2, testPasswd)
	if err != nil {
		t.Errorf("GetSecond account error:%s", err)
	}
	fmt.Printf("second adccount: %s\n", secondAccount.Address.ToBase58())
}

func Test_GetBalanceOf_Wallet(t *testing.T) {
	Init()
	acctCount := testWallet.GetAccountCount()
	for i := 1; i <= acctCount; i++ {
		acctI, err := testWallet.GetAccountByIndex(i, testPasswd)
		if err != nil {
			t.Errorf("GetAccountByIndex error:%s\n", err)
			return
		}
		res, err := testOntSdk.Native.Ont.BalanceOf(acctI.Address)
		if err != nil {
			t.Errorf("get balance error: wallet index = %d, balance of %s, err=%s\n", i, hex.EncodeToString(acctI.Address[:]), err)
			return
		}
		fmt.Printf("walelt index = %d, ont balance of %s = %d\n", i, hex.EncodeToString(acctI.Address[:]), res)
		res, err = testOntSdk.Native.Ong.BalanceOf(acctI.Address)
		if err != nil {
			t.Errorf("get balance error: wallet index = %d, balance of %s, err=%s\n", i, hex.EncodeToString(acctI.Address[:]), err)
			return
		}
		fmt.Printf("walelt index = %d, ong balance of %s = %d\n", i, hex.EncodeToString(acctI.Address[:]), res)
	}
	Test_GetBalanceOf_OntLockProxyContract(t)
}

func Test_GetBalanceOf_OntLockProxyContract(t *testing.T) {
	Init()

	res, err := testOntSdk.Native.Ont.BalanceOf(utils.OntLockContractAddress)
	if err != nil {
		t.Errorf("get balance of ontlockContract err %s\n", err)
	}
	fmt.Printf("balance of ontLockProxyContract = %s = %d\n", hex.EncodeToString(utils.OntLockContractAddress[:]), res)
}

func TestOnt_Lock(t *testing.T) {
	Init()
	toAddressBytes, _ := hex.DecodeString("709c937270e1d5a490718a2b4a230186bdd06a02")
	//toAddressBytes, _ := hex.DecodeString("2186fe74983e0016359c7c1b9063448fc8813b87")
	txHash, err := testOntSdk.Native.OntLock.Lock(testGasPrice, testGasLimit, nil, utils.OntContractAddress, testDefAcc, 0, toAddressBytes, 2)
	if err != nil {
		t.Errorf("NewTransferTransaction error:%s", err)
		return
	}
	testOntSdk.WaitForGenerateBlock(30*time.Second, 1)
	evts, err := testOntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		t.Errorf("GetSmartContractEvent error:%s", err)
		return
	}
	fmt.Printf("TxHash:%s\n", txHash.ToHexString())
	fmt.Printf("State:%d\n", evts.State)
	fmt.Printf("GasConsume:%d\n", evts.GasConsumed)
	for _, notify := range evts.Notify {
		fmt.Printf("ContractAddress:%s\n", notify.ContractAddress)
		fmt.Printf("States:%+v\n", notify.States)
	}

}

func TestOnt_BindProxy(t *testing.T) {
	Init()
	testOntSdk := NewOntologySdk()
	testOntSdk.NewRpcClient().SetAddress(testNetUrl)
	pks, sgners := openWalletForBind()
	txHash, err := testOntSdk.Native.OntLock.BindProxyHash(testGasPrice, testGasLimit, 0, mcutils.OntLockProxyContractAddress[:], pks, sgners)
	if err != nil {
		t.Errorf("BindProxyHash error:%s", err)
		return
	}
	testOntSdk.WaitForGenerateBlock(30*time.Second, 1)
	evts, err := testOntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		t.Errorf("GetSmartContractEvent error:%s", err)
		return
	}
	fmt.Printf("TxHash:%s\n", txHash.ToHexString())
	fmt.Printf("State:%d\n", evts.State)
	fmt.Printf("GasConsume:%d\n", evts.GasConsumed)
	for _, notify := range evts.Notify {
		fmt.Printf("ContractAddress:%s\n", notify.ContractAddress)
		fmt.Printf("States:%+v\n", notify.States)
	}

}

func Test_GetBindProxy(t *testing.T) {
	Init()
	toChainId := 0
	bindProxy, err := getBindProxyHashFromStorage(uint64(toChainId))
	if err != nil && bindProxy != "" {
		t.Errorf("Cannot get bind asset hash, err:%s", err)
	}
	fmt.Printf("GetBindProxyHash(%d) = %s\n", toChainId, bindProxy)
}

func TestOnt_BindAsset(t *testing.T) {
	Init()
	testOntSdk := NewOntologySdk()
	testOntSdk.NewRpcClient().SetAddress(testNetUrl)
	pks, sgners := openWalletForBind()
	txHash, err := testOntSdk.Native.OntLock.BindAssetHash(testGasPrice, testGasLimit, ONT_CONTRACT_ADDRESS, 0, mcutils.OntContractAddress[:], pks, sgners)
	if err != nil {
		t.Errorf("BindAssetHash error:%s", err)
		return
	}
	testOntSdk.WaitForGenerateBlock(30*time.Second, 1)
	evts, err := testOntSdk.GetSmartContractEvent(txHash.ToHexString())
	if err != nil {
		t.Errorf("GetSmartContractEvent error:%s", err)
		return
	}
	fmt.Printf("TxHash:%s\n", txHash.ToHexString())
	fmt.Printf("State:%d\n", evts.State)
	fmt.Printf("GasConsume:%d\n", evts.GasConsumed)
	for _, notify := range evts.Notify {
		fmt.Printf("ContractAddress:%s\n", notify.ContractAddress)
		fmt.Printf("States:%+v\n", notify.States)
	}
}

func Test_GetBindAsset(t *testing.T) {
	Init()
	sourceAssetHash := ONT_CONTRACT_ADDRESS
	toChainId := 0
	bindProxy, err := getBindAssetHashFromStorage(sourceAssetHash, uint64(toChainId))
	if err != nil && bindProxy != "" {
		t.Errorf("Cannot get bind asset hash, err:%s", err)
	}
	fmt.Printf("GetBindAssetHash(%s, %d) = %s\n", hex.EncodeToString(sourceAssetHash[:]), toChainId, bindProxy)
}

func Test_GetSmartContractEvent(t *testing.T) {
	Init()
	hashStr := "3bbd1c40c7d31de0f3f4600fc70c3fdfbbd1d83f7861d0ccd264ac1da4710775"
	testOntSdk := NewOntologySdk()
	testOntSdk.NewRpcClient().SetAddress(testNetUrl)

	evts, err := testOntSdk.GetSmartContractEvent(hashStr)
	if err != nil {
		t.Errorf("GetSmartContractEvent error:%s", err)
		return
	}
	fmt.Printf("TxHash:%s\n", hashStr)
	fmt.Printf("State:%d\n", evts.State)
	fmt.Printf("GasConsume:%d\n", evts.GasConsumed)
	for _, notify := range evts.Notify {
		fmt.Printf("ContractAddress:%s\n", notify.ContractAddress)
		fmt.Printf("States:%+v\n", notify.States)
	}
}

func openWalletForBind() (pubKeys []keypair.PublicKey, singers []*Account) {
	testOntSdk1 := NewOntologySdk()
	accounts := make([]*Account, 0)
	pks := make([]keypair.PublicKey, 0)
	walletPaths := []string{
		"./wallet/peer1.dat",
		"./wallet/peer2.dat",
		"./wallet/peer3.dat",
		"./wallet/peer4.dat",
		"./wallet/peer5.dat",
		"./wallet/peer6.dat",
		"./wallet/peer7.dat",
	}
	for i, walletpath := range walletPaths {
		testWallet, err := testOntSdk1.OpenWallet(walletpath)
		if err != nil {
			fmt.Printf("account.Open index:%d, error:%s\n", i, err)
		}
		testDefAcc, err = testWallet.GetDefaultAccount(testPasswd)
		if err != nil {
			fmt.Printf("account.GetDefaultAccount index:%d, error:%s\n", i, err)
		}
		pks = append(pks, testDefAcc.PublicKey)
		accounts = append(accounts, testDefAcc)
		//fmt.Printf("pk index:%d,  is %v\n", i, pks[i])
		//fmt.Printf("accounts index:%d, is %v\n", i, accounts[i].Address.ToBase58())
	}
	return pks, accounts

}

func getBindProxyHashFromStorage(toChainId uint64) (string, error) {
	bs := make([]byte, 0)
	bs = append(bs, []byte(olp.BIND_PROXY_NAME)...)
	chainIdBytes, _ := utils.GetUint64Bytes(toChainId)
	bs = append(bs, chainIdBytes...)
	proxyStorage, _ := testOntSdk.GetStorage(ONTLOCK_CONTRACT_ADDRESS.ToHexString(), bs)
	ts, err := serialization.ReadVarBytes(bytes.NewBuffer(proxyStorage))
	if err != nil {
		return "", fmt.Errorf("readVarBytes error:%s", err)
	}
	return hex.EncodeToString(ts), nil
}

func getBindAssetHashFromStorage(assetHash common.Address, toChainId uint64) (string, error) {
	bs := make([]byte, 0)
	bs = append(bs, []byte(olp.BIND_ASSET_NAME)...)
	bs = append(bs, assetHash[:]...)
	chainIdBytes, _ := utils.GetUint64Bytes(toChainId)
	bs = append(bs, chainIdBytes...)
	assetStorage, _ := testOntSdk.GetStorage(ONTLOCK_CONTRACT_ADDRESS.ToHexString(), bs)
	ts, err := serialization.ReadVarBytes(bytes.NewBuffer(assetStorage))
	if err != nil {
		return "", fmt.Errorf("readVarBytes error:%s", err)
	}
	return hex.EncodeToString(ts), nil
}
