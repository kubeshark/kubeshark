# 水 mizu
standalone web app traffic viewer for Kubernetes

## Download

Download `mizu` for your platform and operating system

### Latest stable release

* for MacOS - Intel 
`curl -Lo mizu https://github.com/up9inc/mizu/releases/latest/download/mizu_darwin_amd64 && chmod 755 mizu`

* for MacOS - Apple Silicon
 `curl -Lo mizu https://github.com/up9inc/mizu/releases/latest/download/mizu_darwin_arm64 && chmod 755 mizu`
 
* for Linux - Intel 64bit
 `curl -Lo mizu https://github.com/up9inc/mizu/releases/latest/download/mizu_linux_amd64 && chmod 755 mizu` 


SHA256 checksums are available on the [Releases](https://github.com/up9inc/mizu/releases) page.

### Development (unstable) build
Pick one from the [Releases](https://github.com/up9inc/mizu/releases) page.

## How to run

1. Find pod you'd like to tap to in your Kubernetes cluster
2. Run `mizu PODNAME` or `mizu REGEX` 
3. Open browser on `http://localhost:8899` as instructed .. 
4. Watch the WebAPI traffic flowing ..
5. Type ^C to stop

## Examples

Run `mizu help` for usage options


To tap specific pod - 
``` 
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 front-end-649fc5fd6-kqbtn       2/2     Running   0          7m
 ..

 $ mizu tap front-end-649fc5fd6-kqbtn
 +front-end-649fc5fd6-kqbtn
 Web interface is now available at http://localhost:8899
 ^C
```

To tap multiple pods using regex - 
``` 
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 carts-66c77f5fbb-fq65r          2/2     Running   0          20m
 catalogue-5f4cb7cf5-7zrmn       2/2     Running   0          20m
 front-end-649fc5fd6-kqbtn       2/2     Running   0          20m
 ..

 $ mizu tap "^ca.*"
 +carts-66c77f5fbb-fq65r
 +catalogue-5f4cb7cf5-7zrmn
 Web interface is now available at http://localhost:8899
 ^C
```

