package web

import (
	"classroom-analysis/internal/domain"
	"classroom-analysis/internal/events"
	"context"
	"time"
)

func (h *TeacherHandler) ReadKafka(c context.Context, taskId string) (<-chan *domain.KafkaResp, context.Context, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	resultChan := events.GlobalWaiter.Register(taskId)
	go func() {
		<-ctx.Done()
		events.GlobalWaiter.Delete(taskId)
	}()
	return resultChan, ctx, cancel, nil
}
