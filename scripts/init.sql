-- 创建数据库，如果不存在则创建
CREATE
DATABASE IF NOT EXISTS `cmdb` DEFAULT CHARACTER SET utf8mb4 COLLATE=utf8mb4_general_ci;

-- 使用数据库
USE
`cmdb`;

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`roles`;

-- 创建角色表
CREATE TABLE `cmdb`.`roles`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '角色ID',
    `created_at`  datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at`  datetime(3) DEFAULT NULL COMMENT '更新时间',
    `name`        varchar(32) NOT NULL COMMENT '角色名称',
    `code`        varchar(32) NOT NULL COMMENT '角色编码',
    `description` varchar(128) DEFAULT NULL COMMENT '角色描述',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色表';

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`users`;

-- 创建用户表
CREATE TABLE `cmdb`.`users`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
    `username`   varchar(32)  NOT NULL COMMENT '用户名',
    `password`   varchar(128) NOT NULL COMMENT '密码',
    `nickname`   varchar(32)  DEFAULT NULL COMMENT '昵称',
    `email`      varchar(128) DEFAULT NULL COMMENT '邮箱',
    `phone`      varchar(11)  DEFAULT NULL COMMENT '手机号',
    `role_id`    bigint unsigned NOT NULL DEFAULT '2' COMMENT '角色ID',
    `is_enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
    `dept_id`    bigint unsigned NOT NULL COMMENT '部门ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    KEY          `idx_role_id` (`role_id`),
    KEY          `idx_dept_id` (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`servers`;

-- 创建服务器表
CREATE TABLE `cmdb`.`servers`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '服务器ID',
    `created_at`  datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at`  datetime(3) DEFAULT NULL COMMENT '更新时间',
    `name`        varchar(64) NOT NULL COMMENT '服务器名称',
    `ip`          varchar(15) NOT NULL COMMENT 'IP地址',
    `port`        int         NOT NULL DEFAULT '22' COMMENT 'SSH端口',
    `username`    varchar(32) NOT NULL COMMENT 'SSH用户名',
    `password`    varchar(128)         DEFAULT NULL COMMENT 'SSH密码',
    `private_key` text COMMENT 'SSH私钥',
    `type`        varchar(32) NOT NULL COMMENT '服务器类型',
    `status`      varchar(32) NOT NULL DEFAULT 'offline' COMMENT '服务器状态',
    `os`          varchar(64)          DEFAULT NULL COMMENT '操作系统',
    `cpu`         int                  DEFAULT NULL COMMENT 'CPU核数',
    `memory`      int                  DEFAULT NULL COMMENT '内存大小(GB)',
    `disk`        int                  DEFAULT NULL COMMENT '磁盘大小(GB)',
    `comment`     varchar(255)         DEFAULT NULL COMMENT '备注',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_ip` (`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='服务器表';

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`departments`;

-- 创建部门表
CREATE TABLE IF NOT EXISTS departments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3) NULL,
    name VARCHAR(50) NOT NULL COMMENT '部门名称',
    code VARCHAR(50) NOT NULL COMMENT '部门编码',
    description VARCHAR(255) NULL COMMENT '描述',
    parent_id BIGINT UNSIGNED NULL COMMENT '父部门ID',
    sort INT NOT NULL DEFAULT 0 COMMENT '排序',
    is_enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    UNIQUE KEY `idx_code` (`code`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`project_dicts`;

-- 创建项目字典表
CREATE TABLE `cmdb`.`project_dicts`
(
    `project`      varchar(64)  NOT NULL COMMENT '项目代码',
    `project_name` varchar(128) NOT NULL COMMENT '项目中文名称',
    PRIMARY KEY (`project`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目字典表';

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`operations`;

-- 创建操作记录表
CREATE TABLE `cmdb`.`operations`
(
    `id`                  bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `namespace`           varchar(255) NOT NULL COMMENT '命名空间',
    `action`              varchar(255) NOT NULL COMMENT '执行的操作',
    `action_user_name`    varchar(255) NOT NULL COMMENT '执行操作的用户',
    `action_time`         varchar(255) NOT NULL COMMENT '操作时间',
    `action_timestamp`    varchar(255) NOT NULL COMMENT '操作时间戳',
    `operat_user_name`    varchar(255) DEFAULT NULL COMMENT '迭代操作的用户',
    `operation_time`      varchar(255) DEFAULT NULL COMMENT '迭代操作时间，格式为 YYYY/MM/DD HH:mm:ss',
    `operation_timestamp` varchar(255) DEFAULT NULL COMMENT '迭代操作的时间戳，毫秒级别',
    `git_url`            varchar(255) DEFAULT NULL COMMENT '项目git地址',
    `last_git_branch`    varchar(255) DEFAULT NULL COMMENT '最后一次执行的分支名称',
    PRIMARY KEY (`id`),
    KEY `idx_namespace` (`namespace`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='操作记录表';

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`project_git_repos`;

-- 创建项目Git仓库关系表
CREATE TABLE `cmdb`.`project_git_repos` (
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `created_at`  datetime(3) DEFAULT NULL COMMENT '创建时间',
    `updated_at`  datetime(3) DEFAULT NULL COMMENT '更新时间',
    `project_id`  bigint unsigned NOT NULL COMMENT '项目ID',
    `git_url`     varchar(255) NOT NULL COMMENT 'Git仓库地址',
    `description` text COMMENT '仓库描述',
    `created_by`  bigint unsigned NOT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_project_id` (`project_id`),
    KEY `idx_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目Git仓库关系表';

-- 判断并删除表，如果存在则删除
DROP TABLE IF EXISTS `cmdb`.`dict_records`;

-- 创建字典记录表
CREATE TABLE `cmdb`.`dict_records` (
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `dict_name`   varchar(64) NOT NULL COMMENT '字典名称',
    `table_name`  varchar(64) NOT NULL COMMENT '表名',
    `key_name`   varchar(64) NOT NULL COMMENT '键字段名',
    `value_name` varchar(64) NOT NULL COMMENT '值字段名',
    `created_by`  bigint unsigned NOT NULL COMMENT '创建人ID',
    `created_at`  datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_table_name` (`table_name`),
    UNIQUE KEY `uk_dict_name` (`dict_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='字典记录表';

-- 字典模板SQL（用于动态创建字典表）
-- 注意：这个不是要执行的SQL，而是作为模板存在
/*
CREATE TABLE `cmdb`.`dict_${table_name}` (
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `key`         varchar(64) NOT NULL COMMENT '键',
    `value`       varchar(255) NOT NULL COMMENT '值',
    `created_by`  bigint unsigned NOT NULL COMMENT '创建人ID',
    `created_at`  datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='${dict_name}';
*/

-- 创建菜单表
CREATE TABLE IF NOT EXISTS menus (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at  DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    name        VARCHAR(50) NOT NULL COMMENT '菜单名称',
    path        VARCHAR(100) COMMENT '前端路由路径',
    component   VARCHAR(100) COMMENT '前端组件路径',
    permission  VARCHAR(50) COMMENT '权限标识',
    parent_id   BIGINT UNSIGNED COMMENT '父菜单ID',
    sort        INT DEFAULT 0 COMMENT '排序',
    icon        VARCHAR(50) COMMENT '图标',
    is_visible  TINYINT(1) DEFAULT 1 COMMENT '是否可见',
    is_enabled  TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='菜单表';

-- 创建角色-菜单关联表
CREATE TABLE IF NOT EXISTS role_menus (
    role_id     BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
    menu_id     BIGINT UNSIGNED NOT NULL COMMENT '菜单ID',
    created_at  DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (role_id, menu_id),
    KEY `idx_role_id` (`role_id`),
    KEY `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色-菜单关联表';

-- 清空已有的菜单数据
TRUNCATE TABLE menus;
TRUNCATE TABLE role_menus;

-- 插入菜单数据
INSERT INTO menus (id, name, path, component, parent_id, sort, icon, permission, is_visible, is_enabled) VALUES
-- 一级菜单
(1, '资产管理', '/assets', 'Layout', NULL, 1, 'Asset', 'assets', 1, 1),
(2, '监控管理', '/monitor', 'Layout', NULL, 2, 'Monitor', 'monitor', 1, 1),
(3, '知识库', '/knowledge', 'Layout', NULL, 3, 'Document', 'knowledge', 1, 1),
(9, '系统管理', '/system', 'Layout', NULL, 9, 'Setting', 'system', 1, 1),

-- 资产管理子菜单
(11, '线上资产', '/assets/online', 'Layout', 1, 1, 'Cloud', 'assets:online', 1, 1),
(12, '测试资产', '/assets/test', 'Layout', 1, 2, 'Test', 'assets:test', 1, 1),

-- 线上资产子菜单
(111, '数据库', '/assets/online/database', 'assets/online/database/index', 11, 1, 'Database', 'assets:online:database', 1, 1),
(112, '网络', '/assets/online/network', 'assets/online/network/index', 11, 2, 'Network', 'assets:online:network', 1, 1),
(113, '服务器', '/assets/online/server', 'assets/online/server/index', 11, 3, 'Server', 'assets:online:server', 1, 1),

-- 测试资产子菜单
(121, '环境预览', '/assets/test/overview', 'assets/test/overview/index', 12, 1, 'Preview', 'assets:test:overview', 1, 1),
(122, '端口映射', '/assets/test/port-mapping', 'assets/test/port-mapping/index', 12, 2, 'Port', 'assets:test:port-mapping', 1, 1),

-- 监控管理子菜单
(21, '监控总览', '/monitor/overview', 'monitor/overview/index', 2, 1, 'Dashboard', 'monitor:overview', 1, 1),
(22, '告警管理', '/monitor/alarm', 'monitor/alarm/index', 2, 2, 'Alarm', 'monitor:alarm', 1, 1),
(23, '监控配置', '/monitor/config', 'monitor/config/index', 2, 3, 'Setting', 'monitor:config', 1, 1),

-- 知识库子菜单
(31, '文档管理', '/knowledge/document', 'knowledge/document/index', 3, 1, 'Document', 'knowledge:document', 1, 1),

-- 系统管理子菜单
(91, '用户管理', '/system/user', 'system/user/index', 9, 1, 'User', 'system:user', 1, 1),
(92, '角色管理', '/system/role', 'system/role/index', 9, 2, 'Role', 'system:role', 1, 1),
(93, '部门管理', '/system/dept', 'system/dept/index', 9, 3, 'Dept', 'system:dept', 1, 1),
(94, '菜单管理', '/system/menu', 'system/menu/index', 9, 4, 'Menu', 'system:menu', 1, 1),
(95, '字典管理', '/system/dict', 'system/dict/index', 9, 5, 'Dict', 'system:dict', 1, 1);

-- 为超级管理员角色(id=1)分配所有菜单权限
INSERT INTO role_menus (role_id, menu_id)
SELECT 1, id FROM menus;

-- 为普通用户角色(id=2)分配基本菜单权限（除了系统管理）
INSERT INTO role_menus (role_id, menu_id)
SELECT 2, id FROM menus WHERE parent_id IS NULL AND id != 9
UNION
SELECT 2, id FROM menus WHERE parent_id IN (
    SELECT id FROM menus WHERE parent_id IS NULL AND id != 9
)
UNION
SELECT 2, id FROM menus WHERE parent_id IN (
    SELECT id FROM menus WHERE parent_id IN (
        SELECT id FROM menus WHERE parent_id IS NULL AND id != 9
    )
);
