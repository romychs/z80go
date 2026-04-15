package dis

import (
	"fmt"
	"strings"

	"github.com/romychs/z80go"
)

type Disassembler struct {
	pc   uint16
	core z80go.MemIoRW
}

type Disassembly interface {
	Disassm(pc uint16) string
}

func NewDisassembler(core z80go.MemIoRW) *Disassembler {
	d := Disassembler{
		pc:   0,
		core: core,
	}
	return &d
}

// opcode & 0x07
var operands = []string{"B", "C", "D", "E", "H", "L", "(HL)", "A"}
var aluOp = []string{"ADD A" + sep, "ADC A" + sep, "SUB ", "SBC A" + sep, "AND ", "XOR ", "OR ", "CP "}

const sep = ", "

func (d *Disassembler) jp(op, cond string) string {
	addr := d.getW()
	if cond != "" {
		cond += sep
	}
	return fmt.Sprintf("%s %s%s", op, cond, addr)
}

func (d *Disassembler) jr(op, cond string) string {
	addr := d.pc + 1
	offset := d.getByte()
	if offset&0x80 != 0 {
		addr += 0xFF00 | uint16(offset)
	} else {
		addr += uint16(offset)
	}
	if cond != "" {
		cond += sep
	}
	return fmt.Sprintf("%s %s0x%04X", op, cond, addr)
}

func (d *Disassembler) getByte() byte {
	b := d.core.MemRead(d.pc)
	d.pc++
	return b
}

