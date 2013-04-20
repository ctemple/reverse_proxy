/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 13-4-20
 * Time: 上午9:57
 * To change this template use File | Settings | File Templates.
 */
package server

import "net"
import "fmt"
import "strconv"
import "time"
import "strings"


type reverseProxyCmd struct {
	conn net.Conn
	args []string
}

type ReverseProxyServer struct {
	listenerCli net.Listener
	listenerSrv net.Listener
	stop        chan int
	stopNotify 	chan int
	cmds 		chan reverseProxyCmd
	ctrlCmds 	chan string
	ctlConn		*net.Conn
	srvClis		map[string]net.Conn
	lastid		int64
}



func NewServer(srvPort, cliPort int) *ReverseProxyServer {
	var r *ReverseProxyServer = new(ReverseProxyServer)
	var err error
	r.listenerSrv, err = net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(srvPort)))
	if err != nil {
		panic(err.Error())
	}
	r.listenerCli, err = net.Listen("tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(cliPort)))
	if err != nil {
		panic(err.Error())
	}
	r.stop = make(chan int, 10)
	r.stopNotify = make(chan int, 10)
	r.ctrlCmds = make(chan string,100)
	r.cmds = make(chan reverseProxyCmd,10)
	r.srvClis = make(map[string]net.Conn)
	r.lastid = 1
	return r
}

func (self *ReverseProxyServer) MakeId() int64 {
	self.lastid = self.lastid + 1
	return self.lastid
}

func (self *ReverseProxyServer) Startup() {
	go self.runListenerSrv()
	go self.runListenerCli()
	go self.runDispatcher()
}

func (self *ReverseProxyServer) Cleanup() {
	self.stopNotify <- 1

	self.listenerSrv.Close()
	self.listenerCli.Close()

	<-self.stop
	<-self.stop
}

func (self *ReverseProxyServer) runListenerSrv() {
	for {
		conn, err := self.listenerSrv.Accept()
		if err == nil {
			if self.ctlConn == nil {
				go self.runCtlConnection(conn)
				fmt.Println("ctrl connection accepted")
			} else {
				go self.runSrvConnection(conn)
			}
		} else {
			fmt.Println("Srv:", err.Error())
			break
		}
	}
	self.stop <- 1
}

func (self *ReverseProxyServer) redirectTo(src net.Conn,dst net.Conn) {
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
			fmt.Println("redirect closed")
			dst.Close()
			break;
		}
	}
}

func (self *ReverseProxyServer) runDispatcher() {
	for {
		select {
		case cmd := <- self.cmds:
			switch(cmd.args[0]) {
			case "SRV":
				waitConn := self.srvClis[cmd.args[1]]
				if waitConn != nil {
					fmt.Println(cmd.args[1] + " success")
					go self.redirectTo(waitConn,cmd.conn)
					go self.redirectTo(cmd.conn,waitConn)
					delete (self.srvClis,cmd.args[1])
				} else {
					fmt.Print(cmd.args[1] + " failed")
					cmd.conn.Close()
				}
				break
			case "CLI":
				self.srvClis[cmd.args[1]] = cmd.conn
				break;
			}
			break
		case _ = <- self.stopNotify:
			self.stop <- 0
			return
		}

	}
	fmt.Println("runDispatcher ended")
}

func (self *ReverseProxyServer) runListenerCli() {
	for {
		conn, err := self.listenerCli.Accept()
		if err == nil {
			self.runCliConnection(conn)
		} else {
			fmt.Println("Cli:", err.Error())
			break
		}
	}
	self.stop <- 1
}

func (self *ReverseProxyServer) runSrvConnection(conn net.Conn) {
	// conn.SetDeadline(time.Now().Add(10 * time.Second))
	buff := make([]byte, 100)
	bytes, err := conn.Read(buff)
	if err == nil && bytes > 0 {
		cmd := string(buff[0:bytes])
		arg := strings.Split(cmd," ")
		if len(arg) <= 0 {
			conn.Close()
			return
		}
		self.cmds <- reverseProxyCmd{conn,arg}
	} else {
		conn.Close()
	}
}

func (self *ReverseProxyServer) runCtlConnection(conn net.Conn) {
	tmpbuf := make([]byte,4096)
	self.ctlConn = &conn
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
		_,err := conn.Read(tmpbuf)

		if err != nil {
			neterr, ok := err.(net.Error)
			if ok && !neterr.Timeout() {
				self.ctlConn = nil
				fmt.Println("ctrl connection closed ",err)
				break;
			}
		}

		// fmt.Println("process ctrl cmds")

		select {
		case cmd := <-self.ctrlCmds:
			// fmt.Println("ctrlcmd send:" + cmd)
			//(*self.ctlConn).SetWriteDeadline(time.Now().Add(time.Second))
			(*self.ctlConn).Write([]byte(cmd))
			break
		default:
			break;
		}

	}
	fmt.Println("runCtlConnection end")

}

func (self *ReverseProxyServer) runCliConnection(conn net.Conn) {
	if self.ctlConn != nil {
		id := self.MakeId()
		idstr := strconv.FormatInt(id,10)
		cmds := make([]string,2)
		cmds[0] = "CLI"
		cmds[1] = idstr
		self.cmds <- reverseProxyCmd{conn,cmds}

		// fmt.Println("cli connection:" + idstr)

		ctlCmd := "CLI " + idstr + "\n"
		self.ctrlCmds <- ctlCmd
	} else {
		conn.Close()
		fmt.Println("ctrl connection = nil")
	}
}
