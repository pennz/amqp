#!/usr/bin/env node

var amqp = require('amqplib/callback_api');  // ampq Advance Message Queuing Protocol
var args = process.argv.slice(2);

if (args.length == 0) {
    console.log("Usage: receive_logs_direct.js [info] [warning] [error]");
    process.exit(1);
}

amqp.connect('amqp://vtool.duckdns.org', function(error0, connection) {
    if (error0) {
        throw error0;
    }

    connection.createChannel(function(error1, channel) {
        if (error1) {
            throw error1;
        }

        var exchange = 'direct_logs';

        channel.assertExchange(exchange, 'direct', {
            durable: false
        });

        channel.assertQueue('', {
            exclusive: true
        }, // the queue will be deleted after channel closed
            function(error2, q) {
                if (error2) {
                    throw error2;
                }
                console.log(" [*] Waiting for messages in %s. To exit press CTRL+c", q.queue);

                args.forEach(function(serverity) {
                    channel.bindQueue(q.queue, exchange, serverity);
                });

                channel.consume(q.queue, function(msg) {
                    if (msg.content) {
                        console.log(" [x] %s: '%s'", msg.fields.routingKey, msg.content.toString());
                    }
                }, {
                    noAck: true
                });
            }
        );

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 60 * 1000);
});
