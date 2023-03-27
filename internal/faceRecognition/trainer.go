package faceRecognition

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/krassor/GoCV/internal/models"
	//fd "github.com/GoCV/internal/faceDetector/dnnFaceDetector"
	"image"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

var (
	datasetPath string = "dataset"
)

type FaceDetector interface {
	DetectAllFacesOnCapture(src *gocv.Mat) (faces []gocv.Mat, rect []image.Rectangle)
}

type DnnTrainer struct {
	FaceDetector FaceDetector
}

func NewDnnTrainer(faceDetector FaceDetector) *DnnTrainer {
	return &DnnTrainer{
		FaceDetector: faceDetector,
	}
}

func (dnnTrainer *DnnTrainer) TrainModel(tenant *models.Tenant) error {
	var facesToTrain []gocv.Mat
	var labelsToTrain []int
	//var labels = make(map[string]int)
	fr := contrib.NewLBPHFaceRecognizer()

	tenantLabel := fmt.Sprintf("%s %s", tenant.Surname, tenant.Name)
	tenantDatasetPath := path.Join(datasetPath, tenantLabel)

	imageFiles, err := listDir(tenantDatasetPath)

	if err != nil {
		return fmt.Errorf("Error read tenant dir: %w", err)
	}

	roiDir := path.Join("roi", fmt.Sprintf("%s %s", tenant.Surname, tenant.Name))

	if err := createOrCleanPath(roiDir, tenant); err != nil {
		return fmt.Errorf("Error roi path: %w", err)
	}

	for i, file := range imageFiles {
		fmt.Printf("Read %d/%d imageFile: %v\n", i+1, len(imageFiles), file)
		img := gocv.IMRead(file, gocv.IMReadColor)
		faces, _ := dnnTrainer.FaceDetector.DetectAllFacesOnCapture(&img)

		if len(faces) > 1 {
			fmt.Printf("\tThere are %d faces on the foto \"%s\"\n", len(faces), file)
			continue
		}

		face := faces[0]
		gocv.CvtColor(face, &face, gocv.ColorBGRToGray)

		fileName := fmt.Sprintf("%d.jpg", i+1)
		roiFilePath := path.Join(roiDir, fileName)
		gocv.IMWrite(roiFilePath, face)
		facesToTrain = append(facesToTrain, face)
		//TODO: save all trained models into 1 file model.yml
		labelsToTrain = append(labelsToTrain, 1)
	}

	modelFilePath := path.Join("dnnModels", fmt.Sprintf("%s %s.yml", tenant.Surname, tenant.Name))

	fr.Train(facesToTrain, labelsToTrain)
	fr.SaveFile(modelFilePath)

	return nil
}

func createOrCleanPath(cleanPath string, tenant *models.Tenant) error {
	if err := os.Mkdir(cleanPath, 0755); err != nil {
		roiExistFiles, errListDir := listDir(cleanPath)
		if errListDir != nil {
			return fmt.Errorf("Error read roi dir \"%s %s\": %w", tenant.Surname, tenant.Name, err)
		}
		for _, roiExistFile := range roiExistFiles {
			os.Remove(roiExistFile)
		}
	}
	return nil
}

// listDir return slice of images path in srcPath, excluding within directories.
// Return error if directory does not exist
func listDir(srcPath string) ([]string, error) {
	var imageFiles []string

	files, err := os.ReadDir(srcPath) //ioutil.ReadDir(srcPath)
	if err != nil {
		return nil, fmt.Errorf("Error read padth: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			fileName := file.Name()
			if isImage(fileName) {
				imageFiles = append(imageFiles, path.Join(srcPath, fileName))
			}

		}
	}
	return imageFiles, nil
}

func isImage(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}
