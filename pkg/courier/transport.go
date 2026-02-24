package courier

import (
	"log"
	"sync"
)

type Transport interface {
	Serve(router Router) error
}

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
