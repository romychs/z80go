package dis

import "testing"

var disasm *Disassembler

type TestComp struct {
	memory [65536]byte
}

func (t *TestComp) M1MemRead(addr uint16) byte {
	return t.memory[addr]
}
func (t *TestComp) MemRead(addr uint16) byte {
	return t.memory[addr]
}

func (t *TestComp) MemWrite(addr uint16, val byte) {
	t.memory[addr] = val
}

func (t *TestComp) IORead(port uint16) byte {
	return byte(port >> 8)
}

func (t *TestComp) IOWrite(port uint16, val byte) {
	//
}

var testComp *TestComp

func init() {
	testComp = &TestComp{}
	for i := 0; i < 65536; i++ {
		testComp.memory[i] = 0x3f
	}
	disasm = NewDisassembler(testComp)
}

func setMemory(addr uint16, value []byte) {
	for i := 0; i < len(value); i++ {
		testComp.memory[addr+uint16(i)] = value[i]
	}
}

var test = []byte{0x31, 0x2c, 0x05, 0x11, 0x0e, 0x01, 0x0e, 0x09, 0xcd, 0x05, 0x00, 0xc3, 0x00, 0x00}

func Test_LD_SP_nn(t *testing.T) {
	t.Logf("Disassembler Z80 test")
	expected := "  0100 LD SP, 0x052C"
	setMemory(0x100, test)
	res := disasm.Disassm(0x100)
	if res != expected {
		t.Errorf("Error disasm LD SP, nn, result '%s', expected '%s'", res, expected)
	}
}

func Test_LD_DE_nn(t *testing.T) {
	expected := "  0103 LD DE, 0x010E"
	setMemory(0x100, test)
	res := disasm.Disassm(0x103)
	if res != expected {
		t.Errorf("Error disasm LD DE, nn, result '%s', expected '%s'", res, expected)
	}
}

func Test_LD_C_n(t *testing.T) {
	expected := "  0106 LD C, 0x09"
	setMemory(0x100, test)
	res := disasm.Disassm(0x106)
	if res != expected {
		t.Errorf("Error disasm LD C, n, result '%s', expected '%s'", res, expected)
	}
}

func Test_CALL_nn(t *testing.T) {
	expected := "  0108 CALL 0x0005"
	setMemory(0x100, test)
	res := disasm.Disassm(0x108)
	if res != expected {
		t.Errorf("Error disasm CALL nn, result '%s', expected '%s'", res, expected)
	}
}

func Test_JP_nn(t *testing.T) {
	expected := "  010B JP 0x0000"
	setMemory(0x100, test)
	res := disasm.Disassm(0x10b)
	if res != expected {
		t.Errorf("Error disasm JP nn, result '%s', expected '%s'", res, expected)
	}
}