func (d *Disassembler) Disassm(pc uint16) string {
	d.pc = pc
	result := fmt.Sprintf("  %04X ", d.pc)
	op := d.getByte()

	switch {
	// == 00:0F
	case op == 0x00:
		result += "NOP"
	case op == 0x01:
		result += "LD BC" + sep + d.getW()
	case op == 0x02:
		result += "LD (BC)" + sep + "A"
	case op == 0x03:
		result += "INC BC"
	case op == 0x04:
		result += "INC B"
	case op == 0x05:
		result += "DEC B"
	case op == 0x06:
		result += "LD B" + sep + d.getB()
	case op == 0x07:
		result += "RLCA"
	case op == 0x08:
		result += "EX AF, AF'"
	case op == 0x09:
		result += "ADD HL" + sep + "BC"
	case op == 0x0A:
		result += "LD A" + sep + "(BC)"
	case op == 0x0B:
		result += "DEC BC"
	case op == 0x0C:
		result += "INC C"
	case op == 0x0D:
		result += "DEC C"
	case op == 0x0E:
		result += "LD C" + sep + d.getB()
	case op == 0x0F:
		result += "RRCA"
	// 10:1F
	case op == 0x10:
		// DJNZ rel
		result += d.jr("DJNZ", "")
	case op == 0x11:
		result += "LD DE" + sep + d.getW()
	case op == 0x12:
		result += "LD (DE)" + sep + "A"
	case op == 0x13:
		result += "INC DE"
	case op == 0x14:
		result += "INC D"
	case op == 0x15:
		result += "DEC D"
	case op == 0x16:
		result += "LD D" + sep + d.getB()
	case op == 0x17:
		result += "RLA"
	case op == 0x18:
		result += d.jr("JR", "")
	case op == 0x19:
		result += "ADD HL" + sep + "DE"
	case op == 0x1A:
		result += "LD A" + sep + "(DE)"
	case op == 0x1B:
		result += "DEC DE"
	case op == 0x1C:
		result += "INC E"
	case op == 0x1D:
		result += "DEC E"
	case op == 0x1E:
		result += "LD E" + sep + d.getB()
	case op == 0x1F:
		result += "RRA"
	// == 20:2F
	case op == 0x20:
		result += d.jr("JR", "NZ")
	case op == 0x21:
		result += "LD HL" + sep + d.getW()
	case op == 0x22:
		// LD (nn),HL
		result += "LD (" + d.getW() + ")" + sep + "HL"
	case op == 0x23:
		result += "INC HL"
	case op == 0x24:
		result += "INC H"
	case op == 0x25:
		result += "DEC H"
	case op == 0x26:
		result += "LD H" + sep + d.getB()
	case op == 0x27:
		result += "DAA"
	case op == 0x28:
		result += d.jr("JR", "Z")
	case op == 0x29:
		result += "ADD HL" + sep + "HL"
	case op == 0x2A:
		result += "LD HL" + sep + "(" + d.getW() + ")"
	case op == 0x2B:
		result += "DEC HL"
	case op == 0x2C:
		result += "INC L"
	case op == 0x2D:
		result += "DEC L"
	case op == 0x2E:
		result += "LD L" + sep + d.getB()
	case op == 0x2F:
		result += "CPL"
	// == 30:3F
	case op == 0x30:
		result += d.jr("JR", "NC")
	case op == 0x31:
		result += "LD SP" + sep + d.getW()
	case op == 0x32:
		result += "LD (" + d.getW() + ")" + sep + "A"
	case op == 0x33:
		result += "INC SP"
	case op == 0x34:
		result += "INC (HL)"
	case op == 0x35:
		result += "DEC (HL)"
	case op == 0x36:
		result += "LD (HL)" + sep + d.getB()
	case op == 0x37:
		result += "SCF"
	case op == 0x38:
		result += d.jr("JR", "C")
	case op == 0x39:
		result += "ADD HL" + sep + "SP"
	case op == 0x3A:
		result += "LD A" + sep + "(" + d.getW() + ")"
	case op == 0x3B:
		result += "DEC SP"
	case op == 0x3C:
		result += "INC A"
	case op == 0x3D:
		result += "DEC A"
	case op == 0x3E:
		result += "LD A" + sep + d.getB()
	case op == 0x3F:
		result += "CCF"
	case op == 0x76:
		result += "HALT"
	case op >= 0x40 && op <= 0x7F:
		// LD op8, op8
		result += "LD " + operands[(op>>3)&0x07] + sep + operands[op&0x07]
	case op >= 0x80 && op <= 0xBF:
		// ALU op8
		result += aluOp[(op>>3)&0x07] + operands[op&0x07]
	case op == 0xc0:
		result += "RET NZ"
	case op == 0xc1:
		result += "POP BC"
	case op == 0xc2:
		result += d.jp("JP", "NZ")
	case op == 0xc3:
		result += d.jp("JP", "")
	case op == 0xc4:
		result += d.jp("CALL", "NZ")
	case op == 0xc5:
		result += "PUSH BC"
	case op == 0xc6:
		result += "ADD A" + sep + d.getB()
	case op == 0xc7 || op == 0xd7 || op == 0xe7 || op == 0xf7 || op == 0xcf || op == 0xdf || op == 0xef || op == 0xff:
		// RST nnH
		result += fmt.Sprintf("RST %d%dH", (op>>4)&3, (op&1)*8)
	case op == 0xc8:
		result += "RET Z"
	case op == 0xc9:
		result += "RET"
	case op == 0xca:
		result += d.jp("JP", "Z")
	case op == 0xcb:
		result += d.opocodeCB()
	case op == 0xcc:
		result += d.jp("CALL", "Z")
	case op == 0xcd:
		result += d.jp("CALL", "")
	case op == 0xce:
		result += "ADC A" + sep + d.getB()
	case op == 0xd0:
		result += "RET NC"
	case op == 0xd1:
		result += "POP DE"
	case op == 0xd2:
		result += d.jp("JP", "NC")
	case op == 0xd3:
		result += "OUT (" + d.getB() + ")" + sep + "A"
	case op == 0xd4:
		result += d.jp("CALL", "NC")
	case op == 0xd5:
		result += "PUSH DE"
	case op == 0xd6:
		result += "SUB " + d.getB()
	case op == 0xd8:
		result += "RET C"
	case op == 0xd9:
		result += "EXX"
	case op == 0xda:
		result += d.jp("JP", "C")
	case op == 0xdb:
		result += "IN A" + sep + " (" + d.getB() + ")"
	case op == 0xdc:
		result += d.jp("CALL", "C")
	case op == 0xdd:
		result += d.opocodeDD(op)
	case op == 0xde:
		result += "SBC A" + sep + d.getB()
	case op == 0xe0:
		result += "RET PO"
	case op == 0xe1:
		result += "POP HL"
	case op == 0xe2:
		result += d.jp("JP", "PO")
	case op == 0xe3:
		result += "EX (SP)" + sep + "HL"
	case op == 0xe4:
		result += d.jp("CALL", "PO")
	case op == 0xe5:
		result += "PUSH HL"
	case op == 0xe6:
		result += "AND " + d.getB()
	case op == 0xe8:
		result += "RET PE"
	case op == 0xe9:
		result += "JP (HL)"
	case op == 0xea:
		result += d.jp("JP", "PE")
	case op == 0xeb:
		result += "EX DE" + sep + "HL"
	case op == 0xec:
		result += d.jp("CALL", "PE")
	case op == 0xed:
		result += d.opocodeED()
	case op == 0xee:
		result += "XOR " + d.getB()
	case op == 0xf0:
		result += "RET P"
	case op == 0xf1:
		result += "POP AF"
	case op == 0xf2:
		result += d.jp("JP", "P")
	case op == 0xf3:
		result += "DI"
	case op == 0xf4:
		result += d.jp("CALL", "P")
	case op == 0xf5:
		result += "PUSH AF"
	case op == 0xf6:
		result += "OR " + d.getB()
	case op == 0xf8:
		result += "RET M"
	case op == 0xf9:
		result += "LD SP" + sep + "HL"
	case op == 0xfa:
		result += d.jp("JP", "M")
	case op == 0xfb:
		result += "EI"
	case op == 0xfc:
		result += d.jp("CALL", "M")
	case op == 0xfd:
		result += d.opocodeDD(op)
	case op == 0xfe:
		result += "CP " + d.getB()
	default:
		// All unknown as DB
		result += fmt.Sprintf("DB 0x%02X", op)
	}
	return result
}

