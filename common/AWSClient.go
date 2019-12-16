package rgw

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	
	//"strings"
)

type rgwClient struct {
	svc *s3.S3
}

func rgwConfig(endPoint, secretId, secretKey, region string) (*aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}

	if endPoint == "" {
		err = errors.New("error: rgw server address is empty.")
		return nil, err
	}
	//endPoint = "10.61.51.184:81"
	cfg.Credentials = aws.NewStaticCredentialsProvider(secretId, secretKey, "")
	cfg.EndpointResolver = aws.ResolveWithEndpointURL(fmt.Sprintf("http://%s", endPoint))
	cfg.Region = *aws.String(region)

	return &cfg, nil
}

//是否只从etcd获取一次,定义全局变量
//var servers string

//newClient
func NewRgwClient(secretId, secretKey string) (*rgwClient, error) {
	var servers string
	var err error
	//从etcd获取rgwserver地址，但rgw并未提供此功能，若使用，需确认设置的etcd的key,value格式
	if servers == "" {
		err, servers = GetRgwEtcdConfig()
		if err != nil {
			return nil, err
		}
	}

	cfg, err := rgwConfig(servers, secretId, secretKey, "US")
	if err != nil {
		return nil, err
	}

	client := &rgwClient{
		svc: s3.New(*cfg),
	}
	client.svc.ForcePathStyle = *aws.Bool(true)

	return client, nil
}

/*
发布project时指定rgwserver地址，
*/
func NewRgwClient2(secretId, secretKey string) (*rgwClient, error) {

	//从json
	servers := GlobalprojectConfig.RgwConfig.Servers

	cfg, err := rgwConfig(servers, secretId, secretKey, "US")
	if err != nil {
		return nil, err
	}

	client := &rgwClient{
		svc: s3.New(*cfg),
	}
	client.svc.ForcePathStyle = *aws.Bool(true)

	return client, nil
}

//HeadObject
func (rgwClient *rgwClient) HeadObject(bucketName string, objectName string) (int64, int64, string, string, error) {

	req := rgwClient.svc.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})

	resp, err := req.Send()
	if err != nil {
		return 0, 0, "", "", err
	}

	Logger.Debugf("object headInfo:%+v", resp)

	size := *(resp.ContentLength)
	etag := *(resp.ETag)

	lastMoified := *(resp.LastModified)
	createTs := lastMoified.Unix()
	owner := ""

	return size, createTs, owner, etag, err
}
