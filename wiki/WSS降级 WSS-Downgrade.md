# 完整的WSS降级方案（同时修改客户端和服务端） / Complete WSS Downgrade Solution (Modify Both Client and Server)

## 中文 / Chinese

**警告**：降级到WS协议会失去加密保护，仅在受控测试环境中使用！

## 步骤1：修改服务端 [server_api.py](file:///home/zhaobokai/vscode/RATFF/code/server_api.py)

找到服务器主循环函数中的SSL配置部分：
```python
# 服务器主循环
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info("%s: %s, %s: %s", lp.g("certificate_path"), SSL_CERT, lp.g("key_path"), SSL_KEY)

    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info(lp.g("ssl_certificate_loaded_successfully"))
    except FileNotFoundError:
        logging.error(lp.g("certificate_file_not_found"))
        exit(1)
    except Exception as e:
        logging.error("%s: %s", lp.g("ssl_loading_failed"), str(e))
        exit(1)

    logging.info("%s: %s:%s", lp.g("starting_server"), HOST, PORT)
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
```

**修改为非SSL版本**：
```python
# 服务器主循环 - WS版本（无SSL）
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info("注意：已切换为非加密WS协议")

    # 移除SSL配置，使用None表示不启用SSL
    ssl_context = None

    logging.info("%s: %s:%s", lp.g("starting_server"), HOST, PORT)
    try:
        # 注意：ssl参数设为None
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
```

## 步骤2：修改客户端 [client/main.go](file:///home/zhaobokai/vscode/RATFF/code/client/main.go)

**注意**：从v3.0版本开始，客户端代码位于`code/client/`目录下。

找到客户端连接部分：
```go
// 注意：此处为了简便，跳过了证书验证，如果在生产环境和不受信任的环境中，请将INSECURESKIPVERIFY参数值改为false
tls_config := &tls.Config{
	InsecureSkipVerify: INSECURESKIPVERIFY,
}

dialer := websocket.Dialer{
	TLSClientConfig: tls_config,
}

for {
	conn, _, err := dialer.Dial(
		fmt.Sprintf("wss://%s:%d", HOST, PORT),
		nil,
	)
```

**修改为WS版本**：
```go
// 注意：已切换为非加密WS协议，仅适用于测试环境
dialer := websocket.Dialer{}

for {
	conn, _, err := dialer.Dial(
		fmt.Sprintf("ws://%s:%d", HOST, PORT),  // 注意协议从wss改为ws
		nil,
	)
```

## 步骤3：重新编译和部署 / Step 3: Recompile and Deploy

1. **重新编译客户端**：
   ```bash
   cd code/client
   go build -o client .
   ```

2. **重启服务端**：
   ```bash
   cd code
   python server_api.py
   ```

# 安全建议：
1. **生产环境强烈建议使用WSS协议**
2. 只在本地测试或受控环境中考虑使用WS协议
3. 如果使用WS协议，请确保在网络层面有其他安全措施（如VPN、内网隔离等）

如果想要禁用WSS加密并使用WS协议，请严格按照上述完整方案操作。

---

## English / 英文

**Warning**: Downgrading to WS protocol loses encryption protection and should only be used in controlled testing environments!

## Step 1: Modify Server [server_api.py](file:///home/zhaobokai/vscode/RATFF/code/server_api.py)

Find the SSL configuration section in the server main loop function:
```python
# Server main loop
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info("%s: %s, %s: %s", lp.g("certificate_path"), SSL_CERT, lp.g("key_path"), SSL_KEY)

    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info(lp.g("ssl_certificate_loaded_successfully"))
    except FileNotFoundError:
        logging.error(lp.g("certificate_file_not_found"))
        exit(1)
    except Exception as e:
        logging.error("%s: %s", lp.g("ssl_loading_failed"), str(e))
        exit(1)

    logging.info("%s: %s:%s", lp.g("starting_server"), HOST, PORT)
    try:
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
```

**Modify to non-SSL version**:
```python
# Server main loop - WS version (no SSL)
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info("Note: Switched to non-encrypted WS protocol")

    # Remove SSL configuration, use None to indicate no SSL
    ssl_context = None

    logging.info("%s: %s:%s", lp.g("starting_server"), HOST, PORT)
    try:
        # Note: ssl parameter set to None
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
```

## Step 2: Modify Client [client/main.go](file:///home/zhaobokai/vscode/RATFF/code/client/main.go)

**Note**: Starting from v3.0, the client code is located in the `code/client/` directory.

Find the client connection section:
```go
// Note: For convenience, certificate verification is skipped here. In production and untrusted environments, please change the INSECURESKIPVERIFY parameter value to false
tls_config := &tls.Config{
	InsecureSkipVerify: INSECURESKIPVERIFY,
}

dialer := websocket.Dialer{
	TLSClientConfig: tls_config,
}

for {
	conn, _, err := dialer.Dial(
		fmt.Sprintf("wss://%s:%d", HOST, PORT),
		nil,
	)
```

**Modify to WS version**:
```go
// Note: Switched to non-encrypted WS protocol, only suitable for testing environment
dialer := websocket.Dialer{}

for {
	conn, _, err := dialer.Dial(
		fmt.Sprintf("ws://%s:%d", HOST, PORT),  // Note: protocol changed from wss to ws
		nil,
	)
```

## Step 3: Recompile and Deploy

1. **Recompile Client**:
   ```bash
   cd code/client
   go build -o client .
   ```

2. **Restart Server**:
   ```bash
   cd code
   python server_api.py
   ```

# Security Recommendations:
1. **Strongly recommend using WSS protocol in production environment**
2. Only consider using WS protocol in local testing or controlled environments
3. If using WS protocol, ensure other network-level security measures are in place (such as VPN, intranet isolation, etc.)

If you want to disable WSS encryption and use WS protocol, please follow the above complete solution strictly.