func (d *Disassembler) getW() string {
	lo := d.core.MemRead(d.pc)
	d.pc++
	hi := d.core.MemRead(d.pc)
	d.pc++
	return fmt.Sprintf("0x%02X%02X", hi, lo)
}

func (d *Disassembler) getB() string {
	lo := d.core.MemRead(d.pc)
	d.pc++
	return fmt.Sprintf("0x%02X", lo)
}

func (d *Disassembler) getRel() string {
	offset := d.core.MemRead(d.pc)
	var sign string
	if int8(offset) < 0 {
		sign = "-"
	} else {
		sign = "+"
	}
	return sign + fmt.Sprintf("0x%02X", offset&0x7F)
}

var shiftOps = []string{"RLC", "RRC", "RL", "RR", "SLA", "SRA", "SLL", "SRL"}

// opocodeCB disassemble Z80 Opcodes, with CB first byte
func (d *Disassembler) opocodeCB() string {
	op := ""
	opcode := d.getByte()
	if opcode <= 0x3F {
		op = shiftOps[opcode>>3&0x07] + operands[opcode&0x7]
	} else {
		op = shiftOps[(opcode>>6&0x03)-1] + operands[opcode&0x7]
	}
	return op
}

func (d *Disassembler) opocodeDD(op byte) string {
	opcode := d.getByte()
	result := ""
	switch opcode {
	case 0x09:
		result = "ADD ii" + sep + "BC"
	case 0x19:
		result = "ADD ii" + sep + "DE"
	case 0x21:
		result = "LD ii" + sep + d.getW()
	case 0x22:
		result = "LD (" + d.getW() + ")" + sep + "ii"
	case 0x23:
		result = "INC ii"
	case 0x24:
		result = "INC IXH"
	case 0x25:
		result = "DEC IXH"
	case 0x26:
		result = "LD IXH" + sep + "n"
	case 0x29:
		result = "ADD ii" + sep + "ii"
	case 0x2A:
		result = "LD ii" + sep + "(" + d.getW() + ")"
	case 0x2B:
		result = "DEC ii"
	case 0x34:
		result = "INC (ii" + d.getRel() + ")"
	case 0x35:
		result = "DEC (ii" + d.getRel() + ")"
	case 0x36:
		result = "LD (ii" + d.getRel() + ")" + sep + "n"
	case 0x39:
		result = "ADD ii" + sep + "SP"
	case 0x46:
		result = "LD B" + sep + "(ii" + d.getRel() + ")"
	case 0x4E:
		result = "LD C" + sep + "(ii" + d.getRel() + ")"
	case 0x56:
		result = "LD D" + sep + "(ii" + d.getRel() + ")"
	case 0x5E:
		result = "LD E" + sep + "(ii" + d.getRel() + ")"
	case 0x66:
		result = "LD H" + sep + "(ii" + d.getRel() + ")"
	case 0x6E:
		result = "LD L" + sep + "(ii" + d.getRel() + ")"
	case 0x70:
		result = "LD (ii" + d.getRel() + ")" + sep + "B"
	case 0x71:
		result = "LD (ii" + d.getRel() + ")" + sep + "C"
	case 0x72:
		result = "LD (ii" + d.getRel() + ")" + sep + "D"
	case 0x73:
		result = "LD (ii" + d.getRel() + ")" + sep + "E"
	case 0x74:
		result = "LD (ii" + d.getRel() + ")" + sep + "H"
	case 0x75:
		result = "LD (ii" + d.getRel() + ")" + sep + "L"
	case 0x77:
		result = "LD (ii" + d.getRel() + ")" + sep + "A"
	case 0x7E:
		result = "LD A" + sep + "(ii" + d.getRel() + ")"
	case 0x86:
		result = "ADD A" + sep + "(ii" + d.getRel() + ")"
	case 0x8E:
		result = "ADC A" + sep + "(ii" + d.getRel() + ")"
	case 0x96:
		result = "SUB (ii" + d.getRel() + ")"
	case 0x9E:
		result = "SBC A" + sep + "(ii" + d.getRel() + ")"
	case 0xA6:
		result = "AND (ii" + d.getRel() + ")"
	case 0xAE:
		result = "XOR (ii" + d.getRel() + ")"
	case 0xB6:
		result = "OR (ii" + d.getRel() + ")"
	case 0xBE:
		result = "CP (ii" + d.getRel() + ")"
	case 0xCB:
		result = d.opocodeDDCB(op, opcode)
	case 0xE1:
		result = "POP ii"
	case 0xE3:
		result = "EX (SP)" + sep + "ii"
	case 0xE5:
		result = "PUSH ii"
	case 0xE9:
		result = "JP (ii)"
	case 0xF9:
		result = "LD SP" + sep + "ii"
	default:
		return fmt.Sprintf("DB 0x%02X, 0x%02X", op, opcode)
	}

	reg := "IX"
	if op == 0xFD {
		reg = "IY"
	}

	return strings.ReplaceAll(result, "ii", reg)
}

