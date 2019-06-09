package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kurin/blazer/b2"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	// set via build args
	gitVersion = "pre-release"
	gitCommit  string
	buildDate  string
)

// Info is a container for version information.
type Info struct {
	GitVersion string `json:"gitVersion" yaml:"gitVersion"`
	GitCommit  string `json:"gitCommit" yaml:"gitCommit"`
	BuildDate  string `json:"buildDate" yaml:"buildDate"`
	GoVersion  string `json:"goVersion" yaml:"goVersion"`
	Compiler   string `json:"compiler" yaml:"compiler"`
	Platform   string `json:"platform" yaml:"platform"`
}

func main() {

	interval := flag.Duration("interval", time.Second, "defines interval between image fetches")
	version := flag.Bool("version", false, "prints version info")
	url := flag.String("url", "", "url to the image")
	id := flag.String("id", "", "b2 account id")
	key := flag.String("key", "", "b2 api key")
	bucketName := flag.String("bucket", "", "b2 bucket name")

	flag.Parse()

	if *version {
		err := printVersion()
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, k := range []*string{url, id, key, bucketName} {
		if *k == "" {
			log.Println("--url, --id, --key or --bucket cannot be empty")
			flag.Usage()
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	c, err := b2.NewClient(ctx, *id, *key)
	if err != nil {
		fmt.Println(err)
		return
	}

	bucket, err := c.Bucket(ctx, *bucketName)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		log.Fatalf("Could not fetch b2 buckets: %v", err)
	} else {
		log.Printf("Createing new bucket with name %s", *bucketName)
		bucket, err = c.NewBucket(ctx, *bucketName, nil)
		if err != nil {
			log.Fatalf("Failed to create %s bucket: %v", *bucketName, err)
		}
	}

	ticker := time.NewTicker(*interval)

	// Setting up graceful shutdown
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-ticker.C:
			err := saveImageToBucket(ctx, *url, bucket)
			if err != nil {
				log.Printf("Unable to save image %s to bucket %s: %v", *url, bucket.Name(), err)
			}

		case <-gracefulStop:
			log.Println("Shutting down...")
			cancel()
			os.Exit(0)
		}
	}

}

func saveImageToBucket(ctx context.Context, url string, bucket *b2.Bucket) error {
	response, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "Unable to get image at %s", url)
	}
	defer response.Body.Close()

	objectName := fmt.Sprintf("%s.jpg", time.Now().Local().Format("2006-01-02T15:04:05"))

	obj := bucket.Object(objectName)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, response.Body); err != nil {
		w.Close()
		return err
	}
	return w.Close()

}

// Get returns the version info.
func printVersion() error {
	info := &Info{
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(data))

	return nil
}

