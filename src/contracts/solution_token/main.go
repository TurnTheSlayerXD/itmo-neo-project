package solution_token

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/contract"
	"github.com/nspcc-dev/neo-go/pkg/interop/iterator"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/crypto"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/gas"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/management"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
	"github.com/nspcc-dev/neo-go/pkg/interop/util"
)

// Prefixes used for contract data storage.
const (
	balancePrefix = "b"
	accountPrefix = "a"
	tokenPrefix   = "t"

	ownerKey       = 'o'
	totalSupplyKey = 's'
)

var contractTask_hash = interop.Hash160{0xcc, 0x6e, 0xe4, 0x23, 0x67, 0xd9, 0x3f, 0xde, 0x8d, 0x21, 0xf2, 0x30, 0x38, 0x4f, 0xca, 0x54, 0x54, 0x3f, 0x8f, 0x6a}



type NFTSolution struct {
	ID            []byte
	TaskId        []byte
	TaskAssesment int
	Owner         interop.Hash160
	SrcCode       string
	Description   string
}

func _deploy(data any, isUpdate bool) {
	if isUpdate {
		return
	}
	util.AssertMsg(contractTask_hash != nil, "_deploy contractTask_hash == nil")

	ctx := storage.GetContext()
	storage.Put(ctx, ownerKey, runtime.GetCallingScriptHash())
	storage.Put(ctx, totalSupplyKey, 0)
}

// Symbol returns token symbol, it's NICENAMES.
func Symbol() string {
	return "SOLUTION"
}

// Decimals returns token decimals, this NFT is non-divisible, so it's 0.
func Decimals() int {
	return 0
}

// TotalSupply is a contract method that returns the number of tokens minted.
func TotalSupply() int {
	return storage.Get(storage.GetReadOnlyContext(), totalSupplyKey).(int)
}

// BalanceOf returns the number of tokens owned by the specified address.
func BalanceOf(holder interop.Hash160) int {
	if len(holder) != 20 {
		panic("bad owner address")
	}
	ctx := storage.GetReadOnlyContext()
	return getBalanceOf(ctx, mkBalanceKey(holder))
}

// OwnerOf returns the owner of the specified token.
func OwnerOf(token []byte) interop.Hash160 {
	ctx := storage.GetReadOnlyContext()
	return getNFT(ctx, token).Owner
}

// Properties returns properties of the given NFT.
func Properties(token []byte) map[string]string {
	ctx := storage.GetReadOnlyContext()
	nft := getNFT(ctx, token)

	result := map[string]string{
		"id":            string(nft.ID),
		"ownerid":       ownerAddress(nft.Owner),
		"taskid":        string(nft.TaskId),
		"taskassesment": std.Itoa10(nft.TaskAssesment),
		"srccode":       nft.SrcCode,
		"description":   nft.Description,
	}
	return result
}

// Tokens returns an iterator that contains all the tokens minted by the contract.
func Tokens() iterator.Iterator {
	ctx := storage.GetReadOnlyContext()
	key := []byte(tokenPrefix)
	iter := storage.Find(ctx, key, storage.RemovePrefix|storage.KeysOnly)
	return iter
}

func TokensList() []string {
	ctx := storage.GetReadOnlyContext()
	key := []byte(tokenPrefix)
	iter := storage.Find(ctx, key, storage.RemovePrefix|storage.KeysOnly)
	keys := []string{}
	for iterator.Next(iter) {
		k := iterator.Value(iter)
		keys = append(keys, k.(string))
	}
	return keys
}

// TokensOf returns an iterator with all tokens held by the specified address.
func TokensOf(holder interop.Hash160) iterator.Iterator {
	if len(holder) != 20 {
		panic("bad owner address")
	}
	ctx := storage.GetReadOnlyContext()
	key := mkAccountPrefix(holder)
	iter := storage.Find(ctx, key, storage.ValuesOnly)
	return iter
}

func TokensOfList(holder interop.Hash160) [][]byte {
	if len(holder) != 20 {
		panic("bad owner address")
	}
	ctx := storage.GetReadOnlyContext()
	key := mkAccountPrefix(holder)
	res := [][]byte{}
	iter := storage.Find(ctx, key, storage.ValuesOnly)
	for iterator.Next(iter) {
		res = append(res, iterator.Value(iter).([]byte))
	}
	return res
}

func Transfer(to interop.Hash160, token []byte, data any) bool {
	if len(to) != 20 {
		panic("invalid 'to' address")
	}

	ctx := storage.GetContext()
	nft := getNFT(ctx, token)
	from := nft.Owner

	if !runtime.CheckWitness(from) {
		return false
	}

	if !from.Equals(to) {
		nft.Owner = to
		setNFT(ctx, token, nft)

		addToBalance(ctx, from, -1)
		removeToken(ctx, from, token)
		addToBalance(ctx, to, 1)
		addToken(ctx, to, token)
	}

	postTransfer(from, to, token, data)

	return true
}

