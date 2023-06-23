package bp

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"net/http"
	"os"
	"path/filepath"

	//"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

func myCustomResolver(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	if service == endpoints.S3ServiceID {
		return endpoints.ResolvedEndpoint{
			URL:           "https://eu2.contabostorage.com",
			SigningRegion: region,
		}, nil
	}

	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
}
func createSession() (*session.Session, error) {
	region := os.Getenv("AWS_S3_REGION")
	accessKey := os.Getenv("STORE_ACCESS_KEY")
	secretKey := os.Getenv("STORE_SECRET_KEY")

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: "eu2",
		Config: aws.Config{
			Region:           aws.String(region),
			S3ForcePathStyle: aws.Bool(true),
			EndpointResolver: endpoints.ResolverFunc(myCustomResolver),
			Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		},
	})
	if err != nil {
		log.Printf("Error creating session: %v", err)
	}

	return sess, err
}
func uploadFile(uploadFileDir string) error {
	sess, err := createSession()
	if err != nil {
		log.Print(err)

	}

	errChan := make(chan error)
	doneChan := make(chan bool)

	err = filepath.WalkDir(uploadFileDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			errChan <- err
			return err
		}

		if d.IsDir() {
			return nil // Skip directories
		}

		go func(file os.DirEntry) {
			f, err := os.Open(path)
			if err != nil {
				errChan <- err
				return
			}
			defer f.Close()

			doneChan <- true
		}(d)

		return nil
	})

	if err != nil {
		log.Print(err)
		return err
	}

	for range doneChan {
		// Wait for goroutines to finish
	}

	close(errChan)

	for err := range errChan {
		if err != nil {
			log.Print(err)
			return err
		}
	}

	bucket := os.Getenv("BUCKET_NAME")

	err = filepath.WalkDir(uploadFileDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Print(err)
			return err
		}

		if d.IsDir() {
			return nil // Skip directories
		}

		f, err := os.Open(path)
		if err != nil {
			log.Print(err)
			return err
		}
		defer f.Close()

		fileInfo, err := f.Stat()
		if err != nil {
			log.Print(err)
			return err
		}

		fileSize := fileInfo.Size()
		fileBuffer := make([]byte, fileSize)
		if _, err := f.Read(fileBuffer); err != nil {
			log.Print(err)
			return err
		}

		_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
			Bucket:             aws.String(bucket),
			Key:                aws.String(path),
			ACL:                aws.String("public-read-write"),
			Body:               bytes.NewReader(fileBuffer),
			ContentLength:      aws.Int64(fileSize),
			ContentType:        aws.String(http.DetectContentType(fileBuffer)),
			ContentDisposition: aws.String("attachment"),
		})
		if err != nil {
			log.Print(err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
