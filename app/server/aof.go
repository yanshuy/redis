package server

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
)

type AOFFile struct {
	Filename string
	Seq      int
	Type     string
}

func (s *Server) InitAOF() error {
	if s.Config.Appendonly != "yes" {
		return nil
	}

	filePath := filepath.Join(s.Config.Dir, s.Config.Appenddirname, s.Config.Appendfilename)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	manifestPath := filePath + ".manifest"
	aofBaseName := filepath.Base(filePath)

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifestFile, err := os.OpenFile(manifestPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer manifestFile.Close()

		first := aofBaseName + ".1.incr.aof"
		line := fmt.Sprintf("file %s seq 1 type i\n", first)
		if _, err := manifestFile.WriteString(line); err != nil {
			return err
		}

		file, err := os.OpenFile(filepath.Join(dir, first), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		Svr.Aof = file
		return err
	}

	manifestFile, err := os.Open(manifestPath)
	if err != nil {
		return err
	}
	defer manifestFile.Close()

	files, err := AOFFiles(manifestFile)
	if err != nil {
		return err
	}
	err = ReplayAOF(dir, files)
	if err != nil {
		return err
	}

	activeFile := fmt.Sprintf("%s.1.incr.aof", aofBaseName)
	if len(files) > 0 {
		activeFile = files[len(files)-1].Filename
	}
	file, err := os.OpenFile(filepath.Join(dir, activeFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	Svr.Aof = file
	return err
}

func AOFFiles(manifest *os.File) ([]AOFFile, error) {
	var files []AOFFile
	scanner := bufio.NewScanner(manifest)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		// format: "file <filename> seq <seq> type <type>"
		if len(fields) >= 6 && fields[0] == "file" {
			seq, err := strconv.Atoi(fields[3])
			if err != nil {
				continue
			}
			files = append(files, AOFFile{
				Filename: fields[1],
				Seq:      seq,
				Type:     fields[5],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(files, func(a, b AOFFile) int {
		return cmp.Compare(a.Seq, b.Seq)
	})

	return files, nil
}

func ReplayAOF(dir string, files []AOFFile) error {
	aofClient := &client.Client{
		Conn:   nil,
		Role:   client.CLIENT,
		Reader: client.NewReader(),
	}
	for _, file := range files {
		file, err := os.Open(filepath.Join(dir, file.Filename))
		if err != nil {
			return err
		}
		defer file.Close()

		aofClient.Reader.Reset()
		for {
			data, _, err := aofClient.Reader.Read_RESP(file)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			cmd, err := client.ValidateCommand(data)
			if err != nil {
				continue
			}

			aofClient.Command = cmd
			Svr.CmdHandler(aofClient)
		}
	}
	return nil
}
