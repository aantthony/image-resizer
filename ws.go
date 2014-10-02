package main

import "fmt"
import "github.com/nfnt/resize"
import "github.com/rwcarlsen/goexif/exif"
import "net/http"
import "strings"
import "strconv"
import "io/ioutil"
import "image"
import "image/jpeg"
import "image/color"
import "image/draw"

import "bytes"

func fixExif(s image.Image, o int64) image.Image {
	if o == 1 {
		return s
	}
	sr := s.Bounds()
	dWidth := sr.Max.X
	dHeight := sr.Max.Y

	if o == 6 || o == 7 || o == 5 || o == 8 {
		dWidth = sr.Max.Y
		dHeight = sr.Max.X
	}

	d := image.NewRGBA(image.Rect(0, 0, int(dWidth), int(dHeight)))
	dr := d.Bounds()
	ds := d.Stride

	for x := 0; x < dr.Max.X; x++ {
		for y := 0; y < dr.Max.Y; y++ {
			var sColor color.Color
			if o == 1 {
				sColor = s.At(x, y)
			} else if o == 2 {
				sColor = s.At(sr.Max.X-x, y)
			} else if o == 3 {
				sColor = s.At(sr.Max.X-x, sr.Max.Y-y)
			} else if o == 4 {
				sColor = s.At(x, sr.Max.Y-y)
			} else if o == 5 {
				sColor = s.At(y, x)
			} else if o == 6 {
				sColor = s.At(y, sr.Max.Y-x)
			} else if o == 7 {
				sColor = s.At(sr.Max.X-y, sr.Max.Y-x)
			} else if o == 8 {
				sColor = s.At(sr.Max.X-y, x)
			} else {
				sColor = s.At(x, y)
			}
			r, g, b, _ := sColor.RGBA()
			loc := (y-dr.Min.Y)*ds + (x-dr.Min.X)*4
			d.Pix[loc+0] = uint8(r)
			d.Pix[loc+1] = uint8(g)
			d.Pix[loc+2] = uint8(b)
			d.Pix[loc+3] = 255
		}
	}

	return d
}

func handler(w http.ResponseWriter, r *http.Request) {
	h := strings.Split(r.URL.Path, ".jpg/")
	if len(h) != 2 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Invalid Path")
		return
	}
	basePath := h[1]
	filename := strings.Split(strings.Split(h[0], "/")[1], ".")
	size := strings.Split(filename[0], "x")

	outWidth, _ := strconv.ParseInt(size[0], 0, 64)
	outHeight, _ := strconv.ParseInt(size[1], 0, 64)
	if outWidth <= 0 || outHeight <= 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Invalid size")
		return
	}
	if outWidth >= 4096 || outHeight >= 4096 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Invalid size")
		return
	}
	valid := true

	// if (strings.Index(basePath, "example.com/") == 0) {
	//   valid = true
	// }

	if !valid {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Invalid Domain")
		return
	}

	s := []string{"http://", basePath}

	res, err := http.Get(strings.Join(s, ""))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Could not request image")
		return
	}

	if res.StatusCode != 200 {
		if res.StatusCode == 404 {
			w.WriteHeader(404)
			fmt.Fprint(w, "Not Found")
			return
		}
		fmt.Fprint(w, "Requested image but got status: ", res.StatusCode)
		return
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Could not read data")
		return
	}

	reader := bytes.NewReader(buf)

	img, err := jpeg.Decode(reader)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Could not decode JPEG")
		return
	}

	reader.Seek(0, 0)

	x, err := exif.Decode(reader)
	orientation := int64(1)

	if err == nil {
		orientationTag, err := x.Get(exif.Orientation)
		if err == nil {
			orientation = orientationTag.Int(0)
		}
	}

	img = fixExif(img, orientation)

	sourceWidth := img.Bounds().Max.X
	sourceHeight := img.Bounds().Max.Y

	// First attempt to match width:

	aspect := float64(sourceHeight) / float64(sourceWidth)
	width := outWidth
	height := int64(aspect * float64(width))

	portrait := false

	if height >= outHeight {
		// looks good!
		portrait = true
	} else {
		portrait = false
		height = outHeight
		width = int64(float64(outHeight) / aspect)
	}

	var resized image.Image
	n := image.NewRGBA(image.Rect(0, 0, int(outWidth), int(outHeight)))

	if portrait {
		resized = resize.Resize(uint(width), 0, img, resize.Lanczos3)
		heightDiff := int((height - outHeight) / 2)
		draw.Draw(n, n.Bounds(), resized, image.Point{0, heightDiff}, draw.Src)
	} else {
		resized = resize.Resize(0, uint(height), img, resize.Lanczos3)
		widthDiff := int((width - outWidth) / 2)
		draw.Draw(n, n.Bounds(), resized, image.Point{widthDiff, 0}, draw.Src)
	}

	quality := 85
	if width <= 256 {
		quality = 95
	}
	w.Header().Set("Cache-Control", "no-transform,public,max-age=2592000,s-maxage=2592000")
	jpeg.Encode(w, n, &jpeg.Options{quality})
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}

// func init () {
//   http.HandleFunc("/", handler)
// }
