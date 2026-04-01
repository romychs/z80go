package z80go

import log "github.com/sirupsen/logrus"

// jumps to an address
func (z *CPU) jump(addr uint16) {
	z.PC = addr
	z.memPtr = addr
}

// jumps to next word in memory if condition is true
func (z *CPU) condJump(condition bool) {
	addr := z.nextW()
	if condition {
		z.jump(addr)
	}
	z.memPtr = addr
}

// calls to next word in memory
func (z *CPU) call(addr uint16) {
	z.pushW(z.PC)
	z.extendedStack[z.SP] = PushValueTypeCall
	z.PC = addr
	z.memPtr = addr
}

// calls to next word in memory if condition is true
func (z *CPU) condCall(condition bool) {
	addr := z.nextW()
	if condition {
		z.call(addr)
		z.cycleCount += 7
	}
	z.memPtr = addr
}

// returns from subroutine
func (z *CPU) ret() {
	z.PC = z.popW()
	z.memPtr = z.PC
}

// returns from subroutine if condition is true
func (z *CPU) condRet(condition bool) {
	if condition {
		z.ret()
		z.cycleCount += 6
	}
}

func (z *CPU) jr(offset byte) {
	if offset&0x80 != 0 {
		z.PC += 0xFF00 | uint16(offset)
	} else {
		z.PC += uint16(offset)
	}
	z.memPtr = z.PC
}

func (z *CPU) condJr(condition bool) {
	b := z.nextB()
	if condition {
		z.jr(b)
		z.cycleCount += 5
	}
}

func bToByte(cond bool) byte {
	if cond {
		return byte(1)
	}
	return byte(0)
}

// ADD Byte: adds two bytes together
func (z *CPU) addB(a byte, b byte, cy bool) byte {
	result := a + b + bToByte(cy)
	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = carry(4, uint16(a), uint16(b), cy)
	z.Flags.P = carry(7, uint16(a), uint16(b), cy) != carry(8, uint16(a), uint16(b), cy)
	z.Flags.C = carry(8, uint16(a), uint16(b), cy)
	z.Flags.N = false
	z.updateXY(result)
	return result
}

// SUBtract Byte: subtracts two bytes (with optional carry)
func (z *CPU) subB(a byte, b byte, cy bool) byte {
	val := z.addB(a, ^b, !cy)
	z.Flags.C = !z.Flags.C
	z.Flags.H = !z.Flags.H
	z.Flags.N = true
	return val
}

// ADD Word: adds two words together
func (z *CPU) addW(a uint16, b uint16, cy bool) uint16 {
	lsb := z.addB(byte(a), byte(b), cy)
	msb := z.addB(byte(a>>8), byte(b>>8), z.Flags.C)
	result := (uint16(msb) << 8) | uint16(lsb)
	z.Flags.Z = result == 0
	z.memPtr = a + 1
	return result
}

// SUBtract Word: subtracts two words (with optional carry)
func (z *CPU) subW(a uint16, b uint16, cy bool) uint16 {
	lsb := z.subB(byte(a), byte(b), cy)
	msb := z.subB(byte(a>>8), byte(b>>8), z.Flags.C)
	result := (uint16(msb) << 8) | uint16(lsb)
	z.Flags.Z = result == 0
	z.memPtr = a + 1
	return result
}

// Adds A word to HL
func (z *CPU) addHL(val uint16) {
	sf := z.Flags.S
	zf := z.Flags.Z
	pf := z.Flags.P
	result := z.addW(z.hl(), val, false)
	z.setHL(result)
	z.Flags.S = sf
	z.Flags.Z = zf
	z.Flags.P = pf
}

// adds A word to IX or IY
func (z *CPU) addIZ(reg *uint16, val uint16) {
	sf := z.Flags.S
	zf := z.Flags.Z
	pf := z.Flags.P
	result := z.addW(*reg, val, false)
	*reg = result
	z.Flags.S = sf
	z.Flags.Z = zf
	z.Flags.P = pf
}

// adcHL adds A word (+ carry) to HL
func (z *CPU) adcHL(val uint16) {
	result := z.addW(z.hl(), val, z.Flags.C)
	z.Flags.S = result&0x8000 != 0
	z.Flags.Z = result == 0
	z.setHL(result)
}

// sbcHL subtracts A word (+ carry) to HL
func (z *CPU) sbcHL(val uint16) {
	result := z.subW(z.hl(), val, z.Flags.C)
	z.Flags.S = result&0x8000 != 0
	z.Flags.Z = result == 0
	z.setHL(result)
}

