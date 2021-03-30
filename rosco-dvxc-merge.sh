#!/bin/bash
# rosco-dvxc-merge.sh
# @jbuchbinder
#
# Merges together AVI output from Rosco DVXC Dual Vision cameras using ffmpeg and mencoder.
# Run from a directory containing the _front.avi and _rear.avi files from a series of exports.

NAME="$1"

QVAL=15

SAVEIFS=$IFS
IFS=$(echo -en "\n\b")

echo " * Converting files to dual-view format"
for i in *_front.avi; do
        j="${i//_front.avi}_rear.avi"
        k="${i//_front.avi}_dual.mp4"
        echo " - Processing $i / $j ... "
	ffmpeg -i "$i" -ac 2 -cq $QVAL "${i//.avi}.mp4"
	ffmpeg -i "$j" -ac 2 -cq $QVAL "${j//.avi}.mp4"
        ffmpeg -i "$i" -i "$j" -filter_complex \
                "[0:v][1:v]hstack=inputs=2[v];[0:a][1:a]amerge[a]" \
                -map "[v]" -map "[a]" -ac 2 -cq $QVAL "$k"
done

echo " * Concatenating files together"
mencoder -ovc copy -oac pcm -o "${NAME}.mp4" *_dual.mp4
mencoder -ovc copy -oac pcm -o "${NAME}_front.mp4" *_front.mp4
mencoder -ovc copy -oac pcm -o "${NAME}_rear.mp4" *_dual.mp4
echo " * Wrote '${NAME}.mp4'"

IFS=$SAVEIFS

