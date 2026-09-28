<div align="center">
<img src="./assets/logo.svg" width="140" align="center" alt="KomaruGram">

# KomaruGram Go

Telegram Desktop client written from scratch in Go. Feature-rich and security-enhanced foundation. Easy to build and freely distributable, using a permissive license.

[ [English] | [Русский](README_RU.md) ]

[ [Other stuff](./docs/README_FULL.md) | [Widgets and UI](./docs/UI_COMPONENTS.md) ]

</div>

## Why does this exist?

The official Telegram Desktop and its forks have many issues: heavy reliance on the C++ codebase and libraries (TDLib, Qt), a build process that is time-consuming and resource-intensive, and a high barrier to entry for developing your own modifications. In addition, it uses a copyleft license — our repository is completely free in terms of how you distribute it.

## What's inside?

Gio handles the UI. It is a cross-platform, immediate-mode GUI library. It was chosen because it eliminates the typical limitations of frameworks, has no dependencies, doesn't break cross-platform builds, is fully compatible with Wayland, and can be ported to other platforms without any issues.

gotd/td implements MTProto and a little more, which serves as the foundation for all interactions with the Telegram protocol.

wazero is used as a high-performance WebAssembly sandbox for rendering stickers. It ensures cross-platform portability and guarantees code isolation without launching a separate process.

FFmpeg, mpv, VLC and Chromium are available as external integrations that you must provide yourself. However, the application will work even without them, if the user is satisfied with that.

Without FFmpeg, or when chosen in the settings, GIFs and animated avatars play with FFmpeg's H.264 decoder compiled to WebAssembly. It is not built into the application: it is downloaded the first time it is needed from [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm), which holds its sources and build script. To use your own build, point the `KOMARUGRAM_AVCDEC` environment variable at an `avcdec.wasm` file or its URL.

## How to get started

Make a copy of this repository. Ensure that **Go 1.27.1** is installed on your machine — it is the minimum required version. On Linux, Wayland and X11 development libraries may be required.

**Change the directory to the entry point:**
```
cd cmd/messenger
```

### **Linux:**
**Simply with go build:**
```
GOOS=linux go build -ldflags="-s -w" -buildvcs=false -o komarugram
```

### **Windows:**

**Install Gio cmd tools:**:
```
go install gioui.org/cmd/gogio@latest
```
**Building the application with an icon:**
```
gogio -ldflags="-s -w" -icon=../../assets/logo_round.png -target=windows -o komarugram.exe .
```

The build process can consume up to 4 GB of RAM. Keep this in mind and close unnecessary applications during the initial build. All subsequent builds should run instantly.

## Vibecoding

The code in this project was primarily written by the GPT-6 Astra and Claude Opus 5.5 language models. The maintainer is responsible for the concept, selection of the technology stack and libraries, and quality control.

## Thanks

We thank the creator of gotd/td for the excellent library and Gio for an architecture that is both simple and scalable for large applications. Branding is provided free of charge by the [t.me/komarugram](https://t.me/komarugram) project. Thanks augustwise and SvatoshGPT for providing the Claude Code and Codex subscriptions for the needs of this project.

## License

When working with this project, you have no licensing obligations regarding the modification or distribution of the code. Some dependencies require that you respect copyright, but they do not impose any restrictions on the code itself (permissive MIT-compatible licenses).

The one exception is the H.264 decoder, which is FFmpeg's and licensed under the LGPL 2.1 or later. It is kept out of this repository and out of the binary, in [libavcodec-wasm](https://github.com/komarugif/libavcodec-wasm), and loaded at run time, where `KOMARUGRAM_AVCDEC` can replace it.