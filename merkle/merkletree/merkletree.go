package merkletree

import (
	"fmt"

	"github.com/cbergoon/merkletree"
	"github.com/meshplus/bitxhub-kit/merkle"
)

type MerkleTree struct {
	tree *merkletree.MerkleTree
}

var _ merkle.Merkle = &MerkleTree{}

// NewMerkleTree create a empty Merkle Tree
func NewMerkleTree() *MerkleTree {
	return &MerkleTree{
		tree: &merkletree.MerkleTree{},
	}
}

// InitMerkleTree create a new Merkle Tree from the list of Content
func (t *MerkleTree) InitMerkleTree(list []interface{}) error {
	var contents []merkletree.Content
	for _, v := range list {
		c, ok := v.(merkletree.Content)
		if !ok {
			return fmt.Errorf("objet did not implement CalculateHash() and Equals() correctly")
		}
		contents = append(contents, c)
	}
	tree, err := merkletree.NewTree(contents)
	if err != nil {
		return err
	}
	t.tree = tree
	return nil
}

// GetMerkleRoot get the Merkle Root of the tree
func (t *MerkleTree) GetMerkleRoot() []byte {
	return t.tree.MerkleRoot()
}

// VerifyMerkle verify the entire tree (hashes for each node) is valid
func (t *MerkleTree) VerifyMerkle() (bool, error) {
	vt, err := t.tree.VerifyTree()
	if err != nil {
		return false, err
	}
	return vt, err
}

// VerifyContent indicates whether a given content is in the tree and the hashes are valid for that content.
// Returns true if the expected Merkle Root is equivalent to the Merkle root calculated on the critical path
// for a given content. Returns true if valid and false otherwise.
func (t *MerkleTree) VerifyContent(obj interface{}) (bool, error) {
	content, ok := obj.(merkletree.Content)
	if !ok {
		return false, fmt.Errorf("objet did not implement CalculateHash() and Equals() correctly")
	}
	vc, err := t.tree.VerifyContent(content)
	if err != nil {
		return false, err
	}
	return vc, err
}

// GetMerklePath get Merkle path and indexes(left leaf or right leaf)
func (t *MerkleTree) GetMerklePath(obj interface{}) ([][]byte, []int64, error) {
	content, ok := obj.(merkletree.Content)
	if !ok {
		return [][]byte{}, []int64{}, fmt.Errorf("objet did not implement CalculateHash() and Equals() correctly")
	}
	return t.tree.GetMerklePath(content)
}

// String returns a string representation of the tree. Only leaf nodes are included
// in the output.
func (t *MerkleTree) String() string {
	return t.tree.String()
}
