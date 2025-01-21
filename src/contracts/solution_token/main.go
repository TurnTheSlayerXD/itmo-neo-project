package solution_token

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/contract"
	"github.com/nspcc-dev/neo-go/pkg/interop/iterator"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/crypto"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/gas"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/ledger"
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

const (
	forSolutionGas = 1
	minNameLen     = 3
)

type NFTSolution struct {
	ID            []byte
	Taskid        []byte
	Owner         interop.Hash160
	Created       int
	SrcCode       string
	NAssesments   int
	AverAssesment int
	Description   string
}

func _deploy(data interface{}, isUpdate bool) {
	if isUpdate {
		return
	}
	ctx := storage.GetContext()
	storage.Put(ctx, ownerKey, runtime.GetCallingScriptHash())
	storage.Put(ctx, totalSupplyKey, 0)
}

// Symbol returns token symbol, it's NICENAMES.
func Symbol() string {
	return "NICENAMES"
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
		"owner":         ownerAddress(nft.Owner),
		"created":       std.Itoa10(nft.Created),
		"SrcCode":       string(nft.SrcCode),
		"Description":   string(nft.Description),
		"NAssesments":   std.Itoa10(nft.NAssesments),
		"AverAssesment": std.Itoa10(nft.AverAssesment),
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

// Transfer token from its owner to another user, notice that it only has three
// parameters because token owner can be deduced from token ID itself.
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

// postTransfer emits Transfer event and calls onNEP11Payment if needed.
func postTransfer(from interop.Hash160, to interop.Hash160, token []byte, data any) {
	runtime.Notify("Transfer", from, to, 1, token)
	if management.GetContract(to) != nil {
		contract.Call(to, "onNEP11Payment", contract.All, from, 1, token, data)
	}
}

// OnNEP17Payment mints tokens if at least 10 GAS is provided. You don't call
// this method directly, instead it's called by GAS contract when you transfer
// GAS from your address to the address of this NFT contract.
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

	input_data := data.(struct {
		Taskid      []byte
		SrcCode     string
		Description string
	})

	price := 10_0000_0000

	if amount < price {
		panic("insufficient GAS for minting NFT")
	}

	ctx := storage.GetContext()
	tokenID := crypto.Sha256([]byte(input_data.SrcCode))
	if nftExists(ctx, tokenID) {
		panic("Such solution already exists")
	}

	nft := NFTSolution{
		Taskid:        input_data.Taskid,
		ID:            tokenID,
		Owner:         from,
		Created:       ledger.CurrentIndex(),
		SrcCode:       input_data.SrcCode,
		Description:   input_data.Description,
		NAssesments:   0,
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
	if newAssesmentNum < 0 || newAssesmentNum > 10 {
		panic("Wrong assesment num")
	}

	context := storage.GetContext()
	nft := getNFT(context, tokenid)
	prevAver := nft.AverAssesment
	nft.AverAssesment = (nft.AverAssesment*nft.NAssesments + newAssesmentNum) / (nft.NAssesments + 1)
	nft.NAssesments += 1

	setNFT(context, tokenid, nft)

	reward := forSolutionGas
	if prevAver < nft.AverAssesment {
		reward += forSolutionGas
	}
	gas.Transfer(runtime.GetExecutingScriptHash(), runtime.GetCallingScriptHash(), forSolutionGas, nil)
}

// mkAccountPrefix creates DB key-prefix for the account tokens specified
// by concatenating accountPrefix and account address.
func mkAccountPrefix(holder interop.Hash160) []byte {
	res := []byte(accountPrefix)
	return append(res, holder...)
}

// mkBalanceKey creates DB key for the account specified by concatenating balancePrefix
// and account address.
func mkBalanceKey(holder interop.Hash160) []byte {
	res := []byte(balancePrefix)
	return append(res, holder...)
}

// mkTokenKey creates DB key for the token specified by concatenating tokenPrefix
// and token ID.
func mkTokenKey(tokenID []byte) []byte {
	res := []byte(tokenPrefix)
	return append(res, tokenID...)
}

// getBalanceOf returns the balance of an account using database key.
func getBalanceOf(ctx storage.Context, balanceKey []byte) int {
	val := storage.Get(ctx, balanceKey)
	if val != nil {
		return val.(int)
	}
	return 0
}

// addToBalance adds an amount to the account balance. Amount can be negative.
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

// addToken adds a token to the account.
func addToken(ctx storage.Context, holder interop.Hash160, token []byte) {
	key := mkAccountPrefix(holder)
	storage.Put(ctx, append(key, token...), token)
}

// removeToken removes the token from the account.
func removeToken(ctx storage.Context, holder interop.Hash160, token []byte) {
	key := mkAccountPrefix(holder)
	storage.Delete(ctx, append(key, token...))
}

func ownerAddress(owner interop.Hash160) string {
	b := append([]byte{0x35}, owner...)
	return std.Base58CheckEncode(b)
}
