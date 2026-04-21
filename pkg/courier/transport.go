package courier

import (
	"log"
	"sync"
)

// Transport 表示传输层接口，用于承载路由器服务。
type Transport interface {
	Serve(router Router) error
}

// Run 启动路由器服务，使用指定的传输层。
func Run(router Router, transports ...Transport) {
	wg := &sync.WaitGroup{}

	for i := range transports {
		s := transports[i]

		wg.Go(func() {
			if err := s.Serve(router); err != nil {
				log.Println(err)
			}
		})
	}

	wg.Wait()
}
