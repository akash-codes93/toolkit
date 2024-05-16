package toolkit

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// fields hasto be capital
type MerkelNode struct {
	HashValue string
	FileName  string
	Left      *MerkelNode
	Right     *MerkelNode
}

func (m *MerkelNode) IsLeaf() bool {
	if m.Left == nil && m.Right == nil {
		return true
	}
	return false
}

func calculateHashes(files []string) ([]string, error) {
	hashes := []string{}
	// for all the files calculate hashes
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return hashes, err
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return hashes, err
		}
		hash := fmt.Sprintf("%x", h.Sum(nil))
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

func createMerkelTree(files []string) *MerkelNode {
	if len(files) == 0 {
		return nil
	}

	if len(files)%2 != 0 {
		files = append(files, files[len(files)-1])
	}

	hashes, _ := calculateHashes(files)
	hashesNodes := []*MerkelNode{}
	// creating leaves
	for idx, hash := range hashes {
		hashesNodes = append(hashesNodes, &MerkelNode{
			HashValue: hash,
			FileName:  files[idx],
		})
	}

	var prepareMerkel func(nodes []*MerkelNode) []*MerkelNode

	prepareMerkel = func(nodes []*MerkelNode) []*MerkelNode {
		if len(nodes) == 1 {
			return nodes
		}
		newNodes := []*MerkelNode{}

		// now prepare parents
		for i := 0; i < len(nodes); i += 2 {
			combinedHash := nodes[i].HashValue + nodes[i+1].HashValue
			bv := []byte(combinedHash)
			h := sha256.New()
			h.Write(bv)
			hash := fmt.Sprintf("%x", h.Sum(nil))
			// fmt.Println(hash)

			newNodes = append(newNodes, &MerkelNode{
				HashValue: hash,
				Left:      nodes[i],
				Right:     nodes[i+1],
			})
		}
		return prepareMerkel(newNodes)
	}
	return prepareMerkel(hashesNodes)[0]
}

func checkDifferentFiles(oldRoot, newRoot *MerkelNode) []string {
	if oldRoot.HashValue == newRoot.HashValue {
		return []string{}
	}

	if oldRoot.IsLeaf() && newRoot.IsLeaf() {
		return []string{newRoot.FileName}
	}

	left := checkDifferentFiles(oldRoot.Left, newRoot.Left)
	right := checkDifferentFiles(oldRoot.Right, newRoot.Right)

	left = append(left, right...)
	return left
}
