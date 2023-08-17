package bp

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"path/filepath"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"os"
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
		log.Printf("Error creating session: %v", err)
		return err
	}
	log.Print("Session created successfully")

	dir, err := os.Open(uploadFileDir)
	if err != nil {
		log.Printf("Error opening directory: %v", err)
		return err
	}
	defer dir.Close()
	log.Print("Directory opened successfully")

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Printf("Error reading directory: %v", err)
		return err
	}
	log.Printf("Files found in directory: %d", len(files))

	for _, file := range files {
		if file.IsDir() {
			// Skip directories
			continue
		}

		filePath := filepath.Join(uploadFileDir, file.Name())
		f, err := os.Open(filePath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			return err
		}
		defer f.Close()
		log.Printf("File opened: %s", filePath)

		_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
			Bucket: aws.String(os.Getenv("BUCKET_NAME")),
			Key:    aws.String(file.Name()),
			Body:   f,
		})
		if err != nil {
			log.Printf("Error uploading file: %v", err)
			return err
		}
		log.Printf("File uploaded: %s", filePath)
	}
	log.Print("Upload completed successfully")
	return nil
}

//func uploadFile(uploadFileDir string) error {
//	sess, err := createSession()
//	if err != nil {
//		log.Printf("Error creating session: %v", err)
//		return err
//	}
//	log.Print("Session created successfully")
//
//	upFile, err := os.Open(uploadFileDir)
//	if err != nil {
//		if os.IsNotExist(err) {
//			log.Printf("Directory does not exist: %s", uploadFileDir)
//		} else {
//			log.Printf("Error opening directory: %v", err)
//		}
//
//	}
//	defer upFile.Close()
//	log.Print("Directory opened successfully")
//
//	files, err := upFile.Readdirnames(-1)
//	if err != nil {
//		log.Printf("Error reading filenames: %v", err)
//		return err
//	}
//	log.Printf("Filenames read successfully: %v", files)
//
//	errChan := make(chan error)
//	doneChan := make(chan bool)
//
//	for _, file := range files {
//		go func(file string) {
//			f, err := os.Open(filepath.Join(uploadFileDir, file))
//			if err != nil {
//				errChan <- err
//				return
//			}
//			defer f.Close()
//			doneChan <- true
//		}(file)
//	}
//
//	for range files {
//		<-doneChan
//	}
//	log.Print("All files processed successfully")
//
//	bucket := os.Getenv("BUCKET_NAME")
//	_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
//		Bucket: aws.String(bucket),
//		Key:    aws.String(uploadFileDir),
//	})
//	if err != nil {
//		log.Printf("Error uploading file: %v", err)
//		return err
//	}
//	log.Print("File uploaded successfully")
//
//	return nil
//}
