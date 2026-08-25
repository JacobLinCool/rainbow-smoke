# Rainbow Smoke

Rainbow Smoke renders one deterministic image with the Exact Boundary-Color Octree algorithm. For a fixed seed, center, and image size, the raw pixels are reproducible across runs.

The renderer preserves the original boundary-growth semantics:

- the next pixel minimizes squared RGB distance to an active boundary edge;
- unpainted neighbours participate as virtual black;
- equal distances use the canonical frontier position and neighbour slot;
- frontier removal uses deterministic swap-remove ordering.

A fixed-depth dense RGB octree replaces the full-frontier search. A scalar implementation exists only in tests as a correctness oracle.

## Build

Go 1.24 or newer is required.

```sh
make build
```

## Render

```sh
./smoke -width 256 -height 256 -seed 20260825 -output smoke.png
```

The initial pixel defaults to the image center. Override it with `-x` and `-y`.

## Benchmark

Benchmark placement without PNG encoding:

```sh
make benchmark-1024
```

The JSON report separates palette generation, engine initialization, placement, and PNG encoding. The dense RGB index occupies 76,695,844 bytes (about 73.15 MiB) at every image size.

## Verify

```sh
make test
```

The test suite checks octree nearest-neighbour results against a scalar RGB oracle and compares complete rendered pixel buffers against a scalar full-frontier renderer.

This repository is forked from [Ravenslofty/rbsmoke](https://github.com/Ravenslofty/rbsmoke).
