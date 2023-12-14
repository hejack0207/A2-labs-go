package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	// "time"
)

const SIZE_LEN = 8

// Conn 是你需要实现的一种连接类型，它支持下面描述的若干接口；
// 为了实现这些接口，你需要设计一个基于 TCP 的简单协议；
type Conn struct {
	tcpconn net.Conn
	reader  *DataReader
}

// Send size 表示要传输的数据总长度；
// 你需要实现从 reader 读取数据，并将数据通过 TCP 进行传输；
func (conn *Conn) Send(size int, reader io.Reader) (err error) {
	packet := new(bytes.Buffer)
	buf := make([]byte, 2<<10)
	var sentbytes int64 = 0

	binary.Write(packet, binary.BigEndian, int64(size))
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			if sentbytes+int64(n) != int64(size)+SIZE_LEN {
				panic(err)
			} else {
				binary.Write(packet, binary.BigEndian, []byte{'F'})
				_, e := packet.WriteTo(conn.tcpconn)
				return e
			}
		} else if err != nil {
			return err
		}

		if n > 0 {
			if int64(n) > int64(size)+SIZE_LEN-sentbytes {
				binary.Write(packet, binary.BigEndian, buf[:int64(size)-sentbytes])
			} else {
				binary.Write(packet, binary.BigEndian, buf[:n])
			}
			count, e := packet.WriteTo(conn.tcpconn)

			if e != nil {
				return e
			}
			sentbytes += int64(count)
			// fmt.Println("SEND sentbytes:", count, ",size:", size)
			if sentbytes == int64(size)+SIZE_LEN {
				// fmt.Println("SEND sentbytes:", sentbytes, ",size:", size)
				// sentbytes -= (int64(size) + SIZE_LEN)
				binary.Write(packet, binary.BigEndian, []byte{'F'})
				count, e = packet.WriteTo(conn.tcpconn)
				return e
			}
		}
	}
}

type DataReader struct {
	tcpconn       net.Conn
	packet        *bytes.Buffer
	totalsize     int64
	totalconsumed int64
	tcpclosed     bool
}

func (reader *DataReader) Read(buff []byte) (int, error) {
	buf := make([]byte, len(buff))

	for {
		if reader.packet.Len() > 0 && reader.totalconsumed == reader.totalsize {
			var fflag byte
			binary.Read(reader.packet, binary.BigEndian, &fflag)
			if fflag == 'F' {
				// fmt.Println("RECV consumed:", 0, "totalconsumed:", reader.totalconsumed, ",totalsize:", reader.totalsize)
				reader.totalconsumed = 0
				reader.totalsize = -1
				return 0, io.EOF
			} else {
				return 0, errors.New("unexpected finish flag")
			}
		}

		n, err := reader.tcpconn.Read(buf)

		if err != nil {
			if errors.Is(err, io.EOF) {
				reader.tcpclosed = true
				return 0, nil
			} else {
				fmt.Println("err:", err)
				return n, err
			}
		}

		reader.packet.Write(buf[:n])
		if reader.totalsize == -1 {
			if reader.packet.Len() >= SIZE_LEN {
				binary.Read(reader.packet, binary.BigEndian, &reader.totalsize)
				// fmt.Println("totalsize ", reader.totalsize)
			} else {
				continue
			}
		}

		consumed, _ := 0, error(nil)
		if reader.totalsize-reader.totalconsumed <= int64(len(buff)) {
			limitreader := io.LimitReader(reader.packet, reader.totalsize-reader.totalconsumed)
			consumed, _ = limitreader.Read(buff)
		} else {
			consumed, _ = reader.packet.Read(buff)
		}

		reader.totalconsumed += int64(consumed)
		return consumed, nil
	}
}

// Receive 返回的 reader 用于接收数据；
// 你需要实现向 reader 中写入从 TCP 接收到的数据；
func (conn *Conn) Receive() (reader io.Reader, err error) {
	if conn.reader == nil {
		conn.reader = &DataReader{
			tcpconn:       conn.tcpconn,
			packet:        new(bytes.Buffer),
			totalsize:     -1,
			totalconsumed: 0,
			tcpclosed:     false,
		}
	} else {
		if conn.reader.tcpclosed {
			return nil, errors.New("remote connection closed")
		}
	}

	return conn.reader, nil
}

// Close 用于关闭你实现的连接对象及其相关资源
func (conn *Conn) Close() {
	conn.tcpconn.Close()
}

// NewConn 从一个 TCP 连接得到一个你实现的连接对象
func NewConn(conn net.Conn) *Conn {
	return &Conn{
		tcpconn: conn,
	}
}

// 除了上面规定的接口，你还可以自行定义新的类型，变量和函数以满足实现需求

//////////////////////////////////////////////
///////// 接下来的代码为测试代码，请勿修改 /////////
//////////////////////////////////////////////

// 连接到测试服务器，获得一个你实现的连接对象
func dial(serverAddr string) *Conn {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		panic(err)
	}
	return NewConn(conn)
}

