package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	qval    = flag.Int("qval", 15, "Quality constant")
	name    = flag.String("name", "", "Output file name without extension")
	verbose = flag.Bool("v", false, "Verbose")
)

func main() {
	flag.Parse()

	// Locate ffmpeg and mencoder binaries, based on platform
	ffmpeg := (NewCommand("ffmpeg").Find())()
	mencoder := (NewCommand("mencoder").Find())()

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
			if *verbose {
				log.Printf("DEBUG : %s : isDir", dirent.Name())
			}
			continue
		}
		if !strings.HasSuffix(dirent.Name(), "_front.avi") {
			if *verbose {
				log.Printf("DEBUG : %s : does not have _front.avi suffix", dirent.Name())
			}
			continue
		}

		frontAvi := dirent.Name()

		// Determine _rear.avi and _dual.mp4 output names
		frontMp4 := strings.TrimSuffix(frontAvi, ".avi") + ".mp4"
		rearAvi := strings.TrimSuffix(frontAvi, "_front.avi") + "_rear.avi"
		rearMp4 := strings.TrimSuffix(frontAvi, "_front.avi") + "_rear.mp4"
		dualMp4 := strings.TrimSuffix(frontAvi, "_front.avi") + "_dual.mp4"

		// Process all avi files to mp4 files
		{
			log.Printf("Converting %s -> %s", frontAvi, frontMp4)
			args := []string{
				"-i", frontAvi,
				"-ac", "2",
				"-cq", fmt.Sprintf("%d", *qval),
				frontMp4,
			}
			if *verbose {
				log.Printf("DEBUG : exec = %s, args = %#v", ffmpeg, args)
			}
			exec.Command(ffmpeg, args...).Run()
		}
		{
			log.Printf("Converting %s -> %s", rearAvi, rearMp4)
			args := []string{
				"-i", rearAvi,
				"-ac", "2",
				"-cq", fmt.Sprintf("%d", *qval),
				rearMp4,
			}
			if *verbose {
				log.Printf("DEBUG : exec = %s, args = %#v", ffmpeg, args)
			}
			exec.Command(ffmpeg, args...).Run()
		}
		// Combine with ffmpeg to _dual.mp4
		{
			log.Printf("Combining %s + %s -> %s", frontMp4, rearMp4, dualMp4)
			args := []string{
				"-i", frontMp4, "-i", rearMp4,
				"-filter_complex", `"[0:v][1:v]hstack=inputs=2[v];[0:a][1:a]amerge[a]"`,
				"-map", `"[v]"`, "-map", `"[a]"`,
				"-ac", "2", "-cq", fmt.Sprintf("%d", *qval), dualMp4,
			}
			if *verbose {
				log.Printf("DEBUG : exec = %s, args = %#v", ffmpeg, args)
			}
			exec.Command(ffmpeg, args...).Run()
		}
		// Add to list of _dual.mp4 files
		frontFiles = append(frontFiles, frontMp4)
		rearFiles = append(rearFiles, rearMp4)
		dualFiles = append(dualFiles, dualMp4)
	}

	// Concatenate all files together into front/rear/dual files
	{
		log.Printf("Concatenating %d files into %s", len(frontFiles), fmt.Sprintf("%s_front.mp4", *name))
		args := []string{
			"-ovc", "copy", "-oac", "pcm",
			"-o", fmt.Sprintf("%s_front.mp4", *name),
		}
		for _, f := range frontFiles {
			args = append(args, f)
		}
		if *verbose {
			log.Printf("DEBUG : exec = %s, args = %#v", mencoder, args)
		}
		exec.Command(mencoder, args...).Run()
	}
	{
		log.Printf("Concatenating %d files into %s", len(rearFiles), fmt.Sprintf("%s_rear.mp4", *name))
		args := []string{
			"-ovc", "copy", "-oac", "pcm",
			"-o", fmt.Sprintf("%s_rear.mp4", *name),
		}
		for _, f := range rearFiles {
			args = append(args, f)
		}
		if *verbose {
			log.Printf("DEBUG : exec = %s, args = %#v", mencoder, args)
		}
		exec.Command(mencoder, args...).Run()
	}
	{
		log.Printf("Concatenating %d files into %s", len(dualFiles), fmt.Sprintf("%s.mp4", *name))
		args := []string{
			"-ovc", "copy", "-oac", "pcm",
			"-o", fmt.Sprintf("%s.mp4", *name),
		}
		for _, f := range dualFiles {
			args = append(args, f)
		}
		if *verbose {
			log.Printf("DEBUG : exec = %s, args = %#v", mencoder, args)
		}
		exec.Command(mencoder, args...).Run()
	}
}
