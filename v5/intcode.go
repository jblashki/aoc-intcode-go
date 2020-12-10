package intcode

import (
	"fmt"
	"log"
	"os"
	"sync"

	filereader "github.com/jblashki/aoc-filereader-go"
)

//////////////////////
// Consts and types //
//////////////////////

const opSum = 1
const opMul = 2
const opInp = 3
const opOut = 4
const opJpt = 5
const opJpf = 6
const opLst = 7
const opEqu = 8
const opRbs = 9
const opHlt = 99

type paramMode int

const (
	modePos = iota // Postion mode i.e. address
	modeVal        // Value mode i.e. absolute value
	modeRel        // Relative mode i.e. relative address
)

//Signal is signal value returned by read/write methods
type Signal int

const (
	// SigNone means no signal was recieived
	SigNone Signal = iota
	// SigInput means the intcode program requires input
	SigInput
	// SigHalt means intcode program halted successfully
	SigHalt
	// SigError means the intcode program halted with an error which can be read off of error channel
	SigError
)

////////////////////////
// Exported functions //
////////////////////////

// IntCode is the main intcode structure used to define an intcode computer
type IntCode struct {
	memory       []int
	programPos   int
	relativeBase int
	inputChan    chan int
	outputChan   chan int
	signalChan   chan Signal
	errorChan    chan string
	wg           *sync.WaitGroup
	moribund     bool
}

// Create creates a new intcode computer
func Create(wg *sync.WaitGroup, inputBufSize int, outputBufSize int) *IntCode {
	newIC := new(IntCode)

	newIC.memory = make([]int, 0)
	newIC.programPos = 0
	newIC.relativeBase = 0

	if inputBufSize > 0 {
		newIC.inputChan = make(chan int, inputBufSize)
	} else {
		newIC.inputChan = make(chan int)
	}
	if outputBufSize > 0 {
		newIC.outputChan = make(chan int, outputBufSize)
	} else {
		newIC.outputChan = make(chan int)
	}
	newIC.signalChan = make(chan Signal, 1)
	newIC.errorChan = make(chan string, 1)
	newIC.wg = wg
	newIC.moribund = false

	return newIC
}

// CreateLoad creates a new intcode and loads program from filename
func CreateLoad(wg *sync.WaitGroup, filename string, inputBufSize int, outputBufSize int) (*IntCode, error) {
	returnIC := Create(wg, inputBufSize, outputBufSize)

	err := Load(returnIC, filename)
	if err != nil {
		return nil, err
	}

	return returnIC, nil
}

// Close closes and cleans up intcode
func Close(ic *IntCode) {
	ic.moribund = true
	close(ic.inputChan)
	close(ic.outputChan)
	close(ic.signalChan)
	close(ic.errorChan)
}

// Copy does a deep copy of an intcode computer
func Copy(sourceIC *IntCode) *IntCode {
	copiedIC := *sourceIC

	copiedIC.memory = make([]int, len(sourceIC.memory))
	copy(copiedIC.memory, sourceIC.memory)

	return &copiedIC
}

// Set sets an address in an intcode to a specific value
func Set(ic *IntCode, addr int, value int) error {
	if addr >= len(ic.memory) {
		newSpace := addr - len(ic.memory) + 1
		newMem := make([]int, newSpace)
		ic.memory = append(ic.memory, newMem...)
	}

	ic.memory[addr] = value

	return nil
}

// Get returns the value at a specific address in an intocode
func Get(ic *IntCode, addr int) int {
	if addr >= len(ic.memory) {
		return 0
	}

	return ic.memory[addr]
}

