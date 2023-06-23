package bp

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"

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
	log.Print("success creating session")
	if err != nil {
		log.Print(err)
	}

	upFile, err := os.Open(uploadFileDir)
	log.Print("success opening directory")
	if err != nil {
		log.Print(err)
	}

	files, err := upFile.Readdirnames(-1)
	log.Printf("success reading filenames:%v", files)
	if err != nil {
		log.Print(err)
	}
	defer upFile.Close()
	errChan := make(chan error)
	doneChan := make(chan bool)

	for _, file := range files {
		go func(file string) {
			f, err := os.Open(uploadFileDir + file)
			if err != nil {
				errChan <- err
			}
			f.Close()
			doneChan <- true
		}(file)
	}

	for range files {
		<-doneChan
	}

	upFileInfo, err := upFile.Stat()
	if err != nil {
		log.Print(err)
	}
	fileSize := upFileInfo.Size()
	fileBuffer := make([]byte, fileSize)
	_, err = upFile.Read(fileBuffer)
	if err != nil {
		log.Print(err)
	}

	bucket := os.Getenv("BUCKET_NAME")
	_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(bucket),
		Key:                aws.String(uploadFileDir),
		ACL:                aws.String("public-read-write"),
		Body:               bytes.NewReader(fileBuffer),
		ContentLength:      aws.Int64(fileSize),
		ContentType:        aws.String(http.DetectContentType(fileBuffer)),
		ContentDisposition: aws.String("attachment"),
	})

	return err
}
