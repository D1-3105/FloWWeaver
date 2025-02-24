package InputStreamShard

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"gocv.io/x/gocv"
	"io"
	"log/slog"
)

func ToMatrix(shard *StreamShard, mat *gocv.Mat) bool {
	matData := shard
	var newMat gocv.Mat
	var err error

	if !matData.Gzipped {
		newMat, err = gocv.NewMatFromBytes(
			int(matData.Height),
			int(matData.Width),
			gocv.MatType(matData.MatType),
			matData.ImageData,
		)
	} else {
		reader, err := gzip.NewReader(bytes.NewReader(matData.ImageData))
		if err != nil {
			slog.Error(fmt.Sprintf("Error creating gzip reader: %v", err))
			return false
		}
		defer func(reader *gzip.Reader) {
			_ = reader.Close()
		}(reader)

		decompressedData, err := io.ReadAll(reader)
		if err != nil {
			slog.Error(fmt.Sprintf("Error decompressing image data: %v", err))
			return false
		}

		newMat, err = gocv.NewMatFromBytes(
			int(matData.Height),
			int(matData.Width),
			gocv.MatType(matData.MatType),
			decompressedData,
		)
	}

	if err != nil {
		slog.Error(fmt.Sprintf("Error creating Mat from bytes: %v", err))
		return false
	}

	*mat = newMat
	return true
}
