// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && cgo

package appwindow

/*
// __GLIBC__ is defined by glibc's own headers, so one is included first:
// without it the stub below was always compiled.
#include <stdlib.h>
#ifdef __GLIBC__
#include <malloc.h>
static void trimHeap(void) { malloc_trim(0); }
#else
static void trimHeap(void) {}
#endif
*/
import "C"

// trimCHeap gives the system the free pages of every C heap. glibc keeps
// what the GPU driver freed in its arenas and returns it by itself only from
// the top of the main one: with Mesa, every closed window left about 30 MB.
func trimCHeap() { C.trimHeap() }
