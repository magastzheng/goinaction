package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
)

const (
	VERSION         = "1.0.0"
	MAX_BUFFER_SIZE = 1024 * 1024
)

func printVersion() {
	fmt.Println("file sender" + VERSION + " by gerryyang")
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

func handleCommandLine() (file string, service string) {
	var f = flag.String("f", "data2.txt", "file")
	var s = flag.String("s", "127.0.0.1:9001", "host:ip")
	var printVer bool
	flag.BoolVar(&printVer, "v", false, "print version")

	flag.Parse()
	if printVer {
		printVersion()
		os.Exit(0)
	}
	fmt.Println("file:", *f)
	fmt.Println("host:ip:", *s)
	return *f, *s
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8) // int64 is 8 byte
	binary.LittleEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.LittleEndian.Uint64(buf))
}

func writeFile(file string, offset int64, s_real_req string) {

	fout, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	checkError(err)
	defer fout.Close()

	var bytes_buf bytes.Buffer
	bytes_buf.WriteString(s_real_req)

	var write_len int = len(s_real_req)
	cnt, err := fout.WriteAt(bytes_buf.Bytes(), offset)
	if cnt != write_len {
		fmt.Fprintf(os.Stderr, "WriteAt: offset[%d] cnt[%d] !=  write_len[%d] err: %s\n",
			offset, cnt, write_len, err.Error())
		return
	}
	//fmt.Printf("WriteAt: offset[%d] bytes[%d] s_real_req[%#v]\n", offset, cnt, s_real_req)
	fmt.Printf("WriteAt: offset[%d] bytes[%d]\n", offset, cnt)
}

func handleClient(conn net.Conn, file string) {
	defer conn.Close()

	var buf [MAX_BUFFER_SIZE]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Read: %s\n", err.Error())
			return
		}
		//fmt.Printf("handleClient: len[%d] req[%#v]\n", n, buf[0:n])
		fmt.Printf("handleClient: len[%d]\n", n)

		var bytes_buf bytes.Buffer
		bytes_buf.Write(buf[0:n])

		offset := BytesToInt64(bytes_buf.Next(8))
		var s_real_req string = string(bytes_buf.Bytes())
		//fmt.Printf("file[%s] offset[%d] s_real_req[%#v]\n", file, offset, s_real_req)
		fmt.Printf("handleClient: file[%s] offset[%d]\n", file, offset)

		writeFile(file, offset, s_real_req)

		if n > 0 {
			_, err = conn.Write([]byte("ok"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Write: %s\n", err.Error())
				return
			}
		}
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	file, service := handleCommandLine()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Accept: err[%s]\n", err.Error())
			continue
		}
		go handleClient(conn, file)
	}
}
