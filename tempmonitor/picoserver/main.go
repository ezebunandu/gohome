package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"machine"
	"net/netip"
	"time"

	"github.com/soypat/cyw43439"
	"github.com/soypat/cyw43439/examples/common"

	"github.com/soypat/seqs/httpx"
	"github.com/soypat/seqs/stacks"
)

const (
	connTimeout = 3 * time.Second
	maxconns    = 3
	tcpbufsize  = 2030
	hostname    = "picotemp"
	listenPort  = 80
)

type temp struct {
	TempC float64 `json:"tempC"`
	TempF float64 `json:"tempF"`
}

var logger *slog.Logger

func init() {
	logger = slog.New(
		slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
}

func changeLEDState(dev *cyw43439.Device, state bool) {
	if err := dev.GPIOSet(0, state); err != nil {
		logger.Error("failed to change LED state: ", slog.String("err", err.Error()))
	}
}

func setUpDevice() (*stacks.PortStack, *cyw43439.Device) {
	_, stack, dev, err := common.SetupWithDHCP(common.SetupConfig{
		Hostname: hostname,
		Logger:   logger,
		TCPPorts: 1,
	})
	if err != nil {
		panic("setup DHCP:" + err.Error())
	}

	// Turn LED on
	changeLEDState(dev, true)
	return stack, dev
}

func newListener(stack *stacks.PortStack) *stacks.TCPListener {
	// Start TCP server.
	listenAddr := netip.AddrPortFrom(stack.Addr(), listenPort)
	listener, err := stacks.NewTCPListener(
		stack, stacks.TCPListenerConfig{
			MaxConnections: maxconns,
			ConnTxBufSize:  tcpbufsize,
			ConnRxBufSize:  tcpbufsize,
		})
	if err != nil {
		panic("listener create:" + err.Error())
	}

	err = listener.StartListening(listenPort)
	if err != nil {
		panic("listener start:" + err.Error())
	}
	logger.Info("listening",
		slog.String("addr", "http://"+listenAddr.String()),
	)
	return listener
}

func blinkLED(dev *cyw43439.Device, blink chan uint) {
	for {
		select {
		case n := <-blink:
			lastLedState := true
			if n == 0 {
				n = 5
			}
			for i := uint(0); i < n; i++ {
				lastLedState = !lastLedState
				changeLEDState(dev, lastLedState)
				time.Sleep(500 * time.Millisecond)
			}
			// Ensure LED is on at the end
			(changeLEDState(dev, true))
		}
	}
}

func getTemperature() *temp {
	curTemp := machine.ReadTemperature()

	return &temp{
		TempC: float64(curTemp) / 1000,
		TempF: ((float64(curTemp) / 1000) * 9 / 5) + 32,
	}
}

func HTTPHandler(respWriter io.Writer, resp *httpx.ResponseHeader) {
	resp.SetConnectionClose()
	logger.Info("Got temperature request...")
	t := getTemperature()

	body, err := json.Marshal(t)
	if err != nil {
		logger.Error(
			"temperature json:",
			slog.String("err", err.Error()),
		)
		resp.SetStatusCode(500)
	} else {
		resp.SetContentType("application/json")
		resp.SetContentLength(len(body))
	}
	respWriter.Write(resp.Header())
	respWriter.Write(body)
}

func handleConnection(listener *stacks.TCPListener, blink chan uint) {
	// Reuse the same buffers for each
	// connection to avoid heap allocations
	var resp httpx.ResponseHeader
	buf := bufio.NewReaderSize(nil, 1024)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(
				"listener accept:",
				slog.String("err", err.Error()),
			)
			time.Sleep(time.Second)
			continue
		}
		logger.Info(
			"new connection",
			slog.String("remote", conn.RemoteAddr().String()),
		)
		err = conn.SetDeadline(time.Now().Add(connTimeout))
		if err != nil {
			conn.Close()
			logger.Error(
				"conn set deadline:",
				slog.String("err", err.Error()),
			)
			continue
		}
		buf.Reset(conn)
		resp.Reset()
		HTTPHandler(conn, &resp)
		conn.Close()

		blink <- 5
	}
}

func main() {
	stack, dev := setUpDevice()
	listener := newListener(stack)

	blink := make(chan uint, 3)
	go blinkLED(dev, blink)
	go handleConnection(listener, blink)

	for {
		select {
		case <-time.After(1 * time.Minute):
			logger.Info("Waiting for connections...")
		}
	}
}
