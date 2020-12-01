package intcode

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jblashki/aoc-filereader-go"
)

const op_sum = 1
const op_mul = 2
const op_inp = 3
const op_out = 4
const op_jpt = 5
const op_jpf = 6
const op_lst = 7
const op_equ = 8
const op_rbs = 9
const op_hlt = 99

type paramMode int

const (
	mode_pos = iota // Postion mode i.e. address
	mode_val        // Value mode i.e. absolute value
	mode_rel        // Relative mode i.e. relative address
)

type IntCode struct {
	memory       []int
	programPos   int
	relativeBase int
}

func Create() *IntCode {
	newIC := new(IntCode)

	newIC.memory = make([]int, 0)
	newIC.programPos = 0
	newIC.relativeBase = 0

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

func getParamValue(ic *IntCode, param int, mode paramMode) int {
	returnVal := param
	if mode == mode_pos {
		returnVal = Get(ic, param)
	} else if mode == mode_rel {
		returnVal = Get(ic, ic.relativeBase+param)
	}

	return returnVal
}

func Run(ic *IntCode, input chan int, output chan int, haltSignal chan int, errorChan chan string, wg *sync.WaitGroup, debugFile string) {
	ic.programPos = 0
	ic.relativeBase = 0

	debug := false

	var f *os.File = nil
	var err error = nil

	if debugFile != "" {
		f, err = os.OpenFile(debugFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(f)
			debug = true
		}
	}

	defer func() {
		if f != nil {
			f.Close()
		}
		wg.Done()
	}()

	for {
		fullOp := readNextAddr(ic)
		op := fullOp % 100

		switch op {
		case op_sum:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == mode_rel {
				outAddr += ic.relativeBase
			}

			if debug {
				log.Printf("[%v, %v] OP_SUM (%v mode %v, %v mode %v, %v mode %v) %v + %v => 0x%v", ic.programPos, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, param3, param3Mode, val1, val2, outAddr)
			}

			err := Set(ic, outAddr, val1+val2)
			if err != nil {
				haltSignal <- 1
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				errorChan <- errorMsg
				return
			}

		case op_mul:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == mode_rel {
				outAddr += ic.relativeBase
			}

			if debug {
				log.Printf("[%v, %v] OP_MUL (%v mode %v, %v mode %v, %v mode %v) %v * %v => 0x%v", ic.programPos-4, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, param3, param3Mode, val1, val2, outAddr)
			}

			err := Set(ic, outAddr, val1*val2)
			if err != nil {
				haltSignal <- 1
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				errorChan <- errorMsg
				return
			}

		case op_inp:
			param1 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)

			outAddr := param1
			if param1Mode == mode_rel {
				outAddr += ic.relativeBase
			}

			val := <-input

			if debug {
				log.Printf("[%v, %v] OP_INP (%v mode %v) %v => 0x%v", ic.programPos-2, ic.relativeBase,
					param1, param1Mode, val, outAddr)
			}

			err := Set(ic, outAddr, val)
			if err != nil {
				haltSignal <- 1
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param1, ic.programPos-2, err)
				errorChan <- errorMsg
				return
			}

		case op_out:
			param1 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)

			val1 := getParamValue(ic, param1, param1Mode)

			if debug {
				log.Printf("[%v, %v] OP_OUT (%v mode %v) %v => output", ic.programPos-2, ic.relativeBase,
					param1, param1Mode, val1)
			}

			output <- val1

		case op_jpt:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			if debug {
				log.Printf("[%v, %v] OP_JPT (%v mode %v, %v mode %v) jump to 0x%v if %v != 0", ic.programPos-3, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, val2, val1)
			}

			if val1 != 0 {
				ic.programPos = val2
			}

		case op_jpf:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			if debug {
				log.Printf("[%v, %v] OP_JPF (%v mode %v, %v mode %v) jump to 0x%v if %v == 0", ic.programPos-3, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, val2, val1)
			}

			if val1 == 0 {
				ic.programPos = val2
			}

		case op_lst:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == mode_rel {
				outAddr += ic.relativeBase
			}

			outValue := 0
			if val1 < val2 {
				outValue = 1
			}

			if debug {
				log.Printf("[%v, %v] OP_LST (%v mode %v, %v mode %v, %v mode %v) input %v into 0x%v", ic.programPos-4, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, param3, param3Mode, outValue, outAddr)
			}

			err := Set(ic, outAddr, outValue)
			if err != nil {
				haltSignal <- 1
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				errorChan <- errorMsg
				return
			}

		case op_equ:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == mode_rel {
				outAddr += ic.relativeBase
			}

			outValue := 0
			if val1 == val2 {
				outValue = 1
			}

			if debug {
				log.Printf("[%v, %v] OP_EQU (%v mode %v, %v mode %v, %v mode %v) input %v into 0x%v", ic.programPos-4, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, param3, param3Mode, outValue, outAddr)
			}

			err := Set(ic, outAddr, outValue)
			if err != nil {
				haltSignal <- 1
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				errorChan <- errorMsg
				return
			}

		case op_hlt:
			haltSignal <- 0
			if debug {
				log.Printf("[%v, %v] OP_HLT", ic.programPos-1, ic.relativeBase)
			}
			return

		case op_rbs:
			param1 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)

			val1 := getParamValue(ic, param1, param1Mode)

			if debug {
				log.Printf("[%v, %v] OP_RBS (%v mode %v) %v => relativeBase", ic.programPos-2, ic.relativeBase,
					param1, param1Mode, val1)
			}

			ic.relativeBase += val1

		default:
			haltSignal <- 1
			errorMsg := fmt.Sprintf("Unknown operation %v at address %v", op, ic.programPos-1)
			errorChan <- errorMsg
			return
		}
	}

	haltSignal <- 1
	errorMsg := fmt.Sprintf("Invalid HLT")
	errorChan <- errorMsg
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

	mode := op % 10
	if mode == 0 {
		return mode_pos
	} else if mode == 1 {
		return mode_val
	} else {
		return mode_rel
	}
}
