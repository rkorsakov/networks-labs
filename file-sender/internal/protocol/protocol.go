package protocol

import (
	"encoding/binary"
	"io"
)

func WriteFileMetadata(w io.Writer, filename string, fileSize int64) error {
	filenameBytes := []byte(filename)
	if err := binary.Write(w, binary.BigEndian, uint32(len(filenameBytes))); err != nil {
		return err
	}
	if _, err := w.Write(filenameBytes); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, fileSize); err != nil {
		return err
	}
	return nil
}

func ReadFileMetadata(r io.Reader) (string, int64, error) {
	var filenameLen uint32
	if err := binary.Read(r, binary.BigEndian, &filenameLen); err != nil {
		return "", 0, err
	}
	filenameBytes := make([]byte, filenameLen)
	if _, err := io.ReadFull(r, filenameBytes); err != nil {
		return "", 0, err
	}
	var fileSize int64
	if err := binary.Read(r, binary.BigEndian, &fileSize); err != nil {
		return "", 0, err
	}
	return string(filenameBytes), fileSize, nil
}
