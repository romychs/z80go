// Package z80go provides basic Z80 CPU emulation
package z80go

const (
	PushValueTypeDefault = iota // Unknown status of word on stack
	PushValueTypeCall           // Return from call address on stack
	PushValueTypeRst            // Return from RST address on stack
	PushValueTypePush           // 16-bit value, PUSH-ed to stack
	//PushValueTypeMaskableInt
	//PushValueTypeNonMaskableInt
)

type PushValueType byte

// MemIoRW interface for CPU to access memory and io ports of emulated computer
type MemIoRW interface {

	//// M1MemRead Read byte from memory for specified address @ M1 cycle
	//M1MemRead(addr uint16) byte

	// MemRead Read byte from memory for specified address
	MemRead(addr uint16) byte

	// MemWrite Write byte to memory to specified address
	MemWrite(addr uint16, val byte)

	// IORead Read byte from specified port
	IORead(port uint16) byte

	// IOWrite Write byte to specified port
	IOWrite(port uint16, val byte)
}

// CPUInterface Interface to CPU emulator
type CPUInterface interface {
	// Reset CPU to initial state
	Reset()
	// RunInstruction Run single instruction, return number of CPU cycles
	RunInstruction() uint32
	// GetState Get current CPU state
	GetState() *CPU
	// SetState Set current CPU state
	SetState(state *CPU)
	// DebugOutput out current CPU state
	DebugOutput()
}

// FlagsType - Processor flags
type FlagsType struct {
	S bool `json:"s,omitempty"` // Sign flag
	Z bool `json:"z,omitempty"` // Zero flag
	Y bool `json:"y,omitempty"` // Y undocumented flag
	H bool `json:"h,omitempty"` // Half carry flag
	X bool `json:"x,omitempty"` // X undocumented flag
	P bool `json:"p,omitempty"` // Parity/Overflow flag
	N bool `json:"n,omitempty"` // Add/Substract flag
	C bool `json:"c,omitempty"` // Carry flag
}

// CPU - Processor state
type CPU struct {
	A                    byte      `json:"a,omitempty"`
	B                    byte      `json:"b,omitempty"`
	C                    byte      `json:"c,omitempty"`
	D                    byte      `json:"d,omitempty"`
	E                    byte      `json:"e,omitempty"`
	H                    byte      `json:"h,omitempty"`
	L                    byte      `json:"l,omitempty"`
	AAlt                 byte      `json:"AAlt,omitempty"`
	BAlt                 byte      `json:"BAlt,omitempty"`
	CAlt                 byte      `json:"CAlt,omitempty"`
	DAlt                 byte      `json:"DAlt,omitempty"`
	EAlt                 byte      `json:"EAlt,omitempty"`
	HAlt                 byte      `json:"HAlt,omitempty"`
	LAlt                 byte      `json:"LAlt,omitempty"`
	IX                   uint16    `json:"IX,omitempty"`
	IY                   uint16    `json:"IY,omitempty"`
	I                    byte      `json:"i,omitempty"`
	R                    byte      `json:"r,omitempty"`
	SP                   uint16    `json:"SP,omitempty"`
	PC                   uint16    `json:"PC,omitempty"`
	Flags                FlagsType `json:"flags"`
	FlagsAlt             FlagsType `json:"flagsAlt"`
	IMode                byte      `json:"IMode,omitempty"`
	Iff1                 bool      `json:"iff1,omitempty"`
	Iff2                 bool      `json:"iff2,omitempty"`
	Halted               bool      `json:"halted,omitempty"`
	CycleCount           uint32    `json:"cycleCount,omitempty"`
	IntOccurred          bool      `json:"intOccurred,omitempty"`
	NmiOccurred          bool      `json:"interruptOccurred,omitempty"`
	memPtr               uint16
	core                 MemIoRW
	intData              byte
	cycleCount           uint32 // cycle count (t-states)
	memAccess            map[uint16]byte
	codeCoverageEnabled  bool
	codeCoverage         map[uint16]bool
	extendedStackEnabled bool
	extendedStack        map[uint16]PushValueType
	iffDelay             byte
	intPending           bool
	nmiPending           bool
}

// Flags - return flags as byte value
// Used to simplify manipulations with AF register from debugger
func (f *FlagsType) Flags() byte {
	var flags byte = 0
	if f.S {
		flags |= 0x80
	}
	if f.Z {
		flags |= 0x40
	}
	if f.Y {
		flags |= 0x20
	}
	if f.H {
		flags |= 0x10
	}
	if f.X {
		flags |= 0x08
	}
	if f.P {
		flags |= 0x04
	}
	if f.N {
		flags |= 0x02
	}
	if f.C {
		flags |= 0x01
	}
	return flags
}

// String - return string representation of CPU flags
func (f *FlagsType) String() string {
	flags := []byte{'-', '-', '-', '-', '-', '-', '-', '-'}
	if f.S {
		flags[0] = 'S'
	}
	if f.Z {
		flags[1] = 'Z'
	}
	if f.Y {
		flags[2] = '5'
	}
	if f.H {
		flags[3] = 'H'
	}
	if f.X {
		flags[4] = '3'
	}
	if f.P {
		flags[5] = 'P'
	}
	if f.N {
		flags[6] = 'N'
	}
	if f.C {
		flags[7] = 'C'
	}
	return string(flags)
}

func (z *CPU) IIFStr() string {
	flags := []byte{'-', '-'}
	if z.Iff1 {
		flags[0] = '1'
	}
	if z.Iff2 {
		flags[1] = '2'
	}
	return string(flags)
}

//// Flags - return state of CPU flags
//func Flags(f byte) FlagsType {
//	return FlagsType{
//		S: f&0x80 != 0,
//		Z: f&0x40 != 0,
//		Y: f&0x20 != 0,
//		H: f&0x10 != 0,
//		X: f&0x08 != 0,
//		P: f&0x04 != 0,
//		N: f&0x02 != 0,
//		C: f&0x01 != 0,
//	}
//}

// SetFlags - set CPU flags by flags byte.
// Used to simplify manipulations with AF register from debugger
func (f *FlagsType) SetFlags(flags byte) {
	f.S = flags&0x80 != 0
	f.Z = flags&0x40 != 0
	f.Y = flags&0x20 != 0
	f.H = flags&0x10 != 0
	f.X = flags&0x08 != 0
	f.P = flags&0x04 != 0
	f.N = flags&0x02 != 0
	f.C = flags&0x01 != 0
}

// GetPC - return PC register value only, for fast breakpoints test from debugger
func (z *CPU) GetPC() uint16 {
	return z.PC
}
