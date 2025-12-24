package events

import (
	"classroom-analysis/internal/domain"
	"sync"
)

type ResponseWaiter struct {
	sync.Map
}

var GlobalWaiter = &ResponseWaiter{}

func (rw *ResponseWaiter) Register(id string) chan *domain.KafkaResp {
	ch := make(chan *domain.KafkaResp, 1)
	rw.Store(id, ch)
	return ch
}
func (rw *ResponseWaiter) Notify(id string, resp *domain.KafkaResp) {
	if ch, ok := rw.Load(id); ok {
		// 安全的类型断言
		if resultChan, ok := ch.(chan *domain.KafkaResp); ok {
			select {
			case resultChan <- resp:
				// 成功发送
			default:
				// 通道已满或关闭，忽略
			}
		}
		rw.Delete(id)
	}
}
func (rw *ResponseWaiter) Delete(id string) {
	rw.Map.Delete(id)
}
