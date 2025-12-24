package web

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client 封装 WebSocket 连接，解决并发写入问题
type Client struct {
	Conn *websocket.Conn
	// 写锁：确保 Ping 和 业务消息不会并发写入导致 Panic
	mu     sync.Mutex
	TaskID string
}

// 线程安全的发送方法
func (c *Client) SafeWriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.Conn.WriteJSON(v)
}

// 线程安全的发送文本方法
func (c *Client) SafeWriteMessage(messageType int, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.Conn.WriteMessage(messageType, data)
}

type AnalysisWSManager struct {
	// 改变值类型为 *Client
	clients  map[string]*Client
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

// ... AnalysisStatusMessage 保持不变 ...
type AnalysisStatusMessage struct {
	Type      string `json:"type"`
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	ResultURL string `json:"resultUrl,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp int64  `json:"timestamp"`
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

func (m *AnalysisWSManager) HandleConnection(teacherId primitive.ObjectID, w http.ResponseWriter, r *http.Request) {
	if teacherId == primitive.NilObjectID {
		http.Error(w, "缺少taskId参数", http.StatusBadRequest)
		return
	}

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	taskID := teacherId.Hex()

	// 创建封装的 Client 对象
	client := &Client{
		Conn:   conn,
		TaskID: taskID,
	}

	// 注册
	m.registerClient(client)

	// 启动保活循环 (阻塞直到连接断开)
	m.keepAlive(client)

	// 循环结束意味着连接断开，执行清理
	m.unregisterClient(client)
}

func (m *AnalysisWSManager) registerClient(newClient *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	taskID := newClient.TaskID

	// 处理旧连接：如果有旧连接，关闭它
	if oldClient, exists := m.clients[taskID]; exists {
		log.Printf("任务 %s 检测到旧连接，正在关闭...", taskID)
		// 尝试发送关闭消息（尽力而为，不阻塞）
		go func(c *Client) {
			c.SafeWriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "被新连接替换"))
			c.Conn.Close()
		}(oldClient)
	}

	m.clients[taskID] = newClient
	log.Printf("任务 %s 的WebSocket连接已注册", taskID)

	// 发送欢迎消息 (现在是线程安全的，不需要 sleep)
	go func() {
		m.SendTaskStatusUpdate(taskID, "200", "连接已注册", "", "")
	}()
}

func (m *AnalysisWSManager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 只有当 map 里的连接就是当前要注销的连接时才删除
	// 防止删除了已经被新连接替换掉的记录
	if currentClient, exists := m.clients[client.TaskID]; exists && currentClient == client {
		delete(m.clients, client.TaskID)
		client.Conn.Close() // 确保关闭
		log.Printf("任务 %s 的WebSocket连接已注销", client.TaskID)
	}
}

// 保持连接活跃 (Read Loop + Write Ticker)
func (m *AnalysisWSManager) keepAlive(client *Client) {
	// 设置 Pong 处理
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动 Ping Ticker
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	// 用一个 channel 通知 write loop 停止
	done := make(chan struct{})

	// 1. 启动 Ping 发送协程
	go func() {
		for {
			select {
			case <-ticker.C:
				// 这里使用了 SafeWriteMessage，加了锁，不再会 Panic
				// 建议使用标准的 PingMessage 而不是 TextMessage ("ping")
				// 如果前端必须用文本 "ping"，请保留 websocket.TextMessage
				// 这里我演示标准做法：
				if err := client.SafeWriteMessage(websocket.PingMessage, nil); err != nil {
					return // 写入失败，退出
				}
			case <-done:
				return
			}
		}
	}()

	// 2. 主循环作为 Read Loop
	for {
		// 阻塞读取，如果出错或连接关闭，会从这里返回
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("任务 %s 连接异常关闭: %v", client.TaskID, err)
			}
			break // 退出 Read Loop
		}
		// 如果是 TextMessage 的心跳，可以在这里处理
		// 但使用标准 Ping/Pong 机制，ReadMessage 会自动处理 Pong，不需要手动写逻辑
	}

	// Read Loop 结束后，通知 Ping 协程退出
	close(done)
}

func (m *AnalysisWSManager) SendTaskStatusUpdate(taskID, status, message, resultURL, errorMsg string) {
	m.mutex.RLock()
	client, exists := m.clients[taskID]
	m.mutex.RUnlock()

	if !exists {
		// log.Printf("任务 %s 无活跃连接，跳过推送", taskID)
		return
	}

	msg := AnalysisStatusMessage{
		Type:      "analysis",
		TaskID:    taskID,
		Status:    status,
		Message:   message,
		ResultURL: resultURL,
		Error:     errorMsg,
		Timestamp: time.Now().Unix(),
	}

	// 调用线程安全的写方法
	if err := client.SafeWriteJSON(msg); err != nil {
		log.Printf("发送消息到任务 %s 失败: %v", taskID, err)
		// 发送失败通常意味着连接断了，unregisterClient 会在 handleConnection 退出时自动处理
		// 或者在这里主动关闭连接也可以触发清理
		client.Conn.Close()
	}
}
