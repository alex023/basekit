//提供运行程序的启动、停止的框架,标准化信号拦截的运行分块。
// 代码是简化了“github.com/judwhite/go-svc/svc“,不再提供windows服务的特殊处理,从而不再依赖"golang.org/x/sys/windows/svc"。
// 如果需要可以将应用程序替换回去。
package svc

import "os/signal"

// Create variable signal.Notify function so we can mock it in tests
var signalNotify = signal.Notify

// Service interface contains Start and Stop methods which are called
// when the service is started and stopped. The Init method is called
// before the service is started, and after it's determined if the program
// is running as a Windows Service.
//
// The Start method must be non-blocking.
//
// Implement this interface and pass it to Run to start your program.
type Service interface {
	// Init is called before the program/service is started and after it's
	// determined if the program is running as a Windows Service.
	Init(Environment) error
	// Start is called after Init. This method must be non-blocking.
	Start() error
	// Stop is called in response to os.Interrupt, os.Kill, or when a
	// Windows Service is stopped.
	Stop() error
}

// Environment interface contains information about the environment
// your application is running in.
type Environment interface {
	// IsWindowsService returns true if the program is running as a
	// Windows Service.
	IsWindowsService() bool
}
