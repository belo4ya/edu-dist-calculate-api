package repository

// Keys for expressions
func exprKey(id string) []byte {
	return []byte("expr:" + id)
}

func exprListKey(id string) []byte {
	return []byte("expr:list:" + id)
}

func exprTasksPrefix(exprID string) []byte {
	return []byte("expr:" + exprID + ":tasks:")
}

func exprTaskKey(exprID, taskID string) []byte {
	return []byte("expr:" + exprID + ":tasks:" + taskID)
}

// Keys for tasks
func taskKey(id string) []byte {
	return []byte("task:" + id)
}

func taskQueuePendingPrefix() []byte {
	return []byte("task:queue:pending:")
}

func taskQueuePendingKey(id string) []byte {
	return []byte("task:queue:pending:" + id)
}

// Helper functions to extract IDs from keys
func exprIDFromListKey(key []byte) string {
	return string(key)[len("expr:list:"):]
}

func taskIDFromExprTaskKey(key []byte, exprID string) string {
	return string(key)[len("expr:"+exprID+":tasks:"):]
}

func taskIDFromPendingQueueKey(key []byte) string {
	return string(key)[len("task:queue:pending:"):]
}
