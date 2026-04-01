package z80go_test

import (
	"bufio"
	"bytes"
	_ "embed"
	"strconv"
	"strings"
	"testing"

	"github.com/romychs/z80go"
	log "github.com/sirupsen/logrus"
)

const (
	ScanNone int = iota
	ScanDesc
	ScanEvent
	ScanRegs
	ScanState
	ScanMem
	ScanEnd
)

type Registers struct {
	AF     uint16
	BC     uint16
	DE     uint16
	HL     uint16
	AFa    uint16
	BCa    uint16
	DEa    uint16
	HLa    uint16
	IX     uint16
	IY     uint16
	SP     uint16
	PC     uint16
	MemPtr uint16
}

type State struct {
	I        byte
	R        byte
	IFF1     bool
	IFF2     bool
	IM       byte
	isHalted bool
	tStates  uint16
}

type MemorySetup struct {
	addr   uint16
	values []byte
}

type Event struct {
	time uint16
	typ  string
	addr uint16
	data byte
}

type Z80TestIn struct {
	//	description string
	registers   Registers
	state       State
	memorySetup []MemorySetup
}

type Expect struct {
	events    []Event
	registers Registers
	state     State
	memory    []MemorySetup
}

//go:embed tests/tests.in
var testIn []byte

//go:embed tests/tests.expected
var testExpected []byte

type Computer struct {
	cpu    *z80go.CPU
	memory [65536]byte
	ports  [256]byte
}

var z80TestsIn map[string]Z80TestIn
var z80TestsExpected map[string]Expect

var computer Computer

var testNames []string

func init() {
	z80TestsIn = make(map[string]Z80TestIn)
	z80TestsExpected = make(map[string]Expect)
	parseTestIn()
	parseTestExpected()
	for addr := 0; addr < 65535; addr++ {
		computer.memory[addr] = 0x00
	}
	for addr := 0; addr < 255; addr++ {
		computer.ports[addr] = 0
	}
	computer.cpu = z80go.NewCPU(&computer)
}

func (c *Computer) M1MemRead(addr uint16) byte {
	return c.memory[addr]
}

func (c *Computer) MemRead(addr uint16) byte {
	return c.memory[addr]
}

func (c *Computer) MemWrite(addr uint16, val byte) {
	c.memory[addr] = val
}

func (c *Computer) IOWrite(port uint16, val byte) {
	c.ports[port&0x00ff] = val
}

func (c *Computer) IORead(port uint16) byte {
	return byte(port >> 8) //c.ports[port&0x00ff]
}

func parseTestExpected() {
	exScanner := bufio.NewScanner(bytes.NewReader(testExpected))
	scanState := ScanNone

	testName := ""
	var events []Event
	registers := Registers{}
	state := State{}
	var memorySetup []MemorySetup

	for exScanner.Scan() {
		line := exScanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			if scanState == ScanMem {
				scanState = ScanEnd
			} else {
				scanState = ScanNone
				continue
			}
		}
		if ScanNone == scanState {
			scanState = ScanDesc
		} else if len(line) > 0 && line[0] == ' ' {
			scanState = ScanEvent
		}
		//else {
		//	if scanState == ScanEvent {
		//		scanState = ScanRegs
		//	}
		//}

		switch scanState {
		case ScanDesc:
			testName = line
			scanState = ScanRegs
		case ScanEvent:
			events = append([]Event{}, *parseEvent(line))
			scanState = ScanRegs
		case ScanRegs:
			registers = *parseRegisters(line)
			scanState = ScanState
		case ScanState:
			state = *parseState(line)
			scanState = ScanMem
		case ScanMem:
			memorySetup = append(memorySetup, *parseMemory(line))
			//scanState = ScanMem
		case ScanEnd:
			z80TestsExpected[testName] = Expect{
				events:    events,
				registers: registers,
				state:     state,
				memory:    memorySetup,
			}
			events = []Event{}
			memorySetup = []MemorySetup{}
			scanState = ScanNone
		default:
			panic("unhandled default case")
		}

	}
}

