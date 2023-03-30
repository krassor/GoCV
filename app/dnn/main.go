package main

import (
	"context"

	"time"

	fd "github.com/krassor/GoCV/internal/faceDetector/dnnFaceDetector"
	"github.com/krassor/GoCV/internal/graceful"
	"github.com/krassor/GoCV/internal/logger"
	"github.com/krassor/GoCV/internal/services"
	"github.com/krassor/GoCV/internal/transport/httpServer"
	"github.com/krassor/GoCV/internal/transport/httpServer/handlers"
	"github.com/krassor/GoCV/internal/transport/httpServer/routers"
	"github.com/rs/zerolog/log"
	"gocv.io/x/gocv"
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

	faceDetector, err := fd.NewDnnFaceDetector(faceDetectorConfig, dnnObjectConfig)
	defer func() {
		if err := faceDetector.CloseNet(); err != nil {
			log.Error().Msgf("Error close net: %s", err)
		}
	}()
	if err != nil {
		log.Error().Msgf("Error init faceDetector: %s", err)
		return
	}

	trainer := services.NewDnnTrainer(faceDetector)
	handler := handlers.NewFrHandler(trainer)
	router := routers.NewDnnTrainerRouter(handler)
	httpServer := httpServer.NewHttpServer(router)

	maxSecond := 15 * time.Second
	waitShutdown := graceful.GracefulShutdown(
		context.Background(),
		maxSecond,
		map[string]graceful.Operation{
			"http": func(ctx context.Context) error {
				return httpServer.Shutdown(ctx)
			},
		},
	)

	go httpServer.Listen()
	<-waitShutdown

	// fmt.Printf("Start reading device: %v\n", deviceID)

	// if err := os.Mkdir("dataset", 0770); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := os.Mkdir("model", 0770); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }

	// count := 0
	// var (
	// 	img_slice []gocv.Mat

	// 	label  int = 1
	// 	labels []int

	// 	//modelName string = ""
	// 	faces []gocv.Mat

	// 	pr  int
	// 	per contrib.PredictResponse
	// )

	// for {
	// 	if ok := webcam.Read(&img); !ok {
	// 		fmt.Printf("Device closed: %v\n", deviceID)
	// 		return
	// 	}
	// 	if img.Empty() {
	// 		continue
	// 	}

	// 	faces, _ = faceDetector.DetectAllFacesOnCapture(&img)

	// 	img_show := img.Clone()
	// 	gocv.Flip(img_show, &img_show, 1)

	// 	fr := contrib.NewLBPHFaceRecognizer()

	// 	var face gocv.Mat

	// 	font := gocv.FontHersheyComplexSmall
	// 	text := ""

	// 	switch {
	// 	case cap(faces) > 0:
	// 		face = faces[0]
	// 		gocv.CvtColor(face, &face, gocv.ColorRGBToGray)

	// 		if _, err := os.Open("model/modelname_2.yml"); err != nil {
	// 			fmt.Printf("No model file\n")
	// 		} else {
	// 			fr.LoadFile("model/modelname_2.yml")
	// 			per = fr.PredictExtendedResponse(face)
	// 			pr = fr.Predict(face)

	// 			text = fmt.Sprintf("lblpr: %v, label: %v, conf: %.1f", pr, per.Label, 100-per.Confidence)

	// 		}
	// 	case cap(faces) <= 0:
	// 		text = "No faces on the capture"
	// 	}

	// 	gocv.PutText(&img_show, text, image.Point{30, 30}, font, 1, color.RGBA{0, 255, 0, 0}, 1)
	// 	window.IMShow(img_show)

	// 	key := window.WaitKey(20)
	// 	if key == 's' && cap(faces) > 0 {
	// 		gocv.IMWrite(fmt.Sprintf("dataset/screenshot_%d.jpg", count), face)
	// 		img_slice = append(img_slice, face)
	// 		labels = append(labels, label)
	// 		count++
	// 		continue
	// 	} else if key == 't' {
	// 		fr.Train(img_slice, labels)
	// 		fr.SaveFile(fmt.Sprintf("model/modelname_%d.yml", label))
	// 		img_slice = nil
	// 		labels = nil
	// 	} else if key == 'l' {
	// 		label++
	// 	} else if key >= 0 {
	// 		break
	// 	}

	// }
}
