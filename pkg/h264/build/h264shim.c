/* Minimal H.264 decoding surface for the host: send a sample, receive
 * pictures, read back their planes.
 *
 * Everything the host is allowed to do with this module goes through these
 * exports. The module has no main and no reason to touch the outside world. */
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "libavcodec/avcodec.h"

typedef struct {
  AVCodecContext *ctx;
  AVPacket *packet;
  AVFrame *frame;
  int ready; /* a picture was received into frame */
} h264w_decoder;

/* A decoder-only build leaves out the AV1 film grain helpers, which the
 * H.264 SEI code refers to for film grain it does not apply. */
void ff_aom_uninit_film_grain_params(void *params) { (void)params; }

/* Statuses the host sees: 0 done, H264W_AGAIN for "receive pictures" or
 * "send samples", H264W_EOF after the end of the stream, or a negative
 * AVERROR. They do not depend on the errno numbers of this libc. */
#define H264W_AGAIN 1
#define H264W_EOF 2

static int status(int err) {
  if (err == AVERROR(EAGAIN)) return H264W_AGAIN;
  if (err == AVERROR_EOF) return H264W_EOF;
  return err < 0 ? err : 0;
}

__attribute__((export_name("h264w_alloc"))) uint8_t *h264w_alloc(size_t size) {
  return (uint8_t *)malloc(size);
}

__attribute__((export_name("h264w_dealloc"))) void h264w_dealloc(uint8_t *ptr) {
  free(ptr);
}

__attribute__((export_name("h264w_drop"))) void h264w_drop(h264w_decoder *d) {
  if (d == NULL) return;
  av_frame_free(&d->frame);
  av_packet_free(&d->packet);
  avcodec_free_context(&d->ctx);
  free(d);
}

/* Starts a decoder for samples described by config, an avcC record. */
__attribute__((export_name("h264w_new"))) h264w_decoder *h264w_new(const uint8_t *config,
                                                                   size_t size) {
  const AVCodec *codec = avcodec_find_decoder(AV_CODEC_ID_H264);
  if (codec == NULL || size == 0) return NULL;
  h264w_decoder *d = (h264w_decoder *)calloc(1, sizeof(h264w_decoder));
  if (d == NULL) return NULL;
  d->ctx = avcodec_alloc_context3(codec);
  d->packet = av_packet_alloc();
  d->frame = av_frame_alloc();
  if (d->ctx == NULL || d->packet == NULL || d->frame == NULL) goto fail;
  d->ctx->extradata = av_mallocz(size + AV_INPUT_BUFFER_PADDING_SIZE);
  if (d->ctx->extradata == NULL) goto fail;
  memcpy(d->ctx->extradata, config, size);
  d->ctx->extradata_size = (int)size;
  d->ctx->thread_count = 1;
  if (avcodec_open2(d->ctx, codec, NULL) < 0) goto fail;
  return d;
fail:
  h264w_drop(d);
  return NULL;
}

/* Skips the deblocking filter: faster, with blockier pictures. */
__attribute__((export_name("h264w_fast"))) void h264w_fast(h264w_decoder *d, int fast) {
  if (d == NULL) return;
  d->ctx->skip_loop_filter = fast ? AVDISCARD_ALL : AVDISCARD_DEFAULT;
}

/* Sends one sample, or with size 0 the end of the stream. H264W_AGAIN asks
 * to receive pictures first. */
__attribute__((export_name("h264w_send"))) int h264w_send(h264w_decoder *d,
                                                          const uint8_t *data, size_t size) {
  if (d == NULL) return -1;
  if (size == 0) return status(avcodec_send_packet(d->ctx, NULL));
  if (size > INT32_MAX - AV_INPUT_BUFFER_PADDING_SIZE) return AVERROR(EINVAL);
  int err = av_new_packet(d->packet, (int)size);
  if (err < 0) return err;
  memcpy(d->packet->data, data, size);
  err = avcodec_send_packet(d->ctx, d->packet);
  av_packet_unref(d->packet);
  return status(err);
}

/* Receives the next picture. Returns 0 with one, H264W_AGAIN when the
 * decoder needs more samples, H264W_EOF after the end of the stream. */
__attribute__((export_name("h264w_receive"))) int h264w_receive(h264w_decoder *d) {
  if (d == NULL) return -1;
  av_frame_unref(d->frame);
  d->ready = 0;
  int err = avcodec_receive_frame(d->ctx, d->frame);
  if (err == 0) d->ready = 1;
  return status(err);
}

/* Forgets every sample sent, to start again from a keyframe. */
__attribute__((export_name("h264w_flush"))) void h264w_flush(h264w_decoder *d) {
  if (d == NULL) return;
  av_frame_unref(d->frame);
  d->ready = 0;
  avcodec_flush_buffers(d->ctx);
}

__attribute__((export_name("h264w_width"))) int h264w_width(h264w_decoder *d) {
  return d && d->ready ? d->frame->width : 0;
}

__attribute__((export_name("h264w_height"))) int h264w_height(h264w_decoder *d) {
  return d && d->ready ? d->frame->height : 0;
}

/* Reports whether the picture is 8-bit 4:2:0, the only layout read back. */
__attribute__((export_name("h264w_yuv420"))) int h264w_yuv420(h264w_decoder *d) {
  if (!d || !d->ready) return 0;
  return d->frame->format == AV_PIX_FMT_YUV420P || d->frame->format == AV_PIX_FMT_YUVJ420P;
}

/* Plane 0 is luma, 1 and 2 are the chroma planes. */
__attribute__((export_name("h264w_plane"))) uint8_t *h264w_plane(h264w_decoder *d, int plane) {
  if (!d || !d->ready || plane < 0 || plane > 2) return NULL;
  return d->frame->data[plane];
}

__attribute__((export_name("h264w_stride"))) int h264w_stride(h264w_decoder *d, int plane) {
  if (!d || !d->ready || plane < 0 || plane > 2) return 0;
  return d->frame->linesize[plane];
}