func parseTestIn() {
	inScanner := bufio.NewScanner(bytes.NewReader(testIn))
	scanState := ScanNone
	testName := ""
	registers := Registers{}
	state := State{}
	var memorySetup []MemorySetup
	for inScanner.Scan() {
		line := inScanner.Text()
		if len(line) == 0 || strings.TrimSpace(line) == "" {
			scanState = ScanNone
			continue
		}
		if ScanNone == scanState {
			scanState = ScanDesc
		} else if line == "-1" {
			scanState = ScanEnd
		}
		switch scanState {
		case ScanDesc:
			testName = line
			scanState = ScanRegs
		case ScanRegs:
			registers = *parseRegisters(line)
			scanState = ScanState
		case ScanState:
			state = *parseState(line)
			scanState = ScanMem
		case ScanMem:
			memorySetup = append(memorySetup, *parseMemory(line))
			//scanState = ScanMem
		case ScanEnd:
			test := Z80TestIn{
				registers:   registers,
				state:       state,
				memorySetup: memorySetup,
			}
			testNames = append(testNames, testName)
			z80TestsIn[testName] = test
			scanState = ScanNone
		default:
			panic("unhandled default case")
		}

	}
}

func parseEvent(event string) *Event {
	e := Event{}
	line := strings.TrimSpace(event)
	//4 MR 0000 00
	//012345678
	if len(line) < 9 {
		log.Errorf("Invalid event: %s", line)
		return nil
	}
	sp := strings.Index(line, " ")
	if sp == -1 {
		log.Errorf("Invalid event: %s", line)
		return nil
	}
	e.time = parseDecW(line[:sp])
	e.typ = line[sp+1 : sp+3]
	e.addr = parseHexW(line[sp+4 : sp+8])
	if len(line) > sp+9 {
		//println("Event: ", line)

		e.data = parseHexB(line[sp+9:])
	}

	return &e
}

func parseMemory(line string) *MemorySetup {
	m := MemorySetup{}
	//0000 00 -1
	//0123456789
	m.addr = parseHexW(line[0:4])
	mem := line[5:]
	for {
		st := mem[:2]
		if st == "-1" {
			break
		}
		m.values = append(m.values, parseHexB(st))
		mem = strings.TrimSpace(mem[2:])
	}
	return &m
}

func parseRegisters(line string) *Registers {
	r := Registers{}
	line = strings.TrimSpace(line)
	if len(line) != 64 {
		log.Errorf("Invalid register line: %s", line)
	} else {
		// 0000 0000 0000 0000 0000 0000 0000 0000 0000 0000 0000 0000 0000
		for ctr := 0; ctr < 13; ctr++ {
			hexString := line[ctr*5 : ctr*5+4]
			v, err := strconv.ParseUint(hexString, 16, 16)
			if err != nil {
				log.Errorf("Invalid register value: %s in line %s", hexString, line)
				break
			}
			switch ctr {
			case 0:
				r.AF = uint16(v)
			case 1:
				r.BC = uint16(v)
			case 2:
				r.DE = uint16(v)
			case 3:
				r.HL = uint16(v)
			case 4:
				r.AFa = uint16(v)
			case 5:
				r.BCa = uint16(v)
			case 6:
				r.DEa = uint16(v)
			case 7:
				r.HLa = uint16(v)
			case 8:
				r.IX = uint16(v)
			case 9:
				r.IY = uint16(v)
			case 10:
				r.SP = uint16(v)
			case 11:
				r.PC = uint16(v)
			case 12:
				r.MemPtr = uint16(v)
			}
		}

	}
	return &r
}

func parseHexB(line string) byte {
	v, err := strconv.ParseUint(line, 16, 8)
	if err != nil {
		log.Errorf("Invalid HexB value: %s", line)
	}
	return byte(v)
}

func parseHexW(line string) uint16 {
	v, err := strconv.ParseUint(line, 16, 16)
	if err != nil {
		log.Errorf("Invalid HexW value: %s", line)
	}
	return uint16(v)
}

func parseDecB(line string) byte {
	v, err := strconv.ParseUint(line, 10, 8)
	if err != nil {
		log.Errorf("Invalid B value: %s", line)
	}
	return byte(v)
}

func parseDecW(line string) uint16 {
	v, err := strconv.ParseUint(line, 10, 16)
	if err != nil {
		log.Errorf("Invalid W value: %s", line)
	}
	return uint16(v)
}

func parseBoolB(line string) bool {
	v, err := strconv.ParseUint(line, 10, 8)
	if err != nil {
		log.Errorf("Invalid state I value: %s", line)
	}
	return v != 0
}

