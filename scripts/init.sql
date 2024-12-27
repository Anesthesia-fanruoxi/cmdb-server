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
DROP TABLE IF EXISTS `cmdb`.`departments`;
CREATE TABLE IF NOT EXISTS `cmdb`.`departments` (
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `name`        VARCHAR(50) NOT NULL COMMENT '部门名称',
    `code`        VARCHAR(50) NOT NULL COMMENT '部门编码',
    `description` VARCHAR(200) DEFAULT '' COMMENT '部门描述',
    `parent_id`   BIGINT UNSIGNED DEFAULT NULL COMMENT '父部门ID',
    `sort`        INT DEFAULT 0 COMMENT '排序',
    `created_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='部门表';

-- 初始化部门数据
INSERT INTO departments (id, name, code, description, parent_id, sort, created_at, updated_at) VALUES
(1, '总部', 'HQ', '公司总部', NULL, 1, NOW(), NOW()),
(2, '研发中心', 'RD', '研发中心', 1, 1, NOW(), NOW()),
(3, '运维中心', 'OPS', '运维中心', 1, 2, NOW(), NOW()),
(4, '测试中心', 'QA', '测试中心', 1, 3, NOW(), NOW());

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
DROP TABLE IF EXISTS `cmdb`.`git_dict`;
CREATE TABLE IF NOT EXISTS `cmdb`.`git_dict` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `project` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '键',
  `git_url` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '值',
  `created_by` bigint unsigned NOT NULL COMMENT '创建人ID',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`project`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='git地址';

DROP TABLE IF EXISTS `cmdb`.`project_dict`;
CREATE TABLE IF NOT EXISTS `cmdb`.`project_dict` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `project` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '键',
  `project_name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '值',
  `created_by` bigint unsigned NOT NULL COMMENT '创建人ID',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_key` (`project`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='项目字典';

-- 创建菜单表
DROP TABLE IF EXISTS `cmdb`.menus;
CREATE TABLE IF NOT EXISTS `cmdb`.menus (
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
DROP TABLE IF EXISTS `cmdb`.role_menus;
CREATE TABLE IF NOT EXISTS `cmdb`.role_menus (
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
(1, '资产管理', '/assets', 'Layout', NULL, 1, 'Box', 'assets', 1, 1),
(2, '监控管理', '/monitor', 'Layout', NULL, 2, 'Monitor', 'monitor', 1, 1),
(3, '知识库', '/knowledge', 'Layout', NULL, 3, 'Document', 'knowledge', 1, 1),
(9, '系统管理', '/system', 'Layout', NULL, 9, 'Setting', 'system', 1, 1),

-- 资产管理子菜单
(11, '线上资产', '/assets/online', 'Layout', 1, 1, 'Cloud', 'assets:online', 1, 1),
(12, '测试资产', '/assets/test', 'Layout', 1, 2, 'Experiment', 'assets:test', 1, 1),

-- 线上资产子菜单
(111, '数据库', '/assets/online/database', 'assets/online/database/index', 11, 1, 'DataBase', 'assets:online:database', 1, 1),
(112, '网络', '/assets/online/network', 'assets/online/network/index', 11, 2, 'Connection', 'assets:online:network', 1, 1),
(113, '服务器', '/assets/online/server', 'assets/online/server/index', 11, 3, 'Monitor', 'assets:online:server', 1, 1),

-- 测试资产子菜单
(121, '环境预览', '/assets/test/overview', 'assets/test/overview/index', 12, 1, 'View', 'assets:test:overview', 1, 1),
(122, '端口映射', '/assets/test/port-mapping', 'assets/test/port-mapping/index', 12, 2, 'Position', 'assets:test:port-mapping', 1, 1),

-- 监控管理子菜单
(21, '监控总览', '/monitor/overview', 'monitor/overview/index', 2, 1, 'DataBoard', 'monitor:overview', 1, 1),
(22, '告警管理', '/monitor/alarm', 'monitor/alarm/index', 2, 2, 'Bell', 'monitor:alarm', 1, 1),

-- 知识库子菜单
(31, '文档管理', '/knowledge/document', 'knowledge/document/index', 3, 1, 'Document', 'knowledge:document', 1, 1),

-- 系统管理子菜单
(91, '用户管理', '/system/user', 'system/user/index', 9, 1, 'UserFilled', 'system:user', 1, 1),
(92, '角色管理', '/system/role', 'system/role/index', 9, 2, 'User', 'system:role', 1, 1),
(93, '部门管理', '/system/dept', 'system/dept/index', 9, 3, 'OfficeBuilding', 'system:dept', 1, 1),
(94, '菜单管理', '/system/menu', 'system/menu/index', 9, 4, 'Menu', 'system:menu', 1, 1),
(95, '字典管理', '/system/dict', 'system/dict/index', 9, 5, 'List', 'system:dict', 1, 1);

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

-- 清空表数据
TRUNCATE TABLE users;
TRUNCATE TABLE roles;
TRUNCATE TABLE menus;
TRUNCATE TABLE role_menus;
TRUNCATE TABLE user_roles;

-- 初始化角色表
INSERT INTO roles (id, name, code, description, created_at, updated_at) VALUES
(1, '超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', NOW(), NOW()),
(2, '主管', 'manager', '部门主管，管理本部门工作', NOW(), NOW()),
(3, '开发', 'developer', '开发工程师', NOW(), NOW()),
(4, '运维', 'ops', '运维工程师', NOW(), NOW()),
(5, '测试', 'tester', '测试工程师', NOW(), NOW()),
(6, '产品', 'pm', '产品经理', NOW(), NOW());

-- 初始化菜单表
INSERT INTO menus (id, name, path, component, parent_id, sort, icon, permission, is_visible, is_enabled, created_at, updated_at) VALUES
-- 一级菜单
(1, '系统管理', '/system', 'Layout', NULL, 1, 'setting', 'system', 1, 1, NOW(), NOW()),
(2, '资产管理', '/assets', 'Layout', NULL, 2, 'desktop', 'assets', 1, 1, NOW(), NOW()),
(3, '应用管理', '/apps', 'Layout', NULL, 3, 'app-store', 'apps', 1, 1, NOW(), NOW()),

-- 系统管理子菜单
(11, '用户管理', 'users', '/system/users/index', 1, 1, 'user', 'system:users', 1, 1, NOW(), NOW()),
(12, '角色管理', 'roles', '/system/roles/index', 1, 2, 'team', 'system:roles', 1, 1, NOW(), NOW()),
(13, '菜单管理', 'menus', '/system/menus/index', 1, 3, 'menu', 'system:menus', 1, 1, NOW(), NOW()),

-- 资产管理子菜单
(21, '服务器', 'servers', '/assets/servers/index', 2, 1, 'server', 'assets:servers', 1, 1, NOW(), NOW()),
(22, '网络设备', 'network', '/assets/network/index', 2, 2, 'cluster', 'assets:network', 1, 1, NOW(), NOW()),
(23, '存储设备', 'storage', '/assets/storage/index', 2, 3, 'database', 'assets:storage', 1, 1, NOW(), NOW()),

-- 应用管理子菜单
(31, '应用列表', 'list', '/apps/list/index', 3, 1, 'appstore', 'apps:list', 1, 1, NOW(), NOW()),
(32, '部署管理', 'deploy', '/apps/deploy/index', 3, 2, 'rocket', 'apps:deploy', 1, 1, NOW(), NOW()),
(33, '监控告警', 'monitor', '/apps/monitor/index', 3, 3, 'alert', 'apps:monitor', 1, 1, NOW(), NOW());

-- 初始化角色-菜单关联
-- 超级管理员拥有所有权限
INSERT INTO role_menus (role_id, menu_id) VALUES
(1, 1), (1, 11), (1, 12), (1, 13),
(1, 2), (1, 21), (1, 22), (1, 23),
(1, 3), (1, 31), (1, 32), (1, 33);

-- 主管拥有查看权限和部分管理权限
INSERT INTO role_menus (role_id, menu_id) VALUES
(2, 2), (2, 21), (2, 22), (2, 23),
(2, 3), (2, 31), (2, 32), (2, 33);

-- 开发人员权限
INSERT INTO role_menus (role_id, menu_id) VALUES
(3, 2), (3, 21),
(3, 3), (3, 31), (3, 32);

-- 运维人员权限
INSERT INTO role_menus (role_id, menu_id) VALUES
(4, 2), (4, 21), (4, 22), (4, 23),
(4, 3), (4, 31), (4, 32), (4, 33);

-- 测试人员权限
INSERT INTO role_menus (role_id, menu_id) VALUES
(3, 2), (3, 21),
(3, 3), (3, 31), (3, 33);

-- 产品经理权限
INSERT INTO role_menus (role_id, menu_id) VALUES
(6, 2), (6, 21),
(6, 3), (6, 31);

-- 初始化用户表 (添加一个超级管理员用户)
INSERT INTO users (id, username, password, nickname, email, phone, avatar, status, created_at, updated_at) VALUES
(1, 'admin', '$2a$10$n1Ae7S/Hlnd/Xf791WOj0e9imsOCnz.W62eJv1o9//EMth387RES2', '超级管理员', 'admin@example.com', '13800138000', '', 1, NOW(), NOW());
-- admin/Zhang123456
-- 初始化用户-角色关联
INSERT INTO user_roles (user_id, role_id) VALUES (1, 1);

-- 创建部门-项目关联表
DROP TABLE IF EXISTS `cmdb`.`dept_projects`;
CREATE TABLE IF NOT EXISTS `cmdb`.`dept_projects` (
    `dept_id`     BIGINT UNSIGNED NOT NULL COMMENT '部门ID',
    `project`     VARCHAR(64) NOT NULL COMMENT '项目标识',
    PRIMARY KEY (`dept_id`, `project`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='部门-项目关联表';

-- 更新菜单表的权限标识
UPDATE menus SET permission = 'system:users:list' WHERE id = 11;
UPDATE menus SET permission = 'system:roles:list' WHERE id = 12;
UPDATE menus SET permission = 'system:menus:list' WHERE id = 13;

-- 为每个菜单添加操作权限
INSERT INTO menus (name, path, component, parent_id, sort, icon, permission, is_visible, is_enabled) VALUES
-- 用户管理操作权限
('创建用户', NULL, NULL, 11, 1, NULL, 'system:users:create', 0, 1),
('更新用户', NULL, NULL, 11, 2, NULL, 'system:users:update', 0, 1),
('删除用户', NULL, NULL, 11, 3, NULL, 'system:users:delete', 0, 1),
('查看用户', NULL, NULL, 11, 4, NULL, 'system:users:detail', 0, 1),

-- 角色管理操作权限
('创建角色', NULL, NULL, 12, 1, NULL, 'system:roles:create', 0, 1),
('更新角色', NULL, NULL, 12, 2, NULL, 'system:roles:update', 0, 1),
('删除角色', NULL, NULL, 12, 3, NULL, 'system:roles:delete', 0, 1),
('角色菜单', NULL, NULL, 12, 4, NULL, 'system:roles:menu', 0, 1),

-- 菜单管理操作权限
('创建菜单', NULL, NULL, 13, 1, NULL, 'system:menu:create', 0, 1),
('更新菜单', NULL, NULL, 13, 2, NULL, 'system:menu:update', 0, 1),
('删除菜单', NULL, NULL, 13, 3, NULL, 'system:menu:delete', 0, 1),
('菜单树', NULL, NULL, 13, 4, NULL, 'system:menu:tree', 0, 1),
('用户菜单', NULL, NULL, 13, 5, NULL, 'system:menu:user', 0, 1),

-- 部门管理操作权限
('创建部门', NULL, NULL, 93, 1, NULL, 'system:dept:create', 0, 1),
('更新部门', NULL, NULL, 93, 2, NULL, 'system:dept:update', 0, 1),
('删除部门', NULL, NULL, 93, 3, NULL, 'system:dept:delete', 0, 1),
('部门项目', NULL, NULL, 93, 4, NULL, 'system:dept:project', 0, 1),

-- 字典管理操作权限
('创建字典', NULL, NULL, 95, 1, NULL, 'system:dict:create', 0, 1),
('删除字典', NULL, NULL, 95, 2, NULL, 'system:dict:delete', 0, 1),
('查询字典', NULL, NULL, 95, 3, NULL, 'system:dict:query', 0, 1),
('创建字典项', NULL, NULL, 95, 4, NULL, 'system:dict:item:create', 0, 1),
('删除字典项', NULL, NULL, 95, 5, NULL, 'system:dict:item:delete', 0, 1);
