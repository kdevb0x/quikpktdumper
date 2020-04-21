package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
)

var logfile string

func listen(output io.Writer, ip, port string) {

	var f *os.File
	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		panic(err)
	}
	if logfile != "" {
		f, err = os.OpenFile(logfile, 0666, os.ModeAppend)
		if err != nil {
			log.Println(err.Error())
		}
		defer f.Close()

	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error: " + err.Error())
		}
		defer conn.Close()
		_, err = output.Write([]byte("listening for incomming stream..."))
		if err != nil {
			log.Fatal(err)
		}

		r := io.TeeReader(conn, output)

		_, err = io.Copy(f, r)
		if err != nil {
			log.Fatalf("error writing to file %s: %w\n", f.Name(), err)
		}
		if wc, ok := output.(io.Closer); ok {
			defer wc.Close()
		}
	}

}

func main() {
	flag.StringVar(&logfile, "o", "", "filepath for output")
	flag.Parse()
}
