#!/usr/bin/env node

var amqp = require('amqplib/callback_api');  // ampq Advance Message Queuing Protocol

amqp.connect('amqp://vtool.duckdns.org', function(error0, connection) {
    if (error0) {
        throw error0;
    }

    connection.createChannel(function(error1, channel) {
        if (error1) {
            throw error1;
        }
        var queue = 'rpc_queue';

        channel.assertQueue(queue, {
            durable: false});

        channel.prefetch(1);
        console.log(" [X] Awaiting RPC requests");
        channel.consume(queue, function reply(msg) {
            var n = parseInt(msg.content.toString());

            console.log(" [.] fib(%d)", n);

            var r = fibonacci(n);

            channel.sendToQueue(msg.properties.replyTo, Buffer.from(r.toString()), {
                correlationId: msg.properties.correlationId
            });
            console.log(" [x] Received %s", msg.content.toString());

            channel.ack(msg);
        }, {
            noAck: false
        });

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 60 * 1000);
});

function fibonacci(n) {
    if (n == 0 || n == 1)
        return n;
    else
        return fibonacci(n - 1) + fibonacci(n - 2);
}
