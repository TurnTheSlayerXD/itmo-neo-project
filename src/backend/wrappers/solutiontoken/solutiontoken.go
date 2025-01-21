// Code generated by neo-go contract generate-rpcwrapper --manifest <file.json> --out <file.go> [--hash <hash>] [--config <config>]; DO NOT EDIT.

// Package solutiontoken contains RPC wrappers for solution_token contract.
package solutiontoken

import (
	"github.com/nspcc-dev/neo-go/pkg/core/transaction"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep11"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"math/big"
)

// Hash contains contract hash.
var Hash = util.Uint160{0x7d, 0xfe, 0x1d, 0x5a, 0xbe, 0x99, 0x95, 0xa, 0x91, 0x99, 0xcc, 0x7f, 0x9b, 0x2, 0xab, 0x3d, 0x4f, 0xb0, 0x62, 0x76}

// Invoker is used by ContractReader to call various safe methods.
type Invoker interface {
	nep11.Invoker
}

// Actor is used by Contract to call state-changing methods.
type Actor interface {
	Invoker

	nep11.Actor

	MakeCall(contract util.Uint160, method string, params ...any) (*transaction.Transaction, error)
	MakeRun(script []byte) (*transaction.Transaction, error)
	MakeUnsignedCall(contract util.Uint160, method string, attrs []transaction.Attribute, params ...any) (*transaction.Transaction, error)
	MakeUnsignedRun(script []byte, attrs []transaction.Attribute) (*transaction.Transaction, error)
	SendCall(contract util.Uint160, method string, params ...any) (util.Uint256, uint32, error)
	SendRun(script []byte) (util.Uint256, uint32, error)
}

// ContractReader implements safe contract methods.
type ContractReader struct {
	nep11.NonDivisibleReader
	invoker Invoker
	hash    util.Uint160
}

// Contract implements all contract methods.
type Contract struct {
	ContractReader
	nep11.BaseWriter
	actor Actor
	hash  util.Uint160
}

// NewReader creates an instance of ContractReader using Hash and the given Invoker.
func NewReader(invoker Invoker) *ContractReader {
	var hash = Hash
	return &ContractReader{*nep11.NewNonDivisibleReader(invoker, hash), invoker, hash}
}

// New creates an instance of Contract using Hash and the given Actor.
func New(actor Actor) *Contract {
	var hash = Hash
	var nep11ndt = nep11.NewNonDivisible(actor, hash)
	return &Contract{ContractReader{nep11ndt.NonDivisibleReader, actor, hash}, nep11ndt.BaseWriter, actor, hash}
}

// ChangeTaskAssesment creates a transaction invoking `changeTaskAssesment` method of the contract.
// This transaction is signed and immediately sent to the network.
// The values returned are its hash, ValidUntilBlock value and error if any.
func (c *Contract) ChangeTaskAssesment(tokenid []byte, newAssesmentNum *big.Int) (util.Uint256, uint32, error) {
	return c.actor.SendCall(c.hash, "changeTaskAssesment", tokenid, newAssesmentNum)
}

// ChangeTaskAssesmentTransaction creates a transaction invoking `changeTaskAssesment` method of the contract.
// This transaction is signed, but not sent to the network, instead it's
// returned to the caller.
func (c *Contract) ChangeTaskAssesmentTransaction(tokenid []byte, newAssesmentNum *big.Int) (*transaction.Transaction, error) {
	return c.actor.MakeCall(c.hash, "changeTaskAssesment", tokenid, newAssesmentNum)
}

// ChangeTaskAssesmentUnsigned creates a transaction invoking `changeTaskAssesment` method of the contract.
// This transaction is not signed, it's simply returned to the caller.
// Any fields of it that do not affect fees can be changed (ValidUntilBlock,
// Nonce), fee values (NetworkFee, SystemFee) can be increased as well.
func (c *Contract) ChangeTaskAssesmentUnsigned(tokenid []byte, newAssesmentNum *big.Int) (*transaction.Transaction, error) {
	return c.actor.MakeUnsignedCall(c.hash, "changeTaskAssesment", nil, tokenid, newAssesmentNum)
}

// TokensList creates a transaction invoking `tokensList` method of the contract.
// This transaction is signed and immediately sent to the network.
// The values returned are its hash, ValidUntilBlock value and error if any.
func (c *Contract) TokensList() (util.Uint256, uint32, error) {
	return c.actor.SendCall(c.hash, "tokensList")
}

// TokensListTransaction creates a transaction invoking `tokensList` method of the contract.
// This transaction is signed, but not sent to the network, instead it's
// returned to the caller.
func (c *Contract) TokensListTransaction() (*transaction.Transaction, error) {
	return c.actor.MakeCall(c.hash, "tokensList")
}

// TokensListUnsigned creates a transaction invoking `tokensList` method of the contract.
// This transaction is not signed, it's simply returned to the caller.
// Any fields of it that do not affect fees can be changed (ValidUntilBlock,
// Nonce), fee values (NetworkFee, SystemFee) can be increased as well.
func (c *Contract) TokensListUnsigned() (*transaction.Transaction, error) {
	return c.actor.MakeUnsignedCall(c.hash, "tokensList", nil)
}

// TokensOfList creates a transaction invoking `tokensOfList` method of the contract.
// This transaction is signed and immediately sent to the network.
// The values returned are its hash, ValidUntilBlock value and error if any.
func (c *Contract) TokensOfList(holder util.Uint160) (util.Uint256, uint32, error) {
	return c.actor.SendCall(c.hash, "tokensOfList", holder)
}

// TokensOfListTransaction creates a transaction invoking `tokensOfList` method of the contract.
// This transaction is signed, but not sent to the network, instead it's
// returned to the caller.
func (c *Contract) TokensOfListTransaction(holder util.Uint160) (*transaction.Transaction, error) {
	return c.actor.MakeCall(c.hash, "tokensOfList", holder)
}

// TokensOfListUnsigned creates a transaction invoking `tokensOfList` method of the contract.
// This transaction is not signed, it's simply returned to the caller.
// Any fields of it that do not affect fees can be changed (ValidUntilBlock,
// Nonce), fee values (NetworkFee, SystemFee) can be increased as well.
func (c *Contract) TokensOfListUnsigned(holder util.Uint160) (*transaction.Transaction, error) {
	return c.actor.MakeUnsignedCall(c.hash, "tokensOfList", nil, holder)
}
