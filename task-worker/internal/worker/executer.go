package worker

type Executer interface {
	ExecuteTask(rawPayload interface{}) error
}
