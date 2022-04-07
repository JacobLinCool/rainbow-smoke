# Rainbow Smoke

A command line utility for generating beautiful rainbow smoke images.

```sh
❯ ./smoke -help                  
Usage of ./smoke:
  -config string
        Configuration file
  -cpu string
        Write CPU profile to the given file
  -diff string
        Difference algorithm, can be: rgb or lab (default "rgb")
  -dist string
        Output directory (default "img")
  -fit string
        Fitness algorithm, can be: min or sum (default "min")
  -height int
        Rendered image height, must be at least 2 (default -1)
  -help
        Show this help
  -mem string
        Write memory profile to the given file
  -name string
        Name of the creation
  -seed int
        Seed for random number generator (default -9223372036854775808)
  -select string
        Selection algorithm, can be: smallest or greatest (default "smallest")
  -sort string
        Sorting algorithm, can be: hcl, hsv, random, or none (default "hcl")
  -step int
        Step for progress image (default -1)
  -width int
        Rendered image width, must be at least 2 (default -1)
  -x int
        X coordinate of the center of the smoke (default -1)
  -y int
        Y coordinate of the center of the smoke (default -1)
```

> This repository is forked from [Ravenslofty/rbsmoke](https://github.com/Ravenslofty/rbsmoke).
