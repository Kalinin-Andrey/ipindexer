package ipwriter

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// IPNode is the structure for an node in a IP tree
type IPNode struct{
	Value		IP
	Left		*IPNode
	Right		*IPNode
	NextSection	*IPTree
}

// IPTree is a binary tree for IPv4 values
// root always has index[0]
type IPTree []*IPNode

// IP struct
type IP struct {
	s	string
	u	[4]uint
}

// NewIP is return pointer to the new IP struct
func NewIP(s string) (*IP, error) {
	ip := &IP{
		s:	s,
	}
	st := strings.Split(s, ".")
	var i int
	var v string
	for i, v = range st {
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return ip, err
		}
		if i > 3 {
			return ip, errors.New("IPv4 addres can not has more then 4 sections")
		}
		(*ip).u[i] = uint(u)
	}
	if i < 3 {
		return ip, errors.New("IPv4 addres can not has less then 4 sections")
	}
	return ip, nil
}


func (w *IPWriter) makeRootIPNode(ip *IP, startUintNode *UintNode, sectionNum uint) (*IPNode, error) {
	u := ip.u

	for i := sectionNum + 1; i < uint(len(u)); i++ {
		u[i] = 0
	}
	u[sectionNum] = startUintNode.Value

	sls := make([]string, len(u))
	for i := 0; i < len(u); i++ {
		sls[i] = strconv.Itoa(int(u[i]))
	}
	s := strings.Join(sls, ".")
	return &IPNode{
		Value: IP{
			u: u,
			s: s,
		},
	}, nil
}

// MakeIPTree is build an IP tree
func (w *IPWriter) MakeIPTree(ip *IP, startUintNode *UintNode, sectionNum uint) (IPTree, error) {
	var err			error
	var ipTree		IPTree
	var childTree	IPTree
	rootNode, err := w.makeRootIPNode(ip, startUintNode, sectionNum)
	if err != nil {
		return ipTree, err
	}

	if ip.u[sectionNum] == startUintNode.Value {
		if sectionNum != uint(len(ip.u)) - 1 {
			rootUint := (*w).uintTree[0]
			nextSecTree, err := w.MakeIPTree(ip, rootUint, sectionNum + 1)
			if err != nil {
				return ipTree, err
			}
			rootNode.NextSection = &nextSecTree
		}

		ipTree = make(IPTree, 0, 1)
		ipTree = append(ipTree, rootNode)
		return ipTree, nil
	}

	if ip.u[sectionNum] < startUintNode.Value {
		childTree, err = w.MakeIPTree(ip, startUintNode.Left, sectionNum)
		if err != nil {
			return nil, err
		}
		rootNode.Left = childTree[0]
	} else {
		childTree, err = w.MakeIPTree(ip, startUintNode.Right, sectionNum)
		if err != nil {
			return nil, err
		}
		rootNode.Right = childTree[0]
	}

	ipTree = make(IPTree, 0, len(childTree) + 1)
	ipTree = append(ipTree, rootNode)
	ipTree = append(ipTree, childTree...)

	return ipTree, nil
}



