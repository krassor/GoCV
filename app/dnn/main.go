package main

import (
	"context"

	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	fd "github.com/krassor/GoCV/internal/faceDetector/dnnFaceDetector"
	fr "github.com/krassor/GoCV/internal/faceRecognition"
	"github.com/krassor/GoCV/internal/graceful"
	"github.com/krassor/GoCV/internal/logger"
	"github.com/krassor/GoCV/internal/services"
	"github.com/krassor/GoCV/internal/transport/httpServer"
	"github.com/krassor/GoCV/internal/transport/httpServer/handlers"
	"github.com/krassor/GoCV/internal/transport/httpServer/routers"
	"github.com/rs/zerolog/log"
	"gocv.io/x/gocv"

	"github.com/hybridgroup/mjpeg"
)

func main() {
	// deviceID := 0

	// // open capture device
	// webcam, err := gocv.OpenVideoCapture(deviceID)
	// if err != nil {
	// 	fmt.Printf("Error opening video capture device: %v\n", deviceID)
	// 	return
	// }
	// defer webcam.Close()

	// window := gocv.NewWindow("DNN Detection")
	// defer window.Close()

	// img := gocv.NewMat()
	// defer img.Close()

	// img_fd := gocv.NewMat()
	// defer img_fd.Close()
	logger.InitLogger()

	var dnnObjectConfig fd.DnnObjectConfig = fd.DnnObjectConfig{
		SwapRGB: false,
		Ratio:   1.5,
		Mean:    gocv.NewScalar(104, 177, 123, 0),
	}

	var faceDetectorConfig fd.FaceDetectorConfig = fd.FaceDetectorConfig{
		Backend:    gocv.NetBackendDefault,
		Target:     gocv.NetTargetCPU,
		ModelPath:  "data/res10_300x300_ssd_iter_140000.caffemodel",
		ConfigPath: "data/deploy.prototxt",
	}

	var faceRecognitionConfig fr.FaceRecognitionConfig = fr.FaceRecognitionConfig{
		ModelPath:  "dnnModels",
		Confidence: 70,
	}

	faceDetector, err := fd.NewDnnFaceDetector(faceDetectorConfig, dnnObjectConfig)
	if err != nil {
		log.Error().Msgf("Error init faceDetector: %s", err)
		return
	}
	// defer func() {
	// 	if err := faceDetector.CloseNet(); err != nil {
	// 		log.Error().Msgf("Error close net: %s", err)
	// 	}
	// }()

	frFaceDetector, err := fd.NewDnnFaceDetector(faceDetectorConfig, dnnObjectConfig)
	if err != nil {
		log.Error().Msgf("Error init frFaceDetector: %s", err)
		return
	}
	// defer func() {
	// 	if err := frFaceDetector.CloseNet(); err != nil {
	// 		log.Error().Msgf("Error close net: %s", err)
	// 	}
	// }()

	trainer := services.NewDnnTrainer(faceDetector)
	handler := handlers.NewFrHandler(trainer)
	router := routers.NewDnnTrainerRouter(handler)
	httpServer := httpServer.NewHttpServer(router)

	faceRecognition := fr.NewFaceRecognition(frFaceDetector)

	//----------
	/*
		webcam, err := gocv.OpenVideoCapture(0)
		if err != nil {
			fmt.Printf("Error opening capture device: %v\n", 0)
			return
		}
		defer webcam.Close()

		// create the mjpeg stream
		stream := mjpeg.NewStream()

		// start capturing
		go mjpegCapture(webcam, stream)

		fmt.Println("Capturing. Point your browser to 127.0.0.1:8554")

		// start http server
		http.Handle("/", stream)
		log.Err((http.ListenAndServe("127.0.0.1:8554", nil)))
	*/
	//----------
	maxSecond := 15 * time.Second
	waitShutdown := graceful.GracefulShutdown(
		context.Background(),
		maxSecond,
		map[string]graceful.Operation{
			"http": func(ctx context.Context) error {
				return httpServer.Shutdown(ctx)
			},
			"faceDetector": func(ctx context.Context) error {
				return faceDetector.CloseNet(ctx)
			},
			"frfaceDetector": func(ctx context.Context) error {
				return frFaceDetector.CloseNet(ctx)
			},
		},
	)

	newCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go httpServer.Listen()
	go faceRecognition.Recognition(newCtx, faceRecognitionConfig, 0)
	go http.ListenAndServe("localhost:8081", nil)
	<-waitShutdown

}

func mjpegCapture(webcam *gocv.VideoCapture, stream *mjpeg.Stream) {
	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", 0)
			return
		}
		if img.Empty() {
			continue
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf.GetBytes())
		buf.Close()
	}
}
