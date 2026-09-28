/* Minimal setjmp replacement for the WebAssembly build of libvpx.
 *
 * Real setjmp needs the not-yet-standard exception handling proposal. libvpx
 * only uses it to bail out of a frame that failed to decode, so here the
 * escape hatch becomes a trap: the sandbox dies and the host creates a fresh
 * one. Failing closed is the right behaviour for untrusted input anyway. */
#ifndef WASM_SETJMP_SHIM_H
#define WASM_SETJMP_SHIM_H

typedef int jmp_buf[1];

#define setjmp(env) ((void)(env), 0)
#define _setjmp(env) setjmp(env)

static inline void longjmp(jmp_buf env, int val) {
  (void)env;
  (void)val;
  __builtin_trap();
}
#define _longjmp(env, val) longjmp(env, val)

#endif
