-- 口述历史采集工具数据库初始化脚本
-- 由 MySQL 官方镜像 docker-entrypoint-initdb.d 在首次启动时执行
CREATE DATABASE IF NOT EXISTS oralhistory_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建应用账号（若镜像环境变量已创建则忽略）
CREATE USER IF NOT EXISTS 'oralhistory_user'@'%' IDENTIFIED BY 'oralhistory_pwd';
GRANT ALL PRIVILEGES ON oralhistory_db.* TO 'oralhistory_user'@'%';
FLUSH PRIVILEGES;
