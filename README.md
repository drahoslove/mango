# Mango

A visualiser of **Man**delbrot set written in **Go**

 - real-time navigation
 - multithreading support
 - image saving / loading


<img src="images/set_1664718211_(-0.5+0i)_1_0_.png" width="256"
/><img src="images/set_1663476672_(-0.8643325482946538-0.2305787007080872i)_4.745313281212577e+07_2_.png" width="256"
/><img src="images/set_1663582256_(-0.6699236201218569-0.4577691518426872i)_7.774721279938687e+11_1_.png" width="256"
/><img src="images/set_1663536017_(0.2713047044368717-0.5857106239284461i)_2.3726566406062886e+07_3_.png" width="256"
/><img src="images/set_1663549060_(-0.6699236121894319-0.45776914126597795i)_1.7592186044416e+13_2_.png" width="256"
/><img src="images/set_1663528471_(-0.07459391968701061+0.9696561653342303i)_2.9658208007578608e+06_3_.png" width="256"
/>


## Installation

[Download binary for Windows](https://github.com/drahoslove/mango/releases/latest/download/mango.exe)

 &mdash; or &mdash;

`go install github.com/drahoslove/mango@latest`

```cd `go env GOPATH`/bin```

## Examples of usage

`./mango` - default view

`./mango image.png` - to load saved image (you can also drag'n'drop image file to the executable file) - the state si saved in the file name

`WORKERS=16 ./mango`- to change number of rendering gorutines (threads)

## Key binding

### navigation

- `pgDn` / `pgUp` or `wheel ↕` - zoom in / zoom out
- `left click` - center to point
- `←` `→` `↑` `↓` - move left/right/up/down
- `R` - reset zoom and position


### visual
- `1` `2` `3` `4` - switch between coloring modes
- `H` - set number of iterations to High (32k)
- `J` - double the number of iterations
- `K` - halve the number of iterations
- `L` - set number of iterations to Low (1k)

### images
- `ctrl`+`S` - save current image
- `ctrl`+`O` - load saved image
