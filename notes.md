[doc](https://pkg.go.dev/github.com/streadway/amqp)

You can find that peer verification in this section:
[SSL/TLS - Secure connections](https://pkg.go.dev/github.com/streadway/amqp#hdr-Use_Case)

Use **DialTLS** when you wish to provide a client certificate.

## DialTLS
You can check the source code of streadway/amqp.

Also the example is already provided at [this link](https://pkg.go.dev/github.com/streadway/amqp#example-DialTLS)

So here is some points that needs attention:

when setting peer verification, client side will need set cert and the fullchain
cert, i.e., the root ca, intermediate ca, 2nd layer intermediate ca, your ca, the 
full trust train. And for the trust ca store, 


## Knowledge about goroutine

[a medium blog about this](https://medium.com/a-journey-with-go/go-how-does-a-goroutine-start-and-exit-2b3303890452)

**runtime.Caller**

`func Caller(skip int) (pc uintptr, file string, line int, ok bool)`
Caller reports file and line number information about function invocations on
the calling goroutine's stack. The argument skip is the number of stack frames
to ascend, with 0 identifying the caller of Caller. (For historical reasons the
meaning of skip differs between Caller and Callers.) The return values report
the program counter, file name, and line number within the file of the
corresponding call. The boolean ok is false if it was not possible to recover
the information. 

Check this example code:
```go
func main () {
    var wg sync.WaitGroup
    wg.Add(1)

    go func() {
        var skip int
        for {
            if _, file, line, ok := runtime.Caller(skip); ok {
				fmt.Printf("%s: %d\n", file, line)
                skip++
            } else {
                break
            }
        }

        wg.Done()
    } ()
    wg.Wait()
}
```
