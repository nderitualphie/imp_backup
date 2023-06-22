package bp

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
)

// AWS_S3_REGION = "European Union"
// session
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

func UploadObject(localPath string, remoteBucket string, remotePath string) (err error) {
	sess, _ := createSession()
	uploader := s3manager.NewUploader(sess)
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		log.Printf("Error uploading file from local:%v", err)
	}
	defer file.Close()
	// Upload the file to S3
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(remoteBucket),
		Key:    aws.String(remotePath),
		Body:   file,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		log.Printf("Error uploading to bucket: %v", err)
	}
	// Return the uploaded URI

	id := result.UploadID
	log.Printf("upload id:%v", id)
	return err
}
func myCustomResolver(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	region = os.Getenv("AWS_S3_REGION")
	if service == endpoints.S3ServiceID {
		return endpoints.ResolvedEndpoint{
			URL:           os.Getenv("STORE_ENDPOINT"),
			SigningRegion: region,
		}, nil
	}
	return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
}
