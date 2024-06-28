package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type videoInfo struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int
		Height    int
	} `json:"streams"`
}

func checkCorruptVideo(inputVideo string, interval, timeout, concurrency int) error {

	md5hash := md5.New()
	md5hash.Write([]byte(inputVideo))
	hash := md5hash.Sum(nil)
	hex := hex.EncodeToString(hash)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_format", "-show_streams", "-of", "json", inputVideo)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not get video info: %v", err)
	}
	vInfo := &videoInfo{}
	err = json.Unmarshal([]byte(out), vInfo)
	if err != nil {
		return err
	}

	var duration float64
	for _, s := range vInfo.Streams {
		if s.CodecType == "video" {
			duration, _ = strconv.ParseFloat(vInfo.Format.Duration, 64)
		}
	}

	if duration == 0 {
		return fmt.Errorf("could not get video duration")

	}

	g, _ := errgroup.WithContext(ctx)

	g.SetLimit(concurrency)

	for i := interval; i < int(duration); i += interval {
		i := i
		g.Go(func() error {
			defer os.Remove(fmt.Sprintf("%s-%d.jpg", hex, i))
			cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-y", "-ss", strconv.Itoa(i), "-i", inputVideo, "-v", "error", "-xerror", "-vframes", "1", fmt.Sprintf("%s-%d.jpg", hex, i))
			out, err := cmd.CombinedOutput()
			if err != nil && strings.Contains(string(out), "Invalid NAL unit size") {
				cancel()
				return fmt.Errorf("corrupted video")
			} else if err != nil {
				return err
			}

			return nil
		})
	}

	return g.Wait()

}

func main() {
	var (
		filePath    string
		interval    int
		timeout     int
		concurrency int
	)

	flag.StringVar(&filePath, "path", "", "File Path HTTP or Local")
	flag.IntVar(&interval, "interval", 60, "Frame Interval in Seconds")
	flag.IntVar(&timeout, "timeout", 300, "Timeout in Seconds")
	flag.IntVar(&concurrency, "concurrency", 6, "Timeout in Seconds")

	required := []string{"path"}

	flag.Parse()

	seen := make(map[string]bool)

	flag.VisitAll(func(f *flag.Flag) {
		if f.Value.String() != "" {
			seen[f.Name] = true
		}
	})
	for _, req := range required {
		if !seen[req] {
			fmt.Fprintf(os.Stderr, "missing required -%s argument\n", req)
			os.Exit(2)
		}
	}
	err := checkCorruptVideo(filePath, interval, timeout, concurrency)
	if err != nil {
		fmt.Println(err.Error())
	}
}
