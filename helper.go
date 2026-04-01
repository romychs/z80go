package z80go

import log "github.com/sirupsen/logrus"

func (z *CPU) flags() FlagsType {
	return FlagsType{
		S: z.Flags.S,
		Z: z.Flags.Z,
		Y: z.Flags.Y,
		H: z.Flags.H,
		X: z.Flags.X,
		P: z.Flags.P,
		N: z.Flags.N,
		C: z.Flags.C,
	}
}

func (z *CPU) altFlags() FlagsType {
	return FlagsType{
		S: z.FlagsAlt.S,
		Z: z.FlagsAlt.Z,
		Y: z.FlagsAlt.Y,
		H: z.FlagsAlt.H,
		X: z.FlagsAlt.X,
		P: z.FlagsAlt.P,
		N: z.FlagsAlt.N,
		C: z.FlagsAlt.C,
	}
}

func (z *CPU) rb(addr uint16) byte {
	z.memAccess[addr] = MemAccessRead
	return z.core.MemRead(addr)
}

func (z *CPU) wb(addr uint16, val byte) {
	z.memAccess[addr] = MemAccessWrite
	z.core.MemWrite(addr, val)
}

func (z *CPU) rw(addr uint16) uint16 {
	z.memAccess[addr] = MemAccessRead
	z.memAccess[addr+1] = MemAccessRead
	return (uint16(z.core.MemRead(addr+1)) << 8) | uint16(z.core.MemRead(addr))
}

func (z *CPU) ww(addr uint16, val uint16) {
	z.memAccess[addr] = MemAccessWrite
	z.memAccess[addr+1] = MemAccessWrite
	z.core.MemWrite(addr, byte(val))
	z.core.MemWrite(addr+1, byte(val>>8))
}

func (z *CPU) pushW(val uint16) {
	z.SP -= 2
	z.ww(z.SP, val)
	z.extendedStack[z.SP] = PushValueTypePush
}

func (z *CPU) popW() uint16 {
	z.SP += 2
	return z.rw(z.SP - 2)
}

func (z *CPU) nextB() byte {
	b := z.core.MemRead(z.PC)
	z.PC++
	return b
}

func (z *CPU) nextW() uint16 {
	w := (uint16(z.core.MemRead(z.PC+1)) << 8) | uint16(z.core.MemRead(z.PC))
	z.PC += 2
	return w
}

func (z *CPU) bc() uint16 {
	return (uint16(z.B) << 8) | uint16(z.C)
}

func (z *CPU) de() uint16 {
	return (uint16(z.D) << 8) | uint16(z.E)
}

func (z *CPU) hl() uint16 {
	return (uint16(z.H) << 8) | uint16(z.L)
}

func (z *CPU) setBC(val uint16) {
	z.B = byte(val >> 8)
	z.C = byte(val)
}

func (z *CPU) setDE(val uint16) {
	z.D = byte(val >> 8)
	z.E = byte(val)
}

func (z *CPU) setHL(val uint16) {
	z.H = byte(val >> 8)
	z.L = byte(val)
}

func (z *CPU) f() byte {
	val := byte(0)
	if z.Flags.C {
		val |= 0x01
	}
	if z.Flags.N {
		val |= 0x02
	}
	if z.Flags.P {
		val |= 0x04
	}
	if z.Flags.X {
		val |= 0x08
	}
	if z.Flags.H {
		val |= 0x10
	}
	if z.Flags.Y {
		val |= 0x20
	}
	if z.Flags.Z {
		val |= 0x40
	}
	if z.Flags.S {
		val |= 0x80
	}
	return val
}

func (z *CPU) setF(val byte) {
	z.Flags.C = val&1 != 0
	z.Flags.N = (val>>1)&1 != 0
	z.Flags.P = (val>>2)&1 != 0
	z.Flags.X = (val>>3)&1 != 0
	z.Flags.H = (val>>4)&1 != 0
	z.Flags.Y = (val>>5)&1 != 0
	z.Flags.Z = (val>>6)&1 != 0
	z.Flags.S = (val>>7)&1 != 0
}

// increments R, keeping the highest byte intact
func (z *CPU) incR() {
	z.R = (z.R & 0x80) | ((z.R + 1) & 0x7f)
}

func boolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// returns if there was a carry between bit "bit_no" and "bit_no - 1" when
// executing "a + b + cy"
func carry(bitNo int, a uint16, b uint16, cy bool) bool {
	result := int32(a) + int32(b) + boolToInt32(cy)
	carry := result ^ int32(a) ^ int32(b)
	return (carry & (1 << bitNo)) != 0
}

// parity returns the parity of byte: 0 if odd, else 1
func parity(val byte) bool {
	ones := byte(0)
	for i := 0; i < 8; i++ {
		ones += (val >> i) & 1
	}
	return (ones & 1) == 0
}

// updateXY set undocumented 3rd (X) and 5th (Y) flags
func (z *CPU) updateXY(result byte) {
	z.Flags.Y = result&0x20 != 0
	z.Flags.X = result&0x08 != 0
}

func (z *CPU) DebugOutput() {
	log.Debugf("PC: %04X, AF: %04X, BC: %04X, DE: %04X, HL: %04X, SP: %04X, IX: %04X, IY: %04X, I: %02X, R: %02X",
		z.PC, (uint16(z.A)<<8)|uint16(z.f()), z.bc(), z.de(), z.hl(), z.SP,
		z.IX, z.IY, z.I, z.R)

	log.Debugf("\t(%02X %02X %02X %02X), cycleCount: %d\n", z.rb(z.PC), z.rb(z.PC+1),
		z.rb(z.PC+2), z.rb(z.PC+3), z.cycleCount)
}

func (z *CPU) Reset() {
	z.cycleCount = 0
	z.PC = 0
	z.SP = 0xFFFF
	z.IX = 0
	z.IY = 0
	z.memPtr = 0

	z.A = 0xFF
	z.B = 0
	z.C = 0
	z.D = 0
	z.E = 0
	z.H = 0
	z.L = 0

	z.AAlt = 0
	z.BAlt = 0
	z.CAlt = 0
	z.DAlt = 0
	z.EAlt = 0
	z.HAlt = 0
	z.LAlt = 0

	z.I = 0
	z.R = 0

	z.Flags.SetFlags(0xff)
	z.FlagsAlt.SetFlags(0xff)

	z.iffDelay = 0
	z.IMode = 0
	z.Iff1 = false
	z.Iff2 = false
	z.Halted = false
	z.IntOccurred = false
	z.NmiOccurred = false
	z.intData = 0
}
