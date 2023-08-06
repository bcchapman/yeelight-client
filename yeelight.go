package yeelight

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	discoverMSG = "M-SEARCH * HTTP/1.1\r\n HOST:239.255.255.250:1982\r\n MAN:\"ssdp:discover\"\r\n ST:wifi_bulb\r\n"

	// timeout value for TCP and UDP commands
	timeout = time.Second * 30

	//SSDP discover address
	ssdpAddr = "239.255.255.250:1982"

	//CR-LF delimiter
	crlf = "\r\n"

	on  = "on"
	off = "off"
)

type (
	//Command represents COMMAND request to Yeelight device
	Command struct {
		ID     int           `json:"id"`
		Method string        `json:"method"`
		Params []interface{} `json:"params"`
	}

	// CommandResult represents response from Yeelight device
	CommandResult struct {
		ID     int           `json:"id"`
		Result []interface{} `json:"result,omitempty"`
		Error  *Error        `json:"error,omitempty"`
	}

	//Error struct represents error part of response
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	//Yeelight represents device
	Yeelight struct {
		addr string
		rnd  *rand.Rand
	}
)

// Discover discovers device in local network via ssdp
func Discover() (*Yeelight, error) {
	var err error

	ssdp, _ := net.ResolveUDPAddr("udp4", ssdpAddr)
	c, _ := net.ListenPacket("udp4", ":0")
	socket := c.(*net.UDPConn)
	socket.WriteToUDP([]byte(discoverMSG), ssdp)
	socket.SetReadDeadline(time.Now().Add(timeout))

	rsBuf := make([]byte, 1024)
	size, _, err := socket.ReadFromUDP(rsBuf)
	if err != nil {
		return nil, errors.New("no devices found")
	}
	rs := rsBuf[0:size]
	addr, err := parseAddr(string(rs))
	if err != nil {
		return nil, err
	}
	log.Printf("Device with ip %s found\n", addr)
	return New(addr), nil
}

// New creates new device instance for address provided
func New(addr string) *Yeelight {
	return &Yeelight{
		addr: addr,
		rnd:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetProp method is used to retrieve current property of smart LED.
func (y *Yeelight) GetProp(values ...interface{}) ([]interface{}, error) {
	r, err := y.executeCommand("get_prop", values...)
	if nil != err {
		return nil, err
	}
	return r.Result, nil
}

// SetPower is used to switch on or off the smart LED (software managed on/off).
func (y *Yeelight) SetPower(onStatus bool) error {
	var status string
	if onStatus {
		status = on
	} else {
		status = off
	}
	_, err := y.executeCommand("set_power", status)
	return err
}

func (y *Yeelight) randID() int {
	i := y.rnd.Intn(100)
	return i
}

func (y *Yeelight) newCommand(name string, params []interface{}) *Command {
	return &Command{
		Method: name,
		ID:     y.randID(),
		Params: params,
	}
}

// executeCommand executes command with provided parameters
func (y *Yeelight) executeCommand(name string, params ...interface{}) (*CommandResult, error) {
	return y.execute(y.newCommand(name, params))
}

// executeCommand executes command
func (y *Yeelight) execute(cmd *Command) (*CommandResult, error) {
	conn, err := net.Dial("tcp", y.addr)
	if nil != err {
		return nil, fmt.Errorf("cannot open connection to %s. %s", y.addr, err)
	}
	time.Sleep(time.Second)
	conn.SetReadDeadline(time.Now().Add(timeout))

	//write request/command
	b, _ := json.Marshal(cmd)
	fmt.Fprint(conn, string(b)+crlf)

	//	log.Println(string(b))

	//wait and read for response
	res, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("cannot read command result %s", err)
	}
	var rs CommandResult
	err = json.Unmarshal([]byte(res), &rs)
	if nil != err {
		return nil, fmt.Errorf("cannot parse command result %s", err)
	}
	if nil != rs.Error {
		return nil, fmt.Errorf("command execution error. Code: %d, Message: %s", rs.Error.Code, rs.Error.Message)
	}
	return &rs, nil
}

// Code Above is borrowed from https://github.com/avarabyeu/yeelight/

// parseAddr parses address from ssdp response
func parseAddr(msg string) (string, error) {
	exp := regexp.MustCompile(`Location: yeelight:\/\/*([\d|\.|:]*)`)
	matches := exp.FindStringSubmatch(msg)

	if len(matches) != 2 {
		return "", errors.New("UNABLE TO PARSE LOCATION OF DEVICE")
	}

	os.WriteFile("test.txt", []byte(matches[1]), fs.ModeAppend)
	return matches[1], nil
}

// SetPower is used to change color of the smart LED .
func (y *Yeelight) SetColor(hexRGB string) error {
	decimal, err := strconv.ParseInt(hexRGB, 16, 64)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Invalid hexRGB value %s", hexRGB))
	}

	_, err = y.executeCommand("set_rgb", decimal, "smooth", 500)
	return err
}
