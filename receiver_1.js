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
        var queue = 'task_queue';

        channel.assertQueue(queue, {
            durable: true});

        channel.prefetch(1);
        console.log(" [*] Waiting for message in %s. To exit press CTRL+c", queue);

        channel.consume(queue, function(msg) {
            var secs = msg.content.toString().split('.').length - 1;

            console.log(" [x] Received %s", msg.content.toString());
            setTimeout(function() {
                console.log(" [x] Done");
            }, secs * 1000);
        }, {
            noAck: false
        });

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 60 * 1000);
});
