package gateway

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tbtec/vdlg/internal/dto"
	"github.com/tbtec/vdlg/internal/enum"
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

func (gtw *ProcessorGateway) ProcessVideo(ctx context.Context, videoURL dto.Message) dto.OutputMessage {

	bucket := videoURL.BucketName
	key := videoURL.Key

	details, err := gtw.CheckS3Details(ctx, bucket, key)

	if err != nil {
		fmt.Println("Erro ao obter detalhes do objeto S3:", err)
	}
	fmt.Println("Detalhes do objeto S3:", details)

	if details != "OK" {
		gtw.deleteS3Object(ctx, err, bucket, key) // Deleta o vídeo original da pasta input
		return gtw.buildOutputMessage(key, enum.StatusError.String(), details)
	}

	arq, err := gtw.getObjectFromS3(ctx, bucket, key)
	if err != nil {
		fmt.Println("Erro ao buscar arquivo no s3:", err)
	}

	result := processVideoFromURL(videoURL, arq)
	fmt.Printf("✅ Resultado: %+v\n", result)

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

		gtw.deleteS3Object(ctx, err, bucket, key) // Deleta o vídeo original da pasta input
		return gtw.buildOutputMessage(s3OutputKey, enum.StatusCompleted.String(), "")
	}
	gtw.deleteS3Object(ctx, err, bucket, key) // Deleta o vídeo original da pasta input
	return gtw.buildOutputMessage(key, enum.StatusError.String(), enum.StatusErrorProcessing.String())

}

func (*ProcessorGateway) buildOutputMessage(fileName string, status string, reason string) dto.OutputMessage {
	return dto.OutputMessage{
		Filename: fileName,
		Status:   status,
		Reason:   reason,
	}
}

func (gtw *ProcessorGateway) deleteS3Object(ctx context.Context, err error, bucket string, key string) {
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

func processVideoFromURL(message dto.Message, arq []byte) dto.ProcessingResult {

	nome := extractNameFromInputPath(message.Key)
	fmt.Printf("Nome do video: %s\n", nome)

	tempDir := "temp"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	framePattern := filepath.Join(tempDir, "frame_%04d.png")

	// Salva o []byte em um arquivo temporário
	videoPath := filepath.Join(tempDir, "input.mp4")
	err := os.WriteFile(videoPath, arq, 0644)
	if err != nil {
		return dto.ProcessingResult{
			Success: false,
			Message: fmt.Sprintf("Erro ao salvar arquivo temporário: %s", err.Error()),
		}
	}

	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return dto.ProcessingResult{
			Success: false,
			Message: fmt.Sprintf("Erro no ffmpeg: %s\nOutput: %s", err.Error(), string(output)),
		}
	}

	frames, err := filepath.Glob(filepath.Join(tempDir, "*.png"))
	if err != nil || len(frames) == 0 {
		return dto.ProcessingResult{
			Success: false,
			Message: "Nenhum frame foi extraído do vídeo",
		}
	}

	zipFilename := fmt.Sprintf("%s.zip", nome)
	zipPath := filepath.Join("outputs", zipFilename)

	tempZipPath := filepath.Join("outputs")
	os.MkdirAll(tempZipPath, 0755)

	err = createZipFile(frames, zipPath)
	if err != nil {
		return dto.ProcessingResult{
			Success: false,
			Message: "Erro ao criar ZIP: " + err.Error(),
		}
	}

	imageNames := make([]string, len(frames))
	for i, frame := range frames {
		imageNames[i] = filepath.Base(frame)
	}

	return dto.ProcessingResult{
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
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("erro ao fazer upload: %w", err)
	}

	return nil
}

func (gtw *ProcessorGateway) getObjectFromS3(ctx context.Context, bucket, key string) ([]byte, error) {
	resp, err := gtw.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar objeto do S3: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler corpo do objeto: %w", err)
	}

	return data, nil
}

func (gtw *ProcessorGateway) CheckS3Details(ctx context.Context, bucket, key string) (string, error) {
	result, err := gtw.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return enum.StatusErrorFileCheck.String(), fmt.Errorf("erro ao obter detalhes do objeto S3: %w", err)
	}

	const maxSize int64 = 50 * 1024 * 1024 // 50MB em bytes
	//const maxSize int64 = 1024 // 1MB em bytes

	if result.ContentType != nil && *result.ContentType != "video/mp4" {
		return enum.StatusErrorContentType.String(), nil
	}

	if result.ContentLength != nil && *result.ContentLength > maxSize {
		return enum.StatusErrorFileSize.String(), nil
	}
	return "OK", nil
}

func extractNameFromInputPath(path string) string {
	const prefix = "input/"
	const suffix = ".mp4"

	start := strings.Index(path, prefix)
	end := strings.LastIndex(path, suffix)
	if start == -1 || end == -1 || end <= start+len(prefix) {
		return ""
	}
	return path[start+len(prefix) : end]
}