//var bitOps = []string{"BIT", "RES", "SET"}

func (d *Disassembler) opocodeDDCB(op1 byte, op2 byte) string {
	opcode := d.getByte()
	result := ""
	switch opcode {
	case 0x06:
		result = "RLC (ii" + d.getRel() + ")"
	case 0x0E:
		result = "RRC (ii" + d.getRel() + ")"
	case 0x16:
		result = "RL (ii" + d.getRel() + ")"
	case 0x1E:
		result = "RR (ii" + d.getRel() + ")"
	case 0x26:
		result = "SLA (ii" + d.getRel() + ")"
	case 0x2E:
		result = "SRA (ii" + d.getRel() + ")"
	case 0x3E:
		result = "SRL (ii" + d.getRel() + ")"
	case 0x46:
		result = "BIT 0" + sep + "(ii" + d.getRel() + ")"
	case 0x4E:
		result = "BIT 1" + sep + "(ii" + d.getRel() + ")"
	case 0x56:
		result = "BIT 2" + sep + "(ii" + d.getRel() + ")"
	case 0x5E:
		result = "BIT 3" + sep + "(ii" + d.getRel() + ")"
	case 0x66:
		result = "BIT 4" + sep + "(ii" + d.getRel() + ")"
	case 0x6E:
		result = "BIT 5" + sep + "(ii" + d.getRel() + ")"
	case 0x76:
		result = "BIT 6" + sep + "(ii" + d.getRel() + ")"
	case 0x7E:
		result = "BIT 7" + sep + "(ii" + d.getRel() + ")"
	case 0x86:
		result = "RES 0" + sep + "(ii" + d.getRel() + ")"
	case 0x8E:
		result = "RES 1" + sep + "(ii" + d.getRel() + ")"
	case 0x96:
		result = "RES 2" + sep + "(ii" + d.getRel() + ")"
	case 0x9E:
		result = "RES 3" + sep + "(ii" + d.getRel() + ")"
	case 0xA6:
		result = "RES 4" + sep + "(ii" + d.getRel() + ")"
	case 0xAE:
		result = "RES 5" + sep + "(ii" + d.getRel() + ")"
	case 0xB6:
		result = "RES 6" + sep + "(ii" + d.getRel() + ")"
	case 0xBE:
		result = "RES 7" + sep + "(ii" + d.getRel() + ")"
	case 0xC6:
		result = "SET 0" + sep + "(ii" + d.getRel() + ")"
	case 0xCE:
		result = "SET 1" + sep + "(ii" + d.getRel() + ")"
	case 0xD6:
		result = "SET 2" + sep + "(ii" + d.getRel() + ")"
	case 0xDE:
		result = "SET 3" + sep + "(ii" + d.getRel() + ")"
	case 0xE6:
		result = "SET 4" + sep + "(ii" + d.getRel() + ")"
	case 0xEE:
		result = "SET 5" + sep + "(ii" + d.getRel() + ")"
	case 0xF6:
		result = "SET 6" + sep + "(ii" + d.getRel() + ")"
	case 0xFE:
		result = "SET 7" + sep + "(ii" + d.getRel() + ")"
	default:
		result = fmt.Sprintf("DB 0x%02X, 0x%02X, 0x%02X", op1, op2, opcode)
	}
	return result
}

