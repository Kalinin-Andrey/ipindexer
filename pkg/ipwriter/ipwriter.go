package ipwriter

import (
	"bufio"
	"fmt"
	"github.com/Kalinin-Andrey/ipindexer/pkg/apperror"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"strings"
)

// IPWriter is a main struict of this package
type IPWriter struct {
	uintTree 	UintTree
	rws       	io.ReadWriteSeeker
	writer		bufio.Writer
	currentLine	uint
}

// Recordset is a record of index file
type Recordset struct {
	Num			uint
	Value		IP
	Left		uint
	Right		uint
	NextSection	uint
	IsReal		uint8
}

// NotFoundFunc is the type to call if the IPNode was not found
type NotFoundFunc func(parentNode *IPNode, parentRst *Recordset, childPosition uint, notFoundNode *IPNode, currentRecordset *Recordset) (NewRecordset *Recordset, err error)

const (
	// ValPos is the position number in a record for a value
	ValPos			= 0
	// LeftPos is the position number in a record for a line of a left branch
	LeftPos         = 1
	// RightPos is the position number in a record for a line of a right branch
	RightPos		= 2
	// NextSectionPos is the position number in a record for a line of a next section tree
	NextSectionPos	= 3
	// IsRealPos is the position number in a record for IsReal flag
	IsRealPos		= 4

	// BytesInLn is a quantity of bytes in line of index file
	BytesInLn		= 51 // ip:15 + leftLn:10 + rightLn:10 + nextSectionLn:10 + separators:3 + lnEnd:1
)

func (r Recordset) String() string {
	return fmt.Sprintf("%03d.%03d.%03d.%03d %010d %010d %010d %d\n", r.Value.u[0], r.Value.u[1], r.Value.u[2], r.Value.u[3], r.Left, r.Right, r.NextSection, r.IsReal)
}

// New is the constructor for an IPWriter object
func New(rws io.ReadWriteSeeker) *IPWriter {
	uintTree := MakeUintTree(uintTreeCapacity/ 2, uintTreeCapacity)
	w := &IPWriter{
		rws:       rws,
		uintTree: *uintTree,
	}
	return w
}

// SetCurrentByte is the setter for IPWriter.currentLine in bytes
func (w *IPWriter) SetCurrentByte(n uint) {
	w.SetCurrentLine(n / BytesInLn)
}

// SetCurrentLine is the setter for IPWriter.currentLine
func (w *IPWriter) SetCurrentLine(n uint) {
	w.currentLine = n
}

// Write is write given ip in index file
func (w *IPWriter) Write(ips string) error {
	ip, err := NewIP(ips)
	if err != nil {
		return err
	}
	return w.WriteIP(ip)
}

// Find is for searching given ip in index file
func (w *IPWriter) Find(ips string) error {
	ip, err := NewIP(ips)
	if err != nil {
		return err
	}
	return w.FindIP(ip)
}

// FindIP is for searching given IP in index file
func (w *IPWriter) FindIP(ip *IP) error {
	rootUint := (*w).uintTree[0]
	ipTree, err := w.MakeIPTree(ip, rootUint, 0)
	if err != nil {
		return err
	}
	rootRst, err := w.readRecordset(0)
	if err != nil {
		return err
	}
	err = w.bypass(ipTree[0], rootRst, w.returnNotFound)
	return err
}

// WriteIP is write given IP in index file
func (w *IPWriter) WriteIP(ip *IP) error {
	rootUint := (*w).uintTree[0]
	ipTree, err := w.MakeIPTree(ip, rootUint, 0)
	if err != nil {
		return err
	}
	rootRst, err := w.readRecordset(0)
	if err != nil {
		if err != apperror.ErrNotFound {
			return err
		}
		rootRst = &Recordset{
			Value:       ipTree[0].Value,
		}
		err = w.writeRecordset(rootRst, true)
		if err != nil {
			return err
		}
	}
	err = w.bypass(ipTree[0], rootRst, w.writeIfNotFound)
	return err
}

