package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurin/blazer/b2"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

	// Setup configuration
	viper.SetEnvPrefix("JOAQUIN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	pflag.Duration("interval", 10*time.Second, "interval between image fetches")
	pflag.Bool("version", false, "prints version")
	pflag.String("url", "", "url to the image")
	pflag.String("id", "", "b2 account id")
	pflag.String("key", "", "b2 api key")
	pflag.String("bucket", "", "b2 bucket name")

	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Fatal(err)
	}

	// print version if requested
	if viper.GetBool("version") {
		err := printVersion()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	// check existence for all necessary
	for _, k := range []string{"url", "id", "key", "bucket"} {
		if viper.GetString(k) == "" {
			fmt.Printf("%s cannot be empty, set via --%s or env variable JOAQUIN_%s\n", k, k, strings.ToUpper(k))
			pflag.Usage()
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	log.Println("Connecting to b2 cloud storage")
	c, err := b2.NewClient(ctx, viper.GetString("id"), viper.GetString("key"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to authorize/connect to b2"))
	}

	bucketName := viper.GetString("bucket")
	bucket, err := c.Bucket(ctx, bucketName)
	if err != nil {
		log.Fatalf("Could open b2 bucket %s: %v", bucketName, err)
	}

	// create ticker for the regular fetching
	ticker := time.NewTicker(viper.GetDuration("interval"))

	// Setting up graceful shutdown
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)

	url := viper.GetString("url")

	log.Println("Starting image copy process...")
	for {
		select {
		case <-ticker.C:
			err := saveImageToBucket(ctx, url, bucket)
			if err != nil {
				log.Printf("Unable to save image %s to bucket %s: %v", url, bucket.Name(), err)
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

	// make sure correct content type is set
	attrs := &b2.Attrs{ContentType: "image/jpeg"}
	w := obj.NewWriter(ctx, b2.WithAttrsOption(attrs))

	log.Printf("Uploaded image %s, URL: %s", objectName, obj.URL())
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
