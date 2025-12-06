# 启动依赖服务
sudo systemctl start mysql
sudo systemctl start redis-server
sudo systemctl start rabbitmq-server

# 方式1：后台运行（推荐）
nohup go run ./cmd/web > web.log 2>&1 &
nohup go run ./cmd/admin > admin.log 2>&1 &

# 查看日志
tail -f web.log
tail -f admin.log

# 方式2：前台运行（调试用，关闭终端会停止服务）
# go run ./cmd/web
# go run ./cmd/admin

# 检查服务状态
bash check_service.sh

# 访问地址
# Web: http://服务器IP:8080
# Admin: http://服务器IP:8081


//12-4-23-33
bash deploy.sh 
//
这个脚本会自动做这些事：
安装并启动 Go、MySQL、Redis、RabbitMQ
创建数据库 goseckill 和用户 goseckill / goseckill123
执行 go build 编译出 bin/web、bin/admin、bin/seckill-worker
创建并启用 3 个 systemd 服务：
goseckill-web（监听 8080）
goseckill-admin（监听 8081）
goseckill-worker（做异步秒杀任务）
//
sudo systemctl status goseckill-web
sudo systemctl status goseckill-admin
sudo systemctl status goseckill-worker