// Run runs a specific int code
func Run(ic *IntCode, debugFile string) {
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
		ic.wg.Done()
	}()

	for {
		if ic.moribund {
			return
		}
		fullOp := readNextAddr(ic)
		op := fullOp % 100

		switch op {
		case opSum:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == modeRel {
				outAddr += ic.relativeBase
			}

			if debug {
				log.Printf("[%v, %v] OP_SUM (%v mode %v, %v mode %v, %v mode %v) %v + %v => 0x%v", ic.programPos, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, param3, param3Mode, val1, val2, outAddr)
			}

			err := Set(ic, outAddr, val1+val2)
			if err != nil {
				ic.signalChan <- SigError
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				ic.errorChan <- errorMsg
				return
			}

		case opMul:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == modeRel {
				outAddr += ic.relativeBase
			}

			if debug {
				log.Printf("[%v, %v] OP_MUL (%v mode %v, %v mode %v, %v mode %v) %v * %v => 0x%v", ic.programPos-4, ic.relativeBase,
					param1, param1Mode, param2, param2Mode, param3, param3Mode, val1, val2, outAddr)
			}

			err := Set(ic, outAddr, val1*val2)
			if err != nil {
				ic.signalChan <- SigError
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				ic.errorChan <- errorMsg
				return
			}

		case opInp:
			param1 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)

			outAddr := param1
			if param1Mode == modeRel {
				outAddr += ic.relativeBase
			}
			// Signal That input is required
			ic.signalChan <- SigInput

			// Get Input
			val := <-ic.inputChan

			if debug {
				log.Printf("[%v, %v] OP_INP (%v mode %v) %v => 0x%v", ic.programPos-2, ic.relativeBase,
					param1, param1Mode, val, outAddr)
			}

			err := Set(ic, outAddr, val)
			if err != nil {
				ic.signalChan <- SigError
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param1, ic.programPos-2, err)
				ic.errorChan <- errorMsg
				return
			}

		case opOut:
			param1 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)

			val1 := getParamValue(ic, param1, param1Mode)

			if debug {
				log.Printf("[%v, %v] OP_OUT (%v mode %v) %v => output", ic.programPos-2, ic.relativeBase,
					param1, param1Mode, val1)
			}

			ic.outputChan <- val1

		case opJpt:
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

		case opJpf:
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

		case opLst:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == modeRel {
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
				ic.signalChan <- SigError
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				ic.errorChan <- errorMsg
				return
			}

		case opEqu:
			param1 := readNextAddr(ic)
			param2 := readNextAddr(ic)
			param3 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)
			param2Mode := getParamMode(fullOp, 1)
			param3Mode := getParamMode(fullOp, 2)

			val1 := getParamValue(ic, param1, param1Mode)
			val2 := getParamValue(ic, param2, param2Mode)

			outAddr := param3
			if param3Mode == modeRel {
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
				ic.signalChan <- SigError
				errorMsg := fmt.Sprintf("Error setting address %v @ address %v: %v", param3, ic.programPos-4, err)
				ic.errorChan <- errorMsg
				return
			}

		case opHlt:
			ic.signalChan <- SigHalt
			if debug {
				log.Printf("[%v, %v] OP_HLT", ic.programPos-1, ic.relativeBase)
			}
			return

		case opRbs:
			param1 := readNextAddr(ic)
			param1Mode := getParamMode(fullOp, 0)

			val1 := getParamValue(ic, param1, param1Mode)

			if debug {
				log.Printf("[%v, %v] OP_RBS (%v mode %v) %v => relativeBase", ic.programPos-2, ic.relativeBase,
					param1, param1Mode, val1)
			}

			ic.relativeBase += val1

		default:
			ic.signalChan <- SigError
			errorMsg := fmt.Sprintf("Unknown operation %v at address %v", op, ic.programPos-1)
			ic.errorChan <- errorMsg
			return
		}
	}
}

// Load loads an intcode with data from the file specificed
func Load(ic *IntCode, file string) error {
	var err error

	ic.memory, err = filereader.ReadCSVInts(file)

	return err
}

// Read reads value from intcode output. Will value or signal recieved and error if present
func Read(ic *IntCode) (value int, sig Signal, err error) {
	select {
	case value = <-ic.outputChan:
		sig = SigNone
		err = nil

	case sig = <-ic.signalChan:
		if sig == SigError {
			errOutput := <-ic.errorChan
			err = fmt.Errorf("Program error: %v", errOutput)
		}
	}

	return
}

// Write writes input to the intcode. Will return signal recieved and error if present
//func Write(ic *IntCode, input int) (sig Signal, err error) {
func Write(ic *IntCode, input int) {
	// sig = <-ic.signalChan
	// if sig == SigError {
	// 	errOutput := <-ic.errorChan
	// 	err = fmt.Errorf("Program error: %v", errOutput)
	// } else if sig == SigInput {
	ic.inputChan <- input
	// 	sig = SigNone
	// 	err = nil
	// }

	// select {
	// case ic.inputChan <- input:
	// 	sig = SigNone
	// 	err = nil

	// case sig = <-ic.signalChan:
	// 	if sig == SigError {
	// 		errOutput := <-ic.errorChan
	// 		err = fmt.Errorf("Program error: %v", errOutput)
	// 	} else if sig == SigInput {
	// 		ic.inputChan <- input
	// 		sig = SigNone
	// 		err = nil
	// 	}
	// }

	return
}

//////////////////////////
// Unexported functions //
//////////////////////////

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
		return modePos
	} else if mode == 1 {
		return modeVal
	} else {
		return modeRel
	}
}

func getParamValue(ic *IntCode, param int, mode paramMode) int {
	returnVal := param
	if mode == modePos {
		returnVal = Get(ic, param)
	} else if mode == modeRel {
		returnVal = Get(ic, ic.relativeBase+param)
	}

	return returnVal
}