func parseState(line string) *State {
	s := State{}
	line = strings.TrimSpace(line)
	if len(line) < 15 {
		log.Errorf("Invalid state line: %s", line)
	} else {
		//00 00 0 0 0 0     1
		//0123456789012345678
		s.I = parseHexB(line[0:2])
		s.R = parseHexB(line[3:5])
		s.IFF1 = parseBoolB(line[6:7])
		s.IFF2 = parseBoolB(line[8:9])
		s.IM = parseDecB(line[10:11])
		s.isHalted = parseBoolB(line[12:13])
		s.tStates = parseDecW(strings.TrimSpace(line[13:]))
	}
	return &s
}

func TestZ80Fuse(t *testing.T) {
	t.Logf("Fuse-type Z80 emulator test")
	computer.cpu.Reset()
	for _, name := range testNames {
		setComputerState(z80TestsIn[name])
		exp, exists := z80TestsExpected[name]
		if !exists {
			t.Errorf("Expected values for test %s not exists!", name)
			return
		}
		cy := uint32(0)
		for {
			c, _ := computer.cpu.RunInstruction()
			cy += c
			if cy >= uint32(exp.state.tStates) {
				break
			}
		}
		checkComputerState(t, name)
	}
}

func setComputerState(test Z80TestIn) {
	state := z80go.CPU{
		A:           byte(test.registers.AF >> 8),
		B:           byte(test.registers.BC >> 8),
		C:           byte(test.registers.BC),
		D:           byte(test.registers.DE >> 8),
		E:           byte(test.registers.DE),
		H:           byte(test.registers.HL >> 8),
		L:           byte(test.registers.HL),
		AAlt:        byte(test.registers.AFa >> 8),
		BAlt:        byte(test.registers.BCa >> 8),
		CAlt:        byte(test.registers.BCa),
		DAlt:        byte(test.registers.DEa >> 8),
		EAlt:        byte(test.registers.DEa),
		HAlt:        byte(test.registers.HLa >> 8),
		LAlt:        byte(test.registers.HLa),
		IX:          test.registers.IX,
		IY:          test.registers.IY,
		I:           test.state.I,
		R:           test.state.R,
		SP:          test.registers.SP,
		PC:          test.registers.PC,
		Flags:       *z80go.NewFlags(byte(test.registers.AF)),
		FlagsAlt:    *z80go.NewFlags(byte(test.registers.AFa)),
		IMode:       test.state.IM,
		Iff1:        test.state.IFF1,
		Iff2:        test.state.IFF2,
		Halted:      test.state.isHalted,
		CycleCount:  0,
		IntOccurred: false,
		NmiOccurred: false,
		MemPtr:      test.registers.MemPtr,
	}

	// Setup CPU
	computer.cpu.SetState(&state)

	// Setup Memory
	for _, ms := range test.memorySetup {
		addr := ms.addr
		for _, b := range ms.values {
			computer.memory[addr] = b
			addr++
		}
	}
}

func lo(w uint16) byte {
	return byte(w)
}

func hi(w uint16) byte {
	return byte(w >> 8)
}

