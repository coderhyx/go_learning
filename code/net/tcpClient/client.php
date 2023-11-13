<?php

$host = 'localhost';
$port = 8080;

// 要发送的字符串
$str = 'Hello, World! hyx';

// 创建一个TCP连接
$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
if ($socket === false) {
    echo "Error creating socket: " . socket_strerror(socket_last_error()) . PHP_EOL;
    exit(1);
}

// 连接到服务器
$result = socket_connect($socket, $host, $port);
if ($result === false) {
    echo "Error connecting to server: " . socket_strerror(socket_last_error($socket)) . PHP_EOL;
    exit(1);
}

// 将字符串转换为字节数组
$strBytes = str_split($str);

// 获取字节数组的长度
$length = count($strBytes);

// 创建一个包头，使用pack函数将长度转换为大端序的字节表示
$header = pack('N', $length);

// 发送包头
$result = socket_write($socket, $header, 4);
if ($result === false) {
    echo "Error sending header: " . socket_strerror(socket_last_error($socket)) . PHP_EOL;
    exit(1);
}

// 发送字节数据
$result = socket_write($socket, $str, $length);
if ($result === false) {
    echo "Error sending data: " . socket_strerror(socket_last_error($socket)) . PHP_EOL;
    exit(1);
}

echo "Data sent successfully" . PHP_EOL;

// 关闭连接
socket_close($socket);