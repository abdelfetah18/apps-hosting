package buildexecutor

type BuildExecutor interface {
	Execute(srcContext, destination, appId, appName string) error
}
