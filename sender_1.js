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
        var msg = process.argv.slice(2).join(' ') || "Hello World!";

        channel.assertExchange('logs', 'fanout', {durable: false});
        channel.publish('logs', '', Buffer.from(msg));  // temporary queues, fanout e.g. logger, we want to hear about all log messages
        console.log(" [x] Send %s", msg);

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 500);
});
