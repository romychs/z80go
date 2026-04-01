package z80go

import log "github.com/sirupsen/logrus"

// executes A CB opcode
func (z *CPU) execOpcodeCB(opcode byte) {
	z.cycleCount += 8
	z.incR()

	// decoding instructions from http://z80.info/decoding.htm#cb
	x_ := (opcode >> 6) & 3 // 0b11
	y_ := (opcode >> 3) & 7 // 0b111
	z_ := opcode & 7        // 0b111

	var hl byte
	v := byte(0)
	reg := &v
	switch z_ {
	case 0:
		reg = &z.B
	case 1:
		reg = &z.C
	case 2:
		reg = &z.D
	case 3:
		reg = &z.E
	case 4:
		reg = &z.H
	case 5:
		reg = &z.L
	case 6:
		hl = z.rb(z.hl())
		reg = &hl
	case 7:
		reg = &z.A
	}

	switch x_ {
	case 0:
		// rot[y] r[z]
		switch y_ {
		case 0:
			*reg = z.cbRlc(*reg)
		case 1:
			*reg = z.cbRrc(*reg)
		case 2:
			*reg = z.cbRl(*reg)
		case 3:
			*reg = z.cbRr(*reg)
		case 4:
			*reg = z.cbSla(*reg)
		case 5:
			*reg = z.cbSra(*reg)
		case 6:
			*reg = z.cbSll(*reg)
		case 7:
			*reg = z.cbSrl(*reg)
		}

	case 1:
		// BIT y, r[z]
		z.cbBit(*reg, y_)

		// in bit (hl), x/y flags are handled differently:
		if z_ == 6 {
			z.updateXY(byte(z.MemPtr >> 8))
			z.cycleCount += 4
		}

	case 2:
		*reg &= ^(1 << y_) // RES y, r[z]
	case 3:
		*reg |= 1 << y_ // SET y, r[z]
	}

	if (x_ == 0 || x_ == 2 || x_ == 3) && z_ == 6 {
		z.cycleCount += 7
	}

	if reg == &hl {
		z.wb(z.hl(), hl)
	}
}

// execOpcodeDcb executes A displaced CB opcode (DDCB or FDCB)
func (z *CPU) execOpcodeDcb(opcode byte, addr uint16) {
	val := z.rb(addr)
	result := byte(0)

	// decoding instructions from http://z80.info/decoding.htm#ddcb
	x_ := (opcode >> 6) & 3 // 0b11
	y_ := (opcode >> 3) & 7 // 0b111
	z_ := opcode & 7        // 0b111

	switch x_ {
	case 0:
		// rot[y] (iz+d)
		switch y_ {
		case 0:
			result = z.cbRlc(val)
		case 1:
			result = z.cbRrc(val)
		case 2:
			result = z.cbRl(val)
		case 3:
			result = z.cbRr(val)
		case 4:
			result = z.cbSla(val)
		case 5:
			result = z.cbSra(val)
		case 6:
			result = z.cbSll(val)
		case 7:
			result = z.cbSrl(val)
		}

	case 1:
		// bit y,(iz+d)
		result = z.cbBit(val, y_)
		z.updateXY(byte(addr >> 8))
	case 2:
		result = val & ^(1 << y_) // res y, (iz+d)
	case 3:
		result = val | (1 << y_) // set y, (iz+d)

	default:
		log.Errorf("Unknown XYCB opcode: %02X\n", opcode)
	}

	// ld r[z], rot[y] (iz+d)
	// ld r[z], res y,(iz+d)
	// ld r[z], set y,(iz+d)
	if x_ != 1 && z_ != 6 {
		switch z_ {
		case 0:
			z.B = result
		case 1:
			z.C = result
		case 2:
			z.D = result
		case 3:
			z.E = result
		case 4:
			z.H = result
		case 5:
			z.L = result
		// always false
		//case 6:
		//	z.wb(z.hl(), result)
		case 7:
			z.A = result
		}
	}

	if x_ == 1 {
		// bit instructions take 20 cycles, others take 23
		z.cycleCount += 20
	} else {
		z.wb(addr, result)
		z.cycleCount += 23
	}
}
