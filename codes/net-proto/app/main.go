package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

// Conn 是你需要实现的一种连接类型，它支持下面描述的若干接口；
// 为了实现这些接口，你需要设计一个基于 TCP 的简单协议；
type Conn struct {
	tcpconn net.Conn
}

// Send size 表示要传输的数据总长度；
// 你需要实现从 reader 读取数据，并将数据通过 TCP 进行传输；
func (conn *Conn) Send(size int, reader io.Reader) (err error) {
	var status uint8 = 'S'
	packet := new(bytes.Buffer)
	buf := make([]byte, 2<<10)
	var sentbytes int64 = 0
	var remained []byte = nil

	binary.Write(packet, binary.BigEndian, int64(size))
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if err == io.EOF {
				if sentbytes+int64(n) != int64(size) {
					return err // TODO
				} else {
					status = 'F'
				}
			} else if err == nil {
				status = 'C'
			} else {
				return err
			}
		} else {
			if err != nil && err != io.EOF {
				return err
			}
		}

		if n > 0 {
			binary.Write(packet, binary.BigEndian, status)
			binary.Write(packet, binary.BigEndian, n)
			if remained != nil {
				binary.Write(packet, binary.BigEndian, append(remained, buf[:n]...))
			} else {
				binary.Write(packet, binary.BigEndian, buf[:n])
			}
			var count = 0
			count, err = conn.tcpconn.Write(packet.Bytes())
			fmt.Println("sentbytes:", count)
			if count < n {
				remained = buf[count:n]
			}
			sentbytes += int64(count)
		}
		if status == 'F' {
			return nil
		}
	}
}

type DataReader struct {
	tcpconn   net.Conn
	totalsize int64
	status    uint8
	length    int64
	consumed  int64
}

func (reader *DataReader) Read(buff []byte) (int, error) {
	buf := make([]byte, 2<<10)
	packet := new(bytes.Buffer)

	const (
		TOTALSIZE_LEN = 8
		STATUS_LEN    = 1
		LENGTH_LEN    = 8
	)
	var fieldbuf []byte = make([]byte, 8)
	var (
		partial_totalsize []byte = nil
		partial_length    []byte = nil
	)

	for {
		n, err := reader.tcpconn.Read(buf)
		fmt.Println("total ", n, " is read")
		if n > 0 {
			packet = bytes.NewBuffer(buf[:n])

			if reader.totalsize == -1 {
				if len(partial_totalsize) == 0 {
					if packet.Len() > TOTALSIZE_LEN {
						// fieldbuf = packet.Next(TOTALSIZE_LEN)
						c, _ := packet.Read(fieldbuf)
						fmt.Println("c read ", c)
						i, e := binary.Varint(fieldbuf)
						reader.totalsize = i
						fmt.Println("totalsize ", reader.totalsize, " error:", e)
					} else {
						partial_totalsize = packet.Bytes()
						continue
					}
				} else {
					if packet.Len() > TOTALSIZE_LEN-len(partial_totalsize) {
						fieldbuf = packet.Next(TOTALSIZE_LEN - len(partial_totalsize))
						reader.totalsize, _ = binary.Varint(append(partial_totalsize, fieldbuf...))
					} else {
						partial_totalsize = append(partial_totalsize, packet.Bytes()...)
						continue
					}
				}
			}

			if reader.status == 'U' {
				if packet.Len() > STATUS_LEN {
					fieldbuf = packet.Next(STATUS_LEN)
					reader.status = fieldbuf[0]
				} else {
					continue
				}
			}

			if reader.length == -1 {
				if len(partial_length) == 0 {
					if packet.Len() > LENGTH_LEN {
						fieldbuf = packet.Next(LENGTH_LEN)
						reader.length, _ = binary.Varint(fieldbuf)
					} else {
						partial_length = packet.Bytes()
						continue
					}
				} else {
					if packet.Len() > LENGTH_LEN-len(partial_length) {
						fieldbuf = packet.Next(LENGTH_LEN - len(partial_length))
						reader.length, _ = binary.Varint(append(partial_length, fieldbuf...))
					} else {
						partial_length = append(partial_length, packet.Bytes()...)
						continue
					}
				}
			}

			if reader.status == 'C' {
				if packet.Len() > 0 {
					consumed, e := packet.Read(buff)
					reader.consumed += int64(consumed)
					return consumed, e
				} else {
					return 0, nil
				}
			} else if reader.status == 'F' {
				if packet.Len() > 0 {
					consumed, e := packet.Read(buff)
					reader.consumed += int64(consumed)
					return consumed, e
				} else {
					if reader.consumed == reader.length {
						return 0, io.EOF
					} else {
						return 0, nil
					}
				}
			}

			if err != nil {
				break
			}

		}
	}
	return 0, nil
}

// Receive 返回的 reader 用于接收数据；
// 你需要实现向 reader 中写入从 TCP 接收到的数据；
func (conn *Conn) Receive() (reader io.Reader, err error) {
	return &DataReader{
		tcpconn:   conn.tcpconn,
		totalsize: -1,
		status:    'U',
		length:    -1,
		consumed:  0,
	}, nil
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
	// testCase1()
}
