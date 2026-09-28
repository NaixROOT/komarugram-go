/* NEON intrinsics for libvpx's arm64 decoder, as WebAssembly SIMD128
 * through SIMDe; see build.sh. */
#ifndef VPXWASM_ARM_NEON_H
#define VPXWASM_ARM_NEON_H

#define SIMDE_ENABLE_NATIVE_ALIASES
/* wasm32 has no _Float16, which SIMDe takes clang to have; the decoder uses
 * no half floats. 1 is SIMDE_FLOAT16_API_PORTABLE. */
#define SIMDE_FLOAT16_API 1
#include <simde/arm/neon.h>

#endif
