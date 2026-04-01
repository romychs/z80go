package z80go

// executes A DD/FD opcode (IZ = IX or IY)
func (z *CPU) execOpcodeDDFD(opcode byte, iz *uint16) {
	z.cycleCount += uint32(cyclesDDFD[opcode])
	z.incR()

	switch opcode {
	case 0xE1:
		*iz = z.popW() // pop iz
	case 0xE5:
		z.pushW(*iz) // push iz
	case 0xE9:
		// jp iz
		z.PC = *iz
		//z.jump(*iz)
	case 0x09:
		z.addIZ(iz, z.bc()) // add iz,bc
	case 0x19:
		z.addIZ(iz, z.de()) // add iz,de
	case 0x29:
		z.addIZ(iz, *iz) // add iz,iz
	case 0x39:
		z.addIZ(iz, z.SP) // add iz,sp

	case 0x84:
		z.A = z.addB(z.A, byte(*iz>>8), false) // add a,izh
	case 0x85:
		z.A = z.addB(z.A, byte(*iz), false) // add a,izl
	case 0x8C:
		z.A = z.addB(z.A, byte(*iz>>8), z.Flags.C) // adc a,izh
	case 0x8D:
		z.A = z.addB(z.A, byte(*iz), z.Flags.C) // adc a,izl
	case 0x86:
		z.A = z.addB(z.A, z.rb(z.displace(*iz, z.nextB())), false) // add a,(iz+*)
	case 0x8E:
		z.A = z.addB(z.A, z.rb(z.displace(*iz, z.nextB())), z.Flags.C) // adc a,(iz+*)
	case 0x96:
		z.A = z.subB(z.A, z.rb(z.displace(*iz, z.nextB())), false) // sub (iz+*)
	case 0x9E:
		z.A = z.subB(z.A, z.rb(z.displace(*iz, z.nextB())), z.Flags.C) // sbc (iz+*)
	case 0x94:
		z.A = z.subB(z.A, byte(*iz>>8), false) // sub izh
	case 0x95:
		z.A = z.subB(z.A, byte(*iz), false) // sub izl
	case 0x9C:
		z.A = z.subB(z.A, byte(*iz>>8), z.Flags.C) // sbc izh
	case 0x9D:
		z.A = z.subB(z.A, byte(*iz), z.Flags.C) // sbc izl

	case 0xA6:
		z.lAnd(z.rb(z.displace(*iz, z.nextB()))) // and (iz+*)
	case 0xA4:
		z.lAnd(byte(*iz >> 8)) // and izh
	case 0xA5:
		z.lAnd(byte(*iz)) // and izl

	case 0xAE:
		z.lXor(z.rb(z.displace(*iz, z.nextB()))) // xor (iz+*)
	case 0xAC:
		z.lXor(byte(*iz >> 8)) // xor izh
	case 0xAD:
		z.lXor(byte(*iz)) // xor izl
	case 0xB6:
		z.lOr(z.rb(z.displace(*iz, z.nextB()))) // or (iz+*)
	case 0xB4:
		z.lOr(byte(*iz >> 8)) // or izh
	case 0xB5:
		z.lOr(byte(*iz)) // or izl
	case 0xBE:
		z.cp(z.rb(z.displace(*iz, z.nextB()))) // cp (iz+*)
	case 0xBC:
		z.cp(byte(*iz >> 8)) // cp izh
	case 0xBD:
		z.cp(byte(*iz)) // cp izl
	case 0x23:
		*iz += 1 // inc iz
	case 0x2B:
		*iz -= 1 // dec iz
	case 0x34:
		// inc (iz+*)
		addr := z.displace(*iz, z.nextB())
		z.wb(addr, z.inc(z.rb(addr)))
	case 0x35:
		// dec (iz+*)
		addr := z.displace(*iz, z.nextB())
		z.wb(addr, z.dec(z.rb(addr)))
	case 0x24:
		*iz = (*iz & 0x00ff) | (uint16(z.inc(byte(*iz>>8))) << 8) // inc izh
	case 0x25:
		*iz = (*iz & 0x00ff) | (uint16(z.dec(byte(*iz>>8))) << 8) // dec izh
	case 0x2C:
		*iz = (*iz & 0xff00) | uint16(z.inc(byte(*iz))) // inc izl
	case 0x2D:
		*iz = (*iz & 0xff00) | uint16(z.dec(byte(*iz))) // dec izl
	case 0x2A:
		addr := z.nextW()
		*iz = z.rw(addr) // ld iz,(**)
		z.memPtr = addr + 1
	case 0x22:
		addr := z.nextW()
		z.ww(addr, *iz) // ld (**),iz
		z.memPtr = addr + 1
	case 0x21:
		*iz = z.nextW() // ld iz,**
	case 0x36:
		// ld (iz+*),*
		addr := z.displace(*iz, z.nextB())
		z.wb(addr, z.nextB())
	case 0x70:
		z.wb(z.displace(*iz, z.nextB()), z.B) // ld (iz+*),b
	case 0x71:
		z.wb(z.displace(*iz, z.nextB()), z.C) // ld (iz+*),c
	case 0x72:
		z.wb(z.displace(*iz, z.nextB()), z.D) // ld (iz+*),d
	case 0x73:
		z.wb(z.displace(*iz, z.nextB()), z.E) // ld (iz+*),e
	case 0x74:
		z.wb(z.displace(*iz, z.nextB()), z.H) // ld (iz+*),h
	case 0x75:
		z.wb(z.displace(*iz, z.nextB()), z.L) // ld (iz+*),l
	case 0x77:
		z.wb(z.displace(*iz, z.nextB()), z.A) // ld (iz+*),a
	case 0x46:
		z.B = z.rb(z.displace(*iz, z.nextB())) // ld b,(iz+*)
	case 0x4E:
		z.C = z.rb(z.displace(*iz, z.nextB())) // ld c,(iz+*)
	case 0x56:
		z.D = z.rb(z.displace(*iz, z.nextB())) // ld d,(iz+*)
	case 0x5E:
		z.E = z.rb(z.displace(*iz, z.nextB())) // ld e,(iz+*)
	case 0x66:
		z.H = z.rb(z.displace(*iz, z.nextB())) // ld h,(iz+*)
	case 0x6E:
		z.L = z.rb(z.displace(*iz, z.nextB())) // ld l,(iz+*)
	case 0x7E:
		z.A = z.rb(z.displace(*iz, z.nextB())) // ld a,(iz+*)
	case 0x44:
		z.B = byte(*iz >> 8) // ld b,izh
	case 0x4C:
		z.C = byte(*iz >> 8) // ld c,izh
	case 0x54:
		z.D = byte(*iz >> 8) // ld d,izh
	case 0x5C:
		z.E = byte(*iz >> 8) // ld e,izh
	case 0x7C:
		z.A = byte(*iz >> 8) // ld a,izh
	case 0x45:
		z.B = byte(*iz) // ld b,izl
	case 0x4D:
		z.C = byte(*iz) // ld c,izl
	case 0x55:
		z.D = byte(*iz) // ld d,izl
	case 0x5D:
		z.E = byte(*iz) // ld e,izl
	case 0x7D:
		z.A = byte(*iz) // ld a,izl
	case 0x60:
		*iz = (*iz & 0x00ff) | (uint16(z.B) << 8) // ld izh,b
	case 0x61:
		*iz = (*iz & 0x00ff) | (uint16(z.C) << 8) // ld izh,c
	case 0x62:
		*iz = (*iz & 0x00ff) | (uint16(z.D) << 8) // ld izh,d
	case 0x63:
		*iz = (*iz & 0x00ff) | (uint16(z.E) << 8) // ld izh,e
	case 0x64: // ld izh,izh
	case 0x65:
		*iz = ((*iz & 0x00ff) << 8) | (*iz & 0x00ff) // ld izh,izl
	case 0x67:
		*iz = (uint16(z.A) << 8) | (*iz & 0x00ff) // ld izh,a
	case 0x26:
		*iz = (uint16(z.nextB()) << 8) | (*iz & 0x00ff) // ld izh,*
	case 0x68:
		*iz = (*iz & 0xff00) | uint16(z.B) // ld izl,b
	case 0x69:
		*iz = (*iz & 0xff00) | uint16(z.C) // ld izl,c
	case 0x6A:
		*iz = (*iz & 0xff00) | uint16(z.D) // ld izl,d
	case 0x6B:
		*iz = (*iz & 0xff00) | uint16(z.E) // ld izl,e
	case 0x6C:
		*iz = (*iz & 0xff00) | (*iz >> 8) // ld izl,izh
	case 0x6D: // ld izl,izl
	case 0x6F:
		*iz = (*iz & 0xff00) | uint16(z.A) // ld izl,a
	case 0x2E:
		*iz = (*iz & 0xff00) | uint16(z.nextB()) // ld izl,*
	case 0xF9:
		z.SP = *iz // ld sp,iz
	case 0xE3:
		// ex (sp),iz
		val := z.rw(z.SP)
		z.ww(z.SP, *iz)
		*iz = val
		z.memPtr = val
	case 0xCB:
		addr := z.displace(*iz, z.nextB())
		op := z.nextB()
		z.execOpcodeDcb(op, addr)
	default:
		// any other FD/DD opcode behaves as a non-prefixed opcode:
		z.execOpcode(opcode)
		// R should not be incremented twice:
		z.incR()
	}
}
