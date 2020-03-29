package merkle

// Merkle is an interface to verifiable data structures, its implementation can be merkletree, etc.
type Merkle interface {
	GetMerkleRoot() []byte
	VerifyMerkle() (bool, error)
	VerifyContent(interface{}) (bool, error)
	GetMerklePath(interface{}) ([][]byte, []int64, error)
	String() string
}