func checkComputerState(t *testing.T, name string) {
	state := computer.cpu.GetState()
	exp, exists := z80TestsExpected[name]
	if !exists {
		t.Errorf("Expected values for test %s not exists!", name)
		return
	}

	// A,B,C,D,E,H,L
	if hi(exp.registers.AF) != state.A {
		t.Errorf("%s:  Expected A to be %x, got %x", name, hi(exp.registers.AF), state.A)
	}
	if hi(exp.registers.BC) != state.B {
		t.Errorf("%s:  Expected B to be %x, got %x", name, hi(exp.registers.BC), state.B)
		computer.cpu.DebugOutput()
	}

	if lo(exp.registers.BC) != state.C {
		t.Errorf("%s:  Expected C to be %x, got %x", name, lo(exp.registers.BC), state.C)
	}

	if hi(exp.registers.DE) != state.D {
		t.Errorf("%s:  Expected D to be %x, got %x", name, hi(exp.registers.DE), state.D)
	}

	if lo(exp.registers.DE) != state.E {
		t.Errorf("%s:  Expected E to be %x, got %x", name, lo(exp.registers.DE), state.E)
	}

	if hi(exp.registers.HL) != state.H {
		t.Errorf("%s:  Expected H to be %x, got %x", name, hi(exp.registers.HL), state.H)
	}

	if lo(exp.registers.HL) != state.L {
		t.Errorf("%s:  Expected L to be %x, got %x", name, lo(exp.registers.BC), state.L)
	}

	// Alt A,B,C,D,E,H,L
	if hi(exp.registers.AFa) != state.AAlt {
		t.Errorf("%s:  Expected A' to be %x, got %x", name, hi(exp.registers.AFa), state.AAlt)
	}
	if hi(exp.registers.BCa) != state.BAlt {
		t.Errorf("%s:  Expected B' to be %x, got %x", name, hi(exp.registers.BCa), state.BAlt)
	}

	if lo(exp.registers.BCa) != state.CAlt {
		t.Errorf("%s:  Expected C' to be %x, got %x", name, lo(exp.registers.BCa), state.CAlt)
	}

	if hi(exp.registers.DEa) != state.DAlt {
		t.Errorf("%s:  Expected D' to be %x, got %x", name, hi(exp.registers.DEa), state.DAlt)
	}

	if lo(exp.registers.DEa) != state.EAlt {
		t.Errorf("%s:  Expected E' to be %x, got %x", name, lo(exp.registers.DEa), state.EAlt)
	}

	if hi(exp.registers.HLa) != state.HAlt {
		t.Errorf("%s:  Expected H' to be %x, got %x", name, hi(exp.registers.HLa), state.HAlt)
	}

	if lo(exp.registers.HLa) != state.LAlt {
		t.Errorf("%s:  Expected L' to be %x, got %x", name, lo(exp.registers.BCa), state.LAlt)
	}

	// 16b regs PC, SP, Meme, R, I

	if exp.registers.IX != state.IX {
		t.Errorf("%s:  Expected IX to be %x, got %x", name, exp.registers.IX, state.IX)
	}

	if exp.registers.IY != state.IY {
		t.Errorf("%s:  Expected IX to be %x, got %x", name, exp.registers.IX, state.IX)
	}

	if exp.registers.PC != state.PC {
		t.Errorf("%s:  Expected PC to be %x, got %x", name, exp.registers.PC, state.PC)
	}

	if exp.registers.SP != state.SP {
		t.Errorf("%s:  Expected SP to be %x, got %x", name, exp.registers.SP, state.SP)
	}

	if exp.registers.MemPtr != state.MemPtr {
		t.Errorf("%s:  Expected MemPtr to be %x, got %x", name, exp.registers.MemPtr, state.MemPtr)
	}

	// State
	if exp.state.I != state.I {
		t.Errorf("%s:  Expected I to be %x, got %x", name, exp.state.I, state.I)
	}

	if exp.state.IM != state.IMode {
		t.Errorf("%s:  Expected IM to be %d, got %d", name, exp.state.IM, state.IMode)
	}

	if exp.state.IFF1 != state.Iff1 {
		t.Errorf("%s:  Expected IIF1 to be %t, got %t", name, exp.state.IFF1, state.Iff1)
	}

	if exp.state.IFF2 != state.Iff2 {
		t.Errorf("%s:  Expected IIF2 to be %t, got %t", name, exp.state.IFF2, state.Iff2)
	}

	if exp.state.isHalted != state.Halted {
		t.Errorf("%s:  Expected isHalted to be %t, got %t", name, exp.state.isHalted, state.Halted)
	}

	// FLAGS
	if lo(exp.registers.AF) != state.Flags.AsByte() {
		t.Errorf("%s:  Expected Flags to be %08b, got %08b", name, lo(exp.registers.AF), state.Flags.AsByte())
	}

	if lo(exp.registers.AFa) != state.FlagsAlt.AsByte() {
		t.Errorf("%s:  Expected Flags' to be %08b, got %08b", name, lo(exp.registers.AFa), state.FlagsAlt.AsByte())
	}

	// Check memory
	for _, mExpect := range exp.memory {
		addr := mExpect.addr
		for _, b := range mExpect.values {
			if computer.memory[addr] != b {
				t.Errorf("%s:  Expected memory[%x] to be %x, got %x", name, addr, b, computer.memory[addr])
			}
			addr++
		}
	}
}
