package ldcache

import "strconv"

// Flags represents the flags field of an ld.so.cache entry, encoding both the
// library type (ELF, libc version) and required platform/ABI information.
// Flag values are defined in glibc's sysdeps/generic/ldconfig.h.
type Flags int32

// String returns the flags as a human-readable string matching the format
// shown by ldconfig -p (e.g. "libc6,x86-64").
func (f Flags) String() string {
	if f == -1 {
		return "any"
	}

	// flags encode both type and required platform
	typ := byte(f & 0xff)        // FLAG_TYPE_MASK
	req := byte((f >> 8) & 0xff) // FLAG_REQUIRED_MASK

	var res string

	switch typ {
	case 0: // FLAG_LIBC4
		res = "libc4"
	case 1: // FLAG_ELF
		res = "elf"
	case 2: // FLAG_ELF_LIBC5
		res = "libc5"
	case 3: // FLAG_ELF_LIBC6
		res = "libc6"
	default:
		res = strconv.FormatUint(uint64(typ), 10)
	}

	switch req {
	case 0:
		// do nothing
	case 1: // FLAG_SPARC_LIB64
		res += ",64bit"
	case 2: // FLAG_IA64_LIB64
		res += ",IA-64"
	case 3: // FLAG_X8664_LIB64
		res += ",x86-64"
	case 4: // FLAG_S390_LIB64
		res += ",64bit"
	case 5: // FLAG_POWERPC_LIB64
		res += ",64bit"
	case 6: // FLAG_MIPS64_LIBN32
		res += ",N32"
	case 7: // FLAG_MIPS64_LIBN64
		res += ",64bit"
	case 8: // FLAG_X8664_LIBX32
		res += ",x32"
	case 9: // FLAG_ARM_LIBHF
		res += ",hard-float"
	case 10: // FLAG_AARCH64_LIB64
		res += ",AArch64"
	case 11: // FLAG_ARM_LIBSF
		res += ",soft-float"
	case 12: // FLAG_MIPS_LIB32_NAN2008
		res += ",nan2008"
	case 13: // FLAG_MIPS64_LIBN32_NAN2008
		res += ",N32,nan2008"
	case 14: // FLAG_MIPS64_LIBN64_NAN2008
		res += ",64bit,nan2008"
	case 15: // FLAG_RISCV_FLOAT_ABI_SOFT
		res += ",soft-float"
	case 16: // FLAG_RISCV_FLOAT_ABI_DOUBLE
		res += ",double-float"
	case 17: // FLAG_LARCH_FLOAT_ABI_SOFT
		res += ",soft-float"
	case 18: // FLAG_LARCH_FLOAT_ABI_DOUBLE
		res += ",double-float"
	default:
		res += "," + strconv.FormatUint(uint64(req), 10)
	}

	return res
}