func (d *Disassembler) opocodeED() string {
	opcode := d.getByte()
	result := ""
	switch opcode {
	case 0x40:
		result = "IN B" + sep + "(C)"
	case 0x41:
		result = "OUT (C)" + sep + "B"
	case 0x42:
		result = "SBC HL" + sep + "BC"
	case 0x43:
		result = "LD (" + d.getW() + ")" + sep + "BC"
	case 0x44, 0x4C, 0x54, 0x5C, 0x64, 0x6C, 0x74, 0x7C:
		result = "NEG"
	case 0x45, 0x55, 0x5D, 0x65, 0x6D, 0x75, 0x7D:
		result = "RETN"
	case 0x46, 0x4E, 0x66, 0x6E:
		result = "IM 0"
	case 0x47:
		result = "LD I" + sep + "A"
	case 0x48:
		result = "IN C" + sep + "(C)"
	case 0x49:
		result = "OUT (C)" + sep + "C"
	case 0x4A:
		result = "ADC HL" + sep + "BC"
	case 0x4B:
		result = "LD BC" + sep + "(" + d.getW() + ")"
	case 0x4D:
		result = "REТI"
	case 0x4F:
		result = "LD R" + sep + "A"
	case 0x50:
		result = "IN D" + sep + "(C)"
	case 0x51:
		result = "OUT (C)" + sep + "D"
	case 0x52:
		result = "SBC HL" + sep + "DE"
	case 0x53:
		result = "LD (nn)" + sep + "DE"
	case 0x56, 0x76:
		result = "IM 1"
	case 0x57:
		result = "LD A" + sep + "I"
	case 0x58:
		result = "IN E" + sep + "(C)"
	case 0x59:
		result = "OUT (C)" + sep + "E"
	case 0x5A:
		result = "ADC HL" + sep + "DE"
	case 0x5B:
		result = "LD DE" + sep + "(" + d.getW() + ")"
	case 0x5E, 0x7E:
		result = "IM 2"
	case 0x5F:
		result = "LD A" + sep + "R"
	case 0x60:
		result = "IN H" + sep + "(C)"
	case 0x61:
		result = "OUT (C)" + sep + "H"
	case 0x62:
		result = "SBC HL" + sep + "HL"
	case 0x63:
		result = "LD (nn)" + sep + "HL"
	case 0x67:
		result = "RRD"
	case 0x68:
		result = "IN L" + sep + " (C)"
	case 0x69:
		result = "OUT (C)" + sep + "L"
	case 0x6A:
		result = "ADC HL" + sep + " HL"
	case 0x6B:
		result = "LD HL" + sep + " (nn)"
	case 0x6F:
		result = "RLD"
	case 0x70:
		result = "INF"
	case 0x71:
		result = "OUT (C)" + sep + " 0"
	case 0x72:
		result = "SBC HL" + sep + "SP"
	case 0x73:
		result = "LD (nn)" + sep + "SP"
	case 0x78:
		result = "IN A" + sep + "(C)"
	case 0x79:
		result = "OUT (C)" + sep + "A"
	case 0x7A:
		result = "ADC HL" + sep + "SP"
	case 0x7B:
		result = "LD SP" + sep + "(" + d.getW() + ")"
	case 0xA0:
		result = "LDI"
	case 0xA1:
		result = "CPI"
	case 0xA2:
		result = "INI"
	case 0xA3:
		result = "OUTI"
	case 0xA8:
		result = "LDD"
	case 0xA9:
		result = "CPD"
	case 0xAA:
		result = "IND"
	case 0xAB:
		result = "OUTD"
	case 0xB0:
		result = "LDIR"
	case 0xB1:
		result = "CPIR"
	case 0xB2:
		result = "INIR"
	case 0xB3:
		result = "OTIR"
	case 0xB8:
		result = "LDDR"
	case 0xB9:
		result = "CPDR"
	case 0xBA:
		result = "INDR"
	case 0xBB:
		result = "OTDR"
	default:
		result = fmt.Sprintf("DB 0xED, 0x%02X", opcode)
	}
	return result
}
