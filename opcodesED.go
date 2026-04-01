package z80go

import log "github.com/sirupsen/logrus"

// executes A ED opcode
func (z *CPU) execOpcodeED(opcode byte) {
	z.cycleCount += uint32(cyclesED[opcode])
	z.incR()
	switch opcode {
	case 0x47:
		z.I = z.A // ld i,a
	case 0x4F:
		z.R = z.A // ld r,a
	case 0x57:
		// ld a,i
		z.A = z.I
		z.Flags.S = z.A&0x80 != 0
		z.Flags.Z = z.A == 0
		z.Flags.H = false
		z.Flags.N = false
		z.Flags.P = z.Iff2
		z.updateXY(z.A)
	case 0x5F:
		// ld a,r
		z.A = z.R
		z.Flags.S = z.A&0x80 != 0
		z.Flags.Z = z.A == 0
		z.Flags.H = false
		z.Flags.N = false
		z.Flags.P = z.Iff2
	case 0x45, 0x55, 0x5D, 0x65, 0x6D, 0x75, 0x7D:
		// retn
		z.Iff1 = z.Iff2
		z.ret()
	case 0x4D:
		z.ret() // reti

	case 0xA0:
		z.ldi() // ldi
	case 0xB0:
		{
			z.ldi()

			if z.bc() != 0 {
				z.PC -= 2
				z.cycleCount += 5
				z.memPtr = z.PC + 1
			}
		} // ldir

	case 0xA8:
		z.ldd() // ldd
	case 0xB8:
		{
			z.ldd()

			if z.bc() != 0 {
				z.PC -= 2
				z.cycleCount += 5
				z.memPtr = z.PC + 1
			}
		} // lddr

	case 0xA1:
		z.cpi() // cpi
	case 0xA9:
		z.cpd() // cpd
	case 0xB1:
		// cpir
		z.cpi()
		if z.bc() != 0 && !z.Flags.Z {
			z.PC -= 2
			z.cycleCount += 5
			z.memPtr = z.PC + 1
		} else {
			//z.mem_ptr++
		}
		//z.cpir()
	case 0xB9:
		// cpdr
		z.cpd()
		if z.bc() != 0 && !z.Flags.Z {
			z.PC -= 2
			z.cycleCount += 5
			z.memPtr = z.PC + 1
		} else {
			//z.mem_ptr++
		}
	case 0x40:
		z.inRC(&z.B) // in b, (c)
		z.memPtr = z.bc() + 1
	case 0x48:
		z.memPtr = z.bc() + 1
		z.inRC(&z.C) // in c, (c)
		z.updateXY(z.C)
	//case 0x4e:
	// ld c,(iy+dd)

	case 0x50:
		z.inRC(&z.D) // in d, (c)
		z.memPtr = z.bc() + 1
	case 0x58:
		// in e, (c)
		z.inRC(&z.E)
		z.memPtr = z.bc() + 1
		z.updateXY(z.E)
	case 0x60:
		z.inRC(&z.H) // in h, (c)
		z.memPtr = z.bc() + 1
	case 0x68:
		z.inRC(&z.L) // in l, (c)
		z.memPtr = z.bc() + 1
		z.updateXY(z.L)
	case 0x70:
		// in (c)
		var val byte
		z.inRC(&val)
		z.memPtr = z.bc() + 1
	case 0x78:
		// in a, (c)
		z.inRC(&z.A)
		z.memPtr = z.bc() + 1
		z.updateXY(z.A)
	case 0xA2:
		z.ini() // ini
	case 0xB2:
		// inir
		z.ini()
		if z.B > 0 {
			z.PC -= 2
			z.cycleCount += 5
		}
	case 0xAA:
		// ind
		z.ind()
	case 0xBA:
		// indr
		z.ind()
		if z.B > 0 {
			z.PC -= 2
			z.cycleCount += 5
		}
	case 0x41:
		z.core.IOWrite(z.bc(), z.B) // out (c), b
		z.memPtr = z.bc() + 1
	case 0x49:
		z.core.IOWrite(z.bc(), z.C) // out (c), c
		z.memPtr = z.bc() + 1
	case 0x51:
		z.core.IOWrite(z.bc(), z.D) // out (c), d
		z.memPtr = z.bc() + 1
	case 0x59:
		z.core.IOWrite(z.bc(), z.E) // out (c), e
		z.memPtr = z.bc() + 1
	case 0x61:
		z.core.IOWrite(z.bc(), z.H) // out (c), h
		z.memPtr = z.bc() + 1
	case 0x69:
		z.core.IOWrite(z.bc(), z.L) // out (c), l
		z.memPtr = z.bc() + 1
	case 0x71:
		z.core.IOWrite(z.bc(), 0) // out (c), 0
		z.memPtr = z.bc() + 1
	case 0x79:
		// out (c), a
		z.core.IOWrite(z.bc(), z.A)
		z.memPtr = z.bc() + 1
	case 0xA3:
		z.outi() // outi
	case 0xB3:
		// otir
		z.outi()
		if z.B > 0 {
			z.PC -= 2
			z.cycleCount += 5
		}
	case 0xAB:
		z.outd() // outd
	case 0xBB:
		// otdr
		z.outd()
		if z.B > 0 {
			z.cycleCount += 5
			z.PC -= 2
		}

	case 0x42:
		z.sbcHL(z.bc()) // sbc hl,bc
	case 0x52:
		z.sbcHL(z.de()) // sbc hl,de
	case 0x62:
		z.sbcHL(z.hl()) // sbc hl,hl
	case 0x72:
		z.sbcHL(z.SP) // sbc hl,sp
	case 0x4A:
		z.adcHL(z.bc()) // adc hl,bc
	case 0x5A:
		z.adcHL(z.de()) // adc hl,de
	case 0x6A:
		z.adcHL(z.hl()) // adc hl,hl
	case 0x7A:
		z.adcHL(z.SP) // adc hl,sp
	case 0x43:
		// ld (**), bc
		addr := z.nextW()
		z.ww(addr, z.bc())
		z.memPtr = addr + 1
	case 0x53:
		// ld (**), de
		addr := z.nextW()
		z.ww(addr, z.de())
		z.memPtr = addr + 1
	case 0x63:
		// ld (**), hl
		addr := z.nextW()
		z.ww(addr, z.hl())
		z.memPtr = addr + 1
	case 0x73:
		// ld (**), hl
		addr := z.nextW()
		z.ww(addr, z.SP)
		z.memPtr = addr + 1
	case 0x4B:
		// ld bc, (**)
		addr := z.nextW()
		z.setBC(z.rw(addr))
		z.memPtr = addr + 1
	case 0x5B:
		// ld de, (**)
		addr := z.nextW()
		z.setDE(z.rw(addr))
		z.memPtr = addr + 1
	case 0x6B:
		// ld hl, (**)
		addr := z.nextW()
		z.setHL(z.rw(addr))
		z.memPtr = addr + 1
	case 0x7B:
		// ld sp,(**)
		addr := z.nextW()
		z.SP = z.rw(addr)
		z.memPtr = addr + 1
	case 0x44, 0x54, 0x64, 0x74, 0x4C, 0x5C, 0x6C, 0x7C:
		z.A = z.subB(0, z.A, false) // neg
	case 0x46, 0x4e, 0x66, 0x6e:
		z.IMode = 0 // im 0
	case 0x56, 0x76:
		z.IMode = 1 // im 1
	case 0x5E, 0x7E:
		z.IMode = 2 // im 2
	case 0x67:
		// rrd
		a := z.A
		val := z.rb(z.hl())
		z.A = (a & 0xF0) | (val & 0xF)
		z.wb(z.hl(), (val>>4)|(a<<4))

		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A)
		z.Flags.Z = z.A == 0
		z.Flags.S = z.A&0x80 != 0
		z.Flags.P = parity(z.A)
		z.memPtr = z.hl() + 1
	case 0x6F:
		// rld
		a := z.A
		val := z.rb(z.hl())
		z.A = (a & 0xF0) | (val >> 4)
		z.wb(z.hl(), (val<<4)|(a&0xF))
		z.Flags.N = false
		z.Flags.H = false
		z.updateXY(z.A)
		z.Flags.Z = z.A == 0
		z.Flags.S = z.A&0x80 != 0
		z.Flags.P = parity(z.A)
		z.memPtr = z.hl() + 1
	default:
		log.Errorf("Unknown ED opcode: %02X\n", opcode)
	}
}
