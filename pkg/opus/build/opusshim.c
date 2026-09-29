/* Minimal Opus decoding surface for the host: feed a packet, read back
 * 48 kHz mono 16-bit samples.
 *
 * Everything the host is allowed to do with this module goes through these
 * exports. The module has no main and no reason to touch the outside world. */
#include <stdint.h>
#include <stdlib.h>

#include "opus.h"

__attribute__((export_name("opusw_alloc"))) uint8_t *opusw_alloc(size_t size) {
  return (uint8_t *)malloc(size);
}

__attribute__((export_name("opusw_dealloc"))) void opusw_dealloc(uint8_t *ptr) {
  free(ptr);
}

/* A decoder at 48 kHz, mono: a stereo stream is mixed down, as a voice
 * message wants. NULL on failure. */
__attribute__((export_name("opusw_new"))) OpusDecoder *opusw_new(void) {
  int err = 0;
  OpusDecoder *decoder = opus_decoder_create(48000, 1, &err);
  return err == OPUS_OK ? decoder : NULL;
}

/* Forgets what the decoder heard, before decoding from another point of the
 * stream. */
__attribute__((export_name("opusw_reset"))) void opusw_reset(OpusDecoder *decoder) {
  if (decoder != NULL) opus_decoder_ctl(decoder, OPUS_RESET_STATE);
}

__attribute__((export_name("opusw_drop"))) void opusw_drop(OpusDecoder *decoder) {
  if (decoder != NULL) opus_decoder_destroy(decoder);
}

/* Decodes one packet into pcm, which holds room for max samples. Returns the
 * samples decoded, or a negative Opus error. */
__attribute__((export_name("opusw_decode"))) int opusw_decode(OpusDecoder *decoder,
                                                              const uint8_t *data,
                                                              int32_t size,
                                                              int16_t *pcm,
                                                              int32_t max) {
  if (decoder == NULL) return OPUS_BAD_ARG;
  return opus_decode(decoder, data, size, pcm, max, 0);
}