// increments A byte value
func (z *CPU) inc(a byte) byte {
	cf := z.Flags.C
	result := z.addB(a, 1, false)
	z.Flags.C = cf
	return result
}

// decrements A byte value
func (z *CPU) dec(a byte) byte {
	cf := z.Flags.C
	result := z.subB(a, 1, false)
	z.Flags.C = cf
	return result
}

// executes A logic "and" between register A and A byte, then stores the
// result in register A
func (z *CPU) lAnd(val byte) {
	result := z.A & val
	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = true
	z.Flags.P = parity(result)
	z.Flags.N = false
	z.Flags.C = false
	z.updateXY(result)
	z.A = result
}

// executes A logic "xor" between register A and A byte, then stores the
// result in register A
func (z *CPU) lXor(val byte) {
	result := z.A ^ val
	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = false
	z.Flags.P = parity(result)
	z.Flags.N = false
	z.Flags.C = false
	z.updateXY(result)
	z.A = result
}

// executes A logic "or" between register A and A byte, then stores the
// result in register A
func (z *CPU) lOr(val byte) {
	result := z.A | val

	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = false
	z.Flags.P = parity(result)
	z.Flags.N = false
	z.Flags.C = false
	z.updateXY(result)
	z.A = result
}

// compares A value with register A
func (z *CPU) cp(val byte) {
	z.subB(z.A, val, false)

	// the only difference between cp and sub is that
	// the xf/yf are taken from the value to be subtracted,
	// not the result

	z.updateXY(val)
}

// 0xCB opcodes
// rotate left with carry
func (z *CPU) cbRlc(val byte) byte {
	old := val >> 7
	val = (val << 1) | old
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.P = parity(val)
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.C = old != 0
	z.updateXY(val)
	return val
}

