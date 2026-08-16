package server

import (
	"flag"
	"log"
	"path/filepath"
	"strings"
)

var (
	dirFlag            = flag.String("dir", ".", "Directory for RDB persistence")
	dbFileFlag         = flag.String("dbfilename", "rdb.snapshot", "RDB file name")
	appendonlyFlag     = flag.String("appendonly", "no", "Enable append-only mode")
	appenddirnameFlag  = flag.String("appenddirname", "appendonlydir", "Directory for AOF files")
	appendfilenameFlag = flag.String("appendfilename", "appendonly.aof", "Name of the AOF file")
	appendfsyncFlag    = flag.String("appendfsync", "everysec", "Fsync policy for AOF")
)

type Config struct {
	port           string
	Dir            string
	Dbfilename     string
	Appendonly     string
	Appenddirname  string
	Appendfilename string
	Appendfsync    string
}

func NewConfig() Config {
	dir, err := filepath.Abs(*dirFlag)
	if err != nil {
		log.Fatal(err.Error())
	}
	dbfilename := *dbFileFlag
	return Config{
		port:           *portFlag,
		Dir:            dir,
		Dbfilename:     dbfilename,
		Appendonly:     *appendonlyFlag,
		Appenddirname:  *appenddirnameFlag,
		Appendfilename: *appendfilenameFlag,
		Appendfsync:    *appendfsyncFlag,
	}
}

func (config *Config) GetConfig(args []string) ([]string, error) {
	result := make([]string, 0)
	for _, arg := range args {
		var val string
		switch strings.ToLower(arg) {
		case "dir":
			val = config.Dir
		case "dbfilename":
			val = config.Dbfilename
		case "appendonly":
			val = config.Appendonly
		case "appenddirname":
			val = config.Appenddirname
		case "appendfilename":
			val = config.Appendfilename
		case "appendfsync":
			val = config.Appendfsync
		default:
			continue
		}
		result = append(result, arg, val)
	}
	return result, nil
}
