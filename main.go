package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	getSS()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	i := r.Intn(len(servers))

	reconfigSS("/usr/local/etc/shadowsocks-libev.json", i)

	fmt.Println("ss server currently used by shadowsocks-libev")
	fmt.Println(servers[i])
	fmt.Println("please check /usr/local/etc/shadowsocks-libev.json for detail\n")
	fmt.Println("    cat /usr/local/etc/shadowsocks-libev.json\n")

	restartSS()

	fmt.Println("please check log in /tmp/log/ss-auto.log for detail\n")
	fmt.Println("    tail /tmp/log/ss-auto.log\n")

}

const (
	SHADOWSOCKS_ENTRY = "http://www.ishadowsocks.com/"
)

var (
	servers = []string{}

	templateJson = `{
    "server":"%s",
    "server_port":%s,
    "local_port":1080,
    "password":"%s",
    "timeout":600,
    "method":"%s"
}
`
)

func getSS() {
	doc, e := goquery.NewDocument(SHADOWSOCKS_ENTRY)
	if e != nil {
		fmt.Println(e)
	}
	c := doc.Find("#free").Find("div .col-lg-4.text-center")
	c.Each(func(i int, content *goquery.Selection) {
		serverInfo := []string{}
		content.Find("h4").Each(func(i int, content *goquery.Selection) {
			if i < 4 {
				serverInfo = append(serverInfo, strings.Split(content.Text(), ":")[1])
				// fmt.Println(i, serverInfo)
			}
		})
		// 如果密码为空, 不加入到 servers slice 中
		if serverInfo[2] != "" {
			s := fmt.Sprintf(templateJson, serverInfo[0], serverInfo[1], serverInfo[2], serverInfo[3])
			servers = append(servers, s)
		}
	})
}

func reconfigSS(filename string, i int) {
	outputFile, outputError := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)

	outputWriter.WriteString(servers[i])
	outputWriter.Flush()
}

func restartSS() {
	in := bytes.NewBuffer(nil)
	cmd := exec.Command("/bin/bash")

	// Stdout buffer
	cmdOutput := &bytes.Buffer{}
	// Attach buffer to command
	cmd.Stdout = cmdOutput

	cmd.Stdin = in

	cmdLine := "brew services restart shadowsocks-libev | tee -a /tmp/log/ss-auto.log"

	in.WriteString(cmdLine)

	printCommand(cmdLine)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return
	}

	printOutput(cmdOutput.Bytes())

}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

func printCommand(cmd string) {
	fmt.Printf("==> Executing: %s\n", cmd)
}
