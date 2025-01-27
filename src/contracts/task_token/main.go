package task_token

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

const (
	balancePrefix = "b"
	accountPrefix = "a"
	tokenPrefix   = "t"

	ownerKey       = 'o'
	totalSupplyKey = 's'
)

const (
	forSolutionGas = 1
	minNameLen     = 3
)

type NFTTask struct {
	ID            []byte
	Owner         interop.Hash160
	Name          string
	Tests         string
	Description   string
	NSolutions    int
	AverAssesment int
}

func _deploy(data interface{}, isUpdate bool) {
	if isUpdate {
		return
	}

	ctx := storage.GetContext()
	storage.Put(ctx, ownerKey, runtime.GetCallingScriptHash())
	storage.Put(ctx, totalSupplyKey, 0)
}

func Symbol() string {
	return "TASK"
}

func Decimals() int {
	return 0
}

func TotalSupply() int {
	return storage.Get(storage.GetReadOnlyContext(), totalSupplyKey).(int)
}

func BalanceOf(holder interop.Hash160) int {
	if len(holder) != 20 {
		panic("bad owner address")
	}
	ctx := storage.GetReadOnlyContext()
	return getBalanceOf(ctx, mkBalanceKey(holder))
}

func OwnerOf(token []byte) interop.Hash160 {
	ctx := storage.GetReadOnlyContext()
	return getNFT(ctx, token).Owner
}

func Properties(token []byte) map[string]string {
	ctx := storage.GetReadOnlyContext()
	nft := getNFT(ctx, token)

	result := map[string]string{
		"id":            string(nft.ID),
		"owner":         ownerAddress(nft.Owner),
		"name":          nft.Name,
		"tests":         nft.Tests,
		"description":   nft.Description,
		"nsolutions":    std.Itoa10(nft.NSolutions),
		"averassesment": std.Itoa10(nft.AverAssesment),
	}
	return result
}

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

func getNFT(ctx storage.Context, token []byte) NFTTask {
	key := mkTokenKey(token)
	val := storage.Get(ctx, key)

	util.AssertMsg(val != nil, "no token found")

	serializedNFT := val.([]byte)
	deserializedNFT := std.Deserialize(serializedNFT)
	return deserializedNFT.(NFTTask)
}

func nftExists(ctx storage.Context, token []byte) bool {
	key := mkTokenKey(token)
	return storage.Get(ctx, key) != nil
}

func setNFT(ctx storage.Context, token []byte, item NFTTask) {
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
	util.AssertMsg(callingHash.Equals(gas.Hash), "only GAS is accepted")

	dict := std.JSONDeserialize(data.([]byte)).(map[string]any)

	taskData := &struct {
		Name        string
		Tests       string
		Description string
	}{
		Name:        dict["name"].(string),
		Tests:       dict["tests"].(string),
		Description: dict["description"].(string),
	}

	util.AssertMsg(len(taskData.Name) > 3, "name length at least 3 character")

	price := 10_0000_0000

	util.AssertMsg(amount >= price, "insufficient GAS for minting NFT")

	if amount > price {
		gas.Transfer(runtime.GetExecutingScriptHash(),
			runtime.GetCallingScriptHash(), amount-price, nil)
	}

	ctx := storage.GetContext()
	tokenID := crypto.Sha256([]byte(taskData.Name))

	util.AssertMsg(!nftExists(ctx, tokenID), "token already exists")

	nft := NFTTask{
		ID:            tokenID,
		Owner:         from,
		Name:          taskData.Name,
		Tests:         taskData.Tests,
		Description:   taskData.Description,
		NSolutions:    0,
		AverAssesment: 0,
	}
	setNFT(ctx, tokenID, nft)
	addToBalance(ctx, from, 1)
	addToken(ctx, from, tokenID)

	total := storage.Get(ctx, totalSupplyKey).(int) + 1
	storage.Put(ctx, totalSupplyKey, total)

	postTransfer(nil, from, tokenID, nil)
}

func ChangeTaskAssesment(tokenid []byte, newAssesmentNum int) {

	util.AssertMsg(0 < newAssesmentNum && newAssesmentNum <= 10, "Wrong assesment num")

	context := storage.GetContext()

	nft := getNFT(context, tokenid)
	prevAver := nft.AverAssesment
	nft.AverAssesment = (nft.AverAssesment*nft.NSolutions + newAssesmentNum) / (nft.NSolutions + 1)
	nft.NSolutions += 1

	setNFT(context, tokenid, nft)

	reward := forSolutionGas
	if prevAver < nft.AverAssesment {
		reward += forSolutionGas
	}
	util.AssertMsg(gas.Transfer(runtime.GetExecutingScriptHash(), nft.Owner, forSolutionGas, nil),
		"Could not reward for task assesment")
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
