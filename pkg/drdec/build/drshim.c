/* Minimal decoding surface for the host over dr_libs's dr_mp3, dr_flac and
 * dr_wav: open a file the host reads, read back 16-bit samples, move to a
 * sample.
 *
 * Everything the host is allowed to do with this module goes through these
 * exports, and the file is read through host.read, a range at a time, so
 * that a file that downloads as it plays need not be all there. The module
 * has no main and no reason to touch the outside world: the decoders are
 * built without stdio. */
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#define DR_MP3_IMPLEMENTATION
#define DR_MP3_NO_STDIO
#include "dr_mp3.h"
#define DR_FLAC_IMPLEMENTATION
#define DR_FLAC_NO_STDIO
#include "dr_flac.h"
#define DR_WAV_IMPLEMENTATION
#define DR_WAV_NO_STDIO
#include "dr_wav.h"

/* Formats, as the host names them. */
#define DRW_MP3 1
#define DRW_FLAC 2
#define DRW_WAV 3

/* The most frames one read puts out. */
#define DRW_CHUNK 4096

/* Reads n bytes of the file at offset into dst; returns those read, 0 at
 * its end, below 0 when it cannot be read. */
__attribute__((import_module("host"), import_name("read"))) int32_t host_read(int64_t offset, void *dst, int32_t n);

typedef struct {
  int format;
  int64_t cursor, size;
  drmp3 mp3;
  drmp3_seek_point *seek;
  drflac *flac;
  drwav wav;
  uint32_t channels, rate;
  uint64_t frames;
  int16_t *pcm;
} drw_decoder;

/* The file, as the decoders read it: through host_read, at a cursor. The
 * seek origins of the three decoders are the same: set, current, end. */
static size_t file_read(void *user, void *out, size_t n) {
  drw_decoder *d = (drw_decoder *)user;
  size_t got = 0;
  while (got < n && d->cursor < d->size) {
    int32_t want = (int32_t)(n - got > 1 << 20 ? 1 << 20 : n - got);
    int32_t k = host_read(d->cursor, (uint8_t *)out + got, want);
    if (k <= 0) break;
    d->cursor += k;
    got += (size_t)k;
  }
  return got;
}

static int file_seek(void *user, int offset, int origin) {
  drw_decoder *d = (drw_decoder *)user;
  int64_t at = origin == 0 ? offset : origin == 1 ? d->cursor + offset : d->size + offset;
  if (at < 0 || at > d->size) return 0;
  d->cursor = at;
  return 1;
}

static drmp3_bool32 mp3_seek(void *user, int offset, drmp3_seek_origin origin) { return file_seek(user, offset, (int)origin); }
static drflac_bool32 flac_seek(void *user, int offset, drflac_seek_origin origin) { return file_seek(user, offset, (int)origin); }
static drwav_bool32 wav_seek(void *user, int offset, drwav_seek_origin origin) { return file_seek(user, offset, (int)origin); }

static drmp3_bool32 mp3_tell(void *user, drmp3_int64 *cursor) { *cursor = ((drw_decoder *)user)->cursor; return 1; }
static drflac_bool32 flac_tell(void *user, drflac_int64 *cursor) { *cursor = ((drw_decoder *)user)->cursor; return 1; }
static drwav_bool32 wav_tell(void *user, drwav_int64 *cursor) { *cursor = ((drw_decoder *)user)->cursor; return 1; }

__attribute__((export_name("drw_alloc"))) uint8_t *drw_alloc(size_t size) {
  return (uint8_t *)malloc(size);
}

__attribute__((export_name("drw_dealloc"))) void drw_dealloc(uint8_t *ptr) {
  free(ptr);
}

__attribute__((export_name("drw_close"))) void drw_close(drw_decoder *d) {
  if (d == NULL) return;
  switch (d->format) {
  case DRW_MP3:
    drmp3_uninit(&d->mp3);
    break;
  case DRW_FLAC:
    drflac_close(d->flac);
    break;
  case DRW_WAV:
    drwav_uninit(&d->wav);
    break;
  }
  free(d->seek);
  free(d->pcm);
  free(d);
}

