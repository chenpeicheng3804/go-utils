package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"github.com/andybalholm/brotli"
	"io"
)

func CompressBase64(data string) (string, error) {
	var buf bytes.Buffer
	// gw := gzip.NewWriter(&buf)
	gw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	_, err = gw.Write([]byte(data))
	if err != nil {
		return "", err
	}

	err = gw.Close()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func DecompressBase64(data string) (string, error) {
	compressedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(compressedData)
	gr, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	decompressedData, err := io.ReadAll(gr)
	if err != nil {
		return "", err
	}

	return string(decompressedData), nil
}

func CompressBase64WithBrotli(data string) (string, error) {
	var buf bytes.Buffer
	bw := brotli.NewWriter(&buf)

	_, err := bw.Write([]byte(data))
	if err != nil {
		return "", err
	}

	err = bw.Close()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func DecompressBase64WithBrotli(data string) (string, error) {
	compressedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(compressedData)
	br := brotli.NewReader(buf)

	decompressedData, err := io.ReadAll(br)
	if err != nil {
		return "", err
	}

	return string(decompressedData), nil
}
