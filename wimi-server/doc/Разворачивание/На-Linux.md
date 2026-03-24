## Разворачивание WiMi на Linux

Копирование deb-пакета:

```bash
$ scp -r WiMi-0.1.7.deb username1@<ip адрес>:/home/username2
```

Установка пакета `WiMi`:

```bash
$ sudo apt-get update
$ sudo apt-get upgrade
$ sudo dpkg -i WiMi-0.1.7.deb || sudo apt-get -f install
```

Отредактируем порт `WiMi`:

```bash
sudo nano /usr/local/bin/WiMi/MivREST.ini
```

```json
[listener]
port=8092
minThreads=10
maxThreads=100
```

Проверим работоспособность:

```bash
$ cd /usr/local/bin/WiMi/
/usr/local/bin/WiMi$ ./WiMi.run 
```

После установки `WiMi` создаем файлы в `/usr/local/bin/WiMi/`.

```bash
/usr/local/bin/WiMi$ sudo nano environment
```

```bash
LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/usr/local/bin/WiMi/libs"
```

```bash
/usr/local/bin/WiMi$ sudo nano WiMi.service
```

```bash
[Unit]
Description=WiMi
#After=syslog.target
#After=network.target
#After=nginx.service
#After=mysql.service
#Requires=mysql.service
#Wants=redis.service

[Service]
Type=forking
#PIDFile=/usr/local/bin/WiMI/WiMiservice.pid
WorkingDirectory=/usr/local/bin/WiMi

#Environment=LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/usr/local/bin/WiMi/libs"
#Environment=LD_PRELOAD="$LD_PRELOAD:/usr/local/bin/WiMi/libs/libjemalloc.so.2"
EnvironmentFile=/usr/local/bin/WiMi/environment


Environment=RACK_ENV=production

OOMScoreAdjust=-1000

ExecStart=/bin/bash -c ./WiMi start
ExecStop=/bin/bash -c ./WiMi stop
ExecReload=/bin/bash -c ./WiMi restart

TimeoutSec=5

[Install]
WantedBy=multi-user.target
```

Вводим следующие команды:

```bash
/usr/local/bin/WiMi$ sudo ln /usr/local/bin/WiMi/WiMi.service /etc/systemd/system/
```

Запускаем сервис:

```bash
$ sudo service WiMi start
```

Проверяем сервис:

```bash
$ sudo service WiMi status
```

Добавляем в автозапуск:

```bash
$ sudo systemctl enable WiMi
```