/* Opens the file of format, of size bytes, that host_read reads. NULL for a
 * file it cannot decode. An MP3 knows its length only once read through:
 * exact reads it through, to count its frames and make a seek table, so
 * that a move back costs a few frames, not the file; otherwise its length
 * is 0, for the host to know from elsewhere. */
__attribute__((export_name("drw_open"))) drw_decoder *drw_open(int format, int64_t size, int exact) {
  drw_decoder *d = (drw_decoder *)calloc(1, sizeof(drw_decoder));
  if (d == NULL) return NULL;
  d->size = size;
  int ok = 0;
  switch (format) {
  case DRW_MP3:
    if (drmp3_init(&d->mp3, file_read, mp3_seek, mp3_tell, NULL, d, NULL)) {
      d->format = DRW_MP3;
      d->channels = d->mp3.channels;
      d->rate = d->mp3.sampleRate;
      if (exact) {
        d->frames = drmp3_get_pcm_frame_count(&d->mp3);
        drmp3_uint32 count = 256;
        d->seek = (drmp3_seek_point *)malloc(count * sizeof(drmp3_seek_point));
        if (d->seek != NULL && drmp3_calculate_seek_points(&d->mp3, &count, d->seek)) {
          drmp3_bind_seek_table(&d->mp3, count, d->seek);
        }
      }
      ok = 1;
    }
    break;
  case DRW_FLAC:
    d->flac = drflac_open(file_read, flac_seek, flac_tell, d, NULL);
    if (d->flac != NULL) {
      d->format = DRW_FLAC;
      d->channels = d->flac->channels;
      d->rate = d->flac->sampleRate;
      d->frames = d->flac->totalPCMFrameCount;
      ok = 1;
    }
    break;
  case DRW_WAV:
    if (drwav_init(&d->wav, file_read, wav_seek, wav_tell, d, NULL)) {
      d->format = DRW_WAV;
      d->channels = d->wav.channels;
      d->rate = d->wav.sampleRate;
      d->frames = d->wav.totalPCMFrameCount;
      ok = 1;
    }
    break;
  }
  if (ok && d->channels > 0 && d->channels <= 8 && d->rate > 0) {
    d->pcm = (int16_t *)malloc((size_t)DRW_CHUNK * d->channels * sizeof(int16_t));
  }
  if (!ok || d->pcm == NULL) {
    drw_close(d);
    return NULL;
  }
  return d;
}

__attribute__((export_name("drw_channels"))) int drw_channels(drw_decoder *d) { return (int)d->channels; }
__attribute__((export_name("drw_rate"))) int drw_rate(drw_decoder *d) { return (int)d->rate; }
__attribute__((export_name("drw_frames"))) int64_t drw_frames(drw_decoder *d) { return (int64_t)d->frames; }
__attribute__((export_name("drw_pcm"))) int16_t *drw_pcm(drw_decoder *d) { return d->pcm; }

/* Reads up to DRW_CHUNK frames into drw_pcm, interleaved. Returns the
 * frames read, 0 at the end. */
__attribute__((export_name("drw_read"))) int drw_read(drw_decoder *d) {
  switch (d->format) {
  case DRW_MP3:
    return (int)drmp3_read_pcm_frames_s16(&d->mp3, DRW_CHUNK, d->pcm);
  case DRW_FLAC:
    return (int)drflac_read_pcm_frames_s16(d->flac, DRW_CHUNK, d->pcm);
  case DRW_WAV:
    return (int)drwav_read_pcm_frames_s16(&d->wav, DRW_CHUNK, d->pcm);
  }
  return 0;
}

/* Moves to frame; 1 on success. */
__attribute__((export_name("drw_seek"))) int drw_seek(drw_decoder *d, int64_t frame) {
  switch (d->format) {
  case DRW_MP3:
    return drmp3_seek_to_pcm_frame(&d->mp3, (drmp3_uint64)frame);
  case DRW_FLAC:
    return drflac_seek_to_pcm_frame(d->flac, (drflac_uint64)frame);
  case DRW_WAV:
    return drwav_seek_to_pcm_frame(&d->wav, (drwav_uint64)frame);
  }
  return 0;
}
