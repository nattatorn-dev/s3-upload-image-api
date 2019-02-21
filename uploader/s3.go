package uploader

import (
	"fmt"
	"os"
	"path/filepath"

	"s3-upload-image-api/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"s3-upload-image-api/config"
)

//	move the file to s3
func MoveToS3(filePath, fileKey string) (err error) {
	creds := credentials.NewStaticCredentials(*config.AWS_ACCESS_KEY_ID, *config.AWS_SECRET_ACCESS_KEY, "")
	cfg := aws.NewConfig().WithRegion(*config.AWS_REGION).WithCredentials(creds)

	svc := s3.New(
		session.New(),
		cfg,
	)

	file, err := os.Open(filePath)
	if err != nil {
		return
	}

	stat, err := file.Stat()
	if err != nil {
		return
	}

	putReq := s3.PutObjectInput{
		ACL:           aws.String(s3.BucketCannedACLPublicRead),
		Body:          file,
		Bucket:        aws.String(*config.S3_BUCKET),
		Key:           aws.String(fileKey + ".jpeg"),
		ContentLength: aws.Int64(stat.Size()),
		ContentType:   aws.String("image/jpeg"),
	}

	res, err := svc.PutObject(&putReq)
	fmt.Println(res)
	if err != nil {
		return
	}

	//	TODO: manage errors
	//	log.Println("s3 put response: ", resp)

	return
}

//	download from s3 based on bucket and file key.
//	return an error or the tmpFile location
func FetchFromS3(key, bucket string) (tmpFilePath string, err error) {
	//	build our s3 url
	url := fmt.Sprintf("https://%v.s3.amazonaws.com/%v", bucket, key)

	//	our temp file path
	tmpFilePath = filepath.Join(*config.FS_TEMP, RandomHash())
	//	downlod from s3 to our temp file location
	if err = util.HTTPdownload(url, tmpFilePath); err != nil {
		return
	}

	return
}
