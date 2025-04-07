package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

const (
	// 定义一个简单的协议
	protocolID = "/example/1.0.0"
)

// 处理传入的流
func handleStream(stream network.Stream) {
	// 创建一个缓冲区来存储传入的数据
	buf := make([]byte, 1024)

	// 从流中读取数据
	n, err := stream.Read(buf)
	if err != nil {
		fmt.Println("Error reading from stream:", err)
		return
	}

	message := string(buf[:n])
	fmt.Printf("Received message: %s\n", message)

	// 发送响应
	response := "Hello back!"
	_, err = stream.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing to stream:", err)
	}

	// 关闭流
	err = stream.Close()
	if err != nil {
		fmt.Println("Error closing stream:", err)
	}
}

func main() {
	// 创建一个新的 libp2p 主机
	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}

	// 设置流处理器
	h.SetStreamHandler(protocol.ID(protocolID), handleStream)

	// 打印主机信息
	fmt.Printf("Host ID: %s\n", h.ID())
	fmt.Printf("Host Addresses:\n")
	for _, addr := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, h.ID())
	}

	// 如果提供了对等节点地址作为命令行参数，则连接到该节点
	if len(os.Args) > 1 {
		peerAddr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			panic(err)
		}

		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			panic(err)
		}

		// 连接到对等节点
		err = h.Connect(context.Background(), *peerInfo)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Connected to peer: %s\n", peerInfo.ID)

		// 打开一个流
		stream, err := h.NewStream(context.Background(), peerInfo.ID, protocol.ID(protocolID))
		if err != nil {
			panic(err)
		}

		// 发送消息
		message := "Hello from libp2p!"
		_, err = stream.Write([]byte(message))
		if err != nil {
			panic(err)
		}

		// 读取响应
		buf := make([]byte, 1024)
		n, err := stream.Read(buf)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Received response: %s\n", string(buf[:n]))

		// 关闭流
		err = stream.Close()
		if err != nil {
			fmt.Println("Error closing stream:", err)
		}
	}

	// 等待中断信号以优雅地关闭
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	fmt.Println("Shutting down...")
	if err := h.Close(); err != nil {
		panic(err)
	}
}
