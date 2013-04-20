/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 13-4-14
 * Time: 上午11:21
 * To change this template use File | Settings | File Templates.
 */
package main

import "fmt"
import "server"
import "client"
import "time"
import "flag"

var (
	apptype = flag.String("type","server","-type client | server")
	sport = flag.Int("srv_port",0,"srv_port")
	cport = flag.Int("cli_port",0,"cli_port")

	ip = flag.String("ip","ip","srv ip")
	port = flag.Int("port",0,"srv port")
	tip = flag.String("tip","rip","target ip")
	tport = flag.Int("tport",0,"target port")
)

func printUseage() {
	fmt.Println("-type server -srv_port port -cli_port port")
	fmt.Println("-type client -ip ip -port port -tip targetip -tport targetport")
}

func main() {
	flag.Parse()

	switch((*apptype)) {
	case "server":
		fmt.Println("server")
		if *sport == 0 || *cport == 0 {
			printUseage()
			return
		}
		srv := server.NewServer(*sport,*cport)
		srv.Startup()
		for {
			time.Sleep(100 * time.Millisecond)
		}
		break;
	case "client":
		fmt.Println("client")

		if ip == nil || port == nil || tip == nil || tport == nil || *port == 0 || *tport == 0{
			printUseage()
			return
		} else {
			cli := client.NewClient(*ip,*port,*tip,*tport)
			cli.Startup()
			for {
				time.Sleep(100 * time.Millisecond)
			}
		}

		break;
	default:
		printUseage()
		break
	}
}
