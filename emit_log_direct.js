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

        var exchange = 'direct_logs';
        var args = process.argv.slice(2);
        var msg = args.slice(1).join(' ') || "Hello World!";
        var serverity = (args.length > 0) ? args[0] : 'info';

        channel.assertExchange(exchange, 'direct', {durable: false});
        channel.publish(exchange, serverity, Buffer.from(msg));  // temporary queues, fanout e.g. logger, we want to hear about all log messages
        console.log(" [x] Send %s: '%s'", serverity, msg);

    });
    setTimeout(function() {
        connection.close();
        process.exit(0);
    }, 500);
});
