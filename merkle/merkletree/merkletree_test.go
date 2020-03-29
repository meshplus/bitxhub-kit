package merkletree

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	mt "github.com/cbergoon/merkletree"
	"github.com/meshplus/bitxhub-kit/merkle"
	"github.com/pkg/errors"
)

type Transaction struct {
	TransactionHash []byte
}

// CalculateHash hashes the values of a TestContent
func (t *Transaction) CalculateHash() ([]byte, error) {
	return t.TransactionHash, nil
}

// Equals tests for equality of two Contents
func (t *Transaction) Equals(other mt.Content) (bool, error) {
	tOther, ok := other.(*Transaction)
	if !ok {
		return false, errors.New("Parameter should be type Transaction")
	}
	return bytes.Equal(t.TransactionHash, tOther.TransactionHash), nil

}

func newTransaction(hash string) *Transaction {
	var newTx = &Transaction{
		TransactionHash: []byte(hash),
	}
	return newTx
}

var list []interface{}
var other *Transaction
var m merkle.Merkle

func TestQueue_NewMerkleTree(t *testing.T) {
	list = append(list, newTransaction("0001"))
	list = append(list, newTransaction("0003"))
	list = append(list, newTransaction("0004"))
	list = append(list, newTransaction("0005"))
	list = append(list, newTransaction("0002"))
	list = append(list, newTransaction("0006"))
	other = newTransaction("101")

	txTree := NewMerkleTree()
	err := txTree.InitMerkleTree(list)
	if err != nil {
		t.Error("Fail to call NewMerkleTree()")
	}

	m = txTree
	t.Log("Pass NewMerkleTree()")
}

func TestMerkleTree_GetMerkleRoot(t *testing.T) {
	ans := []byte{238, 172, 249, 160, 199, 95, 237, 193, 160, 57, 146, 40, 234, 163, 52, 92, 141, 195, 143, 18, 13, 199, 46, 57, 127, 15, 156, 31, 199, 223, 129, 83}
	if bytes.Equal(m.GetMerkleRoot(), ans) {
		t.Log("Pass GetMerkleRoot()")
	} else {
		t.Errorf("Error GetMerkleRoot()")
	}
}

func TestMerkleTree_VerifyContent(t *testing.T) {
	vc, err := m.VerifyContent(newTransaction("0001"))
	assert.Nil(t, err)
	assert.True(t, vc)

	vc, err = m.VerifyContent(newTransaction("0002"))
	assert.Nil(t, err)
	assert.True(t, vc)

	vc, err = m.VerifyContent(newTransaction("0003"))
	assert.Nil(t, err)
	assert.True(t, vc)

	vc, err = m.VerifyContent(newTransaction("0004"))
	assert.Nil(t, err)
	assert.True(t, vc)

	vc, err = m.VerifyContent(newTransaction("0005"))
	assert.Nil(t, err)
	assert.True(t, vc)

	vc, err = m.VerifyContent(other)
	assert.Nil(t, err)
	assert.True(t, vc)
}
