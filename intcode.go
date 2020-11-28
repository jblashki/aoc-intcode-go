package intcode

import (
	"errors"
	"fmt"

	"github.com/jblashki/aoc-filereader-go"
)

const OP_SUM = 1
const OP_MUL = 2
const OP_HLT = 99

type IntCode struct {
	memory     []int
	programPos int
}

func Create() *IntCode {
	newIC := new(IntCode)

	newIC.memory = make([]int, 0)
	newIC.programPos = 0

	return newIC
}

func Copy(sourceIC *IntCode) *IntCode {
	copiedIC := *sourceIC

	copiedIC.memory = make([]int, len(sourceIC.memory))
	copy(copiedIC.memory, sourceIC.memory)

	return &copiedIC
}

func Set(ic *IntCode, addr int, value int) error {
	if addr >= len(ic.memory) {
		newSpace := addr - len(ic.memory) + 1
		newMem := make([]int, newSpace)
		ic.memory = append(ic.memory, newMem...)
	}

	ic.memory[addr] = value

	return nil
}

func Get(ic *IntCode, addr int) int {
	if addr >= len(ic.memory) {
		return 0
	}

	return ic.memory[addr]
}

func Run(ic *IntCode, returnAddr int) (int, error) {
	ic.programPos = 0
	for {
		op := readNextAddr(ic)

		switch op {
		case OP_SUM:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)

			val1 := Get(ic, param1)
			val2 := Get(ic, param2)

			err := Set(ic, param3, val1+val2)
			if err != nil {
				errormsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				return 0, errors.New(errormsg)
			}

		case OP_MUL:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)

			val1 := Get(ic, param1)
			val2 := Get(ic, param2)

			err := Set(ic, param3, val1*val2)
			if err != nil {
				errormsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				return 0, errors.New(errormsg)
			}

		case OP_HLT:
			return ic.memory[returnAddr], nil

		default:
			errormsg := fmt.Sprintf("Unknown operation %v at address %v", op, ic.programPos-1)
			return 0, errors.New(errormsg)
		}
	}

	return ic.memory[returnAddr], nil
}

func Load(ic *IntCode, file string) error {
	var err error

	ic.memory, err = filereader.ReadCSVInts(file)

	return err
}

func readNextAddr(ic *IntCode) int {
	value := ic.memory[ic.programPos]

	(ic.programPos)++

	return value
}
