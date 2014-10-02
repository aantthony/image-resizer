image-resizer
=============

Image resizing service written in Go


## Install

```bash
go get github.com/nfnt/resize
go get github.com/rwcarlsen/goexif/exif
go build ws.go
./ws
```

## Usage

URL Format:

`http://localhost:8090/{{width}}x{{height}}.jpg/{{original}}`

Example:

`http://localhost:8090/48x48.jpg/lorempixel.com/output/city-q-c-640-480-10.jpg`
