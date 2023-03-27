package dnnFaceDetector

import (
	"fmt"
	"image"
	"image/color"

	"gocv.io/x/gocv"
)

type FaceDetectorConfig struct {
	Backend gocv.NetBackendType
	Target  gocv.NetTargetType
	ModelPath   string
	ConfigPath  string
}

type DnnObjectConfig struct {
	SwapRGB bool
	Ratio   float64
	Mean    gocv.Scalar
}

type DnnFaceDetector struct {
	faceDetectorConfig FaceDetectorConfig
	dnnObjectConfig    DnnObjectConfig
	net                gocv.Net
}

//
func NewDnnFaceDetector(faceDetectorConfig FaceDetectorConfig, dnnObjectConfig DnnObjectConfig) (*DnnFaceDetector, error) {

	var fd DnnFaceDetector

	//TODO: добавить проверку существования файла
	fd.faceDetectorConfig.ModelPath = faceDetectorConfig.ModelPath
	//TODO: добавить проверку существования файла
	fd.faceDetectorConfig.ConfigPath = faceDetectorConfig.ConfigPath

	fd.faceDetectorConfig.Backend = gocv.NetBackendDefault
	if fd.faceDetectorConfig.Backend > gocv.NetBackendDefault {
		fd.faceDetectorConfig.Backend = faceDetectorConfig.Backend
	}

	fd.faceDetectorConfig.Target = gocv.NetTargetCPU
	if fd.faceDetectorConfig.Target > gocv.NetTargetCPU {
		fd.faceDetectorConfig.Target = faceDetectorConfig.Target
	}

	// open DNN object tracking model
	fd.net = gocv.ReadNet(fd.faceDetectorConfig.ModelPath, fd.faceDetectorConfig.ConfigPath)
	if fd.net.Empty() {
		return nil, fmt.Errorf("Error reading network model from : %v %v\n", fd.faceDetectorConfig.ModelPath, fd.faceDetectorConfig.ConfigPath)
	}

	if err := fd.net.SetPreferableBackend(gocv.NetBackendType(fd.faceDetectorConfig.Backend)); err != nil {
		return nil, fmt.Errorf("Error SetPreferableBackend: %w", err)
	}

	if err := fd.net.SetPreferableTarget(gocv.NetTargetType(fd.faceDetectorConfig.Target)); err != nil {
		return nil, fmt.Errorf("Error SetPreferableTarget: %w", err)
	}

	fd.dnnObjectConfig = dnnObjectConfig

	// if filepath.Ext(fd.faceDetectorConfig.Model) == ".caffemodel" {
	// 	fd.dnnObjectConfig.Ratio = 1.5
	// 	fd.dnnObjectConfig.Mean = gocv.NewScalar(104, 177, 123, 0)
	// 	fd.dnnObjectConfig.SwapRGB = false
	// } else {
	// 	fd.dnnObjectConfig.Ratio = 1.0 / 127.5
	// 	fd.dnnObjectConfig.Mean = gocv.NewScalar(127.5, 127.5, 127.5, 0)
	// 	fd.dnnObjectConfig.SwapRGB = true
	// }

	return &fd, nil
}

func (fd *DnnFaceDetector) CloseNet() error {
	if err := fd.net.Close(); err != nil {
		return fmt.Errorf("Error close net: %w", err)
	}
	return nil
}

func (fd *DnnFaceDetector) DetectAllFacesOnCapture(src *gocv.Mat) (faces []gocv.Mat, rect []image.Rectangle) {
	// convert image Mat to 300x300 blob that the object detector can analyze
	blob := gocv.BlobFromImage(*src, fd.dnnObjectConfig.Ratio, image.Pt(300, 300), fd.dnnObjectConfig.Mean, fd.dnnObjectConfig.SwapRGB, false)
	// feed the blob into the detector
	fd.net.SetInput(blob, "")
	// run a forward pass thru the network
	prob := fd.net.Forward("")

	//img_fd := src.Clone()

	faces, rect = performDetection(src, prob)

	prob.Close()
	blob.Close()

	return faces, rect
}

// performDetection analyzes the results from the detector network,
// which produces an output blob with a shape 1x1xNx7
// where N is the number of detections, and each detection
// is a vector of float values
// [batchId, classId, confidence, left, top, right, bottom]
func performDetection(frame *gocv.Mat, results gocv.Mat) (faces []gocv.Mat, rect []image.Rectangle) {
	faces = nil
	rect = nil
	for i := 0; i < results.Total(); i += 7 {
		confidence := results.GetFloatAt(0, i+2)
		if confidence > 0.5 {
			left := int(results.GetFloatAt(0, i+3) * float32(frame.Cols()))
			top := int(results.GetFloatAt(0, i+4) * float32(frame.Rows()))
			right := int(results.GetFloatAt(0, i+5) * float32(frame.Cols()))
			bottom := int(results.GetFloatAt(0, i+6) * float32(frame.Rows()))

			rect = append(rect, image.Rect(left, top, right, bottom))
			faces = append(faces, frame.Region(image.Rect(left, top, right, bottom)))
			gocv.Rectangle(frame, image.Rect(left, top, right, bottom), color.RGBA{0, 255, 0, 0}, 2)
		}
	}
	return faces, rect
}
