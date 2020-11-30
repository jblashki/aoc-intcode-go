package intcode

import (
	"log"
	"sync"

	"github.com/jblashki/aoc-filereader-go"
)

const OP_SUM = 1
const OP_MUL = 2
const OP_INP = 3
const OP_OUT = 4
const OP_JPT = 5
const OP_JPF = 6
const OP_LST = 7
const OP_EQU = 8
const OP_HLT = 99

type paramMode int

const (
	POSITION = iota
	VALUE
)

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

func Run(ic *IntCode, input chan int, output chan int, haltSignal chan int, wg *sync.WaitGroup) {
	ic.programPos = 0

	defer func() {
		haltSignal <- 1
		wg.Done()
	}()

	for {
		fullOp := readNextAddr(ic)
		op := fullOp % 100

		switch op {
		case OP_SUM:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)

			val1 := param1
			if getParamMode(fullOp, 0) == POSITION {
				val1 = Get(ic, param1)
			}
			val2 := param2
			if getParamMode(fullOp, 1) == POSITION {
				val2 = Get(ic, param2)
			}

			err := Set(ic, param3, val1+val2)
			if err != nil {
				log.Printf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				return
			}

		case OP_MUL:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)

			val1 := param1
			if getParamMode(fullOp, 0) == POSITION {
				val1 = Get(ic, param1)
			}

			val2 := param2
			if getParamMode(fullOp, 1) == POSITION {
				val2 = Get(ic, param2)
			}

			err := Set(ic, param3, val1*val2)
			if err != nil {
				log.Printf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				return
			}

		case OP_INP:
			param1 := readNextAddr(ic)

			val := <-input

			err := Set(ic, param1, val)
			if err != nil {
				log.Printf("Error setting address %v @ address %v: %v", param1, ic.programPos-2, err)
				return
			}

		case OP_OUT:
			param1 := readNextAddr(ic)

			val := param1
			if getParamMode(fullOp, 0) == POSITION {
				val = Get(ic, param1)
			}

			output <- val

		case OP_JPT:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)

			val1 := param1
			if getParamMode(fullOp, 0) == POSITION {
				val1 = Get(ic, param1)
			}

			val2 := param2
			if getParamMode(fullOp, 1) == POSITION {
				val2 = Get(ic, param2)
			}

			if val1 != 0 {
				ic.programPos = val2
			}

		case OP_JPF:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)

			val1 := param1
			if getParamMode(fullOp, 0) == POSITION {
				val1 = Get(ic, param1)
			}

			val2 := param2
			if getParamMode(fullOp, 1) == POSITION {
				val2 = Get(ic, param2)
			}

			if val1 == 0 {
				ic.programPos = val2
			}

		case OP_LST:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)

			val1 := param1
			if getParamMode(fullOp, 0) == POSITION {
				val1 = Get(ic, param1)
			}

			val2 := param2
			if getParamMode(fullOp, 1) == POSITION {
				val2 = Get(ic, param2)
			}

			outValue := 0
			if val1 < val2 {
				outValue = 1
			}

			err := Set(ic, param3, outValue)
			if err != nil {
				log.Printf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				return
			}

		case OP_EQU:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)

			val1 := param1
			if getParamMode(fullOp, 0) == POSITION {
				val1 = Get(ic, param1)
			}

			val2 := param2
			if getParamMode(fullOp, 1) == POSITION {
				val2 = Get(ic, param2)
			}

			outValue := 0
			if val1 == val2 {
				outValue = 1
			}

			err := Set(ic, param3, outValue)
			if err != nil {
				log.Printf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				return
			}

		case OP_HLT:
			return

		default:
			log.Printf("Unknown operation %v at address %v", op, ic.programPos-1)
			return
		}
	}

	return
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

func getParamMode(op int, param int) paramMode {
	op /= 100

	for param > 0 {
		op /= 10
		param--
	}

	if op%10 == 0 {
		return POSITION
	} else {
		return VALUE
	}
}
