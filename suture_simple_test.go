package suture

import "fmt"

type Incrementor struct {
	current int
	next    chan int
	stop    chan bool
	state   int
}

func (i *Incrementor) Stop() {
	fmt.Println("Stopping the service")
	i.stop <- true
}

func (i *Incrementor) Serve() {
	i.state = ServiceNotRunning
	for {
		select {
		case i.next <- i.current:
			i.state = ServiceNormal
			i.current++
		case <-i.stop:
			i.state = ServicePaused
			// We sync here just to guarantee the output of "Stopping the service",
			// so this passes the test reliably.
			// Most services would simply "return" here.
			i.stop <- true
			return
		}
	}
}

func (i *Incrementor) State() int {
	return i.state
}

func ExampleNew_simple() {
	supervisor := NewSimple("Supervisor")
	service := &Incrementor{0, make(chan int), make(chan bool), ServiceNotRunning}
	supervisor.Add(service)

	go supervisor.ServeBackground()

	fmt.Println("Got:", <-service.next)
	fmt.Println("Got:", <-service.next)
	supervisor.Stop()

	// We sync here just to guarantee the output of "Stopping the service"
	<-service.stop

	// Output:
	// Got: 0
	// Got: 1
	// Stopping the service
}
