package tasks

import "testing"

// 补扫每秒都会读 pending；重复入队会让队列被同一批任务占满。
func TestEnqueueSchedulesEachTaskOnce(t *testing.T) {
	service := &Service{
		queue:     make(chan uint, 2),
		scheduled: make(map[uint]struct{}),
	}

	if !service.enqueue(7) {
		t.Fatal("first enqueue returned false")
	}
	if !service.enqueue(7) {
		t.Fatal("duplicate enqueue returned false, want true (already scheduled)")
	}
	if len(service.queue) != 1 {
		t.Fatalf("queue length = %d, want 1", len(service.queue))
	}

	<-service.queue
	service.forgetScheduled(7)
	if !service.enqueue(7) {
		t.Fatal("enqueue after the task finished returned false")
	}
}

// 队列满时必须返回 false 并清掉记录，补扫下一轮才会重试。
func TestEnqueueReportsFullQueue(t *testing.T) {
	service := &Service{
		queue:     make(chan uint, 1),
		scheduled: make(map[uint]struct{}),
	}

	if !service.enqueue(1) {
		t.Fatal("enqueue into an empty queue returned false")
	}
	if service.enqueue(2) {
		t.Fatal("enqueue into a full queue returned true")
	}
	if _, scheduled := service.scheduled[2]; scheduled {
		t.Fatal("a task rejected by a full queue must not stay scheduled")
	}
}
