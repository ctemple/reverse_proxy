/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 13-4-20
 * Time: 下午12:47
 * To change this template use File | Settings | File Templates.
 */
package client

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ReverseProxyClient struct {
	serverIp 	string
	serverPort 	int
	targetIp	string
	targetPort	int
	ctrlConn	net.Conn
	remainCmd	string
}

func substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func NewClient(ip string,port int,targetIp string,targetPort int) *ReverseProxyClient {
	var r *ReverseProxyClient = new(ReverseProxyClient)
	var err error
	r.serverIp = ip
	r.serverPort = port
	r.targetIp = targetIp
	r.targetPort = targetPort
	r.ctrlConn,err = net.Dial("tcp",ip + ":" + strconv.Itoa(port))
	fmt.Println("connect to " + ip + ":" + strconv.Itoa(port))
	if err != nil {
		panic(err.Error())
	}
	return r
}

func (self *ReverseProxyClient) Startup()  {
	go self.run()
}

func (self *ReverseProxyClient) run() {
	tmpbuf := make([]byte,4096)
	for {
		// self.ctrlConn.SetReadDeadline(time.Now().Add(3*time.Second))
		bytes,err := self.ctrlConn.Read(tmpbuf)
		if err == nil {
			// fmt.Println("readlen " + strconv.Itoa(bytes))
			if bytes > 0 {
				self.handleCommandBuf(tmpbuf[0:bytes])
			}
		} else {
			panic(err.Error())
		}
	}
}

func (self *ReverseProxyClient) newCli(strid string) {
	fmt.Println("newcli " + strid)
	scon,serr := net.Dial("tcp",self.serverIp + ":" + strconv.Itoa(self.serverPort))
	if serr != nil {
		return
	}
	ccon,cerr := net.Dial("tcp",self.targetIp + ":" + strconv.Itoa(self.targetPort))
	if cerr != nil {
		scon.Close()
		return
	}

	scon.Write([]byte("SRV " + strid))
	go self.redirectTo(scon,ccon)
	go self.redirectTo(ccon,scon)
}

func (self *ReverseProxyClient) handleCommandStr(cmd string) {
	args := strings.Split(cmd," ")
	switch(args[0]) {
	case "CLI":
		self.newCli(args[1])
		break
	}
}

func (self *ReverseProxyClient) handleCommandBuf(buf []byte) {
	self.remainCmd = self.remainCmd + string(buf)
	// fmt.Println(self.remainCmd)
	for {
		pos := strings.Index(self.remainCmd,"\n")
		if pos >= 0 {
			cmd := substr(self.remainCmd,0,pos)
			self.remainCmd = substr(self.remainCmd,pos+1,len(self.remainCmd))
			self.handleCommandStr(cmd)
		} else {
			return
		}
	}
}


func (self *ReverseProxyClient) redirectTo(src net.Conn,dst net.Conn) {
	tmpbuf := make([]byte,4096)
	for {
		//src.SetReadDeadline(time.Now().Add(3 + time.Second))
		//dst.SetWriteDeadline(time.Now().Add(3 + time.Second))
		bytes,err := src.Read(tmpbuf)
		if err == nil {
			// fmt.Println("redirect ",strconv.Itoa(bytes))
			_,werr := dst.Write(tmpbuf[0:bytes])
			if werr != nil {
				src.Close()
				fmt.Println("src redirect closed")
				break;
			}
		} else {
			dst.Close()
			fmt.Println("redirect closed")
			break;
		}
	}
}

