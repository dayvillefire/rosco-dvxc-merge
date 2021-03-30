package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	qval = flag.Int("qval", 15, "Quality constant")
	name = flag.String("name", "", "Output file name without extension")
)

func main() {
	// Locate ffmpeg and mencoder binaries, based on platform
	ffmpeg := "ffmpeg"
	mencoder := "mencoder"

	frontFiles := make([]string, 0)
	rearFiles := make([]string, 0)
	dualFiles := make([]string, 0)

	// Loop through all *_front.avi files
	dirents, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, dirent := range dirents {
		if dirent.IsDir() {
			continue
		}
		if !strings.HasSuffix(dirent.Name(), "_front.avi") {
			continue
		}

		frontAvi := dirent.Name()

		// Determine _rear.avi and _dual.mp4 output names
		frontMp4 := strings.TrimSuffix(frontAvi, ".avi") + ".mp4"
		rearAvi := strings.TrimSuffix(frontAvi, "_front.avi") + "_rear.mp4"
		rearMp4 := strings.TrimSuffix(frontAvi, "_front.avi") + "_rear.mp4"
		dualMp4 := strings.TrimSuffix(frontAvi, "_front.avi") + "_dual.mp4"

		// Process all avi files to mp4 files
		{
			args := []string{
				"-i", frontAvi,
				"-ac", "2",
				"-cq", fmt.Sprintf("%d", *qval),
				frontMp4,
			}
			exec.Command(ffmpeg, args...)
		}
		{
			args := []string{
				"-i", rearAvi,
				"-ac", "2",
				"-cq", fmt.Sprintf("%d", *qval),
				rearMp4,
			}
			exec.Command(ffmpeg, args...)
		}
		// Combine with ffmpeg to _dual.mp4
		{
			args := []string{
				"-i", frontMp4, "-i", rearMp4,
				"-filter_complex", `"[0:v][1:v]hstack=inputs=2[v];[0:a][1:a]amerge[a]"`,
				"-map", `"[v]"`, "-map", `"[a]"`,
				"-ac", "2", "-cq", fmt.Sprintf("%d", *qval), dualMp4,
			}
			exec.Command(ffmpeg, args...)
		}
		// Add to list of _dual.mp4 files
		frontFiles = append(frontFiles, frontMp4)
		rearFiles = append(rearFiles, rearMp4)
		dualFiles = append(dualFiles, dualMp4)
	}

	// Concatenate all files together into front/rear/dual files
	{
		args := []string{
			"-ovc", "copy", "-oac", "pcm",
			"-o", fmt.Sprintf("%s_front.mp4", *name),
		}
		for _, f := range frontFiles {
			args = append(args, f)
		}
		exec.Command(mencoder, args...)
	}
	{
		args := []string{
			"-ovc", "copy", "-oac", "pcm",
			"-o", fmt.Sprintf("%s_rear.mp4", *name),
		}
		for _, f := range rearFiles {
			args = append(args, f)
		}
		exec.Command(mencoder, args...)
	}
	{
		args := []string{
			"-ovc", "copy", "-oac", "pcm",
			"-o", fmt.Sprintf("%s.mp4", *name),
		}
		for _, f := range dualFiles {
			args = append(args, f)
		}
		exec.Command(mencoder, args...)
	}
}
