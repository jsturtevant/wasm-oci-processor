## Prototype

This was a proto type for https://github.com/deislabs/containerd-wasm-shims/issues/89

## set up containerd

Add the following to the containerd config:

```
[stream_processors]
  [stream_processors."io.containerd.oci.wasm.v1.wasm"]
    accepts = ["application/vnd.w3c.wasm.module.v1+wasm"]
    path = "wasm-oci-processor"
    returns = "application/vnd.oci.image.layer.v1.tar"
```

## generate a wasm OCI image

use `oci-tar-builder` crate to build image with custom media types (`image.tar`).

## local testing

use `layer.tar` which is a single layer with wasm file in it extracted from `image.tar`

## upload wasm OCI image to registry

```
regctl image import localhost:5000/wasi-demo:latest /mnt/c/Users/jstur/projects/runwasi/target/wasm32-wasi/debug/img.tar
```
## doesn't work

Files get added to the tar but containerd fails since [diff changed](https://github.com/containerd/containerd/blob/d8b68e3ccc2c859f20f08041024af5be0565601b/rootfs/apply.go#L167) on [extraction](https://github.com/containerd/containerd/blob/06e085c8b50a4953c6a9ea636b459ce3a18964e4/diff/windows/windows.go#L130-L132). It also only works at layer level not the manifest level so wouldn't be able to stich things together.

```
unpacking windows/amd64 sha256:a9bf3c52996ef466e7da0ad4652c6ae7029b6045902161d1f7ee00946fada7a5...
time="2023-05-03T17:12:21-07:00" level=info msg="apply failure, attempting cleanup" error="wrong diff id calculated on extraction \"sha256:201171c178cfb419d4ab25429ffc430eac6e2ed28ec1cc3df4bae5346e20df00\"" key="extract-397179500-PA_p sha256:80aaf4e83cab69cce04481df5000b374834660da8c11493ba1aa380ebc4054cd"
ctr: wrong diff id calculated on extraction "sha256:201171c178cfb419d4ab25429ffc430eac6e2ed28ec1cc3df4bae5346e20df00"
```

# credit
Example heavily modified from https://github.com/containerd/imgcrypt/tree/main/cmd/ctd-decoder under Apache 2.0 license