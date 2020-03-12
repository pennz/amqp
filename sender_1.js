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
        var msg = 'hello world';

        channel.assertQueue(queue, {
            durable: false});
        channel.sendToQueue(queue, Buffer.from(msg));
        console.log(" [x] Send %s", msg);

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 500);
});
