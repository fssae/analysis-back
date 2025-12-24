package web

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/websocket"
)

// 轻量级WebSocket管理器，专门用于分析任务状态推送
type AnalysisWSManager struct {
	// 存储任务ID到连接的映射
	taskConnections map[string]*websocket.Conn
	// 存储连接ID到任务ID的映射，用于清理
	connToTask map[*websocket.Conn]string
	mutex      sync.RWMutex
	upgrader   websocket.Upgrader
}

// 分析任务状态消息
type AnalysisStatusMessage struct {
	Type      string `json:"type"`
	TaskID    string `json:"taskId"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	ResultURL string `json:"resultUrl,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// 全局WebSocket管理器实例
var analysisWSManager *AnalysisWSManager

// 初始化WebSocket管理器
func init() {
	analysisWSManager = &AnalysisWSManager{
		taskConnections: make(map[string]*websocket.Conn),
		connToTask:      make(map[*websocket.Conn]string),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该检查来源
			},
		},
	}
}

// 获取WebSocket管理器实例
func GetAnalysisWSManager() *AnalysisWSManager {
	return analysisWSManager
}

// 处理WebSocket连接
func (m *AnalysisWSManager) HandleConnection(teacherId primitive.ObjectID, w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取任务ID
	if teacherId == primitive.NilObjectID {
		http.Error(w, "缺少taskId参数", http.StatusBadRequest)
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	staskId := teacherId.Hex()

	// 注册连接
	m.registerConnection(staskId, conn)

	// 给连接一点时间稳定，然后再开始keepAlive
	// 这可以避免连接刚建立时就被立即关闭
	time.Sleep(100 * time.Millisecond)

	// 保持连接活跃（keepAlive会负责关闭连接）
	// 当keepAlive返回时，连接已经关闭，只需要清理注册信息
	m.keepAlive(conn)

	// 清理连接注册信息（连接已在keepAlive中关闭）
	log.Printf("即将断开连接")
	m.unregisterConnection(conn)
}

// 注册连接
func (m *AnalysisWSManager) registerConnection(taskID string, conn *websocket.Conn) {
	m.mutex.Lock()

	// 如果该任务已有连接，先清理旧连接
	var oldConn *websocket.Conn
	if existingConn, exists := m.taskConnections[taskID]; exists {
		oldConn = existingConn
		// 先清理映射关系
		delete(m.connToTask, oldConn)
		delete(m.taskConnections, taskID)
	}

	// 注册新连接
	m.taskConnections[taskID] = conn
	m.connToTask[conn] = taskID

	m.mutex.Unlock()

	// 在锁外关闭旧连接，避免阻塞
	if oldConn != nil {
		log.Printf("任务 %s 检测到旧连接，将优雅关闭旧连接并注册新连接", taskID)
		// 优雅关闭旧连接，但使用goroutine避免阻塞新连接注册
		go func() {
			// 先尝试发送关闭消息，给客户端一个正常的关闭信号
			deadline := time.Now().Add(2 * time.Second)
			if err := oldConn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "连接被新连接替换"),
				deadline); err != nil {
				// 如果发送关闭消息失败，直接关闭连接
				log.Printf("任务 %s 发送关闭消息失败，直接关闭旧连接: %v", taskID, err)
			}
			// 等待一小段时间，让客户端收到关闭消息
			time.Sleep(100 * time.Millisecond)
			// 然后关闭连接
			oldConn.Close()
			log.Printf("任务 %s 的旧WebSocket连接已关闭", taskID)
		}()
	}

	log.Printf("任务 %s 的WebSocket连接已注册", taskID)

	// 延迟发送注册消息，确保连接已完全建立
	go func() {
		// 等待一小段时间，确保连接稳定
		time.Sleep(200 * time.Millisecond)
		m.SendTaskStatusUpdate(taskID, "200", "连接已注册", "url", "error")
	}()
}

// 注销连接
func (m *AnalysisWSManager) unregisterConnection(conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if taskID, exists := m.connToTask[conn]; exists {
		delete(m.taskConnections, taskID)
		delete(m.connToTask, conn)
		log.Printf("任务 %s 的WebSocket连接已注销", taskID)
	}
}

