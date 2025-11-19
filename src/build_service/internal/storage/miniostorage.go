package storage

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	minioClient *minio.Client
	bucketName  string
}

func NewMinioStorage(useSSL bool) *MinioStorage {
	ctx := context.Background()

	endpoint := os.Getenv("MINIO_ENDPOINT")
	id := os.Getenv("MINIO_ID")
	secret := os.Getenv("MINIO_SECRET")
	token := os.Getenv("MINIO_TOKEN")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	minioClient, err := minio.New(
		endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(id, secret, token),
			Secure: useSSL,
		},
	)

	if err != nil {
		log.Fatalln(err)
	}

	exists, _ := minioClient.BucketExists(ctx, bucketName)
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln(err)
		}
	}

	return &MinioStorage{
		minioClient: minioClient,
		bucketName:  bucketName,
	}
}

func (m *MinioStorage) GetFile(path string) (io.Reader, error) {
	return m.minioClient.GetObject(
		context.Background(),
		m.bucketName,
		path, minio.GetObjectOptions{},
	)
}

func (m *MinioStorage) PutFile(dstPath string, filePath string) error {
	_, err := m.minioClient.FPutObject(
		context.Background(),
		m.bucketName,
		dstPath,
		filePath,
		minio.PutObjectOptions{},
	)
	return err
}

func (m *MinioStorage) HasFile(path string) bool {
	_, err := m.minioClient.StatObject(
		context.Background(),
		m.bucketName,
		path,
		minio.StatObjectOptions{},
	)

	return err == nil
}

func (m *MinioStorage) ListFiles(path string) []string {
	files := []string{}

	for object := range m.minioClient.ListObjects(context.Background(), m.bucketName, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	}) {
		if object.Err != nil {
			continue
		}
		files = append(files, object.Key)
	}

	return files
}
