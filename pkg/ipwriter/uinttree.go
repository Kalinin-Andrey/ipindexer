package ipwriter

// UintNode is the struct for a nod for UintTree
type UintNode struct{
	Value	uint
	Left	*UintNode
	Right	*UintNode
}

// UintTree is a binary tree for uint values
// root always has index[0]
type UintTree []*UintNode


const uintTreeCapacity = 256

// MakeUintTree is build an UintTree
func MakeUintTree(rootVal uint, cap uint) *UintTree {
	tree := make(UintTree, 0, cap)
	root := &UintNode{Value: rootVal}
	tree = append(tree, root)

	if rootVal % 2 != 0 {
		return &tree
	}
	fourthPart := cap / 4

	leftVal := rootVal - fourthPart
	rightVal := rootVal + fourthPart
	leftTree := MakeUintTree(leftVal, cap / 2)
	rightTree := MakeUintTree(rightVal, cap / 2)
	root.Left = (*leftTree)[0]
	root.Right = (*rightTree)[0]
	tree = append(tree, (*leftTree)...)
	tree = append(tree, (*rightTree)...)
	return &tree
}



