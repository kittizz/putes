package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
)

func GzipUnData(data []byte) (resData []byte) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err := gzip.NewReader(b)
	if err != nil {
		log.Fatalf("gzip decompression failed: %s", err)
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		log.Fatalf("gzip decompression failed: %s", err)
	}

	resData = resB.Bytes()

	return
}

func GzipData(data []byte) (compressedData []byte) {
	var b bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&b, gzip.DefaultCompression)

	_, err := gz.Write(data)
	if err != nil {
		log.Fatalf("gzip compression failed: %s", err)
	}

	if err = gz.Flush(); err != nil {
		log.Fatalf("gzip compression failed: %s", err)
	}

	if err = gz.Close(); err != nil {
		log.Fatalf("gzip compression failed: %s", err)
	}

	compressedData = b.Bytes()

	return
}
