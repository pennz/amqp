# How to

go build .

then run ./amqp


PC1$ ./amqp s anonymous.info  "...." # for sending work with 4 dots then return

PC2$ ./amqp r test-logs_topic anonymous.info # the receiving side