// 发送任务状态更新
// ... existing code ...
// 发送任务状态更新
func (m *AnalysisWSManager) SendTaskStatusUpdate(taskID, status, message, resultURL, errorMsg string) {
	log.Printf("SendMessage to %s", taskID)

	// 获取连接
	m.mutex.RLock()
	conn, exists := m.taskConnections[taskID]
	m.mutex.RUnlock()

	if !exists || conn == nil {
		log.Printf("任务 %s 没有活跃的WebSocket连接", taskID)
		return
	}

	statusMessage := AnalysisStatusMessage{
		Type:      "analysis",
		TaskID:    taskID,
		Status:    status,
		Message:   message,
		ResultURL: resultURL,
		Error:     errorMsg,
		Timestamp: time.Now().Unix(),
	}

	// 发送消息，如果失败则清理连接
	if !m.sendMessage(conn, statusMessage, taskID) {
		// 发送失败，清理无效连接
		m.cleanupInvalidConnection(taskID, conn)
	}
}

// ... existing code ...

// 发送消息到指定连接
// ... existing code ...
// 发送消息到指定连接
// 返回true表示发送成功，false表示发送失败
func (m *AnalysisWSManager) sendMessage(conn *websocket.Conn, message AnalysisStatusMessage, taskID string) bool {
	// 检查连接是否有效
	if conn == nil {
		log.Printf("连接为空，无法发送消息")
		return false
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return false
	}

	// 设置写超时
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	log.Printf("开始发送WebSocket消息到任务 %s", taskID)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("发送WebSocket消息失败: %v", err)
		return false
	}
	return true
}

// 清理无效连接
func (m *AnalysisWSManager) cleanupInvalidConnection(taskID string, conn *websocket.Conn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 再次检查，确保连接确实无效
	if currentConn, exists := m.taskConnections[taskID]; exists && currentConn == conn {
		delete(m.taskConnections, taskID)
		delete(m.connToTask, conn)
		log.Printf("已清理任务 %s 的无效WebSocket连接", taskID)
	}
}

// ... existing code ...

// 保持连接活跃
// 保持连接活跃
func (m *AnalysisWSManager) keepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	// 使用channel来协调goroutine，当连接关闭时通知主循环
	done := make(chan struct{})
	var once sync.Once

	// 关闭连接的统一函数，确保只关闭一次
	closeConn := func() {
		once.Do(func() {
			// 通知主循环连接已关闭
			select {
			case <-done:
				// done channel已关闭，说明已经通知过了
			default:
				close(done)
			}
			// 关闭连接
			conn.Close()
		})
	}

	// 设置初始读取超时（给连接更多时间稳定）
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 在单独的 goroutine 中处理读取
	go func() {
		defer closeConn()

		// 给连接一点时间稳定后再开始读取
		time.Sleep(50 * time.Millisecond)

		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				// 检查是否是正常的关闭错误
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket连接异常关闭: %v", err)
				} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("WebSocket连接正常关闭")
				} else {
					log.Printf("WebSocket读取错误: %v", err)
				}
				return
			}

			// 处理文本心跳消息
			if messageType == websocket.TextMessage {
				message := string(data)
				if message == "pong" {
					// 收到客户端pong响应，更新读取超时
					conn.SetReadDeadline(time.Now().Add(60 * time.Second))
				}
				// 继续处理其他业务消息...
			}
		}
	}()

	// 主循环发送文本心跳ping
	for {
		select {
		case <-done:
			// 连接已关闭，退出循环
			log.Printf("检测到连接已关闭，停止发送ping")
			return
		case <-ticker.C:
			// 在发送前检查连接状态
			select {
			case <-done:
				// 连接已关闭，退出
				return
			default:
				// 连接正常，发送ping
				conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
					log.Printf("发送ping消息失败: %v", err)
					// 发送失败，关闭连接
					closeConn()
					return
				}
			}
		}
	}
}

// 获取活跃连接数量
func (m *AnalysisWSManager) GetActiveConnectionsCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.taskConnections)
}

// 获取所有活跃的任务ID
func (m *AnalysisWSManager) GetActiveTaskIDs() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	taskIDs := make([]string, 0, len(m.taskConnections))
	for taskID := range m.taskConnections {
		taskIDs = append(taskIDs, taskID)
	}
	return taskIDs
}
