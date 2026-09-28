/* Minimal VP9 decoding surface for the host: feed a frame, read back planes.
 *
 * Everything the host is allowed to do with this module goes through these
 * exports. The module has no main and no reason to touch the outside world. */
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "vpx/vpx_decoder.h"
#include "vpx/vp8dx.h"

typedef struct {
  vpx_codec_ctx_t codec;
  vpx_image_t *image;
  int ready;
} vpxw_decoder;

__attribute__((export_name("vpxw_alloc"))) uint8_t *vpxw_alloc(size_t size) {
  return (uint8_t *)malloc(size);
}

__attribute__((export_name("vpxw_dealloc"))) void vpxw_dealloc(uint8_t *ptr) {
  free(ptr);
}

__attribute__((export_name("vpxw_new"))) vpxw_decoder *vpxw_new(void) {
  vpxw_decoder *decoder = (vpxw_decoder *)calloc(1, sizeof(vpxw_decoder));
  if (decoder == NULL) return NULL;
  vpx_codec_dec_cfg_t cfg;
  memset(&cfg, 0, sizeof(cfg));
  cfg.threads = 1;
  if (vpx_codec_dec_init(&decoder->codec, vpx_codec_vp9_dx(), &cfg, 0)) {
    free(decoder);
    return NULL;
  }
  decoder->ready = 1;
  return decoder;
}

__attribute__((export_name("vpxw_drop"))) void vpxw_drop(vpxw_decoder *decoder) {
  if (decoder == NULL) return;
  if (decoder->ready) vpx_codec_destroy(&decoder->codec);
  free(decoder);
}

/* Decodes one compressed frame. Returns 0 on success, or the vpx error code. */
__attribute__((export_name("vpxw_decode"))) int vpxw_decode(vpxw_decoder *decoder,
                                                            const uint8_t *data,
                                                            size_t size) {
  if (decoder == NULL || !decoder->ready) return -1;
  vpx_codec_err_t err = vpx_codec_decode(&decoder->codec, data, (unsigned int)size, NULL, 0);
  if (err != VPX_CODEC_OK) return (int)err;
  vpx_codec_iter_t iter = NULL;
  decoder->image = vpx_codec_get_frame(&decoder->codec, &iter);
  return decoder->image == NULL ? -2 : 0;
}

__attribute__((export_name("vpxw_width"))) int vpxw_width(vpxw_decoder *decoder) {
  return decoder && decoder->image ? (int)decoder->image->d_w : 0;
}

__attribute__((export_name("vpxw_height"))) int vpxw_height(vpxw_decoder *decoder) {
  return decoder && decoder->image ? (int)decoder->image->d_h : 0;
}

/* Plane 0 is luma, 1 and 2 are the chroma planes. */
__attribute__((export_name("vpxw_plane"))) uint8_t *vpxw_plane(vpxw_decoder *decoder, int plane) {
  if (!decoder || !decoder->image || plane < 0 || plane > 2) return NULL;
  return decoder->image->planes[plane];
}

__attribute__((export_name("vpxw_stride"))) int vpxw_stride(vpxw_decoder *decoder, int plane) {
  if (!decoder || !decoder->image || plane < 0 || plane > 2) return 0;
  return decoder->image->stride[plane];
}

/* Chroma subsampling, so the host knows how to read the planes. */
__attribute__((export_name("vpxw_subsampling"))) int vpxw_subsampling(vpxw_decoder *decoder) {
  if (!decoder || !decoder->image) return -1;
  return decoder->image->x_chroma_shift * 10 + decoder->image->y_chroma_shift;
}
