package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/long2ice/s3web/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

var client *minio.Client

func NewCustomHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1024,
		MaxIdleConnsPerHost:   1024,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		DisableCompression:    true,
	}
}

func init() {
	s3 := config.S3Config
	defaultAWSCredProviders := []credentials.Provider{
		&credentials.Static{
			Value: credentials.Value{
				AccessKeyID:     s3.AccessKey,
				SecretAccessKey: s3.SecretKey,
			},
		},
	}
	var err error
	creds := credentials.NewChainCredentials(defaultAWSCredProviders)
	client, err = minio.New(s3.Endpoint, &minio.Options{
		Creds:        creds,
		Secure:       s3.Scheme == "https",
		Region:       config.S3Config.Region,
		BucketLookup: minio.BucketLookupAuto,
		Transport:    NewCustomHTTPTransport(),
	})
	if err != nil {
		log.Fatalln(err)
	}
}

type S3FileSystem struct {
	bucket    string
	subFolder string
}

func NewS3FileSystem(bucket, subFolder string) *S3FileSystem {
	return &S3FileSystem{
		bucket:    bucket,
		subFolder: subFolder,
	}
}

func (s3 *S3FileSystem) pathIsDir(ctx context.Context, name string) bool {
	name = strings.Trim(name, pathSeparator) + pathSeparator
	listCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	objCh := client.ListObjects(listCtx,
		s3.bucket,
		minio.ListObjectsOptions{
			Prefix:  name,
			MaxKeys: 1,
		})
	for range objCh {
		cancel()
		return true
	}
	return false
}

func (s3 *S3FileSystem) Open(name string) (http.File, error) {
	name = path.Join(s3.subFolder, name)
	if name == pathSeparator || s3.pathIsDir(context.Background(), name) {
		return &httpMinioObject{
			client: client,
			object: nil,
			isDir:  true,
			bucket: s3.bucket,
			prefix: strings.TrimSuffix(name, pathSeparator),
		}, nil
	}

	name = strings.TrimPrefix(name, pathSeparator)
	obj, err := s3.getObject(context.Background(), name)
	if err != nil {
		return nil, os.ErrNotExist
	}

	return &httpMinioObject{
		client: client,
		object: obj,
		isDir:  false,
		bucket: s3.bucket,
		prefix: name,
	}, nil
}

func (s3 *S3FileSystem) getObject(ctx context.Context, name string) (*minio.Object, error) {
	names := [4]string{name, name + "/index.html", name + "/index.htm", "/404.html"}
	for _, n := range names {
		obj, err := client.GetObject(ctx, s3.bucket, n, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = obj.Stat()
		if err != nil {
			if minio.ToErrorResponse(err).Code != "NoSuchKey" {
				log.Println(err)
			}
			continue
		}

		return obj, nil
	}

	return nil, os.ErrNotExist
}

type S3Handler struct {
	fsMap map[string]*S3FileSystem
}

func NewS3Handler() *S3Handler {
	fs := make(map[string]*S3FileSystem)
	for _, site := range config.SitesConfig {
		fs[site.Domain] = NewS3FileSystem(config.S3Config.Bucket, site.SubFolder)
	}
	return &S3Handler{
		fsMap: fs,
	}
}

func (s *S3Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	hostWithPort := request.Host
	host := strings.Split(hostWithPort, ":")[0]
	if fs, ok := s.fsMap[host]; ok {
		http.FileServer(fs).ServeHTTP(writer, request)
	} else {
		http.NotFound(writer, request)
	}
}
