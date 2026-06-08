package utils

import (
	"strings"
	"sync"
)

// ringBuffer 环形缓冲区，固定容量，写入自动覆盖最旧条目，无切片重组开销
type ringBuffer struct {
	buf  []string
	cap  int
	head int // 下一个写入位置
	len  int // 当前有效长度
}

func newRingBuffer(cap int) *ringBuffer {
	return &ringBuffer{
		buf: make([]string, cap),
		cap: cap,
	}
}

// Push 写入一条记录，满时自动覆盖最旧的
func (r *ringBuffer) Push(s string) {
	r.buf[r.head] = s
	r.head = (r.head + 1) % r.cap
	if r.len < r.cap {
		r.len++
	}
}

// Snapshot 返回按时间顺序的完整快照（从旧到新）
func (r *ringBuffer) Snapshot() []string {
	if r.len == 0 {
		return nil
	}
	out := make([]string, r.len)
	start := (r.head - r.len + r.cap) % r.cap
	for i := 0; i < r.len; i++ {
		out[i] = r.buf[(start+i)%r.cap]
	}
	return out
}

// Reset 清空缓冲区
func (r *ringBuffer) Reset() {
	r.head = 0
	r.len = 0
}

// Broadcaster 实现了一个简单的字符串消息广播器，支持历史记录和 SSE 同步
type Broadcaster struct {
	clients    map[chan string]bool
	register   chan chan string
	unregister chan chan string
	messages   chan string
	history    *ringBuffer // 环形缓冲区存储最近的 50 条日志
	mu         sync.Mutex
	done       chan struct{} // 优雅关闭信号
}

const historyCapacity = 50

var GlobalBroadcaster *Broadcaster

func init() {
	GlobalBroadcaster = NewBroadcaster()
	go GlobalBroadcaster.run()
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients:    make(map[chan string]bool),
		register:   make(chan chan string),
		unregister: make(chan chan string),
		messages:   make(chan string, 1000),
		history:    newRingBuffer(historyCapacity),
		done:       make(chan struct{}),
	}
}

func (b *Broadcaster) run() {
	defer b.closeAllClients()
	for {
		select {
		case <-b.done:
			return
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.mu.Unlock()
		case message := <-b.messages:
			// 锁内仅做快照操作（更新历史 + 复制客户端列表），缩短持锁时间
			b.mu.Lock()
			// 更新历史记录（过滤掉纯数据事件，只保留文本日志）
			if !strings.HasPrefix(message, "[EVENT:") {
				b.history.Push(message)
			}
			// 快照当前客户端列表
			snapshot := make([]chan string, 0, len(b.clients))
			for client := range b.clients {
				snapshot = append(snapshot, client)
			}
			b.mu.Unlock()

			// 锁外遍历发送，避免阻塞 register/unregister
			for _, client := range snapshot {
				select {
				case client <- message:
				default:
					// 客户端读取太慢则跳过，防止阻塞整个系统
				}
			}
		}
	}
}

// Subscribe 注册一个新客户端
func (b *Broadcaster) Subscribe() chan string {
	client := make(chan string, 500)
	b.register <- client
	return client
}

// Unsubscribe 注销客户端
func (b *Broadcaster) Unsubscribe(client chan string) {
	b.unregister <- client
}

// GetRecent 获取最近的历史日志
func (b *Broadcaster) GetRecent() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.history.Snapshot()
}

// ClearRecent 清空最近的历史日志
func (b *Broadcaster) ClearRecent() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.history.Reset()
}

// Broadcast 发送广播消息（所有模块通过此方法输出实时日志）
func (b *Broadcaster) Broadcast(message string) {
	select {
	case b.messages <- message:
	default:
		// 队列满时忽略，防止极端高频日志影响系统稳定性
	}
}

// closeAllClients 关闭所有客户端 channel（由 run() 的 defer 调用，避免竞态）
func (b *Broadcaster) closeAllClients() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.clients {
		close(ch)
	}
	b.clients = make(map[chan string]bool)
}

// Shutdown 优雅关闭广播器，通知 run goroutine 退出
// 客户端 channel 由 run() 的 defer closeAllClients() 负责关闭，避免与发送循环竞态
func (b *Broadcaster) Shutdown() {
	close(b.done)
}
