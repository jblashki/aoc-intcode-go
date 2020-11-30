package intcode

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestProgram1(t *testing.T) {
	err := testProgram("./test_input/TstProg1", 0, 2)
	if err != nil {
		t.Fatalf(`TestProgram1: returned error: %v`, err)
	}
}

func TestProgram2(t *testing.T) {
	err := testProgram("./test_input/TstProg2", 3, 6)
	if err != nil {
		t.Fatalf(`TestProgram2: returned error: %v`, err)
	}
}

func TestProgram3(t *testing.T) {
	err := testProgram("./test_input/TstProg3", 5, 9801)
	if err != nil {
		t.Fatalf(`TestProgram3: returned error: %v`, err)
	}
}

func TestProgram4(t *testing.T) {
	err := testProgramMultipleOutput("./test_input/TstProg4", []int{0, 4}, []int{30, 2})
	if err != nil {
		t.Fatalf(`TestProgram4: returned error: %v`, err)
	}
}

func TestProgramParamMode(t *testing.T) {
	err := testProgram("./test_input/TstProgParamMode", 4, 99)
	if err != nil {
		t.Fatalf(`TestProgram4: returned error: %v`, err)
	}
}

func TestProgramEq1(t *testing.T) {
	err := testProgram("./test_input/TstProgEq1", 5, 0)
	if err != nil {
		t.Fatalf(`TestProgramEq1: returned error: %v`, err)
	}
}

func TestProgramEq2(t *testing.T) {
	err := testProgram("./test_input/TstProgEq2", 5, 1)
	if err != nil {
		t.Fatalf(`TestProgramEq2: returned error: %v`, err)
	}
}

func TestProgramEq3(t *testing.T) {
	err := testProgram("./test_input/TstProgEq3", 1, 0)
	if err != nil {
		t.Fatalf(`TestProgramEq3: returned error: %v`, err)
	}
}

func TestProgramEq4(t *testing.T) {
	err := testProgram("./test_input/TstProgEq4", 1, 1)
	if err != nil {
		t.Fatalf(`TestProgramEq4: returned error: %v`, err)
	}
}

func TestProgramLt1(t *testing.T) {
	err := testProgram("./test_input/TstProgLt1", 5, 1)
	if err != nil {
		t.Fatalf(`TestProgramLt1: returned error: %v`, err)
	}
}

func TestProgramLt2(t *testing.T) {
	err := testProgram("./test_input/TstProgLt2", 5, 0)
	if err != nil {
		t.Fatalf(`TestProgramLt2: returned error: %v`, err)
	}
}

func TestProgramLt3(t *testing.T) {
	err := testProgram("./test_input/TstProgLt3", 1, 0)
	if err != nil {
		t.Fatalf(`TestProgramLt3: returned error: %v`, err)
	}
}

func TestProgramLt4(t *testing.T) {
	err := testProgram("./test_input/TstProgLt4", 1, 1)
	if err != nil {
		t.Fatalf(`TestProgramLt4: returned error: %v`, err)
	}
}

func TestProgramJmp1(t *testing.T) {
	err := testProgram("./test_input/TstProgJmp1", 13, 20)
	if err != nil {
		t.Fatalf(`TestProgramJmp1: returned error: %v`, err)
	}
}

func TestProgramJmp2(t *testing.T) {
	err := testProgram("./test_input/TstProgJmp2", 13, 9)
	if err != nil {
		t.Fatalf(`TestProgramJmp2: returned error: %v`, err)
	}
}

func TestProgramJmp3(t *testing.T) {
	err := testProgram("./test_input/TstProgJmp3", 8, 0)
	if err != nil {
		t.Fatalf(`TestProgramJmp3: returned error: %v`, err)
	}
}

func TestProgramJmp4(t *testing.T) {
	err := testProgram("./test_input/TstProgJmp4", 8, 1)
	if err != nil {
		t.Fatalf(`TestProgramJmp4: returned error: %v`, err)
	}
}

func TestProgramInputOutput(t *testing.T) {
	err := testInputOutputProgram("./test_input/TstProgInputOutput", []int{10}, []int{10})
	if err != nil {
		t.Fatalf(`TestProgramInputOutput: returned error: %v`, err)
	}
}

func TestProgramInputOutput2(t *testing.T) {
	err := testInputOutputProgram("./test_input/TstProgInputOutput2", []int{5, 3, 10, 4}, []int{8, 40})
	if err != nil {
		t.Fatalf(`TestProgramInputOutput: returned error: %v`, err)
	}
}

func testProgramMultipleOutput(progFile string, outAddrs []int, wantResults []int) error {
	ic := Create()

	err := Load(ic, progFile)
	if err != nil {
		errormsg := fmt.Sprintf("Failed to load program: %v", err)
		return errors.New(errormsg)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	input := make(chan int)
	output := make(chan int)
	halt := make(chan int, 1)

	defer func() {
		close(input)
		close(output)
		close(halt)
	}()

	go Run(ic, input, output, halt, wg)

	wg.Wait()

	for i := 0; i < len(outAddrs); i++ {
		value := Get(ic, outAddrs[i])

		if value != wantResults[i] {
			errormsg := fmt.Sprintf("Program returned %v @ address %v, want %v", value, outAddrs[i], wantResults[i])
			return errors.New(errormsg)
		}
	}

	return nil
}

func testProgram(progFile string, outAddr int, wantResult int) error {
	ic := Create()

	err := Load(ic, progFile)
	if err != nil {
		errormsg := fmt.Sprintf("Failed to load program: %v", err)
		return errors.New(errormsg)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	input := make(chan int)
	output := make(chan int)
	halt := make(chan int, 1)

	defer func() {
		close(input)
		close(output)
		close(halt)
	}()

	go Run(ic, input, output, halt, wg)

	wg.Wait()

	retValue := Get(ic, outAddr)

	if retValue != wantResult {
		errormsg := fmt.Sprintf("Program returned %v, want %v", retValue, wantResult)
		return errors.New(errormsg)
	}

	return nil
}

func testInputOutputProgram(progFile string, input []int, wantOutput []int) error {
	ic := Create()

	err := Load(ic, progFile)
	if err != nil {
		errormsg := fmt.Sprintf("Failed to load program: %v", err)
		return errors.New(errormsg)
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	inputChan := make(chan int, len(input))
	outputChan := make(chan int, len(wantOutput))
	haltSignalChan := make(chan int, 1)

	defer func() {
		close(inputChan)
		close(outputChan)
		close(haltSignalChan)
	}()

	for i := 0; i < len(input); i++ {
		inputChan <- input[i]
	}

	go Run(ic, inputChan, outputChan, haltSignalChan, wg)

	wg.Wait()

	for i := 0; i < len(wantOutput); i++ {
		select {
		case value := <-outputChan:
			if value != wantOutput[i] {
				errormsg := fmt.Sprintf("Program returned %v @ %v, want %v", value, i, wantOutput[i])
				return errors.New(errormsg)
			}
		case _ = <-haltSignalChan:
			errormsg := fmt.Sprintf("Revieved halt signal when expecting result @ address %v in test %v\n", i, progFile)
			return errors.New(errormsg)

		}
	}

	return nil
}
