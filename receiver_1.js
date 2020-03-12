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
        var queue = 'hello';

        channel.assertQueue(queue, {
            durable: false});
        console.log(" [*] Waiting for message in %s. To exit press CTRL+c", queue);
        channel.consume(queue, function(msg) {
            console.log(" [x] Received %s", msg.content.toString());
        }, {
            noAck: true
        });

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 500);
});