// 启动测试服务器
func startServer(handle func(*Conn)) net.Listener {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("[WARNING] ln.Accept", err)
				return
			}
			go handle(NewConn(conn))
		}
	}()
	return ln
}

// 简单断言
func assertEqual(actual string, expected string) {
	if actual != expected {
		panic(fmt.Sprintf("actual:%v expected:%v\n", actual, expected))
	}
}

// 简单 case：单连接，传输少量数据
func testCase0() {
	const data = `Then I heard the voice of the Lord saying, “Whom shall I send? And who will go for us?”
And I said, “Here am I. Send me!”
Isaiah 6:8`

	ln := startServer(func(conn *Conn) {
		defer conn.Close()
		err := conn.Send(len(data), bytes.NewBufferString(data))
		if err != nil {
			panic(err)
		}
	})
	//goland:noinspection GoUnhandledErrorResult
	defer ln.Close()

	conn := dial(ln.Addr().String())
	reader, err := conn.Receive()
	if err != nil {
		panic(err)
	}
	_data, err := io.ReadAll(reader)
	conn.Close()
	if err != nil {
		panic(err)
	}
	assertEqual(string(_data), data)
	fmt.Println("testCase0 PASS")
}

type Pipe struct {
	lock       sync.Mutex
	buf        bytes.Buffer
	blockWrite chan struct{}
	blockRead  chan []byte
}

func newPipe() *Pipe {
	return &Pipe{
		blockWrite: make(chan struct{}),
		blockRead:  make(chan []byte),
	}
}

func (p *Pipe) Read(buf []byte) (n int, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.buf.Len() == 0 {
		_buf, ok := <-p.blockRead
		if ok {
			p.buf.Write(_buf)
			p.blockWrite <- struct{}{}
		}
	}
	return p.buf.Read(buf)
}

func (p *Pipe) Write(buf []byte) {
	p.blockRead <- buf
	<-p.blockWrite
}

func (p *Pipe) Close() {
	close(p.blockRead)
}

// 复杂 case：多连接，传输大量数据
func testCase1() {
	_log := log.New(os.Stdout, "[testCase1] ", log.LstdFlags)
	ln := startServer(func(conn *Conn) {
		defer conn.Close()
		for {
			// 服务端接收数据
			reader, err := conn.Receive()
			if err != nil {
				_log.Println("receive err:", err)
				return
			}
			var (
				_hash = sha256.New()
				buf   = make([]byte, 1<<10)
				total = 0
			)
			for {
				n, err := reader.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					panic(err)
				}
				_hash.Write(buf[:n])
				total += n
			}
			checksum := _hash.Sum(nil)
			_log.Println("server receive data checksum", hex.EncodeToString(_hash.Sum(nil)))
			// 服务端将接收到的数据的 checksum 作为响应发送给客户端
			err = conn.Send(len(checksum), bytes.NewBuffer(checksum))
			if err != nil {
				_log.Println("send err:", err)
				return
			}
		}
	})
	//goland:noinspection GoUnhandledErrorResult
	defer ln.Close()

	const (
		connNum  = 3
		dataNum  = 3
		dataSize = 100 << 20 //也可以是很大的数据，你的实现中不能假定传输数据为固定长度
	)
	var wg sync.WaitGroup
	//同时创建 connNum 个连接进行传输
	for i := 0; i < connNum; i++ {
		wg.Add(1)
		connId := i
		go func() {
			defer wg.Done()
			conn := dial(ln.Addr().String())
			//顺序发送 dataNum 组数据
			for j := 0; j < dataNum; j++ {
				dataId := j
				var (
					_hash    = sha256.New()
					buf      = make([]byte, 2<<10) //也可以是其他大小的 buf，你的实现中不能假定 buf 为固定长度
					pipe     = newPipe()
					checksum []byte
				)
				go func() {
					for j := 0; j < dataSize/len(buf); j++ {
						_, err := rand.Read(buf)
						if err != nil {
							panic(err)
						}
						_hash.Write(buf)
						checksum = _hash.Sum(nil)
						pipe.Write(buf)
					}
					pipe.Close()
					_log.Printf("connId[%v] dataId[%v] send checksum %v\n", connId, dataId, hex.EncodeToString(checksum))
				}()
				err := conn.Send(dataSize, pipe)
				if err != nil {
					panic(err)
				}
				reader, err := conn.Receive() //接收服务端响应其收到的数据的 checksum
				if err != nil {
					panic(err)
				}
				_checksum, err := io.ReadAll(reader)
				if err != nil {
					panic(err)
				}
				//客户端发送数据的 checksum 和服务端接收数据的 checksum 应该一致
				assertEqual(hex.EncodeToString(_checksum), hex.EncodeToString(checksum))
			}
			conn.Close()
		}()
	}
	wg.Wait()
	fmt.Println("testCase1 PASS")
}

func main() {
	testCase0()
	testCase1()
}
