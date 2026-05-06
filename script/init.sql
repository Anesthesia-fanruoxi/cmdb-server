-- 创建数据库
DROP DATABASE IF EXISTS `cmdb`;
CREATE DATABASE `cmdb` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 表结构：服务器资产表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`assets_server`;
CREATE TABLE `cmdb`.`assets_server` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `project` varchar(50) NOT NULL COMMENT '项目',
    `instance_id` varchar(100) NOT NULL COMMENT '实例ID',
    `instance_name` varchar(100) NOT NULL COMMENT '实例名称',
    `hostname` varchar(100) NOT NULL COMMENT '主机名',
    `os` varchar(100) NULL DEFAULT NULL COMMENT '操作系统',
    `status` varchar(50) NULL DEFAULT NULL COMMENT '状态',
    `region` varchar(50) NULL DEFAULT NULL COMMENT '地域',
    `zone` varchar(50) NULL DEFAULT NULL COMMENT '可用区',
    `public_ip` varchar(50) NULL DEFAULT NULL COMMENT '公网IP',
    `private_ip` varchar(50) NULL DEFAULT NULL COMMENT '私网IP',
    `elastic_ip` varchar(50) NULL DEFAULT NULL COMMENT '弹性IP',
    `cpu` int NULL DEFAULT NULL COMMENT 'CPU核数',
    `memory` varchar(50) NULL DEFAULT NULL COMMENT '内存',
    `pay_type` varchar(50) NULL DEFAULT NULL COMMENT '付费类型',
    `network_type` varchar(50) NULL DEFAULT NULL COMMENT '网络类型',
    `bandwidth` int NULL DEFAULT NULL COMMENT '带宽',
    `created_time` datetime NULL DEFAULT NULL COMMENT '创建时间',
    `expire_time` datetime NULL DEFAULT NULL COMMENT '到期时间',
    `instance_type` varchar(50) NULL DEFAULT NULL COMMENT '实例规格',
    `instance_family` varchar(50) NULL DEFAULT NULL COMMENT '实例规格族',
    `image_id` varchar(100) NULL DEFAULT NULL COMMENT '镜像ID',
    `resource_group_id` varchar(100) NULL DEFAULT NULL COMMENT '资源组ID',
    `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_instance_id`(`instance_id`) ,
    INDEX `idx_project`(`project`) ,
    INDEX `idx_hostname`(`hostname`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '服务器资产表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：部门表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_departments`;
CREATE TABLE `cmdb`.`sys_departments` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '部门ID',
    `created_at` datetime(3) NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) NULL DEFAULT NULL COMMENT '更新时间',
    `name` varchar(64) NOT NULL COMMENT '部门名称',
    `code` varchar(64) NOT NULL COMMENT '部门编码',
    `parent_id` bigint NULL DEFAULT NULL COMMENT '父部门ID',
    `description` varchar(255) NULL DEFAULT NULL COMMENT '部门描述',
    `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_name`(`name`) ,
    UNIQUE INDEX `uk_code`(`code`) ,
    INDEX `idx_parent_id`(`parent_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '部门表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：部门-项目关联表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_dept_projects`;
CREATE TABLE `cmdb`.`sys_dept_projects` (
    `dept_id` bigint NOT NULL COMMENT '部门ID',
    `project` varchar(64) NOT NULL COMMENT '项目标识',
    PRIMARY KEY (`dept_id`, `project`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '部门-项目关联表' ROW_FORMAT = DYNAMIC;

-- ------------------------
-- 表结构：用户-项目关联表
-- ------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_user_projects`;
CREATE TABLE IF NOT EXISTS `cmdb`.`sys_user_projects` (
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `project` VARCHAR(64) NOT NULL COMMENT '项目标识',
  `menu_id`  BIGINT NOT NULL COMMENT '关联菜单ID',
  PRIMARY KEY (`user_id`, `project`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_project` (`project`),
  INDEX `idx_menu_id` (`menu_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户项目关联表' ROW_FORMAT = DYNAMIC;
-- ----------------------------
-- 表结构：字典记录表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_dict_records`;
CREATE TABLE `cmdb`.`sys_dict_records` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `dict_name` varchar(64) NOT NULL COMMENT '字典名称',
    `target_table` varchar(64) NOT NULL COMMENT '目标表名',
    `key_name` varchar(64) NOT NULL COMMENT '键字段名',
    `value_name` varchar(64) NOT NULL COMMENT '值字段名',
    `created_by` bigint NOT NULL COMMENT '创建人',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_dict_name`(`dict_name`) ,
    UNIQUE INDEX `uk_target_table`(`target_table`)
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '字典记录表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：文档分类字典
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_category_dict`;
CREATE TABLE `cmdb`.`sys_category_dict` (
    `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `category` varchar(64) NOT NULL COMMENT '键',
    `category_name` varchar(255) NOT NULL COMMENT '值',
    `color` varchar(20) DEFAULT '' COMMENT '字典颜色',
    `created_by` bigint UNSIGNED NOT NULL COMMENT '创建人ID',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_key`(`category`)
) ENGINE = InnoDB AUTO_INCREMENT = 10 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '文档分类字典' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：菜单表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_menus`;
CREATE TABLE `cmdb`.`sys_menus` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
    `parent_id` bigint NULL DEFAULT NULL COMMENT '父菜单ID',
    `name` varchar(50) NOT NULL COMMENT '菜单名称',
    `path` varchar(200) NULL DEFAULT '' COMMENT '路由地址',
    `component` varchar(255) NULL DEFAULT '' COMMENT '组件路径',
    `sort` int NULL DEFAULT 0 COMMENT '显示顺序',
    `icon` varchar(100) NULL DEFAULT '' COMMENT '菜单图标',
    `permission` varchar(100) NULL DEFAULT '' COMMENT '权限标识',
    `is_visible` tinyint(1) NULL DEFAULT 1 COMMENT '是否可见',
    `is_enabled` tinyint(1) NULL DEFAULT 1 COMMENT '是否启用',
    `created_at` datetime(3) NULL DEFAULT NULL,
    `updated_at` datetime(3) NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB AUTO_INCREMENT = 131 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '菜单表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- 表结构：项目字典表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_project_dict`;
CREATE TABLE `cmdb`.`sys_project_dict` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `project` varchar(64) NOT NULL COMMENT '键',
    `project_name` varchar(255) NOT NULL COMMENT '值',
    `agent_url` varchar(255) DEFAULT '' COMMENT 'Agent代理地址',
    `backen_domain` varchar(255) DEFAULT '' COMMENT '后台管理地址',
    `api_domain` varchar(255) DEFAULT '' COMMENT '三方调用地址',
    `git_vue` varchar(255) DEFAULT '' COMMENT '前端git仓库',
    `git_backend` varchar(255) DEFAULT '' COMMENT '后端git仓库',
    `alter_feishu` varchar(255) DEFAULT '' COMMENT '飞书告警地址',
    `update_feishu` varchar(255) DEFAULT '' COMMENT '发版通知地址(开发组内)',
    `notify_feishu` varchar(255) DEFAULT '' COMMENT '通知地址（所有人）',
    `description` varchar(500) DEFAULT '' COMMENT '项目描述',
    `backend_tool` varchar(255) DEFAULT '' COMMENT '后端工具（java17等）',
    `frontend_tool` varchar(255) DEFAULT '' COMMENT '前端工具（node14等）',
    `enable_skywalking` boolean DEFAULT false COMMENT '是否启用skywalking',
    `color` varchar(20) DEFAULT '' COMMENT '字典颜色',
    `created_by` bigint NOT NULL COMMENT '创建人ID',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_project`(`project`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '项目字典' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：agent字典表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_agent_dict`;
CREATE TABLE `cmdb`.`sys_agent_dict` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `project` varchar(64) NOT NULL COMMENT '键',
    `agent_url` varchar(255) NOT NULL COMMENT '值',
    `feishu_url` varchar(255) NOT NULL COMMENT '飞书URL',
    `color` varchar(20) DEFAULT '' COMMENT '字典颜色',
    `created_by` bigint NOT NULL COMMENT '创建人ID',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_project`(`project`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'agent字典表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：视图分类字典表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_view_dict`;
CREATE TABLE `cmdb`.`sys_view_dict` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `category` varchar(64) NOT NULL COMMENT '键',
    `category_name` varchar(255) NOT NULL COMMENT '值',
    `color` varchar(20) DEFAULT '' COMMENT '字典颜色',
    `created_by` bigint NOT NULL COMMENT '创建人ID',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_category`(`category`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '视图分类字典' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：角色-菜单关联表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_role_menus`;
CREATE TABLE `cmdb`.`sys_role_menus` (
    `role_id` bigint NOT NULL COMMENT '角色ID',
    `menu_id` bigint NOT NULL COMMENT '菜单ID',
    `has_menu` tinyint(1) NOT NULL DEFAULT 0 COMMENT '菜单权限：0代表不可见，1代表可见',
    `has_read` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'r权限：0代表没有权限，1代表有权限',
    `has_write` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'w权限：0代表没有权限，1代表有权限',
    `created_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`role_id`, `menu_id`) ,
    INDEX `idx_role_id`(`role_id`) ,
    INDEX `idx_menu_id`(`menu_id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '角色-菜单关联表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：角色表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_roles`;
CREATE TABLE `cmdb`.`sys_roles` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `name` varchar(32) NOT NULL COMMENT '角色名称',
    `code` varchar(32) NOT NULL COMMENT '角色代码',
    `description` varchar(128) NULL DEFAULT NULL COMMENT '角色备注',
    `created_at` datetime(3) NULL DEFAULT NULL,
    `updated_at` datetime(3) NULL DEFAULT NULL,
    `level` bigint NOT NULL COMMENT '角色级别',
    PRIMARY KEY (`id`) ,
    UNIQUE INDEX `uk_name`(`name`) ,
    UNIQUE INDEX `uk_code`(`code`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '角色表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：系统设置表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_settings`;
CREATE TABLE `cmdb`.`sys_settings` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `system_name` varchar(64) NOT NULL COMMENT '系统名称',
    `system_url` varchar(64) COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '平台地址',
    `fs_webhook_url` varchar(255) COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '飞书webhook地址',
    `system_short_name` varchar(32) NOT NULL COMMENT '系统简称',
    `password_min_length` int NOT NULL DEFAULT 8 COMMENT '密码最小长度',
    `password_max_length` int NOT NULL DEFAULT 32 COMMENT '密码最大长度',
    `password_need_number` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否需要数字',
    `password_need_letter` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否需要字母',
    `password_need_case` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否需要大小写',
    `password_need_special` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要特殊字符',
    `login_protection` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否开启登录保护',
    `login_fail_count` int NOT NULL DEFAULT 5 COMMENT '登录失败次数',
    `login_lock_time` int NOT NULL DEFAULT 30 COMMENT '锁定时间(分钟)',
    `default_password` varchar(32) NOT NULL DEFAULT 'Cmdb@123' COMMENT '默认密码',
    `token_expire_time` int NOT NULL DEFAULT 24 COMMENT 'token过期时间(小时)',
    `created_at` datetime(3) NULL DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) NULL DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '系统设置表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：用户表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_users`;
CREATE TABLE `cmdb`.`sys_users` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    `user_name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名',
    `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '密码',
    `nick_name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '昵称',
    `email` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '邮箱',
    `phone` varchar(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '手机号',
    `role_id` bigint NOT NULL DEFAULT '2' COMMENT '角色ID',
    `dept_id` bigint NOT NULL COMMENT '部门ID',
    `online_assets` tinyint(1) NOT NULL COMMENT '生产资源是否忽略权限，0=不忽略，1=忽略',
    `test_assets` tinyint(1) NOT NULL COMMENT '测试资源是否忽略权限，0=不忽略，1=忽略',
    `is_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用：0=禁用，1=启用',
    `allow_password_login` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否允许密码登录：0=禁用，1=启用',
    `is_default_pass` tinyint(1) NOT NULL DEFAULT '1' COMMENT '默认密码是否已修改',
    `otp_secret` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '双因子验证密钥',
    `otp_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '双因子验证是否启用：0=禁用，1=启用',
    `otp_backup_codes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '备用恢复码，JSON格式存储',
    `otp_setup_completed` tinyint(1) NOT NULL DEFAULT '0' COMMENT '双因子设置是否完成：0=未完成，1=已完成',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_user_name`(`user_name`),
    INDEX `idx_role_id`(`role_id`),
    INDEX `idx_dept_id`(`dept_id`),
    INDEX `idx_otp_enabled`(`otp_enabled`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：文档历史版本表     
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_document_histories`;
CREATE TABLE `cmdb`.`kb_document_histories` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '历史记录ID',
    `doc_id` bigint NOT NULL COMMENT '关联的文档ID',
    `title` varchar(255) NOT NULL COMMENT '历史版本标题',
    `content` text NOT NULL COMMENT '历史版本内容',
    `category` varchar(64) NOT NULL COMMENT '关联的分类',
    `project` varchar(64) NOT NULL COMMENT '所属项目代码',
    `version` int NOT NULL COMMENT '版本号',
    `creator` varchar(64) NOT NULL COMMENT '创建者名称',
    `user_name` varchar(64) NOT NULL COMMENT '更新者名称',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_doc_version`(`doc_id`, `version`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='文档历史版本表';

-- ----------------------------
-- 表结构：文档回收站表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_document_recycle`;
CREATE TABLE `cmdb`.`kb_document_recycle` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '回收记录ID',
    `doc_id` bigint NOT NULL COMMENT '文档ID',
    `title` varchar(255) NOT NULL COMMENT '文档标题',
    `content` text NOT NULL COMMENT '文档内容',
    `category` varchar(64) NOT NULL COMMENT '关联的分类',
    `project` varchar(64) NOT NULL COMMENT '所属项目代码',
    `creator` varchar(64) NOT NULL COMMENT '创建者名称',
    `deleter_name` varchar(64) NOT NULL COMMENT '删除者名称',
    `deleted_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_doc_id`(`doc_id`),
    INDEX `idx_project`(`project`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='文档回收站表';

-- ----------------------------
-- 表结构：知识库文档表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_documents`;
CREATE TABLE `cmdb`.`kb_documents` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '文档ID',
    `title` varchar(255) NOT NULL COMMENT '文档标题',
    `content` text NOT NULL COMMENT '文档内容(Markdown格式)',
    `project` varchar(64) NOT NULL COMMENT '关联的项目代码',
    `category` varchar(64) NOT NULL COMMENT '关联的分类',
    `creator` varchar(64) NOT NULL COMMENT '创建者名称',
    `user_name` varchar(64) NOT NULL COMMENT '最后更新者名称',
    `version` int NOT NULL DEFAULT 1 COMMENT '当前版本号',
    `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态(1:正常,0:回收站)',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_project`(`project`),
    INDEX `idx_creator`(`creator`),
    INDEX `idx_status`(`status`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='知识库文档表';

-- ----------------------------
-- 个人文档表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_personal`;
CREATE TABLE IF NOT EXISTS `cmdb`.`kb_personal` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `title` varchar(200) NOT NULL COMMENT '文档标题',
    `category` varchar(64) NOT NULL COMMENT '分类',
    `content` text COMMENT '文档内容',
    `user_name` varchar(64) NOT NULL COMMENT '最后更新者名称',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='个人文档表';

-- ----------------------------
-- 个人文档回收站表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_personal_recycle`;
CREATE TABLE IF NOT EXISTS `cmdb`.`kb_personal_recycle` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `doc_id` bigint unsigned NOT NULL COMMENT '文档ID',
    `title` varchar(200) NOT NULL COMMENT '文档标题',
    `category` varchar(64) NOT NULL COMMENT '分类',
    `content` text COMMENT '文档内容',
    `user_name` varchar(64) NOT NULL COMMENT '删除者名称',
    `deleted_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '删除时间',
    PRIMARY KEY (`id`),
    INDEX `idx_doc_id`(`doc_id`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='个人文档回收站表';

-- ----------------------------
-- 公开文档表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_public`;
CREATE TABLE `cmdb`.`kb_public` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `title` varchar(200) NOT NULL COMMENT '文档标题',
    `content` text NOT NULL COMMENT '文档内容',
    `category` varchar(50) NOT NULL COMMENT '文档分类',
    `project` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属项目'
    `creator` varchar(50) NOT NULL COMMENT '创建者名称',
    `share_code` varchar(50) DEFAULT NULL COMMENT '分享code',
    `logo` VARCHAR(500) DEFAULT '' COMMENT '项目Logo地址'
    `user_name` varchar(50) NOT NULL COMMENT '最后修改者名称',
    `version` bigint NOT NULL DEFAULT 1 COMMENT '版本号',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_category` (`category`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='公开文档表';
-- ----------------------------
-- 公开文档分享表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_public_share`;
CREATE TABLE `cmdb`.`kb_public_share` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `doc_id` bigint unsigned NOT NULL COMMENT '文档ID',
    `share_code` varchar(32) NOT NULL COMMENT '分享码',
    `share_url` varchar(255) NOT NULL COMMENT '分享链接',
    `creator` varchar(50) NOT NULL COMMENT '分享者名称',
    `project` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属项目'
    `expired_at` datetime NOT NULL COMMENT '过期时间',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_share_code` (`share_code`),
    KEY `idx_doc_id` (`doc_id`),
    KEY `idx_expired_at` (`expired_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='公开文档分享表';

-- ----------------------------
-- 公开文档历史表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_public_history`;
CREATE TABLE `cmdb`.`kb_public_history` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `doc_id` bigint unsigned NOT NULL COMMENT '文档ID',
    `title` varchar(200) NOT NULL COMMENT '文档标题',
    `content` text NOT NULL COMMENT '文档内容',
    `category` varchar(50) NOT NULL COMMENT '文档分类',
    `project` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属项目',
    `version` bigint NOT NULL COMMENT '版本号',
    `creator` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '创建者名称',
    `user_name` varchar(50) NOT NULL COMMENT '操作者名称',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_doc_id` (`doc_id`),
    KEY `idx_version` (`version`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='公开文档历史表';

-- ----------------------------
-- 公开文档回收站表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`kb_public_recycle`;
CREATE TABLE `cmdb`.`kb_public_recycle` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `doc_id` bigint unsigned NOT NULL COMMENT '文档ID',
    `title` varchar(200) NOT NULL COMMENT '文档标题',
    `content` text NOT NULL COMMENT '文档内容',
    `category` varchar(50) NOT NULL COMMENT '文档分类',
    `project` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属项目'
    `creator` varchar(50) NOT NULL COMMENT '创建者名称',
    `user_name` varchar(50) NOT NULL COMMENT '删除者名称',
    `deleted_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '删除时间',
    PRIMARY KEY (`id`),
    KEY `idx_doc_id` (`doc_id`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='公开文档回收站表';

-- ----------------------------
-- 任务定义表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`job_info`;
CREATE TABLE IF NOT EXISTS `cmdb`.`job_info` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` varchar(100) NOT NULL COMMENT '任务名称',
    `task_key` varchar(100) NOT NULL COMMENT '任务标识符',
    `cron` varchar(100) NOT NULL COMMENT 'cron表达式',
    `description` varchar(255) DEFAULT NULL COMMENT '任务描述',
    `script_type` VARCHAR(50) NOT NULL DEFAULT 'shell' COMMENT '脚本类型(shell/python/etc)'
    `script_content` LONGTEXT NOT NULL COMMENT '脚本内容'
    `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态 0:停止 1:运行',
    `creator` varchar(50) NOT NULL COMMENT '创建人',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_task_key` (`task_key`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='任务定义表';

-- ----------------------------
-- 任务-项目关联表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`job_projects`;
CREATE TABLE IF NOT EXISTS `cmdb`.`job_projects` (
    `job_id` bigint NOT NULL COMMENT '任务ID',
    `project` varchar(64) NOT NULL COMMENT '项目标识',
    `script_params` JSON COMMENT '脚本参数（JSON格式，用于模板渲染）'
    PRIMARY KEY (`job_id`,`project`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='任务-项目关联表';

-- ----------------------------
-- 任务执行记录表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`job_execution`;
CREATE TABLE IF NOT EXISTS `cmdb`.`job_execution` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `job_id` bigint NOT NULL COMMENT '任务ID',
    `project` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '项目标识',
    `task_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务名称',
    `task_key` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务标识符',
    `exec_log` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '执行日志',
    `exec_status` tinyint NOT NULL DEFAULT '0' COMMENT '执行状态 0:执行中 1:成功 2:失败',
    `result` text COMMENT '执行结果',
    `error_msg` text COMMENT '错误信息',
    `start_time` datetime NOT NULL COMMENT '开始时间',
    `end_time` datetime DEFAULT NULL COMMENT '结束时间',
    `duration` int DEFAULT '0' COMMENT '执行时长(秒)',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_job_project` (`job_id`,`project`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='任务执行记录表';


-- ----------------------------
-- 操作日志表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`audit_operation_log`;
CREATE TABLE `cmdb`.`audit_operation_log` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `user_name` varchar(50) NOT NULL COMMENT '用户名',
    `module` varchar(50) NOT NULL COMMENT '操作模块',
    `action` varchar(50) NOT NULL COMMENT '操作类型',
    `method` varchar(10) NOT NULL COMMENT '请求方法',
    `url` varchar(255) NOT NULL COMMENT '请求URL',
    `ip` varchar(50) NOT NULL COMMENT '请求IP',
    `city` varchar(50) NOT NULL COMMENT '城市',
    `country` varchar(50) NOT NULL COMMENT '国家',
    `region` varchar(50) NOT NULL COMMENT '省份',
    `request` text COMMENT '请求参数',
    `response` text COMMENT '响应结果',
    `status` int NOT NULL COMMENT '响应状态码',
    `error_msg` text COMMENT '错误信息',
    `duration` int NOT NULL COMMENT '执行时长(ms)',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='操作日志表';

-- ----------------------------
-- 登录日志表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`audit_login_log`;
CREATE TABLE `cmdb`.`audit_login_log` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `user_name` varchar(50) NOT NULL COMMENT '用户名',
    `ip` varchar(50) NOT NULL COMMENT '登录IP',
    `city` varchar(50) NOT NULL COMMENT '城市',
    `country` varchar(50) NOT NULL COMMENT '国家',
    `region` varchar(50) NOT NULL COMMENT '省份',
    `browser` varchar(50) DEFAULT NULL COMMENT '浏览器类型',
    `os` varchar(50) DEFAULT NULL COMMENT '操作系统',
    `status` tinyint NOT NULL COMMENT '登录状态(0:失败,1:成功)',
    `msg` varchar(255) DEFAULT NULL COMMENT '提示消息',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='登录日志表';


-- ----------------------------
-- 视图表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`es_view`;
CREATE TABLE IF NOT EXISTS `cmdb`.`es_view` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` varchar(100) NOT NULL COMMENT '视图名称',
    `project` varchar(50) NOT NULL COMMENT '所属项目',
    `category` varchar(50) NOT NULL COMMENT '视图分类',
    `index_pattern` varchar(255) NOT NULL COMMENT '索引模式',
    `time_field` varchar(50) NOT NULL COMMENT '时间字段名称',
    `time_format` varchar(60) NOT NULL COMMENT '时间格式',
    `description` varchar(255) DEFAULT NULL COMMENT '视图描述',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_project` (`project`),
    UNIQUE KEY `uk_project_name` (`project`, `name`) COMMENT '同一项目下视图名称唯一'
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='视图表';

-- ----------------------------
-- Elasticsearch审计日志表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_es_log`;
CREATE TABLE IF NOT EXISTS `cmdb`.`sys_es_log` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `query_id` VARCHAR(64) COMMENT '查询ID',
    `user_id` VARCHAR(64) COMMENT '用户ID',
    `user_name` VARCHAR(100) COMMENT '用户名',
    `client_ip` VARCHAR(64) COMMENT '客户端IP',
    `city` VARCHAR(64) COMMENT '城市',
    `country` VARCHAR(64) COMMENT '国家',
    `region` VARCHAR(64) COMMENT '省份',
    `project` VARCHAR(255) COMMENT '项目名称',
    `view_name` VARCHAR(255) COMMENT '视图名称',
    `operation` VARCHAR(255) COMMENT '操作类型',
    `page` INT COMMENT '页码',
    `q_index_pattern` VARCHAR(255) COMMENT '查询索引模式',
    `q_time_field` VARCHAR(255) COMMENT '查询时间字段',
    `q_start_time` VARCHAR(255) COMMENT '查询开始时间',
    `q_end_time` VARCHAR(255) COMMENT '查询结束时间',
    `q_keyword` VARCHAR(255) COMMENT '搜索关键词',
    `response_code` INT COMMENT '响应状态码',
    `query_time_ms` INT COMMENT '查询耗时(毫秒)',
    `doc_count` INT COMMENT '查询文档数量',
    `error_message` TEXT COMMENT '错误信息',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_project` (`project`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_view_name` (`view_name`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='Elasticsearch审计日志表';

-- ----------------------------
-- MySQL 审计表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`audit_es_log`;
CREATE TABLE IF NOT EXISTS `cmdb`.`audit_es_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` varchar(64) NOT NULL COMMENT '用户ID',
  `user_name` varchar(64) NOT NULL COMMENT '用户名',
  `client_ip` varchar(64) NOT NULL COMMENT '客户端IP',
  `city` varchar(64) NOT NULL COMMENT '城市',
  `country` varchar(64) NOT NULL COMMENT '国家',
  `region` varchar(64) NOT NULL COMMENT '省份',
  `db_name` varchar(64) NOT NULL COMMENT '数据库名称',
  `query_sql` text NOT NULL COMMENT '查询语句',
  `query_id` varchar(64) DEFAULT NULL COMMENT '查询ID',
  `operation` varchar(20) DEFAULT NULL COMMENT '操作类型',
  `page` int(11) DEFAULT 0 COMMENT '页码',
  `affected_rows` int(11) DEFAULT 0 COMMENT '影响行数',
  `execution_time` int(11) DEFAULT 0 COMMENT '执行时间(毫秒)',
  `status` varchar(16) NOT NULL DEFAULT 'success' COMMENT '执行状态: success/failed',
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_db_name` (`db_name`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_query_id` (`query_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='SQL审计日志表';

-- ----------------------------
-- 创建SQL审批表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sql_apply`;
CREATE TABLE IF NOT EXISTS `cmdb`.`sql_apply` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `project` VARCHAR(255) NOT NULL COMMENT '项目名称',
    `submitter_id` BIGINT NOT NULL COMMENT '提交人ID',
    `submitter_name` VARCHAR(50) NOT NULL COMMENT '提交人姓名',
    `submitter_status` INT NOT NULL DEFAULT 0 COMMENT '提交人状态: 0-不能处理，1-可以处理（可撤销）',
    `sql_content` TEXT NOT NULL COMMENT 'SQL内容',
    `database_name` VARCHAR(100) NOT NULL COMMENT '目标数据库名称',
    `execution_time` DATETIME COMMENT '预定执行时间，为空表示立即执行',
    `description` TEXT COMMENT '任务描述或说明',
    `url_key` VARCHAR(32) DEFAULT NULL COMMENT 'URL密钥，用于移动端审批链接',
    `rule_check_result` TEXT COMMENT '规则检查结果，JSON格式',
    `has_violation` TINYINT DEFAULT 0 COMMENT '是否有违规: 0-否，1-是',
    `has_blocker` TINYINT DEFAULT 0 COMMENT '是否有阻断: 0-否，1-是',
    `apply_status` INT NOT NULL DEFAULT 0 COMMENT '审批人状态: 0-不能处理，1-可以处理',
    `apply_id` BIGINT COMMENT '审批人ID',
    `apply_name` VARCHAR(50) COMMENT '审批人姓名',
    `apply_comment` TEXT COMMENT '审批意见',
    `apply_time` DATETIME COMMENT '审批时间',
    `apply_source` VARCHAR(50) COMMENT '审批来源: web/url',
    `executor_status` INT NOT NULL DEFAULT 0 COMMENT '执行人状态: 0-不能处理，1-可以处理',
    `executor_id` BIGINT COMMENT '执行人ID',
    `executor_name` VARCHAR(50) COMMENT '执行人姓名',
    `executor_source` VARCHAR(50) COMMENT '执行人来源: web/url',
    `execution_result` TEXT COMMENT '执行结果',
    `actual_execution_time` DATETIME COMMENT '实际执行时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `status` INT NOT NULL DEFAULT 0 COMMENT '整体状态: 0-待审批，1-待执行，2-执行中，3-执行成功，4-执行失败，5-已驳回，6-已撤销',
    INDEX `idx_submitter_id` (`submitter_id`),
    INDEX `idx_approver_id` (`apply_id`),
    INDEX `idx_executor_id` (`executor_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='SQL审批表';

-- ----------------------------
-- SQL审批规则表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sql_rule`;
CREATE TABLE IF NOT EXISTS `cmdb`.`sql_rule` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `rule_name` VARCHAR(100) NOT NULL COMMENT '规则名称',
    `rule_category` VARCHAR(20) NOT NULL COMMENT '规则分类：security(安全)/standard(规范)',
    `rule_type` VARCHAR(50) NOT NULL COMMENT '规则类型：field_check',
    `sql_tag` VARCHAR(20) NOT NULL COMMENT 'SQL标签：CREATE/ALTER/DROP/TRUNCATE/INSERT/UPDATE/DELETE/SELECT',
    `rule_description` TEXT COMMENT '规则描述',
    `is_enabled` TINYINT NOT NULL DEFAULT 1 COMMENT '是否启用：0-禁用，1-启用（启用即强制）',
    `rule_content` TEXT COMMENT '规则配置（JSON格式）',
    `error_message` TEXT COMMENT '规则匹配失败时的错误消息',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_sql_tag` (`sql_tag`),
    INDEX `idx_is_enabled` (`is_enabled`),
    INDEX `idx_rule_category` (`rule_category`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='SQL审批规则表（新版-基于解析结果）';

-- ----------------------------
-- SQL审批流程配置表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sql_apply_flow`;
CREATE TABLE IF NOT EXISTS `cmdb`.`sql_apply_flow` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `project_id` BIGINT COMMENT '项目名称',
    `apply_id` BIGINT COMMENT '审批人ID',
    `executor_id` BIGINT COMMENT '执行人ID',
    INDEX `idx_project_id` (`project_id`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='SQL审批流程配置表';

-- ----------------------------
-- 表结构：双因子验证日志表     
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_factor_log`;
CREATE TABLE `cmdb`.`sys_factor_log` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `action` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '操作类型：setup=设置，verify=验证，disable=禁用，use_backup=使用备用码',
    `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '状态：success=成功，failed=失败',
    `ip_address` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'IP地址',
    `user_agent` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '用户代理',
    `details` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '详细信息，JSON格式',
    PRIMARY KEY (`id`),
    INDEX `idx_user_id`(`user_id`),
    INDEX `idx_created_at`(`created_at`),
    INDEX `idx_action`(`action`),
    INDEX `idx_status`(`status`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '双因子验证日志表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：安全配置表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_security_config`;
CREATE TABLE `cmdb`.`sys_security_config` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `blacklist_count` int NOT NULL DEFAULT 5 COMMENT '黑名单最大错误次数',
    `blacklist_lock` int NOT NULL DEFAULT 30 COMMENT '黑名单锁定时间(分钟)',
    `ratelimit_cd` int NOT NULL DEFAULT 60 COMMENT '限流冷却时间(秒)',
    `ratelimit_ttl` int NOT NULL DEFAULT 3600 COMMENT 'Redis存储时间(秒)',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '安全配置表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：监控视图表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`monitor_views`;
CREATE TABLE `cmdb`.`monitor_views` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '视图ID',
    `view_name` varchar(100) NOT NULL COMMENT '视图名称',
    `query` text NOT NULL COMMENT '查询语句',
    `description` varchar(255) DEFAULT NULL COMMENT '视图备注',
    `chart_type` varchar(50) DEFAULT 'line' COMMENT '图表类型：line, bar, pie等',
    `category` varchar(50) DEFAULT 'default' COMMENT '视图分类：hardware, container, ssl等',
    `sort` int DEFAULT 0 COMMENT '排序',
    `standard` varchar(50) DEFAULT 'default' COMMENT '标准：custom, percent',
    `created_by` bigint NOT NULL COMMENT '创建人ID',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` datetime(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_view_name`(`view_name`),
    INDEX `idx_created_by`(`created_by`)
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '监控视图表' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- 表结构：SQL导出表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sql_export`;
CREATE TABLE `cmdb`.`sql_export` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `project` varchar(255) NOT NULL COMMENT '项目名称',
    `submitter_id` bigint NOT NULL COMMENT '提交人ID',
    `submitter_name` varchar(50) NOT NULL COMMENT '提交人姓名',
    `submitter_status` tinyint DEFAULT '0' COMMENT '提交人状态: 0-不能处理，1-可以处理',
    `submitter_remark` text NOT NULL COMMENT '申请说明',
    `apply_id` bigint DEFAULT NULL COMMENT '审批人ID',
    `apply_name` varchar(50) DEFAULT NULL COMMENT '审批人姓名',
    `apply_time` datetime DEFAULT NULL COMMENT '审批时间',
    `apply_source` varchar(50) DEFAULT NULL COMMENT '审批来源',
    `apply_status` tinyint DEFAULT '0' COMMENT '审批人状态：0-不能处理，1-可以处理',
    `sql_content` text NOT NULL COMMENT 'SQL内容',
    `database_name` varchar(100) NOT NULL COMMENT '目标数据库名称',
    `url_key` varchar(100) NOT NULL COMMENT 'URL键',
    `reviewer_id` bigint DEFAULT NULL COMMENT '审核人ID',
    `reviewer_name` varchar(50) DEFAULT NULL COMMENT '审核人姓名',
    `reviewer_time` datetime DEFAULT NULL COMMENT '审核时间',
    `reviewer_source` varchar(50) DEFAULT NULL COMMENT '审核来源',
    `reviewer_status` tinyint DEFAULT '0' COMMENT '审核人状态：0-不能处理，1-可以处理',
    `executor_id` bigint DEFAULT NULL COMMENT '执行人ID',
    `executor_name` varchar(50) DEFAULT NULL COMMENT '执行人姓名',
    `executor_source` varchar(50) DEFAULT NULL COMMENT '执行人来源',
    `execution_result` text COMMENT '执行结果',
    `executor_status` tinyint DEFAULT '0' COMMENT '执行人状态：0-不能处理，1-可以处理',
    `file_path` varchar(500) DEFAULT NULL COMMENT '文件存储路径',
    `file_name` varchar(255) DEFAULT NULL COMMENT '文件名',
    `file_size` bigint DEFAULT '0' COMMENT '文件大小(字节)',
    `download_count` int DEFAULT '0' COMMENT '下载次数',
    `download_limit` int DEFAULT '3' COMMENT '下载次数限制',
    `expire_time` datetime DEFAULT NULL COMMENT '文件过期时间',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `status` tinyint DEFAULT '0' COMMENT '状态：0-待提交，1-待审核，2-待执行，3-执行中，4-执行完成，5-执行失败，6-已驳回，7-已撤销',
    PRIMARY KEY (`id`),
    KEY `idx_submitter_id` (`submitter_id`),
    KEY `idx_approver_id` (`apply_id`),
    KEY `idx_executor_id` (`executor_id`),
    KEY `idx_status` (`status`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='SQL导出审批表';

-- ----------------------------
-- 表结构：发版记录表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`asset_release_record`;
CREATE TABLE `cmdb`.`asset_release_record` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `task_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务ID',
    `project` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '项目名称',
    `type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'default' COMMENT '发版类型',
    `status` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT 'pending' COMMENT '任务状态',
    `step` int DEFAULT '0' COMMENT '发版步骤',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `started_at` datetime DEFAULT NULL COMMENT '开始时间',
    `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
    `git_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT 'Git仓库地址',
    `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '任务描述',
    `update_feishu` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '任务通知飞书URL',
    `notify_feishu` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '步骤通知飞书URL',
    `result` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '任务结果',
    `error_message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '错误信息',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_task_id` (`task_id`) USING BTREE,
    KEY `idx_project` (`project`) USING BTREE,
    KEY `idx_status` (`status`) USING BTREE,
    KEY `idx_created_at` (`created_at`) USING BTREE
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='发版记录表';

-- ----------------------------
-- 表结构：插件表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`agent_store`;
CREATE TABLE IF NOT EXISTS `cmdb`.`agent_store` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
  `name` VARCHAR(64) NOT NULL COMMENT '插件名称(对应agent的name)',
  `version` VARCHAR(32) NOT NULL COMMENT '版本号(如2.0, 1.0.0)',
  `display_name` VARCHAR(128) NOT NULL COMMENT '显示名称',
  `plugin_type` VARCHAR(32) NOT NULL COMMENT '插件类型(container/binary)',
  `description` TEXT COMMENT '插件描述',
  `port` INT COMMENT '插件服务端口(容器内端口,仅container类型)',
  `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY `uk_name_version` (`name`, `version`),
  INDEX `idx_name` (`name`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='插件商店';

-- ----------------------------
-- 表结构：ip审计表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`audit_ip_log`;
CREATE TABLE `cmdb`.`audit_ip_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `request_ip` varchar(45) NOT NULL COMMENT '请求来源IP',
  `query_ip` varchar(45) NOT NULL COMMENT '查询的IP（入参）',
  `status` int(11) NOT NULL DEFAULT '200' COMMENT '响应状态 200:成功 400:失败',
  `created_at` datetime NOT NULL COMMENT '请求时间',
  `date` varchar(10) NOT NULL COMMENT '日期(YYYYMMDD)便于统计',
  PRIMARY KEY (`id`),
  KEY `idx_request_ip` (`request_ip`),
  KEY `idx_query_ip` (`query_ip`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_date` (`date`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='IP位置查询审计记录';

-- ----------------------------
-- 表结构：加解密审计表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`audit_key_log`;
CREATE TABLE `cmdb`.`audit_key_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` int(11) NOT NULL COMMENT '用户ID',
  `user_name` varchar(64) NOT NULL COMMENT '用户名',
  `operation_type` varchar(20) NOT NULL COMMENT '操作类型: encrypt/decrypt/batchDecrypt',
  `project` varchar(64) NOT NULL COMMENT '项目标识',
  `batch_count` int(11) DEFAULT 1 COMMENT '批量操作数量',
  `status` varchar(20) NOT NULL COMMENT '操作状态: success/failed',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
PRIMARY KEY (`id`),
KEY `idx_user_id` (`user_id`),
KEY `idx_operation_type` (`operation_type`),
KEY `idx_project` (`project`),
KEY `idx_created_at` (`created_at`),
KEY `idx_user_operation` (`user_id`, `operation_type`, `created_at`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='加解密操作审计表';

-- ----------------------------
-- 用户设备绑定表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sys_user_device`;
CREATE TABLE `cmdb`.`sys_user_device` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `user_name` varchar(64) NOT NULL COMMENT '用户名',
    `machine_id` varchar(128) NOT NULL COMMENT '设备硬件指纹',
    `device_secret` varchar(128) NOT NULL COMMENT '设备密钥',
    `device_name` varchar(64) DEFAULT NULL COMMENT '设备名称',
    `version` varchar(32) DEFAULT NULL COMMENT '客户端版本号',
    `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    `expire_at` datetime DEFAULT NULL COMMENT '过期时间',
    `status` tinyint DEFAULT 1 COMMENT '状态:1-正常,0-禁用',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_name` (`user_name`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_machine_id` (`machine_id`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='用户设备绑定表';

-- ----------------------------
-- ES共享查询关键词表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`es_shared_keyword`;
CREATE TABLE `cmdb`.`es_shared_keyword` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `project` varchar(64) NOT NULL COMMENT '项目',
    `category` varchar(64) NOT NULL COMMENT '分类',
    `view_id` bigint NOT NULL COMMENT '视图ID',
    `keyword` varchar(500) NOT NULL COMMENT '关键词',
    `remark` varchar(255) DEFAULT NULL COMMENT '备注',
    `creator` varchar(64) NOT NULL COMMENT '添加人',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_project_category_view_keyword` (`project`, `category`, `view_id`, `keyword`),
    KEY `idx_project` (`project`),
    KEY `idx_view_id` (`view_id`),
    KEY `idx_creator` (`creator`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='ES共享查询关键词表';


-- ----------------------------
-- SQL共享查询表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`sql_shared_query`;
CREATE TABLE `cmdb`.`sql_shared_query` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `project` varchar(64) NOT NULL COMMENT '项目',
    `db_name` varchar(64) NOT NULL COMMENT '数据库',
    `query` text NOT NULL COMMENT '查询语句',
    `remark` varchar(255) DEFAULT NULL COMMENT '备注',
    `creator` varchar(64) NOT NULL COMMENT '添加人',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_project` (`project`),
    KEY `idx_creator` (`creator`)
) ENGINE=InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT='SQL共享查询表';

-- ----------------------------
-- 任务中心表
-- ----------------------------
DROP TABLE IF EXISTS `cmdb`.`task_center`;
CREATE TABLE IF NOT EXISTS `cmdb`.`task_center` (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    task_id VARCHAR(64) NOT NULL COMMENT '任务ID',
    type VARCHAR(32) NOT NULL COMMENT '任务类型: analysis/es_export/sql_export',
    status VARCHAR(16) NOT NULL DEFAULT 'pending' COMMENT '状态: pending/running/success/failed/canceled',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    user_name VARCHAR(64) COMMENT '用户名',
    title VARCHAR(128) COMMENT '任务标题',
    file_path VARCHAR(512) COMMENT '文件路径',
    file_name VARCHAR(128) COMMENT '文件名',
    file_size BIGINT DEFAULT 0 COMMENT '文件大小(字节)',
    total_count INT DEFAULT 0 COMMENT '总记录数',
    error_message TEXT COMMENT '错误信息',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    completed_at DATETIME COMMENT '完成时间',
    expire_at DATETIME COMMENT '过期时间',
    INDEX idx_task_id (task_id),
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务中心表';

-- ----------------------------
-- 初始化数据
-- ----------------------------
-- 初始化监控视图表

INSERT INTO `cmdb`.`monitor_views` (`view_name`, `query`, `description`, `chart_type`, `category`, `sort`, `standard`, `created_by`, `created_at`, `updated_at`) VALUES 
('cpu1分钟负载', 'cpu_load_1', '', 'line', 'hardware', 6, '', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),   
('cpu5分钟负载', 'cpu_load_5', '', 'line', 'hardware', 7, 'default', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('cpu15分钟负载', 'cpu_load_15', '', 'line', 'hardware', 8, 'default', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('内存使用百分比', 'memory_used_percent', '', 'line', 'hardware', 3, 'percent(1-100)', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),   
('空闲内存', 'memory_free', '', 'line', 'hardware', 4, 'kb', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('磁盘百分比', 'disk_used_percent', '', 'line', 'hardware', 2, 'percent(1-100)', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('磁盘空闲', 'disk_free', '', 'line', 'hardware', 5, 'b', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('cpu使用百分比', 'cpu_percent', '', 'line', 'hardware', 1, 'percent(1-100)', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('ssl证书', 'ssl_domain_days_left', '', 'table', 'ssl', 1, 'default', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('【总】容器百分比', 'sum by(container) (container_memory_usage{project="$project",namespace="$namespace",container!="filebeat"}) / sum by(container) (container_memory_limit{project="$project",namespace="$namespace",container!="filebeat"})', '', 'line', 'container', 1, 'percent(0-1)', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('【总】容器内存使用', 'sum by(container) (container_memory_usage{project="$project",namespace="$namespace"})', '', 'line', 'container', 2, 'data(b)', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543'),
('单个容器使用内存百分比', 'sum by(podName) (container_memory_usage{project="$project",namespace="$namespace",container="$container"}) / sum by(podName) (container_memory_limit{project="$project",namespace="$namespace",container="$container"})', '', 'line', 'container', 2, 'data(b)', 1, '2025-05-23 16:39:51.543', '2025-05-23 16:39:51.543');

('实时QPS', 'trafficswitching_realtime_qps{project="$project",service="$service"}', '实时每秒请求数', 'line', 'trafficswitching', 1, 'default', 1),
('实时成功QPS', 'trafficswitching_realtime_success_qps{project="$project",service="$service"}', '实时每秒成功请求数', 'line', 'trafficswitching', 2, 'default', 1),
('实时错误QPS', 'trafficswitching_realtime_error_qps{project="$project",service="$service"}', '实时每秒错误请求数', 'line', 'trafficswitching', 3, 'default', 1),
('实时活跃连接', 'trafficswitching_realtime_active_connections{project="$project",service="$service"}', '当前活跃连接数', 'line', 'trafficswitching', 4, 'default', 1),
('实时平均延迟', 'trafficswitching_realtime_avg_latency_ms{project="$project",service="$service"}', '实时平均响应延迟(ms)', 'line', 'trafficswitching', 5, 'default', 1),
('实时最大延迟', 'trafficswitching_realtime_max_latency_ms{project="$project",service="$service"}', '实时最大响应延迟(ms)', 'line', 'trafficswitching', 6, 'default', 1),

-- ==================== 运行时指标（折线图）====================
('Goroutine数量', 'trafficswitching_runtime_goroutines{project="$project",service="$service"}', '当前Goroutine数量', 'line', 'trafficswitching', 10, 'default', 1),
('内存使用', 'trafficswitching_runtime_memory_mb{project="$project",service="$service"}', 'Go程序内存使用(MB)', 'line', 'trafficswitching', 11, 'default', 1),
('GC运行次数', 'trafficswitching_runtime_gc_cycles{project="$project",service="$service"}', 'GC运行次数', 'line', 'trafficswitching', 12, 'default', 1),

-- ==================== 累计统计（单值）====================
('累计请求总数', 'trafficswitching_total_requests{project="$project",service="$service"}', '累计请求总数', 'stat', 'trafficswitching', 20, 'default', 1),
('累计成功请求', 'trafficswitching_total_success{project="$project",service="$service"}', '累计成功请求数', 'stat', 'trafficswitching', 21, 'default', 1),
('累计失败请求', 'trafficswitching_total_errors{project="$project",service="$service"}', '累计失败请求数', 'stat', 'trafficswitching', 22, 'default', 1),
('累计成功率', 'trafficswitching_total_success_rate{project="$project",service="$service"}', '累计成功率', 'gauge', 'trafficswitching', 23, 'percent', 1),
('累计取消请求', 'trafficswitching_total_canceled{project="$project",service="$service"}', '累计取消请求数', 'stat', 'trafficswitching', 24, 'default', 1),

-- ==================== 缓存指标 ====================
('代理缓存使用', 'trafficswitching_proxy_cache_size{project="$project",service="$service"}', '当前代理缓存大小', 'gauge', 'trafficswitching', 30, 'default', 1),
('代理缓存上限', 'trafficswitching_proxy_max_cache_size{project="$project",service="$service"}', '代理缓存最大限制', 'stat', 'trafficswitching', 31, 'default', 1),

-- ==================== 配置信息（单值）====================
('CPU核心数', 'trafficswitching_runtime_cpu_cores{project="$project",service="$service"}', '机器CPU核心数', 'stat', 'trafficswitching', 40, 'default', 1),
('GOMAXPROCS', 'trafficswitching_runtime_gomaxprocs{project="$project",service="$service"}', 'GOMAXPROCS值', 'stat', 'trafficswitching', 41, 'default', 1),
('每主机最大连接', 'trafficswitching_transport_max_conns_per_host{project="$project",service="$service"}', '每个主机最大连接数', 'stat', 'trafficswitching', 42, 'default', 1),
('全局最大空闲连接', 'trafficswitching_transport_max_idle_conns{project="$project",service="$service"}', '全局最大空闲连接数', 'stat', 'trafficswitching', 43, 'default', 1),
('每主机最大空闲连接', 'trafficswitching_transport_max_idle_conns_per_host{project="$project",service="$service"}', '每个主机最大空闲连接数', 'stat', 'trafficswitching', 44, 'default', 1),

-- ==================== 错误类指标（折线图）====================
('后端错误', 'trafficswitching_error_backend_error{project="$project",service="$service"}', '后端服务错误次数', 'line', 'trafficswitching', 50, 'default', 1),
('管道断开', 'trafficswitching_error_broken_pipe{project="$project",service="$service"}', 'Broken Pipe错误次数', 'line', 'trafficswitching', 51, 'default', 1),
('连接拒绝', 'trafficswitching_error_connection_refused{project="$project",service="$service"}', '连接拒绝错误次数', 'line', 'trafficswitching', 52, 'default', 1),
('连接重置', 'trafficswitching_error_connection_reset{project="$project",service="$service"}', '连接重置错误次数', 'line', 'trafficswitching', 53, 'default', 1),
('DNS错误', 'trafficswitching_error_dns_error{project="$project",service="$service"}', 'DNS解析错误次数', 'line', 'trafficswitching', 54, 'default', 1),
('EOF错误', 'trafficswitching_error_eof{project="$project",service="$service"}', 'EOF错误次数', 'line', 'trafficswitching', 55, 'default', 1),
('超时错误', 'trafficswitching_error_timeout{project="$project",service="$service"}', '请求超时错误次数', 'line', 'trafficswitching', 56, 'default', 1),

-- ==================== 状态码指标（折线图）====================
('2xx响应', 'trafficswitching_status_2xx{project="$project",service="$service"}', '2xx成功响应数', 'line', 'trafficswitching', 60, 'default', 1),
('3xx响应', 'trafficswitching_status_3xx{project="$project",service="$service"}', '3xx重定向响应数', 'line', 'trafficswitching', 61, 'default', 1),
('4xx响应', 'trafficswitching_status_4xx{project="$project",service="$service"}', '4xx客户端错误数', 'line', 'trafficswitching', 62, 'default', 1),
('5xx响应', 'trafficswitching_status_5xx{project="$project",service="$service"}', '5xx服务端错误数', 'line', 'trafficswitching', 63, 'default', 1);
-- 初始化部门表
INSERT INTO `cmdb`.`sys_departments` (`id`, `created_at`, `updated_at`, `name`, `code`, `parent_id`, `description`, `sort`) VALUES
(1, '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', '系统部门', 'SYS', NULL, '系统部门', 1),
(2, '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', '开发部门', 'DEV', NULL, '开发部门', 2),
(3, '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', '测试部门', 'TEST', NULL, '测试部门', 3);
-- 初始化字典表
INSERT INTO `cmdb`.`sys_dict_records` (`id`, `dict_name`, `target_table`, `key_name`, `value_name`, `created_by`, `created_at`) VALUES
(1, '项目字典', 'sys_project_dict', 'project', 'project_name', 1, '2024-12-23 18:37:44.184'),
(2, '文档分类字典', 'sys_category_dict', 'category', 'category_name', 1, '2024-12-23 18:37:44.184'),
(3, 'agent地址字典', 'sys_agent_dict', 'project', 'agent_url', 1, '2024-12-23 18:37:44.184'),
(4, '视图分类', 'sys_view_dict', 'category', 'category_name', 1, '2024-12-23 18:37:44.184');
-- 初始化角色表
INSERT INTO `cmdb`.`sys_roles` (`id`, `name`, `code`, `description`, `created_at`, `updated_at`, `level`) VALUES
(1, '超级管理员', 'admin', '系统超级管理员', '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', 0),
(2, '运维', 'ops', '运维技术部门', '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', 1),
(3, '主管', 'task', '主管', '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', 1),
(4, '普通用户', 'user', '普通用户', '2024-12-19 18:04:23.000', '2024-12-19 18:04:23.000', 3);
-- 初始化用户表
INSERT INTO `cmdb`.`sys_users` (`id`, `created_at`, `updated_at`, `user_name`, `password`, `nick_name`, `email`, `phone`, `role_id`, `dept_id`, `online_assets`, `test_assets`, `is_enabled`, `is_default_pass`) VALUES
(1, '2024-12-23 14:29:24.193', '2025-02-20 09:46:07.762', 'admin', '$2a$12$BtS9HZ4POePnxAZei1fCiuMAncUXRDPSs5anGewVoJAy0/Yl389.6', '管理员', 'zs@example.com', '13800138001', 1, 1, 1, 1, 1, 0);
-- 初始化系统设置
INSERT INTO `cmdb`.`sys_settings` (`id`, `system_name`, `system_url`, `system_short_name`, `password_min_length`, `password_max_length`, `password_need_number`, `password_need_letter`, `password_need_case`, `password_need_special`, `login_protection`, `login_fail_count`, `login_lock_time`, `default_password`, `token_expire_time`, `created_at`, `updated_at`) VALUES
(1, 'CMDB资产管理系统', 'http://cmdb.hzbxhd.com', 'CMDB', 8, 32, 1, 1, 1, 0, 1, 5, 1440, 'Cmdb@123', 168, '2025-03-03 15:41:40.329', '2025-03-03 09:37:27.290');
-- 初始化SQL规则数据
INSERT INTO `cmdb`.`sql_rule` (`rule_name`, `rule_category`, `rule_type`, `sql_tag`, `rule_description`, `is_enabled`, `rule_content`, `error_message`) VALUES
('禁止删除字段', 'security', 'field_check', 'ALTER', 'ALTER TABLE不允许删除字段，防止数据丢失', 1, '{"check_field": "alter_info.drop_field_count", "check_type": "ratio_gt", "check_value": 0}', '禁止删除字段，删除字段将导致数据丢失'),
('禁止TRUNCATE', 'security', 'field_check', 'TRUNCATE', '禁止使用TRUNCATE清空表，防止数据丢失', 1, '{"check_field": "simple_info.operation", "check_type": "equals", "check_value": "TRUNCATE"}', '禁止使用TRUNCATE清空表，请使用DELETE'),
('UPDATE必须有WHERE', 'security', 'field_check', 'UPDATE', 'UPDATE语句必须包含WHERE条件，防止全表更新', 1, '{"check_field": "has_where", "check_type": "equals", "check_value": false}', 'UPDATE语句必须包含WHERE条件，以避免全表更新'),
('DELETE必须有WHERE', 'security', 'field_check', 'DELETE', 'DELETE语句必须包含WHERE条件，防止全表删除', 1, '{"check_field": "has_where", "check_type": "equals", "check_value": false}', 'DELETE语句必须包含WHERE条件，以避免全表删除'),
('禁止删除表', 'security', 'field_check', 'DROP', '禁止使用DROP TABLE删除表', 1, '{"check_field": "simple_info.object_type", "check_type": "equals", "check_value": "TABLE"}', '禁止删除表，删除表将导致数据丢失'),
('禁止删除数据库', 'security', 'field_check', 'DROP', '禁止使用DROP DATABASE删除数据库', 1, '{"check_field": "simple_info.object_type", "check_type": "equals", "check_value": "DATABASE"}', '禁止删除数据库，删除数据库将导致所有数据丢失'),
('主键必需', 'standard', 'field_check', 'CREATE', '创建表时必须定义主键', 1, '{"check_field": "table_info.primary_key", "check_type": "empty"}', '表必须定义主键'),
('表注释必需', 'standard', 'field_check', 'CREATE', '创建表时必须添加表注释', 1, '{"check_field": "table_info.comment", "check_type": "empty"}', '表必须添加注释(COMMENT)'),
('字段注释必需', 'standard', 'field_check', 'CREATE', '所有字段必须添加注释', 1, '{"check_field": "field_info.comment_summary", "check_type": "percent_lt", "check_value": 100}', '所有字段必须添加注释(COMMENT)'),
('禁止自定义存储引擎', 'standard', 'field_check', 'CREATE', '不允许自定义存储引擎', 1, '{"check_field": "table_info.engine", "check_type": "not_empty"}', '不允许自定义存储引擎，请使用数据库默认引擎'),
('禁止自定义字符集', 'standard', 'field_check', 'CREATE', '不允许自定义字符集', 1, '{"check_field": "table_info.charset", "check_type": "not_empty"}', '不允许自定义字符集，请使用数据库默认字符集'),
('禁止自定义排序规则', 'standard', 'field_check', 'CREATE', '不允许自定义排序规则', 1, '{"check_field": "table_info.collation", "check_type": "not_empty"}', '不允许自定义排序规则，请使用数据库默认排序规则'),
('索引必需', 'standard', 'field_check', 'CREATE', '创建表时除主键外必须至少定义一个索引（索引总数≥2）', 1, '{"check_field": "table_info.index_count", "check_type": "lt", "check_value": 2}', '表必须至少定义2个索引（主键+至少一个额外索引）'),
('ALTER操作数量限制', 'standard', 'field_check', 'ALTER', '单条ALTER语句操作数量不超过10个', 0, '{"check_field": "alter_info.total_actions", "check_type": "gt", "check_value": 10}', 'ALTER操作数量超过限制(10个)'),
('禁止添加字段时指定位置', 'security', 'field_check', 'ALTER', '添加字段时不允许使用FIRST/AFTER指定位置，会导致表重建锁表', 1, '{"check_field": "alter_info.add_field_with_position_count", "check_type": "ratio_gt", "check_value": 0}', '禁止添加字段时指定位置(FIRST/AFTER)，会导致表重建锁表，请让字段添加到表末尾'),
('ALTER添加字段必须有注释', 'standard', 'field_check', 'ALTER', 'ALTER TABLE添加字段时必须添加注释', 1, '{"check_field": "alter_info.add_field_without_comment_count", "check_type": "ratio_gt", "check_value": 0}', '添加字段时必须添加注释(COMMENT)'),
('禁止字段自定义排序规则', 'standard', 'field_check', 'ALTER', 'ALTER TABLE时禁止字段自定义COLLATE排序规则，应使用表级别统一设置', 1, '{"check_field": "alter_info.field_with_collate_count", "check_type": "ratio_gt", "check_value": 0}', '禁止字段自定义排序规则(COLLATE)，请使用表级别统一设置'),
('禁止字段自定义排序规则', 'standard', 'field_check', 'CREATE', 'CREATE TABLE时禁止字段自定义COLLATE排序规则，应使用表级别统一设置', 1, '{"check_field": "field_info.collate_summary", "check_type": "ratio_gt", "check_value": 0}', '禁止字段自定义排序规则(COLLATE)，请使用表级别统一设置');

-- 初始化菜单表
INSERT INTO `cmdb`.`sys_menus` (id, parent_id, name, path, component, sort, icon, permission, is_visible, is_enabled, created_at, updated_at) VALUES
(1, NULL, '资产管理', '/assets', 'Layout', 1, 'Box', 'assets', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(2, NULL, '监控管理', '/monitor', 'Layout', 2, 'Monitor', 'monitor', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(3, NULL, '知识库', '/knowledge', 'Layout', 3, 'Document', 'knowledge', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(4, NULL, '定时任务', '/job', 'Layout', 4, 'Collection', 'job', 1, 1, '2025-03-06 09:21:05.932', '2025-03-06 09:21:05.932'),
(5, NULL, '日志分析', '/elfk', 'Layout', 5, 'Tools', 'elfk', 1, 1, '2025-03-10 16:54:08.757', '2025-03-12 10:37:52.937'),
(6, NULL, 'sql管理', '/sql', 'Layout', 6, 'Tools', 'sql', 1, 1, '2025-03-10 16:54:08.757', '2025-03-12 10:37:52.937'),
(7, Null, 'agent管理', '/agent', 'Layout', 7, 'CoffeeCup', 'agent', 1, 1, '2025-03-06 09:21:05.932', '2025-03-06 09:21:05.932'),
(8, NULL, '审计管理', '/audit', 'Layout', 8, 'Folder', 'audit', 1, 1, '2025-03-12 10:34:50.887', '2025-03-12 10:34:50.887'),
(9, NULL, '系统管理', '/system', 'Layout', 9, 'Setting', 'system', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),

(11, 1, '测试环境管理', '/assets/kubePod', 'assets/kubePod/index', 1, 'ChromeFilled', 'assets:kubePod', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),
(12, 1, '测试环境信息', '/assets/kubeServices', 'assets/kubeServices/index', 2, 'ChromeFilled', 'assets:kubeServices', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),
(13, 1, '测试环境域名', '/assets/domain', 'assets/domain/index', 3, 'ChromeFilled', 'assets:domain', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),
(14, 1, '生产发版记录', '/assets/proUpdate', 'assets/proUpdate/index', 4, 'ChromeFilled', 'assets:proUpdate', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),
(15, 1, '生产加解密', '/assets/key', 'assets/key/index', 5, 'ChromeFilled', 'assets:key', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),
(16, 1, '生成文件上传', '/assets/file', 'assets/file/index', 6, 'ChromeFilled', 'assets:file', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),
(17, 1, '阿里云资产信息', '/assets/al', 'assets/al/index', 7, 'ChromeFilled', 'assets:al', 1, 1, '2025-06-23 16:01:19.804', '2025-06-23 16:01:19.804'),

(21, 2, '硬件监控', '/monitor/hard', 'monitor/hard/index', 1, 'Monitor', 'monitor:hard', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(22, 2, '容器监控', '/monitor/container', 'monitor/container/index', 2, 'Box', 'monitor:container', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(23, 2, '证书监控', '/monitor/ssl', 'monitor/ssl/index', 3, 'Lock', 'monitor:ssl', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(24, 2, '流量监控', '/monitor/trafficswitching', 'monitor/trafficswitching/index', 4, 'Lock', 'monitor:trafficswitching', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(29, 2, '告警管理', '/monitor/alarm', 'monitor/alarm/index', 9, 'Bell', 'monitor:alarm', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),

(31, 3, '对外文档', '/knowledge/public', 'knowledge/public/index', 1, 'Box', 'knowledge:public', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(32, 3, '内部文档', '/knowledge/document', 'knowledge/document/index', 2, 'Document', 'knowledge:document', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(33, 3, '个人文档', '/knowledge/personal', 'knowledge/personal/index', 3, 'Cellphone', 'knowledge:personal', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(34, 3, '回收站', '/knowledge/recycle', 'knowledge/recycle/index', 4, 'Coffee', 'knowledge:recycle', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),

(41, 4, '任务管理', '/job/task', 'job/task/index', 1, 'House', 'job:task', 1, 1, '2025-03-06 09:24:25.380', '2025-03-06 09:36:58.032'),
(42, 4, '节点管理', '/job/agent', 'job/agent/index', 2, 'Coordinate', 'job:agent', 1, 1, '2025-03-06 09:23:41.554', '2025-03-06 09:23:41.554'),
(43, 4, '执行记录', '/job/exec', 'job/exec/index', 3, 'Calendar', 'job:exec', 1, 1, '2025-03-06 10:49:18.716', '2025-03-06 10:49:18.716'),

(51, 5, '日志搜索', '/elfk/search', 'elfk/search/index', 1, 'Search', 'elfk:search', 1, 1, '2025-03-10 16:54:41.795', '2025-03-12 10:38:50.366'),
(52, 5, '视图管理', '/elfk/view', 'elfk/view/index', 2, 'Reading', 'elfk:view', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),

(71, 7, '插件市场', '/agent/store', 'agent/store/index', 1, 'Monitor', 'agent:store', 1, 1, '2025-10-16 14:32:48.434', '2025-10-16 14:32:48.434'),
(72, 7, '项目插件', '/agent/project', 'agent/project/index', 2, 'Monitor', 'agent:project', 1, 1, '2025-10-16 14:32:48.434', '2025-10-16 14:32:48.434'),
(73, 7, '插件管理', '/agent/plugins', 'agent/plugins/index', 3, 'Monitor', 'agent:plugins', 1, 1, '2025-10-16 14:32:48.434', '2025-10-16 14:32:48.434'),

(61, 6, 'sql查询', '/sql/search', 'sql/search/index', 1, 'Search', 'sql:search', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),
(62, 6, 'sql执行', '/sql/apply', 'sql/apply/index', 2, 'Cellphone', 'sql:apply', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),
(63, 6, 'sql导出', '/sql/export', 'sql/export/index', 3, 'Cellphone', 'sql:export', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),
(64, 6, 'sql流程', '/sql/process', 'sql/process/index', 4, 'Promotion', 'sql:process', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),
(65, 6, 'sql规则', '/sql/rules', 'sql/rules/index', 5, 'Promotion', 'sql:rules', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),
(66, 6, 'bi查询', '/sql/databi', 'sql/databi/index', 5, 'Promotion', 'sql:databi', 1, 1, '2025-03-12 10:38:32.830', '2025-03-12 10:38:32.830'),

(80, 8, '审计分析', '/audit/analysis', 'audit/analysis/index', 0, 'DataAnalysis', 'audit:analysis', 1, 1, '2025-03-12 10:36:13.482', '2025-03-12 10:35:48.583'),
(81, 8, '操作日志', '/audit/operation', 'audit/operation/index', 1, 'ScaleToOriginal', 'audit:operation', 1, 1, '2025-03-12 10:36:13.482', '2025-03-12 10:35:48.583'),
(82, 8, '登录日志', '/audit/login', 'audit/login/index', 2, 'Platform', 'audit:login', 1, 1, '2025-03-12 10:36:13.482', '2025-03-12 10:36:13.482'),
(83, 8, 'es审计', '/audit/search', 'audit/search/index', 3, 'Reading', 'audit:search', 1, 1, '2025-03-12 10:37:05.452', '2025-03-12 10:37:05.452'),
(84, 8, 'SQL审计', '/audit/sql', 'audit/sql/index', 4, 'DataAnalysis', 'audit:sql', 1, 1, '2025-03-12 10:37:05.452', '2025-03-12 10:37:05.452'),
(85, 8, 'ip访问审计', '/audit/ip', 'audit/ip/index', 5, 'DataAnalysis', 'audit:ip', 1, 1, '2025-03-12 10:37:05.452', '2025-03-12 10:37:05.452'),
(86, 8, '加解密审计', '/audit/key', 'audit/key/index', 6, 'DataAnalysis', 'audit:key', 1, 1, '2025-03-12 10:37:05.452', '2025-03-12 10:37:05.452');

(91, 9, '用户管理', '/system/user', 'system/user/index', 1, 'UserFilled', 'system:user', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(92, 9, '角色管理', '/system/role', 'system/role/index', 2, 'User', 'system:role', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(93, 9, '部门管理', '/system/dept', 'system/dept/index', 3, 'OfficeBuilding', 'system:dept', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(94, 9, '项目管理', '/system/project', 'system/project/index', 4, 'OfficeBuilding', 'system:project', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(95, 9, '菜单管理', '/system/menu', 'system/menu/index', 5, 'Menu', 'system:menu', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(96, 9, '字典管理', '/system/dict', 'system/dict/index', 6, 'List', 'system:dict', 1, 1, '2025-02-10 11:51:30.494', '2025-02-10 11:51:30.494'),
(97, 9, '系统设置', '/system/setting', 'system/setting/index', 7, 'Setting', 'system:setting', 1, 1, '2025-02-18 14:31:19.794', '2025-02-18 14:31:19.794');

-- 初始化系统设置
INSERT INTO `cmdb`.`sys_settings` ( `id`, `system_name`, `system_short_name`, `system_url`, `fs_webhook_url`, `password_min_length`, `password_max_length`, `password_need_number`, `password_need_letter`, `password_need_case`, `password_need_special`, `login_protection`, `login_fail_count`, `login_lock_time`, `default_password`, `token_expire_time`, `created_at`, `updated_at` ) VALUES 
( 1, 'CMDB系统', 'CMDB', 'http://localhost:8080', 'https://open.notice.cn/open-apis/bot/v2/hook/your-webhook-url', 8, 32, 1, 1, 1, 0, 1, 5, 30, 'Cmdb@123', 24, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP ) ON DUPLICATE KEY UPDATE `fs_webhook_url` = VALUES(`fs_webhook_url`);

INSERT INTO `cmdb`.`sys_security_config` (blacklist_count, blacklist_lock, ratelimit_cd, ratelimit_ttl) VALUES (5, 30, 60, 3600);

SET FOREIGN_KEY_CHECKS = 1;