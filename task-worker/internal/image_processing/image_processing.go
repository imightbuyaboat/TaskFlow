package image_processing

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/imightbuyaboat/TaskFlow/pkg/task"
)

type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

func (ip *ImageProcessor) ExecuteTask(rawPayload interface{}) error {
	data, err := json.Marshal(rawPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal rawPayload: %v", err)
	}

	var payload task.ImageProcessingPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload to ImageProcessingPayload: %v", err)
	}

	src, err := imaging.Open(payload.Path)
	if err != nil {
		return fmt.Errorf("failed to open source image: %v", err)
	}

	if payload.Grayscale {
		src = imaging.Grayscale(src)
	}
	if payload.Invert {
		src = imaging.Invert(src)
	}
	src = imaging.Blur(src, payload.Blur)
	src = imaging.Sharpen(src, payload.Sharpen)
	src = imaging.AdjustGamma(src, payload.Gamma)
	src = imaging.AdjustContrast(src, payload.Contrast)
	src = imaging.AdjustBrightness(src, payload.Brightness)
	src = imaging.AdjustSaturation(src, payload.Saturation)

	lastPointIndex := strings.LastIndex(payload.Path, ".")

	dstPath := payload.Path[:lastPointIndex] + "_" + uuid.New().String() + payload.Path[lastPointIndex:]
	err = imaging.Save(src, dstPath)
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	return nil
}
