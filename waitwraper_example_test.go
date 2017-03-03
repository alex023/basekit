package basetool

import "fmt"

func ExampleWaitWraper_Wrap() {
	fn := func() {
		fmt.Println("this just example for wrap!")
	}
	ww := WaitWraper{}
	ww.Wrap(fn)
	ww.Wait()
	//output:
	//
	//this just example for wrap!
}
