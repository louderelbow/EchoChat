-- EchoChat go-zero 版数据库结构（与 Gin 原版 6 张表对齐）
SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `user_info` (
  `id`             bigint       NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uuid`           char(20)     NOT NULL COMMENT '用户唯一id',
  `nickname`       varchar(20)  NOT NULL COMMENT '昵称',
  `telephone`      varchar(64)  NOT NULL COMMENT '账号（邮箱）',
  `email`          char(30)     DEFAULT NULL COMMENT '邮箱',
  `avatar`         char(255)    NOT NULL DEFAULT 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png' COMMENT '头像',
  `gender`         tinyint      DEFAULT 0 COMMENT '性别，0.未知，1.男，2.女',
  `signature`      varchar(100) DEFAULT NULL COMMENT '个性签名',
  `password`       varchar(60)  NOT NULL COMMENT '密码（bcrypt 哈希）',
  `birthday`       char(8)      DEFAULT NULL COMMENT '生日',
  `created_at`     datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at`     datetime     DEFAULT NULL COMMENT '删除时间',
  `last_online_at` datetime     DEFAULT NULL COMMENT '上次登录时间',
  `last_offline_at` datetime    DEFAULT NULL COMMENT '最近离线时间',
  `is_admin`       tinyint      NOT NULL DEFAULT 0 COMMENT '是否是管理员，0.不是，1.是',
  `status`         tinyint      NOT NULL DEFAULT 0 COMMENT '状态，0.正常，1.禁用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_uuid` (`uuid`),
  KEY `idx_user_telephone` (`telephone`),
  KEY `idx_user_created_at` (`created_at`),
  KEY `idx_user_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `group_info` (
  `id`         bigint       NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uuid`       char(20)     NOT NULL COMMENT '群组唯一id',
  `name`       varchar(20)  NOT NULL COMMENT '群名称',
  `notice`     varchar(500) DEFAULT NULL COMMENT '群公告',
  `members`    json         DEFAULT NULL COMMENT '群组成员',
  `member_cnt` int          NOT NULL DEFAULT 1 COMMENT '群人数',
  `owner_id`   char(20)     NOT NULL COMMENT '群主uuid',
  `add_mode`   tinyint      NOT NULL DEFAULT 0 COMMENT '加群方式，0.直接，1.审核',
  `avatar`     char(255)    NOT NULL DEFAULT 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png' COMMENT '头像',
  `status`     tinyint      NOT NULL DEFAULT 0 COMMENT '状态，0.正常，1.禁用，2.解散',
  `created_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime     DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_uuid` (`uuid`),
  KEY `idx_group_created_at` (`created_at`),
  KEY `idx_group_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `user_contact` (
  `id`           bigint   NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `user_id`      char(20) NOT NULL COMMENT '用户唯一id',
  `contact_id`   char(20) NOT NULL COMMENT '对应联系id',
  `contact_type` tinyint  NOT NULL COMMENT '联系类型，0.用户，1.群聊',
  `status`       tinyint  NOT NULL COMMENT '联系状态',
  `created_at`   datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_at`    datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_uc_user_id` (`user_id`),
  KEY `idx_uc_contact_id` (`contact_id`),
  KEY `idx_uc_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `session` (
  `id`              bigint       NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uuid`            char(20)     NOT NULL COMMENT '会话uuid',
  `send_id`         char(20)     NOT NULL COMMENT '创建会话人id',
  `receive_id`      char(20)     NOT NULL COMMENT '接受会话人id',
  `receive_name`    varchar(20)  NOT NULL COMMENT '名称',
  `avatar`          char(255)    NOT NULL DEFAULT 'default_avatar.png' COMMENT '头像',
  `last_message`    text         DEFAULT NULL COMMENT '最新的消息',
  `last_message_at` datetime     DEFAULT NULL COMMENT '最近接收时间',
  `created_at`      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted_at`      datetime     DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_session_uuid` (`uuid`),
  KEY `idx_session_send_id` (`send_id`),
  KEY `idx_session_receive_id` (`receive_id`),
  KEY `idx_session_created_at` (`created_at`),
  KEY `idx_session_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `contact_apply` (
  `id`            bigint       NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uuid`          char(20)     NOT NULL COMMENT '申请id',
  `user_id`       char(20)     NOT NULL COMMENT '申请人id',
  `contact_id`    char(20)     NOT NULL COMMENT '被申请id',
  `contact_type`  tinyint      NOT NULL COMMENT '被申请类型，0.用户，1.群聊',
  `status`        tinyint      NOT NULL COMMENT '申请状态，0.申请中，1.通过，2.拒绝，3.拉黑',
  `message`       varchar(100) DEFAULT NULL COMMENT '申请信息',
  `last_apply_at` datetime     NOT NULL COMMENT '最后申请时间',
  `deleted_at`    datetime     DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_apply_uuid` (`uuid`),
  KEY `idx_apply_user_id` (`user_id`),
  KEY `idx_apply_contact_id` (`contact_id`),
  KEY `idx_apply_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `message` (
  `id`          bigint       NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uuid`        char(20)     NOT NULL COMMENT '消息uuid',
  `session_id`  char(20)     NOT NULL COMMENT '会话uuid',
  `type`        tinyint      NOT NULL COMMENT '消息类型，0.文本，1.语音，2.文件，3.通话',
  `content`     text         DEFAULT NULL COMMENT '消息内容',
  `url`         char(255)    DEFAULT NULL COMMENT '消息url',
  `send_id`     char(20)     NOT NULL COMMENT '发送者uuid',
  `send_name`   varchar(20)  NOT NULL COMMENT '发送者昵称',
  `send_avatar` varchar(255) NOT NULL COMMENT '发送者头像',
  `receive_id`  char(20)     NOT NULL COMMENT '接受者uuid',
  `file_type`   char(10)     DEFAULT NULL COMMENT '文件类型',
  `file_name`   varchar(50)  DEFAULT NULL COMMENT '文件名',
  `file_size`   char(20)     DEFAULT NULL COMMENT '文件大小',
  `status`      tinyint      NOT NULL COMMENT '状态，0.未发送，1.已发送',
  `created_at`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `send_at`     datetime     DEFAULT NULL COMMENT '发送时间',
  `av_data`     text         DEFAULT NULL COMMENT '通话传递数据',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_uuid` (`uuid`),
  KEY `idx_message_session_id` (`session_id`),
  KEY `idx_message_send_id` (`send_id`),
  KEY `idx_message_receive_id` (`receive_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
