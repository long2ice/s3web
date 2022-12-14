package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
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
	spa       bool
}

func NewS3FileSystem(bucket, subFolder string, spa bool) *S3FileSystem {
	return &S3FileSystem{
		bucket:    bucket,
		subFolder: subFolder,
		spa:       spa,
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
	var names []string
	if s3.spa {
		names = []string{name, path.Join(s3.subFolder, "index.html"), "/404.html"}
	} else {
		names = []string{name, path.Join(name, "index.html"), "/404.html"}
	}
	for _, n := range names {
		obj, err := client.GetObject(ctx, s3.bucket, n, minio.GetObjectOptions{})
		if err != nil {
			log.Error(err)
			continue
		}

		_, err = obj.Stat()
		if err != nil {
			if minio.ToErrorResponse(err).Code != "NoSuchKey" {
				log.Error(err)
			}
			continue
		}

		return obj, nil
	}

	return nil, os.ErrNotExist
}

func NewS3Handler() fiber.Handler {
	fs := make(map[string]*S3FileSystem)
	for _, site := range config.SitesConfig {
		fs[site.Domain] = NewS3FileSystem(config.S3Config.Bucket, site.SubFolder, site.Spa)
	}
	return func(c *fiber.Ctx) (err error) {
		hostname := c.Hostname()
		domain := strings.Split(hostname, ":")[0]
		if fs[domain] == nil {
			return c.Next()
		}
		return filesystem.New(filesystem.Config{
			Root: fs[domain],
		})(c)
	}
}