// rotate right with carry
func (z *CPU) cbRrc(val byte) byte {
	old := val & 1
	val = (val >> 1) | (old << 7)
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.C = old != 0
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// rotate left (simple)
func (z *CPU) cbRl(val byte) byte {
	cf := z.Flags.C
	z.Flags.C = val>>7 != 0
	val = (val << 1) | bToByte(cf)
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// rotate right (simple)
func (z *CPU) cbRr(val byte) byte {
	c := z.Flags.C
	z.Flags.C = (val & 1) != 0
	val = (val >> 1) | (bToByte(c) << 7)
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// shift left preserving sign
func (z *CPU) cbSla(val byte) byte {
	z.Flags.C = (val >> 7) != 0
	val <<= 1
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// SLL (exactly like SLA, but sets the first bit to 1)
func (z *CPU) cbSll(val byte) byte {
	z.Flags.C = val&0x80 != 0
	val <<= 1
	val |= 1
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// shift right preserving sign
func (z *CPU) cbSra(val byte) byte {
	z.Flags.C = (val & 1) != 0
	val = (val >> 1) | (val & 0x80) // 0b10000000
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// shift register right
func (z *CPU) cbSrl(val byte) byte {
	z.Flags.C = (val & 1) != 0
	val >>= 1
	z.Flags.S = val&0x80 != 0
	z.Flags.Z = val == 0
	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = parity(val)
	z.updateXY(val)
	return val
}

// tests bit "n" from A byte
func (z *CPU) cbBit(val byte, n byte) byte {
	result := val & (1 << n)
	z.Flags.S = result&0x80 != 0
	z.Flags.Z = result == 0
	z.Flags.H = true
	z.updateXY(val)
	z.Flags.P = z.Flags.Z
	z.Flags.N = false
	return result
}

func (z *CPU) ldi() {
	de := z.de()
	hl := z.hl()
	val := z.rb(hl)
	z.wb(de, val)

	z.setHL(z.hl() + 1)
	z.setDE(z.de() + 1)
	z.setBC(z.bc() - 1)

	// see https://wikiti.brandonw.net/index.php?title=Z80_Instruction_Set
	// for the calculation of xf/yf on LDI
	result := val + z.A

	z.Flags.X = result&0x08 != 0 // bit 3
	z.Flags.Y = result&0x02 != 0 // bit 1

	z.Flags.N = false
	z.Flags.H = false
	z.Flags.P = z.bc() > 0

}

func (z *CPU) ldd() {
	z.ldi()
	// same as ldi but HL and DE are decremented instead of incremented
	z.setHL(z.hl() - 2)
	z.setDE(z.de() - 2)
}

func (z *CPU) cpi() {
	cf := z.Flags.C
	result := z.subB(z.A, z.rb(z.hl()), false)
	z.setHL(z.hl() + 1)
	z.setBC(z.bc() - 1)

	val := result - bToByte(z.Flags.H)
	z.Flags.X = val&0x08 != 0
	z.Flags.Y = val&0x02 != 0
	z.Flags.P = z.bc() != 0
	z.Flags.C = cf
	z.memPtr += 1
}

func (z *CPU) cpd() {
	z.cpi()
	// same as cpi but HL is decremented instead of incremented
	z.setHL(z.hl() - 2)
	z.memPtr -= 2
}

func (z *CPU) inRC(r *byte) {
	*r = z.core.IORead(z.bc())
	z.Flags.Z = *r == 0
	z.Flags.S = *r&0x80 != 0
	z.Flags.P = parity(*r)
	z.Flags.N = false
	z.Flags.H = false
}

func (z *CPU) ini() {
	val := z.core.IORead(z.bc())
	z.wb(z.hl(), val)
	z.memPtr = z.bc() + 1
	z.B--

	other := val + z.C + 1
	if other < val {
		z.Flags.H = true
		z.Flags.C = true
	} else {
		z.Flags.H = false
		z.Flags.C = false
	}
	z.Flags.N = val&0x80 != 0
	z.Flags.P = parity((other & 0x07) ^ z.B)
	z.Flags.S = z.B&0x80 != 0
	z.Flags.Z = z.B == 0
	z.updateXY(z.B)
	z.setHL(z.hl() + 1)
}

func (z *CPU) ind() {
	val := z.core.IORead(z.bc())
	z.wb(z.hl(), val)
	z.memPtr = z.bc() - 1
	z.B--

	other := val + z.C - 1
	z.Flags.N = val&0x80 != 0
	if other < val {
		z.Flags.H = true
		z.Flags.C = true
	} else {
		z.Flags.H = false
		z.Flags.C = false
	}
	z.Flags.P = parity((other & 0x07) ^ z.B)

	z.Flags.S = z.B&0x80 != 0
	z.Flags.Z = z.B == 0
	z.updateXY(z.B)
	z.setHL(z.hl() - 1)
}

func (z *CPU) outi() {
	val := z.rb(z.hl())
	z.B--
	z.memPtr = z.bc() + 1
	z.core.IOWrite(z.bc(), val)
	z.setHL(z.hl() + 1)
	other := val + z.L
	z.Flags.N = val&0x80 != 0
	if other < val {
		z.Flags.H = true
		z.Flags.C = true
	} else {
		z.Flags.H = false
		z.Flags.C = false
	}
	z.Flags.P = parity((other & 0x07) ^ z.B)
	z.Flags.Z = z.B == 0
	z.Flags.S = z.B&0x80 != 0
	z.updateXY(z.B)
}

func (z *CPU) outd() {
	val := z.rb(z.hl())
	z.B--
	z.memPtr = z.bc() - 1
	z.core.IOWrite(z.bc(), val)
	z.setHL(z.hl() - 1)
	other := val + z.L
	z.Flags.N = val&0x80 != 0
	if other < val {
		z.Flags.H = true
		z.Flags.C = true
	} else {
		z.Flags.H = false
		z.Flags.C = false
	}
	z.Flags.P = parity((other & 0x07) ^ z.B)
	z.Flags.Z = z.B == 0
	z.Flags.S = z.B&0x80 != 0
	z.updateXY(z.B)
}

func (z *CPU) daa() {
	correction := byte(0)

	if (z.A&0x0F) > 0x09 || z.Flags.H {
		correction += 0x06
	}

	if z.A > 0x99 || z.Flags.C {
		correction += 0x60
		z.Flags.C = true
	}

	substraction := z.Flags.N
	if substraction {
		z.Flags.H = z.Flags.H && (z.A&0x0F) < 0x06
		z.A -= correction
	} else {
		z.Flags.H = (z.A & 0x0F) > 0x09
		z.A += correction
	}

	z.Flags.S = z.A&0x80 != 0
	z.Flags.Z = z.A == 0
	z.Flags.P = parity(z.A)
	z.updateXY(z.A)
}

func (z *CPU) displace(baseAddr uint16, offset byte) uint16 {
	addr := baseAddr
	if offset&0x80 == 0x80 {
		addr += 0xff00 | uint16(offset)
	} else {
		addr += uint16(offset)
	}
	//addr := baseAddr + uint16(displacement)
	z.memPtr = addr
	return addr
}

func (z *CPU) processInterrupts() {
	// "When an EI instruction is executed, any pending interrupt request
	// is not accepted until after the instruction following EI is executed."
	if z.iffDelay > 0 {
		z.iffDelay -= 1
		if z.iffDelay == 0 {
			z.Iff1 = true
			z.Iff2 = true
		}
		return
	}

	if z.nmiPending {
		z.nmiPending = false
		z.Halted = false
		z.Iff1 = false
		z.incR()

		z.cycleCount += 11
		z.call(0x0066)
		return
	}

	if z.intPending && z.Iff1 {
		z.intPending = false
		z.Halted = false
		z.Iff1 = false
		z.Iff2 = false
		z.incR()

		switch z.IMode {
		case 0:
			z.cycleCount += 11
			z.execOpcode(z.intData)
		case 1:
			z.cycleCount += 13
			z.call(0x38)
		case 2:
			z.cycleCount += 19
			z.call(z.rw((uint16(z.I) << 8) | uint16(z.intData)))
		default:
			log.Errorf("Unsupported interrupt mode %d\n", z.IMode)
		}
		return
	}
}

// GenNMI function to call when an NMI is to be serviced
func (z *CPU) GenNMI() {
	z.nmiPending = true
}

// GenINT function to call when an INT is to be serviced
func (z *CPU) GenINT(data byte) {
	z.intPending = true
	z.intData = data
}

// executes A non-prefixed opcode
func (z *CPU) execOpcode(opcode byte) {
	z.cycleCount += uint32(cycles00[opcode])
	z.incR()

	switch opcode {
	case 0x7F:
		//z.a = z.a // ld a,a
	case 0x78:
		z.A = z.B // ld a,b
	case 0x79:
		z.A = z.C // ld a,c
	case 0x7A:
		z.A = z.D // ld a,d
	case 0x7B:
		z.A = z.E // ld a,e
	case 0x7C:
		z.A = z.H // ld a,h
	case 0x7D:
		z.A = z.L // ld a,l

	case 0x47:
		z.B = z.A // ld b,a
	case 0x40:
		//z.b = z.b // ld b,b
	case 0x41:
		z.B = z.C // ld b,c
	case 0x42:
		z.B = z.D // ld b,d
	case 0x43:
		z.B = z.E // ld b,e
	case 0x44:
		z.B = z.H // ld b,h
	case 0x45:
		z.B = z.L // ld b,l

	case 0x4F:
		z.C = z.A // ld c,a
	case 0x48:
		z.C = z.B // ld c,b
	case 0x49:
		//z.c = z.c // ld c,c
	case 0x4A:
		z.C = z.D // ld c,d
	case 0x4B:
		z.C = z.E // ld c,e
	case 0x4C:
		z.C = z.H // ld c,h
	case 0x4D:
		z.C = z.L // ld c,l

	case 0x57:
		z.D = z.A // ld d,a
	case 0x50:
		z.D = z.B // ld d,b
	case 0x51:
		z.D = z.C // ld d,c
	case 0x52:
		//z.d = z.d // ld d,d
	case 0x53:
		z.D = z.E // ld d,e
	case 0x54:
		z.D = z.H // ld d,h
	case 0x55:
		z.D = z.L // ld d,l

	case 0x5F:
		z.E = z.A // ld e,a
	case 0x58:
		z.E = z.B // ld e,b
	case 0x59:
		z.E = z.C // ld e,c
	case 0x5A:
		z.E = z.D // ld e,d
	case 0x5B:
		//z.e = z.e // ld e,e
	case 0x5C:
		z.E = z.H // ld e,h
	case 0x5D:
		z.E = z.L // ld e,l

	case 0x67:
		z.H = z.A // ld h,a
	case 0x60:
		z.H = z.B // ld h,b
	case 0x61:
		z.H = z.C // ld h,c
	case 0x62:
		z.H = z.D // ld h,d
	case 0x63:
		z.H = z.E // ld h,e
	case 0x64:
		//z.h = z.h // ld h,h
	case 0x65:
		z.H = z.L // ld h,l

	case 0x6F:
		z.L = z.A // ld l,a
	case 0x68:
		z.L = z.B // ld l,b
	case 0x69:
		z.L = z.C // ld l,c
	case 0x6A:
		z.L = z.D // ld l,d
	case 0x6B:
		z.L = z.E // ld l,e
	case 0x6C:
		z.L = z.H // ld l,h
	case 0x6D:
		//z.l = z.l // ld l,l

	case 0x7E:
		z.A = z.rb(z.hl()) // ld a,(hl)
	case 0x46:
		z.B = z.rb(z.hl()) // ld b,(hl)
	case 0x4E:
		z.C = z.rb(z.hl()) // ld c,(hl)
	case 0x56:
		z.D = z.rb(z.hl()) // ld d,(hl)
	case 0x5E:
		z.E = z.rb(z.hl()) // ld e,(hl)
	case 0x66:
		z.H = z.rb(z.hl()) // ld h,(hl)
	case 0x6E:
		z.L = z.rb(z.hl()) // ld l,(hl)

	case 0x77:
		z.wb(z.hl(), z.A) // ld (hl),a
	case 0x70:
		z.wb(z.hl(), z.B) // ld (hl),b
	case 0x71:
		z.wb(z.hl(), z.C) // ld (hl),c
	case 0x72:
		z.wb(z.hl(), z.D) // ld (hl),d
	case 0x73:
		z.wb(z.hl(), z.E) // ld (hl),e
	case 0x74:
		z.wb(z.hl(), z.H) // ld (hl),h
	case 0x75:
		z.wb(z.hl(), z.L) // ld (hl),l

	case 0x3E:
		z.A = z.nextB() // ld a,*
	case 0x06:
		z.B = z.nextB() // ld b,*
	case 0x0E:
		z.C = z.nextB() // ld c,*
	case 0x16:
		z.D = z.nextB() // ld d,*
	case 0x1E:
		z.E = z.nextB() // ld e,*
	case 0x26:
		z.H = z.nextB() // ld h,*
	case 0x2E:
		z.L = z.nextB() // ld l,*
	case 0x36:
		z.wb(z.hl(), z.nextB()) // ld (hl),*
	case 0x0A:
		// ld a,(bc)
		z.A = z.rb(z.bc())
		z.memPtr = z.bc() + 1
	case 0x1A:
		// ld a,(de)
		z.A = z.rb(z.de())
		z.memPtr = z.de() + 1
	case 0x3A:
		// ld a,(**)
		addr := z.nextW()
		z.A = z.rb(addr)
		z.memPtr = addr + 1
	case 0x02:
		// ld (bc),a
		z.wb(z.bc(), z.A)
		z.memPtr = (uint16(z.A) << 8) | ((z.bc() + 1) & 0xFF)
	case 0x12:
		// ld (de),a
		z.wb(z.de(), z.A)
		z.memPtr = (uint16(z.A) << 8) | ((z.de() + 1) & 0xFF)
	case 0x32:
		// ld (**),a
		addr := z.nextW()
		z.wb(addr, z.A)
		z.memPtr = (uint16(z.A) << 8) | ((addr + 1) & 0xFF)
	case 0x01:
		z.setBC(z.nextW()) // ld bc,**
	case 0x11:
		z.setDE(z.nextW()) // ld de,**
	case 0x21:
		z.setHL(z.nextW()) // ld hl,**
	case 0x31:
		z.SP = z.nextW() // ld sp,**

	case 0x2A:
		// ld hl,(**)
		addr := z.nextW()
		z.setHL(z.rw(addr))
		z.memPtr = addr + 1
	case 0x22:
		// ld (**),hl
		addr := z.nextW()
		z.ww(addr, z.hl())
		z.memPtr = addr + 1
	case 0xF9:
		z.SP = z.hl() // ld sp,hl

	case 0xEB:
		// ex de,hl
		de := z.de()
		z.setDE(z.hl())
		z.setHL(de)
	case 0xE3:
		// ex (sp),hl
		val := z.rw(z.SP)
		z.ww(z.SP, z.hl())
		z.setHL(val)
		z.memPtr = val
	case 0x87:
		z.A = z.addB(z.A, z.A, false) // add a,a
	case 0x80:
		z.A = z.addB(z.A, z.B, false) // add a,b
	case 0x81:
		z.A = z.addB(z.A, z.C, false) // add a,c
	case 0x82:
		z.A = z.addB(z.A, z.D, false) // add a,d
	case 0x83:
		z.A = z.addB(z.A, z.E, false) // add a,e
	case 0x84:
		z.A = z.addB(z.A, z.H, false) // add a,h
	case 0x85:
		z.A = z.addB(z.A, z.L, false) // add a,l
	case 0x86:
		z.A = z.addB(z.A, z.rb(z.hl()), false) // add a,(hl)
	case 0xC6:
		z.A = z.addB(z.A, z.nextB(), false) // add a,*

	case 0x8F:
		z.A = z.addB(z.A, z.A, z.Flags.C) // adc a,a
	case 0x88:
		z.A = z.addB(z.A, z.B, z.Flags.C) // adc a,b
	case 0x89:
		z.A = z.addB(z.A, z.C, z.Flags.C) // adc a,c
	case 0x8A:
		z.A = z.addB(z.A, z.D, z.Flags.C) // adc a,d
	case 0x8B:
		z.A = z.addB(z.A, z.E, z.Flags.C) // adc a,e
	case 0x8C:
		z.A = z.addB(z.A, z.H, z.Flags.C) // adc a,h
	case 0x8D:
		z.A = z.addB(z.A, z.L, z.Flags.C) // adc a,l
	case 0x8E:
		z.A = z.addB(z.A, z.rb(z.hl()), z.Flags.C) // adc a,(hl)
	case 0xCE:
		z.A = z.addB(z.A, z.nextB(), z.Flags.C) // adc a,*

	case 0x97:
		z.A = z.subB(z.A, z.A, false) // sub a,a
	case 0x90:
		z.A = z.subB(z.A, z.B, false) // sub a,b
	case 0x91:
		z.A = z.subB(z.A, z.C, false) // sub a,c
	case 0x92:
		z.A = z.subB(z.A, z.D, false) // sub a,d
	case 0x93:
		z.A = z.subB(z.A, z.E, false) // sub a,e
	case 0x94:
		z.A = z.subB(z.A, z.H, false) // sub a,h
	case 0x95:
		z.A = z.subB(z.A, z.L, false) // sub a,l
	case 0x96:
		z.A = z.subB(z.A, z.rb(z.hl()), false) // sub a,(hl)
	case 0xD6:
		z.A = z.subB(z.A, z.nextB(), false) // sub a,*

	case 0x9F:
		z.A = z.subB(z.A, z.A, z.Flags.C) // sbc a,a
	case 0x98:
		z.A = z.subB(z.A, z.B, z.Flags.C) // sbc a,b
	case 0x99:
		z.A = z.subB(z.A, z.C, z.Flags.C) // sbc a,c
	case 0x9A:
		z.A = z.subB(z.A, z.D, z.Flags.C) // sbc a,d
	case 0x9B:
		z.A = z.subB(z.A, z.E, z.Flags.C) // sbc a,e
	case 0x9C:
		z.A = z.subB(z.A, z.H, z.Flags.C) // sbc a,h
	case 0x9D:
		z.A = z.subB(z.A, z.L, z.Flags.C) // sbc a,l
	case 0x9E:
		z.A = z.subB(z.A, z.rb(z.hl()), z.Flags.C) // sbc a,(hl)
	case 0xDE:
		z.A = z.subB(z.A, z.nextB(), z.Flags.C) // sbc a,*

	case 0x09:
		z.addHL(z.bc()) // add hl,bc
	case 0x19:
		z.addHL(z.de()) // add hl,de
	case 0x29:
		z.addHL(z.hl()) // add hl,hl
	case 0x39:
		z.addHL(z.SP) // add hl,sp

	case 0xF3:
		z.Iff1 = false
		z.Iff2 = false // di
	case 0xFB:
		z.iffDelay = 1 // ei
	case 0x00: // nop
	case 0x76:
		z.Halted = true // halt
		z.PC--
	case 0x3C:
		z.A = z.inc(z.A) // inc a
	case 0x04:
		z.B = z.inc(z.B) // inc b
	case 0x0C:
		z.C = z.inc(z.C) // inc c
	case 0x14:
		z.D = z.inc(z.D) // inc d
	case 0x1C:
		z.E = z.inc(z.E) // inc e
	case 0x24:
		z.H = z.inc(z.H) // inc h
	case 0x2C:
		z.L = z.inc(z.L) // inc l
	case 0x34:
		// inc (hl)
		result := z.inc(z.rb(z.hl()))
		z.wb(z.hl(), result)
	case 0x3D:
		z.A = z.dec(z.A) // dec a
	case 0x05:
		z.B = z.dec(z.B) // dec b
	case 0x0D:
		z.C = z.dec(z.C) // dec c
	case 0x15:
		z.D = z.dec(z.D) // dec d
	case 0x1D:
		z.E = z.dec(z.E) // dec e
	case 0x25:
		z.H = z.dec(z.H) // dec h
	case 0x2D:
		z.L = z.dec(z.L) // dec l
	case 0x35:
		// dec (hl)
		result := z.dec(z.rb(z.hl()))
		z.wb(z.hl(), result)
	case 0x03:
		z.setBC(z.bc() + 1) // inc bc
	case 0x13:
		z.setDE(z.de() + 1) // inc de
	case 0x23:
		z.setHL(z.hl() + 1) // inc hl
	case 0x33:
		z.SP = z.SP + 1 // inc sp
	case 0x0B:
		z.setBC(z.bc() - 1) // dec bc
	case 0x1B:
		z.setDE(z.de() - 1) // dec de
	case 0x2B:
		z.setHL(z.hl() - 1) // dec hl
	case 0x3B:
		z.SP = z.SP - 1 // dec sp
	case 0x27:
		z.daa() // daa
	case 0x2F:
		// cpl
		z.A = ^z.A
		z.Flags.N = true
		z.Flags.H = true
		z.updateXY(z.A)
	case 0x37:
		// scf
		z.Flags.C = true
		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A | z.f())
	case 0x3F:
		// ccf
		z.Flags.H = z.Flags.C
		z.Flags.C = !z.Flags.C
		z.Flags.N = false
		z.updateXY(z.A | z.f())
	case 0x07:
		// rlca (rotate left)
		z.Flags.C = z.A&0x80 != 0
		z.A = (z.A << 1) | bToByte(z.Flags.C)
		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A)
	case 0x0F:
		// rrca (rotate right)
		z.Flags.C = z.A&1 != 0
		z.A = (z.A >> 1) | (bToByte(z.Flags.C) << 7)
		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A)
	case 0x17:
		// rla
		cy := bToByte(z.Flags.C)
		z.Flags.C = z.A&0x80 != 0
		z.A = (z.A << 1) | cy
		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A)
	case 0x1F:
		// rra
		cy := bToByte(z.Flags.C)
		z.Flags.C = z.A&1 != 0
		z.A = (z.A >> 1) | (cy << 7)
		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A)
	case 0xA7:
		z.lAnd(z.A) // and a
	case 0xA0:
		z.lAnd(z.B) // and b
	case 0xA1:
		z.lAnd(z.C) // and c
	case 0xA2:
		z.lAnd(z.D) // and d
	case 0xA3:
		z.lAnd(z.E) // and e
	case 0xA4:
		z.lAnd(z.H) // and h
	case 0xA5:
		z.lAnd(z.L) // and l
	case 0xA6:
		z.lAnd(z.rb(z.hl())) // and (hl)
	case 0xE6:
		z.lAnd(z.nextB()) // and *

	case 0xAF:
		z.lXor(z.A) // xor a
	case 0xA8:
		z.lXor(z.B) // xor b
	case 0xA9:
		z.lXor(z.C) // xor c
	case 0xAA:
		z.lXor(z.D) // xor d
	case 0xAB:
		z.lXor(z.E) // xor e
	case 0xAC:
		z.lXor(z.H) // xor h
	case 0xAD:
		z.lXor(z.L) // xor l
	case 0xAE:
		z.lXor(z.rb(z.hl())) // xor (hl)
	case 0xEE:
		z.lXor(z.nextB()) // xor *

	case 0xB7:
		z.lOr(z.A) // or a
	case 0xB0:
		z.lOr(z.B) // or b
	case 0xB1:
		z.lOr(z.C) // or c
	case 0xB2:
		z.lOr(z.D) // or d
	case 0xB3:
		z.lOr(z.E) // or e
	case 0xB4:
		z.lOr(z.H) // or h
	case 0xB5:
		z.lOr(z.L) // or l
	case 0xB6:
		z.lOr(z.rb(z.hl())) // or (hl)
	case 0xF6:
		z.lOr(z.nextB()) // or *

	case 0xBF:
		z.cp(z.A) // cp a
	case 0xB8:
		z.cp(z.B) // cp b
	case 0xB9:
		z.cp(z.C) // cp c
	case 0xBA:
		z.cp(z.D) // cp d
	case 0xBB:
		z.cp(z.E) // cp e
	case 0xBC:
		z.cp(z.H) // cp h
	case 0xBD:
		z.cp(z.L) // cp l
	case 0xBE:
		z.cp(z.rb(z.hl())) // cp (hl)
	case 0xFE:
		z.cp(z.nextB()) // cp *

	case 0xC3:
		z.jump(z.nextW()) // jm **
	case 0xC2:
		z.condJump(!z.Flags.Z) // jp nz, **
	case 0xCA:
		z.condJump(z.Flags.Z) // jp z, **
	case 0xD2:
		z.condJump(!z.Flags.C) // jp nc, **
	case 0xDA:
		z.condJump(z.Flags.C) // jp c, **
	case 0xE2:
		z.condJump(!z.Flags.P) // jp po, **
	case 0xEA:
		z.condJump(z.Flags.P) // jp pe, **
	case 0xF2:
		z.condJump(!z.Flags.S) // jp p, **
	case 0xFA:
		z.condJump(z.Flags.S) // jp m, **

	case 0x10:
		z.B--
		z.condJr(z.B != 0) // djnz *
	case 0x18:
		z.PC += uint16(z.nextB()) // jr *
		z.memPtr = z.PC
	case 0x20:
		z.condJr(!z.Flags.Z) // jr nz, *
	case 0x28:
		z.condJr(z.Flags.Z) // jr z, *
	case 0x30:
		z.condJr(!z.Flags.C) // jr nc, *
	case 0x38:
		z.condJr(z.Flags.C) // jr c, *

	case 0xE9:
		z.PC = z.hl() // jp (hl)
	case 0xCD:
		z.call(z.nextW()) // call

	case 0xC4:
		z.condCall(!z.Flags.Z) // cnz
	case 0xCC:
		z.condCall(z.Flags.Z) // cz
	case 0xD4:
		z.condCall(!z.Flags.C) // cnc
	case 0xDC:
		z.condCall(z.Flags.C) // cc
	case 0xE4:
		z.condCall(!z.Flags.P) // cpo
	case 0xEC:
		z.condCall(z.Flags.P) // cpe
	case 0xF4:
		z.condCall(!z.Flags.S) // cp
	case 0xFC:
		z.condCall(z.Flags.S) // cm

	case 0xC9:
		z.ret() // ret
	case 0xC0:
		z.condRet(!z.Flags.Z) // ret nz
	case 0xC8:
		z.condRet(z.Flags.Z) // ret z
	case 0xD0:
		z.condRet(!z.Flags.C) // ret nc
	case 0xD8:
		z.condRet(z.Flags.C) // ret c
	case 0xE0:
		z.condRet(!z.Flags.P) // ret po
	case 0xE8:
		z.condRet(z.Flags.P) // ret pe
	case 0xF0:
		z.condRet(!z.Flags.S) // ret p
	case 0xF8:
		z.condRet(z.Flags.S) // ret m

	case 0xC7:
		z.call(0x00) // rst 0
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xCF:
		z.call(0x08) // rst 1
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xD7:
		z.call(0x10) // rst 2
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xDF:
		z.call(0x18) // rst 3
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xE7:
		z.call(0x20) // rst 4
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xEF:
		z.call(0x28) // rst 5
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xF7:
		z.call(0x30) // rst 6
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xFF:
		z.call(0x38) // rst 7
		z.extendedStack[z.SP] = PushValueTypeRst
	case 0xC5:
		z.pushW(z.bc()) // push bc
	case 0xD5:
		z.pushW(z.de()) // push de
	case 0xE5:
		z.pushW(z.hl()) // push hl
	case 0xF5:
		z.pushW((uint16(z.A) << 8) | uint16(z.f())) // push af

	case 0xC1:
		z.setBC(z.popW()) // pop bc
	case 0xD1:
		z.setDE(z.popW()) // pop de
	case 0xE1:
		z.setHL(z.popW()) // pop hl
	case 0xF1:
		// pop af
		val := z.popW()
		z.A = byte(val >> 8)
		z.setF(byte(val))
	case 0xDB:
		// in a,(n)
		port := (uint16(z.A) << 8) | uint16(z.nextB())
		z.A = z.core.IORead(port)
		z.memPtr = port + 1 // (uint16(a) << 8) | (uint16(z.a+1) & 0x00ff)
	case 0xD3:
		// out (n), a
		port := uint16(z.nextB())
		z.core.IOWrite(port, z.A)
		z.memPtr = ((port + 1) & 0x00ff) | (uint16(z.A) << 8)
	case 0x08:
		// ex af,af'
		a := z.A
		f := z.Flags.Flags()

		z.A = z.AAlt
		z.Flags.SetFlags(z.FlagsAlt.Flags())

		z.AAlt = a
		z.FlagsAlt.SetFlags(f)
	case 0xD9:
		// exx
		b := z.B
		c := z.C
		d := z.D
		e := z.E
		h := z.H
		l := z.L

		z.B = z.BAlt
		z.C = z.CAlt
		z.D = z.DAlt
		z.E = z.EAlt
		z.H = z.HAlt
		z.L = z.LAlt

		z.BAlt = b
		z.CAlt = c
		z.DAlt = d
		z.EAlt = e
		z.HAlt = h
		z.LAlt = l
	case 0xCB:
		z.execOpcodeCB(z.nextB())
	case 0xED:
		z.execOpcodeED(z.nextB())
	case 0xDD:
		z.execOpcodeDDFD(z.nextB(), &z.IX)
	case 0xFD:
		z.execOpcodeDDFD(z.nextB(), &z.IY)

	default:
		log.Errorf("Unknown opcode %02X\n", opcode)
	}
}
