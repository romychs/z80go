package z80go

import (
	"errors"
)

const (
	MemAccessRead  = 1
	MemAccessWrite = 2
)

// NewCPU initializes a Z80 CPU instance and return pointer to it
func NewCPU(core MemIoRW) *CPU {
	z := CPU{}
	z.Reset()
	z.core = core
	z.cycleCount = 0
	z.codeCoverageEnabled = false
	// z.codeCoverage = make(map[uint16]bool)
	// z.memAccess =
	z.codeCoverageEnabled = false
	// z.codeCoverage         map[uint16]bool
	z.extendedStackEnabled = false
	//z.extendedStack        [65536]uint8
	return &z
}

// RunInstruction executes the next instruction in memory + handles interrupts
func (z *CPU) RunInstruction() (uint32, *map[uint16]byte) {
	z.memAccess = map[uint16]byte{}
	if z.codeCoverageEnabled {
		z.codeCoverage[z.PC] = true
	}
	pre := z.cycleCount
	if z.Halted {
		z.execOpcode(0x00)
	} else {
		opcode := z.nextB()
		z.execOpcode(opcode)
	}
	z.processInterrupts()
	return z.cycleCount - pre, &z.memAccess
}

// SetState set current CPU state
// Used by debuggers to override CPU state, set new PC, for example
func (z *CPU) SetState(state *CPU) {
	z.A = state.A
	z.B = state.B
	z.C = state.C
	z.D = state.D
	z.E = state.E
	z.H = state.H
	z.L = state.L

	z.AAlt = state.AAlt
	z.BAlt = state.BAlt
	z.CAlt = state.CAlt
	z.DAlt = state.DAlt
	z.EAlt = state.EAlt
	z.HAlt = state.HAlt
	z.LAlt = state.LAlt

	z.PC = state.PC
	z.SP = state.SP
	z.IX = state.IX
	z.IY = state.IY
	z.I = state.I
	z.R = state.R

	z.Flags.S = state.Flags.S
	z.Flags.Z = state.Flags.Z
	z.Flags.Y = state.Flags.Y
	z.Flags.H = state.Flags.H
	z.Flags.X = state.Flags.X
	z.Flags.P = state.Flags.P
	z.Flags.N = state.Flags.N
	z.Flags.C = state.Flags.C

	z.IMode = state.IMode
	z.Iff1 = state.Iff1
	z.Iff2 = state.Iff2
	z.Halted = state.Halted
	z.IntOccurred = state.IntOccurred
	z.NmiOccurred = false
}
func (z *CPU) GetState() *CPU {
	return &CPU{
		A:    z.A,
		B:    z.B,
		C:    z.C,
		D:    z.D,
		E:    z.E,
		H:    z.H,
		L:    z.L,
		AAlt: z.AAlt,
		BAlt: z.BAlt,
		CAlt: z.CAlt,
		DAlt: z.DAlt,
		EAlt: z.EAlt,
		HAlt: z.HAlt,
		LAlt: z.LAlt,

		IX: z.IX,
		IY: z.IY,
		I:  z.I,
		R:  z.R,
		SP: z.SP,
		PC: z.PC,

		Flags:       z.flags(),
		FlagsAlt:    z.altFlags(),
		IMode:       z.IMode,
		Iff1:        z.Iff1,
		Iff2:        z.Iff2,
		Halted:      z.Halted,
		CycleCount:  z.cycleCount,
		IntOccurred: z.IntOccurred,
		NmiOccurred: z.NmiOccurred,
		memPtr:      z.memPtr,
	}
}

//func (z *CPU) PC() uint16 {
//	return z.PC
//}

// ClearCodeCoverage - clears code coverage journal
func (z *CPU) ClearCodeCoverage() {
	clear(z.codeCoverage)
}

// SetCodeCoverage - enable of disable code coverage journal
func (z *CPU) SetCodeCoverage(enabled bool) {
	z.codeCoverageEnabled = enabled
	if !enabled {
		clear(z.codeCoverage)
	}
}

// CodeCoverage - return list of addresses executed by CPU
func (z *CPU) CodeCoverage() map[uint16]bool {
	return z.codeCoverage
}

// SetExtendedStack - enable or disable marking stack values by PushValueType*
func (z *CPU) SetExtendedStack(enabled bool) {
	z.extendedStackEnabled = enabled
	if enabled {
		for addr := 0; addr < 65536; addr++ {
			z.extendedStack[addr] = PushValueTypeDefault
		}
	}
}

// ExtendedStack - return array with markers of PushValueType* for each byte of memory
func (z *CPU) ExtendedStack() ([]byte, error) {
	var err error
	if !z.extendedStackEnabled {
		err = errors.New("error, z80: ExtendedStack disabled")
	}
	return z.extendedStack[:], err
}
