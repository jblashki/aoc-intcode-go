package intcode

import (
	"errors"
	"fmt"
	"regexp"
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

func TestProgramInvalidOp(t *testing.T) {
	err := testProgram("./test_input/TstProgInvalidOp", 0, 30)
	if err == nil {
		t.Fatalf(`TestProgramInvalidOp: failed to return error on invalid operation`)
	}

	want := regexp.MustCompile(`Unknown operation 98 at address 0`)

	if !want.MatchString(err.Error()) {
		t.Fatalf(`TestProgramInvalidOp: error: %q, want match for %#q`, err.Error(), want)
	}
}

func testProgramMultipleOutput(progFile string, outAddrs []int, wantResults []int) error {
	ic := Create()

	err := Load(ic, progFile)
	if err != nil {
		errormsg := fmt.Sprintf("Failed to load program: %v", err)
		return errors.New(errormsg)
	}

	retValue, err := Run(ic, 0)
	if err != nil {
		errormsg := fmt.Sprintf("Failed to run program: %v", err)
		return errors.New(errormsg)
	}

	for i := 0; i < len(outAddrs); i++ {
		value := Get(ic, outAddrs[i])

		if value != wantResults[i] {
			errormsg := fmt.Sprintf("Program returned %v @ address %v, want %v", retValue, outAddrs[i], wantResults[i])
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

	retValue, err := Run(ic, outAddr)
	if err != nil {
		errormsg := fmt.Sprintf("Failed to run program: %v", err)
		return errors.New(errormsg)
	}

	if retValue != wantResult {
		errormsg := fmt.Sprintf("Program returned %v, want %v", retValue, wantResult)
		return errors.New(errormsg)
	}

	return nil
}
