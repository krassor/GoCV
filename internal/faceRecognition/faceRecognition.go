package faceRecognition

import (
	"context"
	"fmt"
	"image"
	"os"

	"image/color"
	"path/filepath"
	"strings"

	"github.com/krassor/GoCV/internal/pkg/utils"
	"github.com/rs/zerolog/log"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

type FaceDetector interface {
	DetectAllFacesOnCapture(src *gocv.Mat) (faces []gocv.Mat, rect []image.Rectangle, err error)
}

type faceRecognition struct {
	faceDetector FaceDetector
}

func NewFaceRecognition(fd FaceDetector) *faceRecognition {
	return &faceRecognition{
		faceDetector: fd,
	}
}

type FaceRecognitionConfig struct {
	ModelPath  string
	Confidence float32
}

// OpenVideoCapture return VideoCapture specified by device ID if inputVideo is a number.
// Return VideoCapture created from video file, URL, or GStreamer pipeline if inputVideo is a string.
func (fr *faceRecognition) Recognition(ctx context.Context, config FaceRecognitionConfig, inputVideo interface{}) {

	log.Info().Msgf("Start reading stream \"%s\"", inputVideo)

	if config.ModelPath == "" {
		config.ModelPath = "dnnModels"
	}
	if config.Confidence <= 0 {
		log.Error().Msgf("Confidence must be above zero")
		return
	}

	if info, err := os.Stat(config.ModelPath); err != nil && !info.IsDir() {
		log.Error().Msgf("Cannot find model path: %w", err)
		return
	}

	// open capture device
	inputVideoDevice, err := gocv.OpenVideoCapture(inputVideo)
	if err != nil {
		log.Error().Msgf("Error open video capture device \"%v\": %w", inputVideo, err)
		return
	}
	defer inputVideoDevice.Close()

	img := gocv.NewMat()
	img_show := gocv.NewMat()
	faceRecognizer := contrib.NewLBPHFaceRecognizer()
	face := gocv.NewMat()
	defer img.Close()
	defer img_show.Close()
	defer face.Close()
	var (
		faces      []gocv.Mat
		per        contrib.PredictResponse
		font       = gocv.FontHersheyComplexSmall
		modelFiles []string
	)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Stop recognition")
			return

		default:
			if ok := inputVideoDevice.Read(&img); !ok {
				log.Error().Msgf("Cannot read stream \"%s\": %w", inputVideo, err)
				return
			}
			if img.Empty() {
				continue
			}

			faces, _, err = fr.faceDetector.DetectAllFacesOnCapture(&img)
			if err != nil {
				log.Err(err)
			}

			img_show = img.Clone()
			gocv.Flip(img_show, &img_show, 1)

			textOnCapture := make([]string, cap(faces))

			switch {
			case cap(faces) > 0:
				for i, f := range faces {
					face = f.Clone()
					faceNumber := i
					gocv.CvtColor(face, &face, gocv.ColorRGBToGray)

					modelFiles, err = utils.ListDir(config.ModelPath, isModel)
					if err != nil {
						log.Error().Msgf("Error read model files: %w", err)
						return
					}

					for _, model := range modelFiles {
						faceRecognizer.LoadFile(model)
						per = faceRecognizer.PredictExtendedResponse(face)
						//percentConfidence := 100 - per.Confidence
						if per.Confidence >= config.Confidence {
							textOnCapture[faceNumber] = fmt.Sprintf("%s, conf: %.1f", model, per.Confidence)
							break
						} else {
							continue
						}
					}
				}

			case cap(faces) <= 0:
				textOnCapture = append(textOnCapture, "No faces on the capture")
			}

			for i, text := range textOnCapture {
				gocv.PutText(&img_show, text, image.Point{30, 30 + 60*(i+1)}, font, 1, color.RGBA{0, 255, 0, 0}, 1)
				fmt.Printf("%s\n", text)
			}

			gocv.WaitKey(200)
		}
	}

}

func isModel(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yml", ".yaml":
		return true
	}
	return false
}
