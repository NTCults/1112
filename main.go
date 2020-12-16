package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var filename = "playlist.m3u8"
var period = 5

func main() {
	data, err := os.Open("./" + filename)
	if err != nil {
		panic(err)
	}
	processPlaylist(data, os.Stdout)
}

func processPlaylist(data io.Reader, w io.Writer) {
	var block []string
	var blockItemIndex = 0

	var mainBlockStarted bool
	var groupCounter = period
	var previousChunkNumber = 0
	var versionRewrited bool

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()

		if !versionRewrited && strings.Contains(line, "EXT-X-VERSION") {
			fmt.Fprintf(w, "#EXT-X-VERSION:4\n")
			versionRewrited = true
			continue
		}

		// process meta headers
		if !mainBlockStarted {
			if strings.Contains(line, "#EXTINF") {
				mainBlockStarted = true
			} else {
				fmt.Fprintf(w, "%s\n", line)
			}
		}

		if mainBlockStarted {
			block = append(block, line)
			if checkIfBlockEnded(line) {
				block, groupCounter, previousChunkNumber = processBlock(block, groupCounter, previousChunkNumber)
				writeBlock(os.Stdout, block)

				blockItemIndex = 0
				block = []string{}
			}
			blockItemIndex++
		}
	}
}

func writeBlock(w io.Writer, block []string) {
	for i := 0; i < len(block); i++ {
		fmt.Fprintf(w, "%s\n", block[i])
	}
}

func checkIfBlockEnded(line string) bool {
	if strings.Contains(line, ".ts") {
		return true
	}
	return false
}

func processBlock(block []string, groupCounter, previousChunkNumber int) ([]string, int, int) {
	lastElIndex := len(block) - 1
	chunkNumStr := block[lastElIndex]

	chunkNum := getChunkNumber(chunkNumStr)
	keyNumber := chunkNum / period

	if groupCounter == (period - 1) {
		key := generateKey(keyNumber)
		block = append([]string{key}, block...)
		groupCounter = 0
	} else if discontinuityCheck(block, previousChunkNumber, chunkNum) {
		key := generateKey(keyNumber)
		//add key after dicontinuity tag
		index := 1
		block = append(block[:index+1], block[index:]...)
		block[index] = key

		groupCounter = 0
	} else {
		groupCounter++
	}
	return block, groupCounter, chunkNum
}

func discontinuityCheck(block []string, previousChunkNumber, currentChunkNumber int) bool {
	if strings.Contains(block[0], "EXT-X-DISCONTINUITY") {
		if (currentChunkNumber - previousChunkNumber) != 1 {
			return true
		}
	}
	return false
}

var chunkNumberRegExp = regexp.MustCompile(`[0-9]+`)

func getChunkNumber(line string) int {
	result := chunkNumberRegExp.FindString(line)
	numChunk, err := strconv.Atoi(result)
	if err != nil {
		panic(err)
	}
	return numChunk
}

func generateKey(keyNum int) string {
	return fmt.Sprintf("ZALOOPA!!!!!!!!!!!!!!!!!!!!!!!!!%d", keyNum)
}