func getNFT(ctx storage.Context, token []byte) NFTSolution {
	key := mkTokenKey(token)
	val := storage.Get(ctx, key)
	if val == nil {
		panic("no token found")
	}

	serializedNFT := val.([]byte)
	deserializedNFT := std.Deserialize(serializedNFT)
	return deserializedNFT.(NFTSolution)
}

func nftExists(ctx storage.Context, token []byte) bool {
	key := mkTokenKey(token)
	return storage.Get(ctx, key) != nil
}

func setNFT(ctx storage.Context, token []byte, item NFTSolution) {
	key := mkTokenKey(token)
	val := std.Serialize(item)
	storage.Put(ctx, key, val)
}

func postTransfer(from interop.Hash160, to interop.Hash160, token []byte, data any) {
	runtime.Notify("Transfer", from, to, 1, token)
	if management.GetContract(to) != nil {
		contract.Call(to, "onNEP11Payment", contract.All, from, 1, token, data)
	}
}

func OnNEP17Payment(from interop.Hash160, amount int, data any) {
	defer func() {
		if r := recover(); r != nil {
			runtime.Log(r.(string))
			util.Abort()
		}
	}()

	callingHash := runtime.GetCallingScriptHash()
	if !callingHash.Equals(gas.Hash) {
		panic("only GAS is accepted")
	}
	
	dict := std.JSONDeserialize(data.([]byte)).(map[string]any)
	solutionData := &struct {
		TaskId        []byte
		SrcCode       string
		TaskAssesment int
		Description   string
	}{
		TaskId:        std.Base64Decode(dict["taskid"].([]byte)),
		SrcCode:       dict["srccode"].(string),
		Description:   dict["description"].(string),
		TaskAssesment: dict["taskassesment"].(int),
	}

	price := 1_0000_0000

	if amount < price {
		panic("insufficient GAS for minting NFT")
	} else if amount > price {
		gas.Transfer(runtime.GetExecutingScriptHash(),
			from, amount-price, nil)
	}

	ctx := storage.GetContext()
	tokenID := crypto.Sha256([]byte(solutionData.SrcCode))
	if nftExists(ctx, tokenID) {
		panic("Such solution already exists")
	}

	nft := NFTSolution{
		TaskId:        solutionData.TaskId,
		TaskAssesment: solutionData.TaskAssesment,
		ID:            tokenID,
		Owner:         from,
		SrcCode:       solutionData.SrcCode,
		Description:   solutionData.Description,
	}
	setNFT(ctx, tokenID, nft)
	addToBalance(ctx, from, 1)
	addToken(ctx, from, tokenID)

	total := storage.Get(ctx, totalSupplyKey).(int) + 1
	storage.Put(ctx, totalSupplyKey, total)

	postTransfer(nil, from, tokenID, nil)

	util.AssertMsg(contractTask_hash != nil, "CONTRACT TASK HASh EQUALS NIL WTF")

	contract.Call(contractTask_hash, "changeTaskAssesment", contract.All,
		nft.TaskId, nft.TaskAssesment)

}

func mkAccountPrefix(holder interop.Hash160) []byte {
	res := []byte(accountPrefix)
	return append(res, holder...)
}

func mkBalanceKey(holder interop.Hash160) []byte {
	res := []byte(balancePrefix)
	return append(res, holder...)
}

func mkTokenKey(tokenID []byte) []byte {
	res := []byte(tokenPrefix)
	return append(res, tokenID...)
}
func getBalanceOf(ctx storage.Context, balanceKey []byte) int {
	val := storage.Get(ctx, balanceKey)
	if val != nil {
		return val.(int)
	}
	return 0
}

func addToBalance(ctx storage.Context, holder interop.Hash160, amount int) {
	key := mkBalanceKey(holder)
	old := getBalanceOf(ctx, key)
	old += amount
	if old > 0 {
		storage.Put(ctx, key, old)
	} else {
		storage.Delete(ctx, key)
	}
}

func addToken(ctx storage.Context, holder interop.Hash160, token []byte) {
	key := mkAccountPrefix(holder)
	storage.Put(ctx, append(key, token...), token)
}

func removeToken(ctx storage.Context, holder interop.Hash160, token []byte) {
	key := mkAccountPrefix(holder)
	storage.Delete(ctx, append(key, token...))
}

func ownerAddress(owner interop.Hash160) string {
	b := append([]byte{0x35}, owner...)
	return std.Base58CheckEncode(b)
}
