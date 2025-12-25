package web

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/events"
	"context"
	"time"
)

func (h *TeacherHandler) ReadKafka(c context.Context, imageId string) (<-chan *domain.KafkaResp, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	resultChan := events.GlobalWaiter.Register(imageId)
	// 3. 自动清理逻辑：如果超时了，把 Map 里的 channel 删掉防止内存泄漏
	go func() {
		<-ctx.Done()
		// 超时清理
		events.GlobalWaiter.Delete(imageId)
	}()
	return resultChan, ctx, cancel, nil
}
