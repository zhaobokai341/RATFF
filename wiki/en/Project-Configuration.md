# Project Configuration

This project supports both Chinese and English languages. Please choose the corresponding documentation version according to your needs:
- Chinese Documentation: [Environment Setup](zh/准备环境.md) | [Running](zh/运行.md)
- English Documentation: [Environment Setup](en/Environment-Setup.md) | [Running](en/Running.md)

Enter the `RATFF/code` directory.

## server_api.py
Open `server_api.py` and find code similar to this:
```python
# Server configuration
HOST = '0.0.0.0'
PORT = 8765
WEB_HOST = '0.0.0.0'
WEB_PORT = 5000
SSL_CERT = 'cert.pem'
SSL_KEY = 'key.pem'
LANGUAGE = 'zh'
SECURITY_PATH = 'fuck'
SECURITY_PASSWORD_HASH = b'$2b$04$T8NZ.WUIuO05WyVpLrQYiOdgqc2zbx7E9ysF03696dYvwGohCFzwC'
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|HOST/PORT|IP address and port where websocket server runs|
|WEB_HOST/WEB_PORT|IP address and port where quart server runs|
|SSL_CERT/SSL_KEY|Path to certificate key required when websocket server starts|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|SECURITY_PATH|Security path when quart server starts|
|SECURITY_PASSWORD_HASH|Security password bcrypt hash value when quart server starts (default password is fuck)|

Modify the configuration as needed.

There are two other files `server.py` and `server_web.py`

Used for command-line control and web control respectively.

Choose as needed.

## server.py
Open `server.py` and find code similar to this:
```python
# Server configuration
LANGUAGE = "zh"
API_SITE = "http://localhost:5000"  # Server address
API_PATH = "fuck"  # API path
APT_PASSWORD = "fuck"  # Access password
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|API_SITE|Configure the address of the connected quart server (format `http/https://{HOST}:{PORT}`)|
|API_PATH/APT_PASSWORD|Security path and password required by the connected server|

All parameters correspond to the configuration in `server_api.py`

Modify as needed.

## server_web.py
Open `server_web.py` and find code similar to this:
```python
LANGUAGE = "zh"
WEB_HOST = "0.0.0.0"
WEB_PORT = 8000
API_SITE = 'http://localhost:5000'
SECURITY_PATH = 'fuck'  # Security path
SECURITY_PASSWORD_HASH = '6ac3c336e4094835293a3fed8a4b5fedde1b5e2626d9838fed50693bba00af0e'  # Password hash value
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|LANGUAGE|Interface language setting, 'zh' for Chinese, 'en' for English|
|WEB_HOST/WEB_PORT|IP address and port where quart server runs|
|API_SITE|Configure the address of the connected quart server (format `http/https://{HOST}:{PORT}`) (should match the corresponding configuration in `server_api.py`)|
|SECURITY_PATH|Security path when quart server starts (should match the corresponding configuration in `server_api.py`)|
|SECURITY_PASSWORD_HASH|Security password SHA256 hash value when quart server starts (default password is fuck) (should match the corresponding configuration in `server_api.py`)|

Note: The password hash format in `server_web.py` is different from `server_api.py`. The former uses SHA256 string format, while the latter uses bcrypt byte format.

Modify as needed.

## client.go
Go back one level and open `client.go`
Find code similar to this:
```go
const (
	HOST string = "127.0.0.1"
	PORT int = 8765
	INSECURESKIPVERIFY bool = true
)
```
Below are explanations of these parameters:
|Parameter|Description|
|---|---|
|HOST/PORT|Address and port of the server to connect to (should match HOST/PORT configuration in `server_api.py`)|
|INSECURESKIPVERIFY|Whether to skip certificate verification, set to true in development environment, recommend false in production environment|

When modifying configuration, please note:
1. Ensure client's HOST/PORT matches server-side configuration
2. In production environment, it's recommended to set INSECURESKIPVERIFY to false to enable certificate verification

## WSS/WS Protocol Switching Instructions

**Important Security Reminder**: This project uses WSS (WebSocket Secure) protocol by default, providing encrypted communication.

### Complete WSS Downgrade Solution (Modify Both Client and Server)

**Warning**: Downgrading to WS protocol loses encryption protection and should only be used in controlled testing environments!

#### Step 1: Modify Server [server_api.py](file:///home/zhaobokai/vscode/RATFF/code/server_api.py)

Find the SSL configuration section in the server main loop function:
```python
# Server main loop
async def server_loop():
    logging.info(lp.g("initializing_server"))
    logging.info(f"{lp.g('certificate_path')}: {SSL_CERT}，{lp.g('key_path')}: {SSL_KEY}")
    
    try:
        ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
        ssl_context.load_cert_chain(SSL_CERT, SSL_KEY)
        logging.info(lp.g("ssl_certificate_loaded_successfully"))
    except FileNotFoundError:
        logging.error(lp.g("certificate_file_not_found"))
        exit(1)
    except Exception as e:
        logging.error(f"{lp.g('ssl_loading_failed')}: {str(e)}")
        exit(1)

    logging.info(f"{lp.g('starting_server')}: {HOST}:{PORT}")
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
    
    logging.info(f"{lp.g('starting_server')}: {HOST}:{PORT}")
    try:
        # Note: ssl parameter set to None
        async with websockets.serve(handle_client, HOST, PORT, ssl=ssl_context):
            logging.info(lp.g("server_started_successfully"))
```

#### Step 2: Modify Client [client.go](file:///home/zhaobokai/vscode/RATFF/code/client.go)

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

#### Step 3: Recompile and Deploy

1. **Recompile Server**:
   ```bash
   # server_api.py is a Python script, no compilation needed, run directly
   ```

2. **Recompile Client**:
   ```bash
   cd ..  # Back to project root directory
   go build -o client client.go
   ```

### Security Recommendations:
1. **Strongly recommend using WSS protocol in production environment**
2. Only consider using WS protocol in local testing or controlled environments
3. If using WS protocol, ensure other network-level security measures are in place (such as VPN, intranet isolation, etc.)

If you want to disable WSS encryption and use WS protocol, please follow the above complete solution strictly.