func (w *IPWriter) bypass(parentNode *IPNode, parentRst *Recordset, notFoundFunc NotFoundFunc) error {
	var lineNum, childPosition uint
	var currentNode *IPNode
	var currentRecordset *Recordset
	var err error

	switch {
	case parentNode.Left != nil:
		childPosition = LeftPos
		lineNum = parentRst.Left
		currentNode = parentNode.Left
	case parentNode.Right != nil:
		childPosition = RightPos
		lineNum = parentRst.Right
		currentNode = parentNode.Right
	case parentNode.NextSection != nil:
		childPosition = NextSectionPos
		lineNum = parentRst.NextSection
		currentNode = (*parentNode.NextSection)[0]
	default:
		if parentRst.IsReal != 1 {
			return errors.New("bypass has done, but recordset is not real ip")
		}
		return nil
	}

	if lineNum == 0 {
		currentRecordset, err = notFoundFunc(parentNode, parentRst, childPosition, currentNode, nil)
		if err != nil {
			return err
		}
	} else {
		currentRecordset, err = w.readRecordset(lineNum)
		if err != nil {
			if err != apperror.ErrNotFound {
				return err
			}
			// if situation after a failure
			currentRecordset, err = notFoundFunc(parentNode, parentRst, childPosition, currentNode, nil)
			if err != nil {
				return err
			}
		} else if currentNode.Left == nil && currentNode.Right == nil && currentNode.NextSection == nil && currentRecordset.IsReal != 1 {
			currentRecordset.IsReal = 1
			currentRecordset, err = notFoundFunc(parentNode, parentRst, childPosition, currentNode, currentRecordset)
			if err != nil {
				return err
			}
		}
	}
	return w.bypass(currentNode, currentRecordset, notFoundFunc)
}

func (w *IPWriter) returnNotFound(parentNode *IPNode, parentRst *Recordset, childPosition uint, currentNode *IPNode, currentRecordset *Recordset) (NewRecordset *Recordset, err error) {
	return nil, apperror.ErrNotFound
}

func (w *IPWriter) writeIfNotFound(parentNode *IPNode, parentRst *Recordset, childPosition uint, currentNode *IPNode, currentRecordset *Recordset) (NewRecordset *Recordset, err error) {
	if currentRecordset != nil {
		err = w.writeRecordset(currentRecordset, false)
		if err != nil {
			return nil, err
		}
		return currentRecordset, nil
	}

	switch childPosition {
	case LeftPos:
		parentRst.Left = w.currentLine
	case RightPos:
		parentRst.Right = w.currentLine
	case NextSectionPos:
		parentRst.NextSection = w.currentLine
	}
	err = w.writeRecordset(parentRst, false)
	if err != nil {
		return nil, err
	}
	newRecordset := &Recordset{
		Value:       currentNode.Value,
	}

	if currentNode.Left == nil && currentNode.Right == nil && currentNode.NextSection == nil {
		newRecordset.IsReal = 1
	}

	err = w.writeRecordset(newRecordset, true)
	if err != nil {
		return nil, err
	}
	return newRecordset, nil
}

func (w *IPWriter) goTo(lineNum uint) error {
	_, err := w.rws.Seek(int64(lineNum) * BytesInLn, io.SeekStart)
	return err
}

func (w *IPWriter) readRecordset(lineNum uint) (*Recordset, error) {
	err := w.goTo(lineNum)
	if err != nil {
		return nil, err
	}
	r			:= bufio.NewReader(w.rws)
	ln, err		:= r.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	if ln == "" {
		return nil, apperror.ErrNotFound
	}
	fields := strings.Fields(ln)
	ip, err := NewIP(fields[0])
	if err != nil {
		return nil, err
	}
	left, err := strconv.ParseUint(fields[LeftPos], 10, 64)
	if err != nil {
		return nil, err
	}
	right, err := strconv.ParseUint(fields[RightPos], 10, 64)
	if err != nil {
		return nil, err
	}
	nextSec, err := strconv.ParseUint(fields[NextSectionPos], 10, 64)
	if err != nil {
		return nil, err
	}
	isReal, err := strconv.ParseUint(fields[IsRealPos], 10, 8)
	if err != nil {
		return nil, err
	}
	rst := &Recordset{
		Num:			lineNum,
		Value:			*ip,
		Left:			uint(left),
		Right:			uint(right),
		NextSection:	uint(nextSec),
		IsReal:			uint8(isReal),
	}
	return rst, nil
}

func (w *IPWriter) writeRecordset(recordset *Recordset, isNew bool) error {
	if isNew {
		recordset.Num = w.currentLine
	}

	err := w.goTo(recordset.Num)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w.rws, recordset.String())
	if err == nil && isNew {
		w.currentLine ++
	}
	return err
}



