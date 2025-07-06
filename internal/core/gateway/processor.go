package gateway

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tbtec/vdlg/internal/dto"
)

type ProcessorGateway struct {
	S3Client *s3.Client
}

type IProcessorGateway interface {
	ProcessaVideo(ctx context.Context, videoURL string) (*string, error)
	uploadToS3(ctx context.Context, bucketName, key, filePath string) (*string, error)
	notifyProcessingResultSNS(ctx context.Context, topicARN, key, messageBody string) (*string, error)
}

func NewProcessorGateway(config aws.Config) *ProcessorGateway {
	return &ProcessorGateway{
		S3Client: s3.NewFromConfig(config),
	}
}

type ProcessingResult struct {
	Success    bool
	Message    string
	ZipPath    string
	FrameCount int
	Images     []string
}

func (gtw *ProcessorGateway) ProcessVideo(ctx context.Context, videoURL dto.Message) ProcessingResult {

	timestamp := time.Now().Format("20060102_150405")

	result := processVideoFromURL(videoURL.Url, timestamp)
	fmt.Printf("✅ Resultado: %+v\n", result)

	bucket := videoURL.BucketName
	key := videoURL.Key

	if result.Success {
		s3OutputKey := "output/" + filepath.Base(result.ZipPath)
		err := gtw.uploadToS3(ctx, bucket, s3OutputKey, result.ZipPath)

		if err != nil {
			fmt.Println("Erro ao fazer upload:", err)
		} else {
			fmt.Printf("📤 ZIP enviado para S3 em: s3://%s/%s\n", bucket, s3OutputKey)

			// 🧹 Remove o ZIP local após o upload com sucesso
			if err := os.Remove(result.ZipPath); err != nil {
				fmt.Printf("⚠️ Erro ao remover ZIP local: %v\n", err)
			} else {
				fmt.Printf("🧹 ZIP local removido: %s\n", result.ZipPath)
			}
		}

		// Deleta o vídeo original da pasta input
		_, err = gtw.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

		if err != nil {
			fmt.Printf("❌ Erro ao deletar %s: %v\n", key, err)
		} else {
			fmt.Printf("🗑️  Arquivo deletado: %s/%s\n", bucket, key)
		}
	}

	return result

}

func processVideoFromURL(videoURL, timestamp string) ProcessingResult {

	fmt.Printf("⏳ Processando vídeo: %s\n", videoURL)

	tempDir := filepath.Join("temp", timestamp)
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	framePattern := filepath.Join(tempDir, "frame_%04d.png")

	cmd := exec.Command("ffmpeg",
		"-i", videoURL,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return ProcessingResult{
			Success: false,
			Message: fmt.Sprintf("Erro no ffmpeg: %s\nOutput: %s", err.Error(), string(output)),
		}
	}

	frames, err := filepath.Glob(filepath.Join(tempDir, "*.png"))
	if err != nil || len(frames) == 0 {
		return ProcessingResult{
			Success: false,
			Message: "Nenhum frame foi extraído do vídeo",
		}
	}

	zipFilename := fmt.Sprintf("frames_%s.zip", timestamp)
	zipPath := filepath.Join("outputs", zipFilename)

	err = createZipFile(frames, zipPath)
	if err != nil {
		return ProcessingResult{
			Success: false,
			Message: "Erro ao criar ZIP: " + err.Error(),
		}
	}

	imageNames := make([]string, len(frames))
	for i, frame := range frames {
		imageNames[i] = filepath.Base(frame)
	}

	return ProcessingResult{
		Success:    true,
		Message:    fmt.Sprintf("%d frames extraídos.", len(frames)),
		ZipPath:    zipPath,
		FrameCount: len(frames),
		Images:     imageNames,
	}

}

func createZipFile(files []string, output string) error {
	zipFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		w, err := writer.Create(filepath.Base(file))
		if err != nil {
			return err
		}

		_, err = io.Copy(w, f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (gtw *ProcessorGateway) uploadToS3(ctx context.Context, bucketName, key, filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer file.Close()

	_, err = gtw.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key), // Ex: "output/frames_20250703_223000.zip"
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("erro ao fazer upload: %w", err)
	}

	return nil
}
