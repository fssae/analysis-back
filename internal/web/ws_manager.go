package web

import (
	"classroom-analysis/internal/domain"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client 封装 WebSocket 连接，解决并发写入问题
type Client struct {
	Hub  *AnalysisWSManager
	Conn *websocket.Conn
	//灰度测试pump
	send   chan []byte
	TaskID string
}
type AnalysisWSManager struct {
	clients  map[string]*Client
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

var analysisWSManager *AnalysisWSManager

func init() {
	analysisWSManager = &AnalysisWSManager{
		clients: make(map[string]*Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}
func GetAnalysisWSManager() *AnalysisWSManager {
	return analysisWSManager
}

// HandleConnection 处理连接的核心入口
func (m *AnalysisWSManager) HandleConnection(taskId string, w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 创建 Client 对象
	client := &Client{
		Hub:    m,
		Conn:   conn,
		TaskID: taskId,
		send:   make(chan []byte, 256), // 256 是缓冲大小
	}

	// 1. 注册 (清理旧连接)
	m.registerClient(client)

	// 2. 启动写协程 (WritePump)
	// 专门负责把 send channel 里的数据和 Ping 写给客户端
	go client.writePump()

	// 3. 启动读协程 (ReadPump)
	// 专门负责处理 Pong 和检测连接断开。
	// 注意：这里我们让 ReadPump 阻塞当前函数，直到连接断开。
	// 这样当连接断开时，HandleConnection 才会返回。
	client.readPump()
}

func (m *AnalysisWSManager) registerClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskID := client.TaskID

	// 处理旧连接
	if oldClient, exists := m.clients[taskID]; exists {
		log.Printf("任务 %s 检测到旧连接，正在断开旧连接...", taskID)
		//触发 oldClient.writePump 中的 !ok 分支,
		//发送 CloseMessage 关闭 TCP 连接
		close(oldClient.send)
		delete(m.clients, taskID)
	}

	// 注册新连接
	m.clients[taskID] = client
	log.Printf("任务 %s 的WebSocket连接已注册", taskID)

	// 发送欢迎消息
	select {
	case client.send <- []byte("{\"type\":\"analysis\",\"status\":\"200\",\"message\":\"连接已注册\"}"):
	default:
	}
}

func (m *AnalysisWSManager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 双重检查：确保我们要删除的确实是 map 里存的那个对象
	// 防止：新连接刚注册(registerClient)，旧连接的 readPump 此时断开触发注销，
	// 如果不检查，旧连接可能会误删新连接的记录。
	if currentClient, exists := m.clients[client.TaskID]; exists && currentClient == client {
		delete(m.clients, client.TaskID)

		// 【关键修改】：关闭通道，通知 writePump 退出
		close(client.send)

		log.Printf("任务 %s 的WebSocket连接已注销", client.TaskID)
	}
}
func (c *Client) readPump() {
	defer c.Hub.unregisterClient(c)
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	//设置 Pong 处理器
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}

}
func (c *Client) writePump() {
	ticker := time.NewTicker(20 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel 被关闭，发送 Close 帧
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 获取 Writer
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// 写入消息
			w.Write(message)

			// 【修改点】：删除 "for i < n" 的合并循环
			// 确保每一条 send 消息对应前端一个 onmessage 事件，兼容 JSON.parse

			// 关闭 Writer，发送当前帧
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendTaskStatusUpdate: 对外暴露的方法
func (m *AnalysisWSManager) SendTaskStatusUpdate(taskID, status, message, resultURL, errorMsg string) {
	// 1. 获取连接对象
	m.mutex.RLock()
	client, exists := m.clients[taskID]
	m.mutex.RUnlock()

	if !exists {
		return
	}

	// 2. 构造消息
	msg := domain.AnalysisStatusMessage{
		Type:      "analysis",
		TaskID:    taskID,
		Status:    status,
		Message:   message,
		ResultURL: resultURL,
		Error:     errorMsg,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON序列化失败: %v", err)
		return
	}

	// 3发送到 Channel
	select {
	case client.send <- data:
		// 发送成功
	default:
		// Channel 满了（客户端阻塞），断开连接防止服务器内存泄漏
		log.Printf("任务 %s 客户端阻塞，断开连接", taskID)
		m.unregisterClient(client)
	}
}
