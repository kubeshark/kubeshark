# 水 mizu
standalone web app traffic viewer for Kubernetes

## Download

Download `mizu` for your platform and operating system

### Latest stable release

* for MacOS - `curl -o mizu https://github.com/up9inc/mizu/releases/download/latest/mizu_darwin_amd64 && chmod 755 mizu`
* for Linux - `curl -o mizu https://github.com/up9inc/mizu/releases/download/latest/mizu_linux_amd64 && chmod 755 mizu`

### Development (unstable) build
Pick one from the [Releases](https://github.com/up9inc/mizu/releases) page.

## How to run

1. Find pod you'd like to tap to in your Kubernetes cluster
2. Run `mizu PODNAME` or `mizu REGEX` 
3. Open browser on `http://localhost:8899` as instructed .. 
4. Watch the WebAPI traffic flowing ..

## Examples
TBD
