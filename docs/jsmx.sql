/*
 Navicat Premium Data Transfer

 Source Server         : 172.16.2.8
 Source Server Type    : MySQL
 Source Server Version : 50740
 Source Host           : 172.16.2.8:3306
 Source Schema         : jsmx

 Target Server Type    : MySQL
 Target Server Version : 50740
 File Encoding         : 65001

 Date: 13/11/2024 16:01:34
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for aiconfig
-- ----------------------------
DROP TABLE IF EXISTS `aiconfig`;
CREATE TABLE `aiconfig`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `monster_id` bigint(20) NOT NULL,
  `patrol_range` int(255) NOT NULL DEFAULT 0,
  `chase_range` int(255) NOT NULL,
  `alert_range` int(255) NOT NULL DEFAULT 0 COMMENT '警戒范围',
  `auto_beatback` int(255) NOT NULL DEFAULT 1 COMMENT '1自动反击，0不反击',
  `spells` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '拥有哪些技能',
  `conds` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '# 条件节点，满足条件时，会执行后面的act\n        \'conds\':{\n            # 配置方式：\'节点名\': [判断类型(>, <, =, %), 数值, 时间， 概率](未使用填-1)\n            \'once\': [[\'=\', -1, \'act_1\', -1, -1]],  # 立即执行act_1\n            \'blood\':[[\'<\', \'0.5\', \'act_2\', -1, -1]], # 血量小于0.5时执行act2\n            \'time\':[[\'%\' ,\'10\', \'act_3\', -1, 80]],  # 每隔10秒有0.8的概率执行一次act_3\n            ...\n        }\r\n',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of aiconfig
-- ----------------------------
INSERT INTO `aiconfig` VALUES (1, 1, 30, 50, 20, 1, '1,2', '[]', '2024-10-21 14:27:15', '2024-11-12 18:54:35');
INSERT INTO `aiconfig` VALUES (2, 2, 30, 50, 10, 1, '1,2', '[]', '2024-10-21 14:27:15', '2024-11-12 18:54:40');

-- ----------------------------
-- Table structure for buffer_state
-- ----------------------------
DROP TABLE IF EXISTS `buffer_state`;
CREATE TABLE `buffer_state`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `animation` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `buf_type` tinyint(255) NOT NULL DEFAULT 0 COMMENT '0 伤害类型，1治疗类型',
  `damage` int(255) NOT NULL DEFAULT 0 COMMENT '负数就是治疗，正数为伤害',
  `effect_duration_time` smallint(255) NOT NULL DEFAULT 0 COMMENT '每次持续时间',
  `effect_disappear_time` smallint(6) NOT NULL COMMENT '每次结束后的消失时间',
  `effect_cnt` smallint(255) NOT NULL DEFAULT 1 COMMENT '生效次数',
  `cd_time` int(11) NOT NULL DEFAULT 0 COMMENT 'cd时间',
  `stackable` tinyint(255) NOT NULL DEFAULT 0 COMMENT '是否可叠加',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of buffer_state
-- ----------------------------
INSERT INTO `buffer_state` VALUES (1, '', 'buf1', 0, 60, 1000, 1000, 5, 10000, 0, '2024-11-13 15:29:48', '2024-11-13 15:47:15');
INSERT INTO `buffer_state` VALUES (2, '', 'buf2', 1, -10, 1000, 1000, 10, 10000, 0, '2024-11-13 15:34:24', '2024-11-13 15:34:24');

-- ----------------------------
-- Table structure for hero
-- ----------------------------
DROP TABLE IF EXISTS `hero`;
CREATE TABLE `hero`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '资源',
  `attr_type` tinyint(255) NOT NULL DEFAULT 0 COMMENT '属性类型:0 力量，1敏捷, 2智慧',
  `uid` bigint(20) NOT NULL,
  `experience` bigint(255) NOT NULL DEFAULT 0,
  `level` int(255) NOT NULL DEFAULT 1,
  `max_life` bigint(255) NOT NULL DEFAULT 0,
  `max_mana` bigint(255) NOT NULL DEFAULT 0,
  `defense` bigint(255) NOT NULL DEFAULT 0,
  `attack` bigint(255) NOT NULL DEFAULT 0,
  `base_life` bigint(255) NOT NULL,
  `base_mana` bigint(255) NOT NULL,
  `base_defense` bigint(255) NOT NULL,
  `base_attack` bigint(255) NOT NULL,
  `strength` bigint(255) NOT NULL,
  `agility` bigint(255) NOT NULL,
  `intelligence` bigint(255) NOT NULL,
  `step_time` int(11) NOT NULL DEFAULT 0 COMMENT '移动速度',
  `scene_id` int(11) NOT NULL DEFAULT 0,
  `init_posx` int(255) NOT NULL DEFAULT 0,
  `init_posy` int(255) NOT NULL DEFAULT 0,
  `init_posz` int(255) NOT NULL DEFAULT 0,
  `attack_range` int(255) NOT NULL DEFAULT 0,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `uid_nk`(`uid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 23 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '角色' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of hero
-- ----------------------------
INSERT INTO `hero` VALUES (1, '陶醉的永恩', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 1, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 12:13:04', '2024-11-13 12:13:04');
INSERT INTO `hero` VALUES (2, '肖申克在巴黎徒步', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 2, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 12:25:50', '2024-11-13 12:25:50');
INSERT INTO `hero` VALUES (3, '呆萌的乔布斯', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 3, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 14:42:35', '2024-11-13 14:42:35');
INSERT INTO `hero` VALUES (4, '科比打豆豆', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 4, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 14:58:35', '2024-11-13 14:58:35');
INSERT INTO `hero` VALUES (5, '细腻的普拉蒂尼', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 5, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 14:59:41', '2024-11-13 14:59:41');
INSERT INTO `hero` VALUES (6, '风中的哈维', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 6, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:00:19', '2024-11-13 15:00:19');
INSERT INTO `hero` VALUES (7, '野性的雅典娜', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 7, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:00:39', '2024-11-13 15:00:39');
INSERT INTO `hero` VALUES (8, '柔弱的齐达內', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 8, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:01:22', '2024-11-13 15:01:22');
INSERT INTO `hero` VALUES (9, '粗犷的姆巴佩', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 9, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:01:56', '2024-11-13 15:01:56');
INSERT INTO `hero` VALUES (10, '一休走向人生巅峰', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 10, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:02:14', '2024-11-13 15:02:14');
INSERT INTO `hero` VALUES (11, '欧文完成了帽子戏法', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 11, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:32:55', '2024-11-13 15:32:55');
INSERT INTO `hero` VALUES (12, '听话的鲁尼', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 12, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:36:49', '2024-11-13 15:36:49');
INSERT INTO `hero` VALUES (13, '尤西比奥一眼定情', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 13, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:37:03', '2024-11-13 15:37:03');
INSERT INTO `hero` VALUES (14, '约翰·查尔斯在武汉看电影', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 14, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:45:45', '2024-11-13 15:45:45');
INSERT INTO `hero` VALUES (15, '罗马里奥有亿点点忧伤', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 15, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:47:33', '2024-11-13 15:47:33');
INSERT INTO `hero` VALUES (16, '巴乔横扫六合', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 16, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:48:42', '2024-11-13 15:48:42');
INSERT INTO `hero` VALUES (17, '加林查吃爆米花', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 17, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:49:37', '2024-11-13 15:49:37');
INSERT INTO `hero` VALUES (18, '懵懂的伊布', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 18, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:49:53', '2024-11-13 15:49:53');
INSERT INTO `hero` VALUES (19, '普拉蒂尼掐指一算', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 19, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:51:38', '2024-11-13 15:51:38');
INSERT INTO `hero` VALUES (20, '知性的贝利', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 20, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:51:49', '2024-11-13 15:51:49');
INSERT INTO `hero` VALUES (21, '美好的永恩', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 21, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:59:13', '2024-11-13 15:59:13');
INSERT INTO `hero` VALUES (22, '永恩求而不得', 'https://img2.baidu.com/it/u=3171875674,3530712457&fm=253&fmt=auto&app=120&f=JPEG?w=800&h=800', 0, 22, 0, 1, 1420, 1300, 44, 78, 1000, 1000, 5, 22, 28, 22, 20, 300, 1, 0, 0, 0, 3, '2024-11-13 15:59:32', '2024-11-13 15:59:32');

-- ----------------------------
-- Table structure for login
-- ----------------------------
DROP TABLE IF EXISTS `login`;
CREATE TABLE `login`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` bigint(20) NOT NULL,
  `remote` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `ip` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `model` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `imei` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `os` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `appid` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `login_at` bigint(255) NOT NULL DEFAULT 0,
  `logout_at` bigint(255) NOT NULL DEFAULT 0,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `uid_nk`(`uid`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 288 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of login
-- ----------------------------
INSERT INTO `login` VALUES (1, 1, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688091, 0, '2024-11-04 10:41:31');
INSERT INTO `login` VALUES (2, 2, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688105, 0, '2024-11-04 10:41:45');
INSERT INTO `login` VALUES (3, 3, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688122, 0, '2024-11-04 10:42:02');
INSERT INTO `login` VALUES (4, 4, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688135, 0, '2024-11-04 10:42:15');
INSERT INTO `login` VALUES (5, 5, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688136, 0, '2024-11-04 10:42:16');
INSERT INTO `login` VALUES (6, 6, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688142, 0, '2024-11-04 10:42:22');
INSERT INTO `login` VALUES (7, 7, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688146, 0, '2024-11-04 10:42:26');
INSERT INTO `login` VALUES (8, 8, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688152, 0, '2024-11-04 10:42:32');
INSERT INTO `login` VALUES (9, 9, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688156, 0, '2024-11-04 10:42:36');
INSERT INTO `login` VALUES (10, 10, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688161, 0, '2024-11-04 10:42:41');
INSERT INTO `login` VALUES (11, 11, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688165, 0, '2024-11-04 10:42:45');
INSERT INTO `login` VALUES (12, 12, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688177, 0, '2024-11-04 10:42:57');
INSERT INTO `login` VALUES (13, 13, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688185, 0, '2024-11-04 10:43:05');
INSERT INTO `login` VALUES (14, 14, '127.0.0.1:53315', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688189, 0, '2024-11-04 10:43:09');
INSERT INTO `login` VALUES (15, 15, '127.0.0.1:53568', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688192, 0, '2024-11-04 10:43:12');
INSERT INTO `login` VALUES (16, 16, '127.0.0.1:53314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688195, 0, '2024-11-04 10:43:15');
INSERT INTO `login` VALUES (17, 17, '127.0.0.1:53568', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730688203, 0, '2024-11-04 10:43:23');
INSERT INTO `login` VALUES (18, 18, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690146, 0, '2024-11-04 11:15:46');
INSERT INTO `login` VALUES (19, 19, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690147, 0, '2024-11-04 11:15:47');
INSERT INTO `login` VALUES (20, 20, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690183, 0, '2024-11-04 11:16:23');
INSERT INTO `login` VALUES (21, 21, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690191, 0, '2024-11-04 11:16:31');
INSERT INTO `login` VALUES (22, 22, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690197, 0, '2024-11-04 11:16:37');
INSERT INTO `login` VALUES (23, 23, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690203, 0, '2024-11-04 11:16:43');
INSERT INTO `login` VALUES (24, 24, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690211, 0, '2024-11-04 11:16:51');
INSERT INTO `login` VALUES (25, 25, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690218, 0, '2024-11-04 11:16:58');
INSERT INTO `login` VALUES (26, 26, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690223, 0, '2024-11-04 11:17:03');
INSERT INTO `login` VALUES (27, 27, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690227, 0, '2024-11-04 11:17:07');
INSERT INTO `login` VALUES (28, 28, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690233, 0, '2024-11-04 11:17:13');
INSERT INTO `login` VALUES (29, 29, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690238, 0, '2024-11-04 11:17:18');
INSERT INTO `login` VALUES (30, 30, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690245, 0, '2024-11-04 11:17:25');
INSERT INTO `login` VALUES (31, 31, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690249, 0, '2024-11-04 11:17:29');
INSERT INTO `login` VALUES (32, 32, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690261, 0, '2024-11-04 11:17:41');
INSERT INTO `login` VALUES (33, 33, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690325, 0, '2024-11-04 11:18:45');
INSERT INTO `login` VALUES (34, 34, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690329, 0, '2024-11-04 11:18:49');
INSERT INTO `login` VALUES (35, 35, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690332, 0, '2024-11-04 11:18:52');
INSERT INTO `login` VALUES (36, 36, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690335, 0, '2024-11-04 11:18:55');
INSERT INTO `login` VALUES (37, 37, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690341, 0, '2024-11-04 11:19:01');
INSERT INTO `login` VALUES (38, 38, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690344, 0, '2024-11-04 11:19:04');
INSERT INTO `login` VALUES (39, 39, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690348, 0, '2024-11-04 11:19:08');
INSERT INTO `login` VALUES (40, 40, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690352, 0, '2024-11-04 11:19:12');
INSERT INTO `login` VALUES (41, 41, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690356, 0, '2024-11-04 11:19:16');
INSERT INTO `login` VALUES (42, 42, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690370, 0, '2024-11-04 11:19:30');
INSERT INTO `login` VALUES (43, 43, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690439, 0, '2024-11-04 11:20:39');
INSERT INTO `login` VALUES (44, 44, '127.0.0.1:61422', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690445, 0, '2024-11-04 11:20:45');
INSERT INTO `login` VALUES (45, 45, '127.0.0.1:61423', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730690450, 0, '2024-11-04 11:20:50');
INSERT INTO `login` VALUES (46, 46, '127.0.0.1:62842', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730692705, 0, '2024-11-04 11:58:25');
INSERT INTO `login` VALUES (47, 47, '127.0.0.1:62843', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730692721, 0, '2024-11-04 11:58:41');
INSERT INTO `login` VALUES (48, 48, '127.0.0.1:62843', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730692725, 0, '2024-11-04 11:58:45');
INSERT INTO `login` VALUES (49, 49, '127.0.0.1:62842', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730692729, 0, '2024-11-04 11:58:49');
INSERT INTO `login` VALUES (50, 50, '127.0.0.1:62843', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730692734, 0, '2024-11-04 11:58:54');
INSERT INTO `login` VALUES (51, 51, '127.0.0.1:62842', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730692799, 0, '2024-11-04 11:59:59');
INSERT INTO `login` VALUES (52, 52, '127.0.0.1:62843', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693038, 0, '2024-11-04 12:03:58');
INSERT INTO `login` VALUES (53, 53, '127.0.0.1:62842', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693272, 0, '2024-11-04 12:07:52');
INSERT INTO `login` VALUES (54, 54, '127.0.0.1:62843', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693317, 0, '2024-11-04 12:08:37');
INSERT INTO `login` VALUES (55, 57, '127.0.0.1:59020', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693545, 0, '2024-11-04 12:12:25');
INSERT INTO `login` VALUES (56, 58, '127.0.0.1:59059', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693546, 0, '2024-11-04 12:12:26');
INSERT INTO `login` VALUES (57, 55, '127.0.0.1:62843', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693528, 0, '2024-11-04 12:12:08');
INSERT INTO `login` VALUES (58, 56, '127.0.0.1:62842', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693558, 0, '2024-11-04 12:12:38');
INSERT INTO `login` VALUES (59, 59, '127.0.0.1:55653', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693564, 0, '2024-11-04 12:12:44');
INSERT INTO `login` VALUES (60, 60, '127.0.0.1:55654', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693576, 0, '2024-11-04 12:12:56');
INSERT INTO `login` VALUES (61, 61, '127.0.0.1:55653', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693607, 0, '2024-11-04 12:13:27');
INSERT INTO `login` VALUES (62, 62, '127.0.0.1:55654', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693621, 0, '2024-11-04 12:13:41');
INSERT INTO `login` VALUES (63, 63, '127.0.0.1:55653', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693756, 0, '2024-11-04 12:15:56');
INSERT INTO `login` VALUES (64, 64, '127.0.0.1:55654', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730693889, 0, '2024-11-04 12:18:09');
INSERT INTO `login` VALUES (65, 65, '127.0.0.1:55653', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694024, 0, '2024-11-04 12:20:24');
INSERT INTO `login` VALUES (66, 66, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694567, 0, '2024-11-04 12:29:27');
INSERT INTO `login` VALUES (67, 67, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694570, 0, '2024-11-04 12:29:30');
INSERT INTO `login` VALUES (68, 68, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694571, 0, '2024-11-04 12:29:31');
INSERT INTO `login` VALUES (69, 69, '127.0.0.1:57807', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694573, 0, '2024-11-04 12:29:33');
INSERT INTO `login` VALUES (70, 70, '127.0.0.1:57807', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694575, 0, '2024-11-04 12:29:35');
INSERT INTO `login` VALUES (71, 71, '127.0.0.1:57807', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694577, 0, '2024-11-04 12:29:37');
INSERT INTO `login` VALUES (72, 72, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694578, 0, '2024-11-04 12:29:38');
INSERT INTO `login` VALUES (73, 73, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694580, 0, '2024-11-04 12:29:40');
INSERT INTO `login` VALUES (74, 74, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694583, 0, '2024-11-04 12:29:43');
INSERT INTO `login` VALUES (75, 75, '127.0.0.1:57807', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694583, 0, '2024-11-04 12:29:43');
INSERT INTO `login` VALUES (76, 76, '127.0.0.1:57807', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694586, 0, '2024-11-04 12:29:46');
INSERT INTO `login` VALUES (77, 77, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694588, 0, '2024-11-04 12:29:48');
INSERT INTO `login` VALUES (78, 78, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694591, 0, '2024-11-04 12:29:51');
INSERT INTO `login` VALUES (79, 79, '127.0.0.1:57806', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730694593, 0, '2024-11-04 12:29:53');
INSERT INTO `login` VALUES (80, 80, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703677, 0, '2024-11-04 15:01:17');
INSERT INTO `login` VALUES (81, 81, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703684, 0, '2024-11-04 15:01:24');
INSERT INTO `login` VALUES (82, 82, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703686, 0, '2024-11-04 15:01:26');
INSERT INTO `login` VALUES (83, 83, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703709, 0, '2024-11-04 15:01:49');
INSERT INTO `login` VALUES (84, 84, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703713, 0, '2024-11-04 15:01:53');
INSERT INTO `login` VALUES (85, 85, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703730, 0, '2024-11-04 15:02:10');
INSERT INTO `login` VALUES (86, 86, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703840, 0, '2024-11-04 15:04:00');
INSERT INTO `login` VALUES (87, 87, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730703841, 0, '2024-11-04 15:04:01');
INSERT INTO `login` VALUES (88, 88, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704057, 0, '2024-11-04 15:07:37');
INSERT INTO `login` VALUES (89, 89, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704060, 0, '2024-11-04 15:07:40');
INSERT INTO `login` VALUES (90, 90, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704160, 0, '2024-11-04 15:09:20');
INSERT INTO `login` VALUES (91, 91, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704161, 0, '2024-11-04 15:09:21');
INSERT INTO `login` VALUES (92, 92, '127.0.0.1:55924', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704273, 0, '2024-11-04 15:11:13');
INSERT INTO `login` VALUES (93, 93, '127.0.0.1:55925', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704274, 0, '2024-11-04 15:11:14');
INSERT INTO `login` VALUES (94, 94, '127.0.0.1:49768', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704539, 0, '2024-11-04 15:15:39');
INSERT INTO `login` VALUES (95, 95, '127.0.0.1:50213', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730704544, 0, '2024-11-04 15:15:44');
INSERT INTO `login` VALUES (96, 96, '127.0.0.1:52271', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706054, 0, '2024-11-04 15:40:54');
INSERT INTO `login` VALUES (97, 97, '127.0.0.1:52271', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706056, 0, '2024-11-04 15:40:56');
INSERT INTO `login` VALUES (98, 98, '127.0.0.1:52272', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706061, 0, '2024-11-04 15:41:01');
INSERT INTO `login` VALUES (99, 99, '127.0.0.1:52271', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706150, 0, '2024-11-04 15:42:30');
INSERT INTO `login` VALUES (100, 100, '127.0.0.1:52272', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706278, 0, '2024-11-04 15:44:38');
INSERT INTO `login` VALUES (101, 101, '127.0.0.1:62797', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706696, 0, '2024-11-04 15:51:36');
INSERT INTO `login` VALUES (102, 102, '127.0.0.1:62798', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706713, 0, '2024-11-04 15:51:53');
INSERT INTO `login` VALUES (103, 103, '127.0.0.1:62797', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730706724, 0, '2024-11-04 15:52:04');
INSERT INTO `login` VALUES (104, 104, '127.0.0.1:61185', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730775730, 0, '2024-11-05 11:02:10');
INSERT INTO `login` VALUES (105, 105, '127.0.0.1:50130', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730776396, 0, '2024-11-05 11:13:16');
INSERT INTO `login` VALUES (106, 106, '127.0.0.1:50131', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730776443, 0, '2024-11-05 11:14:03');
INSERT INTO `login` VALUES (107, 107, '127.0.0.1:50703', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730776640, 0, '2024-11-05 11:17:20');
INSERT INTO `login` VALUES (108, 108, '127.0.0.1:50717', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730780324, 0, '2024-11-05 12:18:44');
INSERT INTO `login` VALUES (109, 109, '127.0.0.1:50719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730780332, 0, '2024-11-05 12:18:52');
INSERT INTO `login` VALUES (110, 110, '127.0.0.1:50717', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730780345, 0, '2024-11-05 12:19:05');
INSERT INTO `login` VALUES (111, 111, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730790392, 0, '2024-11-05 15:06:32');
INSERT INTO `login` VALUES (112, 112, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730790623, 0, '2024-11-05 15:10:23');
INSERT INTO `login` VALUES (113, 113, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730790738, 0, '2024-11-05 15:12:18');
INSERT INTO `login` VALUES (114, 114, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730790864, 0, '2024-11-05 15:14:24');
INSERT INTO `login` VALUES (115, 115, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791075, 0, '2024-11-05 15:17:55');
INSERT INTO `login` VALUES (116, 116, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791164, 0, '2024-11-05 15:19:24');
INSERT INTO `login` VALUES (117, 117, '127.0.0.1:61719', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791249, 0, '2024-11-05 15:20:49');
INSERT INTO `login` VALUES (118, 118, '127.0.0.1:64029', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791564, 0, '2024-11-05 15:26:04');
INSERT INTO `login` VALUES (119, 119, '127.0.0.1:64029', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791690, 0, '2024-11-05 15:28:10');
INSERT INTO `login` VALUES (120, 120, '127.0.0.1:64029', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791880, 0, '2024-11-05 15:31:20');
INSERT INTO `login` VALUES (121, 121, '127.0.0.1:64520', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730791911, 0, '2024-11-05 15:31:51');
INSERT INTO `login` VALUES (122, 122, '127.0.0.1:65443', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792318, 0, '2024-11-05 15:38:38');
INSERT INTO `login` VALUES (123, 123, '127.0.0.1:65444', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792346, 0, '2024-11-05 15:39:06');
INSERT INTO `login` VALUES (124, 124, '127.0.0.1:65443', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792366, 0, '2024-11-05 15:39:26');
INSERT INTO `login` VALUES (125, 125, '127.0.0.1:65444', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792381, 0, '2024-11-05 15:39:41');
INSERT INTO `login` VALUES (126, 126, '127.0.0.1:65443', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792488, 0, '2024-11-05 15:41:28');
INSERT INTO `login` VALUES (127, 127, '127.0.0.1:50353', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792981, 0, '2024-11-05 15:49:41');
INSERT INTO `login` VALUES (128, 128, '127.0.0.1:50354', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730792997, 0, '2024-11-05 15:49:57');
INSERT INTO `login` VALUES (129, 129, '127.0.0.1:57314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730794926, 0, '2024-11-05 16:22:06');
INSERT INTO `login` VALUES (130, 130, '127.0.0.1:57314', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730795077, 0, '2024-11-05 16:24:37');
INSERT INTO `login` VALUES (131, 131, '127.0.0.1:57629', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730795087, 0, '2024-11-05 16:24:47');
INSERT INTO `login` VALUES (132, 132, '127.0.0.1:58160', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730795408, 0, '2024-11-05 16:30:08');
INSERT INTO `login` VALUES (133, 133, '127.0.0.1:58161', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730795425, 0, '2024-11-05 16:30:25');
INSERT INTO `login` VALUES (134, 134, '127.0.0.1:64851', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730800778, 0, '2024-11-05 17:59:38');
INSERT INTO `login` VALUES (135, 135, '127.0.0.1:64851', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730800860, 0, '2024-11-05 18:01:00');
INSERT INTO `login` VALUES (136, 136, '127.0.0.1:50151', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730802889, 0, '2024-11-05 18:34:49');
INSERT INTO `login` VALUES (137, 137, '127.0.0.1:50152', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730802898, 0, '2024-11-05 18:34:58');
INSERT INTO `login` VALUES (138, 138, '127.0.0.1:50151', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730802913, 0, '2024-11-05 18:35:13');
INSERT INTO `login` VALUES (139, 139, '127.0.0.1:58937', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730860854, 0, '2024-11-06 10:40:54');
INSERT INTO `login` VALUES (140, 140, '127.0.0.1:52948', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730861354, 0, '2024-11-06 10:49:14');
INSERT INTO `login` VALUES (141, 141, '127.0.0.1:52949', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730861375, 0, '2024-11-06 10:49:35');
INSERT INTO `login` VALUES (142, 142, '127.0.0.1:52948', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730861396, 0, '2024-11-06 10:49:56');
INSERT INTO `login` VALUES (143, 143, '127.0.0.1:60118', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730861788, 0, '2024-11-06 10:56:28');
INSERT INTO `login` VALUES (144, 144, '127.0.0.1:60118', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730862044, 0, '2024-11-06 11:00:44');
INSERT INTO `login` VALUES (145, 145, '127.0.0.1:64605', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730864322, 0, '2024-11-06 11:38:42');
INSERT INTO `login` VALUES (146, 146, '127.0.0.1:51874', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730879494, 0, '2024-11-06 15:51:34');
INSERT INTO `login` VALUES (147, 147, '127.0.0.1:51874', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730879599, 0, '2024-11-06 15:53:19');
INSERT INTO `login` VALUES (148, 148, '127.0.0.1:52282', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730879606, 0, '2024-11-06 15:53:26');
INSERT INTO `login` VALUES (149, 149, '127.0.0.1:51874', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730879620, 0, '2024-11-06 15:53:40');
INSERT INTO `login` VALUES (150, 150, '127.0.0.1:52282', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730879780, 0, '2024-11-06 15:56:20');
INSERT INTO `login` VALUES (151, 151, '127.0.0.1:51874', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730879788, 0, '2024-11-06 15:56:28');
INSERT INTO `login` VALUES (152, 152, '127.0.0.1:52282', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730880038, 0, '2024-11-06 16:00:38');
INSERT INTO `login` VALUES (153, 153, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881310, 0, '2024-11-06 16:21:50');
INSERT INTO `login` VALUES (154, 154, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881468, 0, '2024-11-06 16:24:28');
INSERT INTO `login` VALUES (155, 155, '127.0.0.1:56153', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881507, 0, '2024-11-06 16:25:07');
INSERT INTO `login` VALUES (156, 156, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881595, 0, '2024-11-06 16:26:35');
INSERT INTO `login` VALUES (157, 157, '127.0.0.1:56153', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881842, 0, '2024-11-06 16:30:42');
INSERT INTO `login` VALUES (158, 158, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881850, 0, '2024-11-06 16:30:50');
INSERT INTO `login` VALUES (159, 159, '127.0.0.1:56153', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730881980, 0, '2024-11-06 16:33:00');
INSERT INTO `login` VALUES (160, 160, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882088, 0, '2024-11-06 16:34:48');
INSERT INTO `login` VALUES (161, 161, '127.0.0.1:56153', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882183, 0, '2024-11-06 16:36:23');
INSERT INTO `login` VALUES (162, 162, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882230, 0, '2024-11-06 16:37:10');
INSERT INTO `login` VALUES (163, 163, '127.0.0.1:55727', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882234, 0, '2024-11-06 16:37:14');
INSERT INTO `login` VALUES (164, 164, '127.0.0.1:56153', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882238, 0, '2024-11-06 16:37:18');
INSERT INTO `login` VALUES (165, 165, '127.0.0.1:52820', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882754, 0, '2024-11-06 16:45:54');
INSERT INTO `login` VALUES (166, 166, '127.0.0.1:52821', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882784, 0, '2024-11-06 16:46:24');
INSERT INTO `login` VALUES (167, 167, '127.0.0.1:52820', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882807, 0, '2024-11-06 16:46:47');
INSERT INTO `login` VALUES (168, 168, '127.0.0.1:52821', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882823, 0, '2024-11-06 16:47:03');
INSERT INTO `login` VALUES (169, 169, '127.0.0.1:52820', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730882847, 0, '2024-11-06 16:47:27');
INSERT INTO `login` VALUES (170, 170, '127.0.0.1:51840', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730884890, 0, '2024-11-06 17:21:30');
INSERT INTO `login` VALUES (171, 171, '127.0.0.1:51841', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730884920, 0, '2024-11-06 17:22:00');
INSERT INTO `login` VALUES (172, 172, '127.0.0.1:52565', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730885443, 0, '2024-11-06 17:30:43');
INSERT INTO `login` VALUES (173, 173, '127.0.0.1:52565', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730885594, 0, '2024-11-06 17:33:14');
INSERT INTO `login` VALUES (174, 174, '127.0.0.1:53619', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886342, 0, '2024-11-06 17:45:42');
INSERT INTO `login` VALUES (175, 175, '127.0.0.1:53620', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886368, 0, '2024-11-06 17:46:08');
INSERT INTO `login` VALUES (176, 176, '127.0.0.1:54218', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886808, 0, '2024-11-06 17:53:28');
INSERT INTO `login` VALUES (177, 177, '127.0.0.1:54219', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886810, 0, '2024-11-06 17:53:30');
INSERT INTO `login` VALUES (178, 178, '127.0.0.1:54218', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886812, 0, '2024-11-06 17:53:32');
INSERT INTO `login` VALUES (179, 179, '127.0.0.1:54219', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886814, 0, '2024-11-06 17:53:34');
INSERT INTO `login` VALUES (180, 180, '127.0.0.1:54218', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886845, 0, '2024-11-06 17:54:05');
INSERT INTO `login` VALUES (181, 181, '127.0.0.1:54219', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886848, 0, '2024-11-06 17:54:08');
INSERT INTO `login` VALUES (182, 182, '127.0.0.1:54218', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886853, 0, '2024-11-06 17:54:13');
INSERT INTO `login` VALUES (183, 183, '127.0.0.1:54219', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886866, 0, '2024-11-06 17:54:26');
INSERT INTO `login` VALUES (184, 184, '127.0.0.1:54218', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730886873, 0, '2024-11-06 17:54:33');
INSERT INTO `login` VALUES (185, 185, '127.0.0.1:59239', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730887546, 0, '2024-11-06 18:05:46');
INSERT INTO `login` VALUES (186, 186, '127.0.0.1:59239', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730887711, 0, '2024-11-06 18:08:31');
INSERT INTO `login` VALUES (187, 187, '127.0.0.1:59884', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888064, 0, '2024-11-06 18:14:24');
INSERT INTO `login` VALUES (188, 188, '127.0.0.1:59884', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888233, 0, '2024-11-06 18:17:13');
INSERT INTO `login` VALUES (189, 189, '127.0.0.1:59884', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888349, 0, '2024-11-06 18:19:09');
INSERT INTO `login` VALUES (190, 190, '127.0.0.1:56284', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888359, 0, '2024-11-06 18:19:19');
INSERT INTO `login` VALUES (191, 191, '127.0.0.1:59884', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888387, 0, '2024-11-06 18:19:47');
INSERT INTO `login` VALUES (192, 192, '127.0.0.1:56284', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888411, 0, '2024-11-06 18:20:11');
INSERT INTO `login` VALUES (193, 193, '127.0.0.1:59884', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888518, 0, '2024-11-06 18:21:58');
INSERT INTO `login` VALUES (194, 194, '127.0.0.1:56284', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730888780, 0, '2024-11-06 18:26:20');
INSERT INTO `login` VALUES (195, 195, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889194, 0, '2024-11-06 18:33:14');
INSERT INTO `login` VALUES (196, 196, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889365, 0, '2024-11-06 18:36:05');
INSERT INTO `login` VALUES (197, 197, '127.0.0.1:59662', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889375, 0, '2024-11-06 18:36:15');
INSERT INTO `login` VALUES (198, 198, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889381, 0, '2024-11-06 18:36:21');
INSERT INTO `login` VALUES (199, 199, '127.0.0.1:59662', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889388, 0, '2024-11-06 18:36:28');
INSERT INTO `login` VALUES (200, 200, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889585, 0, '2024-11-06 18:39:45');
INSERT INTO `login` VALUES (201, 201, '127.0.0.1:59662', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889674, 0, '2024-11-06 18:41:14');
INSERT INTO `login` VALUES (202, 202, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730889802, 0, '2024-11-06 18:43:22');
INSERT INTO `login` VALUES (203, 203, '127.0.0.1:59662', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730890018, 0, '2024-11-06 18:46:58');
INSERT INTO `login` VALUES (204, 204, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730890234, 0, '2024-11-06 18:50:34');
INSERT INTO `login` VALUES (205, 205, '127.0.0.1:59662', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730890401, 0, '2024-11-06 18:53:21');
INSERT INTO `login` VALUES (206, 206, '127.0.0.1:59274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730890563, 0, '2024-11-06 18:56:03');
INSERT INTO `login` VALUES (207, 207, '127.0.0.1:64331', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730947914, 0, '2024-11-07 10:51:54');
INSERT INTO `login` VALUES (208, 208, '127.0.0.1:65329', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730948436, 0, '2024-11-07 11:00:36');
INSERT INTO `login` VALUES (209, 209, '127.0.0.1:55548', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730962326, 0, '2024-11-07 14:52:06');
INSERT INTO `login` VALUES (210, 210, '127.0.0.1:56096', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730962621, 0, '2024-11-07 14:57:01');
INSERT INTO `login` VALUES (211, 211, '127.0.0.1:56097', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730962643, 0, '2024-11-07 14:57:23');
INSERT INTO `login` VALUES (212, 212, '127.0.0.1:56096', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730962671, 0, '2024-11-07 14:57:51');
INSERT INTO `login` VALUES (213, 213, '127.0.0.1:65365', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730966048, 0, '2024-11-07 15:54:08');
INSERT INTO `login` VALUES (214, 214, '127.0.0.1:65455', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730966065, 0, '2024-11-07 15:54:25');
INSERT INTO `login` VALUES (215, 215, '127.0.0.1:65457', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730966096, 0, '2024-11-07 15:54:56');
INSERT INTO `login` VALUES (216, 216, '127.0.0.1:65082', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730967164, 0, '2024-11-07 16:12:44');
INSERT INTO `login` VALUES (217, 217, '127.0.0.1:59633', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1730969329, 0, '2024-11-07 16:48:49');
INSERT INTO `login` VALUES (218, 218, '127.0.0.1:52798', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731059566, 0, '2024-11-08 17:52:46');
INSERT INTO `login` VALUES (219, 219, '127.0.0.1:52797', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731059671, 0, '2024-11-08 17:54:31');
INSERT INTO `login` VALUES (220, 220, '127.0.0.1:52795', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731059719, 0, '2024-11-08 17:55:19');
INSERT INTO `login` VALUES (221, 221, '127.0.0.1:51853', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731059793, 0, '2024-11-08 17:56:33');
INSERT INTO `login` VALUES (222, 222, '127.0.0.1:51853', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731059858, 0, '2024-11-08 17:57:38');
INSERT INTO `login` VALUES (223, 223, '127.0.0.1:51853', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731060130, 0, '2024-11-08 18:02:10');
INSERT INTO `login` VALUES (224, 224, '127.0.0.1:52443', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731060139, 0, '2024-11-08 18:02:19');
INSERT INTO `login` VALUES (225, 225, '127.0.0.1:51853', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731060354, 0, '2024-11-08 18:05:54');
INSERT INTO `login` VALUES (226, 226, '127.0.0.1:52443', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731060360, 0, '2024-11-08 18:06:00');
INSERT INTO `login` VALUES (227, 227, '127.0.0.1:52443', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731060364, 0, '2024-11-08 18:06:04');
INSERT INTO `login` VALUES (228, 228, '127.0.0.1:51853', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731060464, 0, '2024-11-08 18:07:44');
INSERT INTO `login` VALUES (229, 229, '127.0.0.1:59815', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061250, 0, '2024-11-08 18:20:50');
INSERT INTO `login` VALUES (230, 230, '127.0.0.1:59816', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061255, 0, '2024-11-08 18:20:55');
INSERT INTO `login` VALUES (231, 231, '127.0.0.1:59815', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061381, 0, '2024-11-08 18:23:01');
INSERT INTO `login` VALUES (232, 232, '127.0.0.1:59816', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061385, 0, '2024-11-08 18:23:05');
INSERT INTO `login` VALUES (233, 233, '127.0.0.1:59815', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061436, 0, '2024-11-08 18:23:56');
INSERT INTO `login` VALUES (234, 234, '127.0.0.1:52270', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061884, 0, '2024-11-08 18:31:24');
INSERT INTO `login` VALUES (235, 235, '127.0.0.1:52271', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731061887, 0, '2024-11-08 18:31:27');
INSERT INTO `login` VALUES (236, 236, '127.0.0.1:52270', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731062066, 0, '2024-11-08 18:34:26');
INSERT INTO `login` VALUES (237, 237, '127.0.0.1:52271', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731062069, 0, '2024-11-08 18:34:29');
INSERT INTO `login` VALUES (238, 238, '127.0.0.1:64581', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731403988, 0, '2024-11-12 17:33:08');
INSERT INTO `login` VALUES (239, 239, '127.0.0.1:49844', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731404633, 0, '2024-11-12 17:43:53');
INSERT INTO `login` VALUES (240, 240, '127.0.0.1:49845', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731404654, 0, '2024-11-12 17:44:14');
INSERT INTO `login` VALUES (241, 241, '127.0.0.1:49844', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731404661, 0, '2024-11-12 17:44:21');
INSERT INTO `login` VALUES (242, 242, '127.0.0.1:49845', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731404691, 0, '2024-11-12 17:44:51');
INSERT INTO `login` VALUES (243, 243, '127.0.0.1:51090', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731405026, 0, '2024-11-12 17:50:26');
INSERT INTO `login` VALUES (244, 244, '127.0.0.1:53730', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731406220, 0, '2024-11-12 18:10:20');
INSERT INTO `login` VALUES (245, 245, '127.0.0.1:53730', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731406337, 0, '2024-11-12 18:12:17');
INSERT INTO `login` VALUES (246, 246, '127.0.0.1:54090', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731406368, 0, '2024-11-12 18:12:48');
INSERT INTO `login` VALUES (247, 247, '127.0.0.1:58102', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731407222, 0, '2024-11-12 18:27:02');
INSERT INTO `login` VALUES (248, 248, '127.0.0.1:60781', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731408328, 0, '2024-11-12 18:45:28');
INSERT INTO `login` VALUES (249, 249, '127.0.0.1:60781', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731408606, 0, '2024-11-12 18:50:06');
INSERT INTO `login` VALUES (250, 250, '127.0.0.1:60781', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731408707, 0, '2024-11-12 18:51:47');
INSERT INTO `login` VALUES (251, 251, '127.0.0.1:60781', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731408831, 0, '2024-11-12 18:53:51');
INSERT INTO `login` VALUES (252, 252, '127.0.0.1:60781', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731408903, 0, '2024-11-12 18:55:03');
INSERT INTO `login` VALUES (253, 253, '127.0.0.1:60781', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731409129, 0, '2024-11-12 18:58:49');
INSERT INTO `login` VALUES (254, 254, '127.0.0.1:50544', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731463672, 0, '2024-11-13 10:07:52');
INSERT INTO `login` VALUES (255, 255, '127.0.0.1:50544', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731463851, 0, '2024-11-13 10:10:51');
INSERT INTO `login` VALUES (256, 256, '127.0.0.1:63050', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731464577, 0, '2024-11-13 10:22:57');
INSERT INTO `login` VALUES (257, 257, '127.0.0.1:63050', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731464821, 0, '2024-11-13 10:27:01');
INSERT INTO `login` VALUES (258, 258, '127.0.0.1:63705', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731464826, 0, '2024-11-13 10:27:06');
INSERT INTO `login` VALUES (259, 259, '127.0.0.1:63050', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731464839, 0, '2024-11-13 10:27:19');
INSERT INTO `login` VALUES (260, 260, '127.0.0.1:64875', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731465341, 0, '2024-11-13 10:35:41');
INSERT INTO `login` VALUES (261, 261, '127.0.0.1:64875', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731465405, 0, '2024-11-13 10:36:45');
INSERT INTO `login` VALUES (262, 262, '127.0.0.1:64875', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731465579, 0, '2024-11-13 10:39:39');
INSERT INTO `login` VALUES (263, 263, '127.0.0.1:49907', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731466054, 0, '2024-11-13 10:47:34');
INSERT INTO `login` VALUES (264, 264, '127.0.0.1:50692', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731466427, 0, '2024-11-13 10:53:47');
INSERT INTO `login` VALUES (265, 265, '127.0.0.1:53887', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731468335, 0, '2024-11-13 11:25:35');
INSERT INTO `login` VALUES (266, 1, '127.0.0.1:59679', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731471182, 0, '2024-11-13 12:13:02');
INSERT INTO `login` VALUES (267, 2, '127.0.0.1:61116', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731471949, 0, '2024-11-13 12:25:49');
INSERT INTO `login` VALUES (268, 3, '127.0.0.1:56743', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731480153, 0, '2024-11-13 14:42:33');
INSERT INTO `login` VALUES (269, 4, '127.0.0.1:58171', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481114, 0, '2024-11-13 14:58:34');
INSERT INTO `login` VALUES (270, 5, '127.0.0.1:58171', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481180, 0, '2024-11-13 14:59:40');
INSERT INTO `login` VALUES (271, 6, '127.0.0.1:58274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481217, 0, '2024-11-13 15:00:17');
INSERT INTO `login` VALUES (272, 7, '127.0.0.1:58171', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481237, 0, '2024-11-13 15:00:37');
INSERT INTO `login` VALUES (273, 8, '127.0.0.1:58274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481281, 0, '2024-11-13 15:01:21');
INSERT INTO `login` VALUES (274, 9, '127.0.0.1:58171', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481315, 0, '2024-11-13 15:01:55');
INSERT INTO `login` VALUES (275, 10, '127.0.0.1:58274', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731481333, 0, '2024-11-13 15:02:13');
INSERT INTO `login` VALUES (276, 11, '127.0.0.1:61304', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731483174, 0, '2024-11-13 15:32:54');
INSERT INTO `login` VALUES (277, 12, '127.0.0.1:61304', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731483407, 0, '2024-11-13 15:36:47');
INSERT INTO `login` VALUES (278, 13, '127.0.0.1:61776', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731483421, 0, '2024-11-13 15:37:01');
INSERT INTO `login` VALUES (279, 14, '127.0.0.1:63022', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731483943, 0, '2024-11-13 15:45:43');
INSERT INTO `login` VALUES (280, 15, '127.0.0.1:63022', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484052, 0, '2024-11-13 15:47:32');
INSERT INTO `login` VALUES (281, 16, '127.0.0.1:63022', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484121, 0, '2024-11-13 15:48:41');
INSERT INTO `login` VALUES (282, 17, '127.0.0.1:63429', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484175, 0, '2024-11-13 15:49:35');
INSERT INTO `login` VALUES (283, 18, '127.0.0.1:63022', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484191, 0, '2024-11-13 15:49:51');
INSERT INTO `login` VALUES (284, 19, '127.0.0.1:63429', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484296, 0, '2024-11-13 15:51:36');
INSERT INTO `login` VALUES (285, 20, '127.0.0.1:63022', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484307, 0, '2024-11-13 15:51:47');
INSERT INTO `login` VALUES (286, 21, '127.0.0.1:64717', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484752, 0, '2024-11-13 15:59:12');
INSERT INTO `login` VALUES (287, 22, '127.0.0.1:64718', '127.0.0.1', '', '', '', 'aagame', 'yyb', 1731484771, 0, '2024-11-13 15:59:31');

-- ----------------------------
-- Table structure for monster
-- ----------------------------
DROP TABLE IF EXISTS `monster`;
CREATE TABLE `monster`  (
  `id` bigint(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '模型',
  `monster_type` tinyint(255) NOT NULL DEFAULT 0 COMMENT '0-怪物, 1-npc',
  `level` int(255) NOT NULL DEFAULT 1,
  `grade` tinyint(255) NOT NULL COMMENT '级别：0\"普通怪\",1\"小头目\",2\"精英怪\",3\"大BOSS\",\r\n            4\"变态怪\", 5 \"变态怪\"',
  `attr_type` tinyint(255) NOT NULL DEFAULT 0 COMMENT '属性类型:0 力量，1敏捷, 2智慧',
  `base_life` bigint(255) NOT NULL,
  `base_mana` bigint(255) NOT NULL,
  `base_defense` bigint(255) NOT NULL,
  `base_attack` bigint(255) NOT NULL,
  `attach_attack_random` int(255) NOT NULL DEFAULT 0 COMMENT '附带的随机值',
  `strength` bigint(255) NOT NULL,
  `agility` bigint(255) NOT NULL,
  `intelligence` bigint(255) NOT NULL,
  `run_step_time` smallint(6) NOT NULL COMMENT '跑速度',
  `idle_step_time` smallint(6) NOT NULL COMMENT '正常速度',
  `chase_step_time` smallint(6) NOT NULL COMMENT '追击速度',
  `escape_step_time` smallint(6) NOT NULL COMMENT '逃跑速度',
  `attack_range` smallint(255) NOT NULL DEFAULT 5 COMMENT '攻击范围',
  `attack_duration` smallint(255) NOT NULL DEFAULT 0 COMMENT '攻击间隔',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '简介',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of monster
-- ----------------------------
INSERT INTO `monster` VALUES (1, 'm1', 'monster1', 0, 1, 0, 0, 100, 100, 2, 5, 5, 2, 2, 2, 250, 300, 200, 250, 3, 2000, 'monster1', '2024-10-17 18:46:38', '2024-11-13 15:49:16');
INSERT INTO `monster` VALUES (2, 'm2', 'monster2', 0, 2, 0, 1, 150, 120, 4, 10, 0, 3, 3, 3, 250, 300, 200, 250, 3, 2000, 'monster2', '2024-10-25 15:29:50', '2024-11-13 15:49:19');
INSERT INTO `monster` VALUES (3, 'n1', 'npc1', 1, 100, 0, 0, 10000, 10000, 1000, 5000, 0, 200, 200, 200, 250, 300, 200, 250, 3, 1000, 'npc1', '2024-10-17 18:46:38', '2024-11-13 10:43:56');

-- ----------------------------
-- Table structure for online
-- ----------------------------
DROP TABLE IF EXISTS `online`;
CREATE TABLE `online`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_count` int(255) NOT NULL DEFAULT 0,
  `scenes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `time` bigint(20) NOT NULL DEFAULT 0,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of online
-- ----------------------------

-- ----------------------------
-- Table structure for register
-- ----------------------------
DROP TABLE IF EXISTS `register`;
CREATE TABLE `register`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` bigint(20) NOT NULL,
  `remote` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `ip` varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `imei` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `os` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `model` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `appid` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `channel_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `register_type` int(255) NOT NULL,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk`(`appid`, `imei`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 23 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of register
-- ----------------------------
INSERT INTO `register` VALUES (1, 1, '', '', '31d5-2473-48ca-bf44-901c-aaa8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 12:13:02', '2024-11-13 12:13:02');
INSERT INTO `register` VALUES (2, 2, '', '', 'fa79-4666-4579-9806-9df2-9ba9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 12:25:49', '2024-11-13 12:25:49');
INSERT INTO `register` VALUES (3, 3, '', '', 'b777-e00d-4fa9-9c57-63fe-b88b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 14:42:33', '2024-11-13 14:42:33');
INSERT INTO `register` VALUES (4, 4, '', '', '370c-e9e6-4737-9a0e-40c5-89ba', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 14:58:34', '2024-11-13 14:58:34');
INSERT INTO `register` VALUES (5, 5, '', '', '6a88-4cd6-4d65-a3b5-9391-898b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 14:59:40', '2024-11-13 14:59:40');
INSERT INTO `register` VALUES (6, 6, '', '', 'd64a-81b9-4486-bac0-bbd5-b9ab', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:00:17', '2024-11-13 15:00:17');
INSERT INTO `register` VALUES (7, 7, '', '', '2b3e-dab8-4df2-b8d0-3f59-8b89', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:00:37', '2024-11-13 15:00:37');
INSERT INTO `register` VALUES (8, 8, '', '', 'b57a-b316-495a-8d37-d474-b989', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:01:21', '2024-11-13 15:01:21');
INSERT INTO `register` VALUES (9, 9, '', '', '7d59-5045-4a83-9477-6dbb-8abb', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:01:55', '2024-11-13 15:01:55');
INSERT INTO `register` VALUES (10, 10, '', '', 'b78e-966e-47bd-9653-b55b-a98b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:02:13', '2024-11-13 15:02:13');
INSERT INTO `register` VALUES (11, 11, '', '', 'abd5-5f6e-405b-b57b-4cfc-8898', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:32:54', '2024-11-13 15:32:54');
INSERT INTO `register` VALUES (12, 12, '', '', '0cec-fad9-4960-bc24-e0bc-898a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:36:47', '2024-11-13 15:36:47');
INSERT INTO `register` VALUES (13, 13, '', '', '00a7-71d8-4a70-b1e9-8285-a89b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:37:01', '2024-11-13 15:37:01');
INSERT INTO `register` VALUES (14, 14, '', '', '6ae4-8592-4622-8516-94a7-998a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:45:43', '2024-11-13 15:45:43');
INSERT INTO `register` VALUES (15, 15, '', '', '40ba-5581-4ff5-9508-4630-8ab9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:47:32', '2024-11-13 15:47:32');
INSERT INTO `register` VALUES (16, 16, '', '', '31ae-a0a2-4e90-b815-770c-88aa', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:48:41', '2024-11-13 15:48:41');
INSERT INTO `register` VALUES (17, 17, '', '', '2ead-fffd-4f46-8d8d-f551-b899', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:49:35', '2024-11-13 15:49:35');
INSERT INTO `register` VALUES (18, 18, '', '', '79ae-743e-4bbc-86da-f8cd-9bba', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:49:51', '2024-11-13 15:49:51');
INSERT INTO `register` VALUES (19, 19, '', '', '4b15-fe22-4ef8-9ba3-9930-99a9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:51:36', '2024-11-13 15:51:36');
INSERT INTO `register` VALUES (20, 20, '', '', 'abce-aef0-46a6-9858-d7ce-898a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:51:47', '2024-11-13 15:51:47');
INSERT INTO `register` VALUES (21, 21, '', '', '5d74-20cd-44d3-beeb-c185-88b9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:59:12', '2024-11-13 15:59:12');
INSERT INTO `register` VALUES (22, 22, '', '', '4eea-5479-449f-b607-0796-89ba', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-13 15:59:31', '2024-11-13 15:59:31');

-- ----------------------------
-- Table structure for scene
-- ----------------------------
DROP TABLE IF EXISTS `scene`;
CREATE TABLE `scene`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '场景名称',
  `scene_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '场景类型',
  `map_file` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '场景资源名称',
  `enterx` int(255) NOT NULL,
  `entery` int(255) NOT NULL,
  `enterz` int(255) NOT NULL,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of scene
-- ----------------------------
INSERT INTO `scene` VALUES (1, 'xinshoucun', '0', 'xinshoucun', 10, 140, 0, '2024-10-17 18:35:00', '2024-10-17 18:35:00');

-- ----------------------------
-- Table structure for scene_door
-- ----------------------------
DROP TABLE IF EXISTS `scene_door`;
CREATE TABLE `scene_door`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `scene_id` int(11) NOT NULL COMMENT '所在场景',
  `posx` int(255) NOT NULL,
  `posy` int(255) NOT NULL,
  `posz` int(255) NOT NULL DEFAULT 0,
  `target_scene_id` int(11) NOT NULL COMMENT '目标场景',
  `dest_posx` int(255) NOT NULL,
  `dest_posy` int(255) NOT NULL,
  `dest_posz` int(255) NOT NULL,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `scene_id_nk`(`scene_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of scene_door
-- ----------------------------

-- ----------------------------
-- Table structure for scene_monster_config
-- ----------------------------
DROP TABLE IF EXISTS `scene_monster_config`;
CREATE TABLE `scene_monster_config`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `scene_id` int(11) NOT NULL,
  `monster_id` bigint(20) NOT NULL,
  `total` int(11) NOT NULL,
  `reborn` int(11) NOT NULL DEFAULT 60 COMMENT '重生间隔',
  `bornx` int(11) NOT NULL,
  `borny` int(11) NOT NULL,
  `bornz` int(11) NOT NULL,
  `a_range` int(11) NOT NULL COMMENT '活动范围',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `scene_id_nk`(`scene_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of scene_monster_config
-- ----------------------------
INSERT INTO `scene_monster_config` VALUES (1, 1, 1, 10, 60, 15, 140, 0, 50, '2024-10-17 18:47:30', '2024-11-08 18:50:08');
INSERT INTO `scene_monster_config` VALUES (2, 1, 2, 10, 60, 100, 150, 0, 50, '2024-10-25 15:32:54', '2024-11-12 17:26:24');
INSERT INTO `scene_monster_config` VALUES (3, 1, 3, 0, 60, 22, 140, 0, 1, '2024-10-17 18:47:30', '2024-11-08 17:53:13');

-- ----------------------------
-- Table structure for scene_npc_config
-- ----------------------------
DROP TABLE IF EXISTS `scene_npc_config`;
CREATE TABLE `scene_npc_config`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `scene_id` int(11) NOT NULL,
  `npc_id` bigint(20) NOT NULL,
  `bornx` int(255) NOT NULL,
  `borny` int(255) NOT NULL,
  `bornz` int(255) NOT NULL,
  `a_range` int(255) NOT NULL COMMENT '活动范围',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `scene_id_nk`(`scene_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of scene_npc_config
-- ----------------------------

-- ----------------------------
-- Table structure for spell
-- ----------------------------
DROP TABLE IF EXISTS `spell`;
CREATE TABLE `spell`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `fly_animation` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `damage` bigint(255) NOT NULL DEFAULT 0,
  `mana` bigint(255) NOT NULL DEFAULT 0 COMMENT '消耗',
  `fly_step_time` int(11) NOT NULL COMMENT '飞行速度',
  `cd_time` int(11) NOT NULL DEFAULT 0 COMMENT 'cd间隔',
  `buf_id` int(11) NOT NULL DEFAULT 0,
  `is_range_attack` tinyint(255) NOT NULL DEFAULT 0 COMMENT '是否范围攻击',
  `attack_range` smallint(255) NOT NULL DEFAULT 0 COMMENT '攻击范围，单体为0',
  `spell_type` tinyint(255) NOT NULL DEFAULT 0 COMMENT '技能类型: 0 对敌人，1，对自己, 2，对友军',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of spell
-- ----------------------------
INSERT INTO `spell` VALUES (1, '飞火术', 'spell1', 20, 2, 500, 30000, 1, 1, 10, 0, '飞火远程攻击', '2024-11-12 16:04:49', '2024-11-13 16:00:30');
INSERT INTO `spell` VALUES (2, '恢复buf', 'spell2', 0, 10, 0, 20000, 2, 0, 0, 1, '给自己加恢复buf', '2024-11-12 16:20:23', '2024-11-13 16:00:35');

-- ----------------------------
-- Table structure for third_account
-- ----------------------------
DROP TABLE IF EXISTS `third_account`;
CREATE TABLE `third_account`  (
  `id` bigint(20) NOT NULL,
  `third_account` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'wx的openid',
  `uid` bigint(11) UNSIGNED NOT NULL,
  `platform` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'wx',
  `third_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `head_url` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `sex` tinyint(4) NOT NULL DEFAULT 0,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of third_account
-- ----------------------------

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `algo` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `hash` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `role` tinyint(100) NOT NULL COMMENT '账号类型: 1-管理员账号，2- 第三方平台账号',
  `coin` bigint(255) NOT NULL DEFAULT 0 COMMENT '金币',
  `is_online` tinyint(4) NOT NULL DEFAULT 0,
  `salt` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '盐值',
  `last_loginat` bigint(255) NOT NULL DEFAULT 0,
  `priv_keyey` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `pub_key` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `debug` tinyint(255) NOT NULL DEFAULT 1,
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '0禁用，1启用',
  `is_guest` tinyint(255) NOT NULL DEFAULT 0,
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 23 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------
INSERT INTO `user` VALUES (1, '', '', 2, 10, 2, '', 1731471182, '', '', 1, 1, 1, '2024-11-13 12:13:03', '2024-11-13 12:13:03');
INSERT INTO `user` VALUES (2, '', '', 2, 10, 2, '', 1731471949, '', '', 1, 1, 1, '2024-11-13 12:25:49', '2024-11-13 12:25:49');
INSERT INTO `user` VALUES (3, '', '', 2, 10, 2, '', 1731480153, '', '', 1, 1, 1, '2024-11-13 14:42:34', '2024-11-13 14:42:34');
INSERT INTO `user` VALUES (4, '', '', 2, 10, 2, '', 1731481114, '', '', 1, 1, 1, '2024-11-13 14:58:34', '2024-11-13 14:58:34');
INSERT INTO `user` VALUES (5, '', '', 2, 10, 2, '', 1731481180, '', '', 1, 1, 1, '2024-11-13 14:59:40', '2024-11-13 14:59:40');
INSERT INTO `user` VALUES (6, '', '', 2, 10, 2, '', 1731481217, '', '', 1, 1, 1, '2024-11-13 15:00:18', '2024-11-13 15:00:18');
INSERT INTO `user` VALUES (7, '', '', 2, 10, 2, '', 1731481237, '', '', 1, 1, 1, '2024-11-13 15:00:38', '2024-11-13 15:00:38');
INSERT INTO `user` VALUES (8, '', '', 2, 10, 2, '', 1731481281, '', '', 1, 1, 1, '2024-11-13 15:01:21', '2024-11-13 15:01:21');
INSERT INTO `user` VALUES (9, '', '', 2, 10, 2, '', 1731481315, '', '', 1, 1, 1, '2024-11-13 15:01:55', '2024-11-13 15:01:55');
INSERT INTO `user` VALUES (10, '', '', 2, 10, 2, '', 1731481333, '', '', 1, 1, 1, '2024-11-13 15:02:13', '2024-11-13 15:02:13');
INSERT INTO `user` VALUES (11, '', '', 2, 10, 2, '', 1731483174, '', '', 1, 1, 1, '2024-11-13 15:32:54', '2024-11-13 15:32:54');
INSERT INTO `user` VALUES (12, '', '', 2, 10, 2, '', 1731483407, '', '', 1, 1, 1, '2024-11-13 15:36:48', '2024-11-13 15:36:48');
INSERT INTO `user` VALUES (13, '', '', 2, 10, 2, '', 1731483421, '', '', 1, 1, 1, '2024-11-13 15:37:02', '2024-11-13 15:37:02');
INSERT INTO `user` VALUES (14, '', '', 2, 10, 2, '', 1731483943, '', '', 1, 1, 1, '2024-11-13 15:45:44', '2024-11-13 15:45:44');
INSERT INTO `user` VALUES (15, '', '', 2, 10, 2, '', 1731484052, '', '', 1, 1, 1, '2024-11-13 15:47:32', '2024-11-13 15:47:32');
INSERT INTO `user` VALUES (16, '', '', 2, 10, 2, '', 1731484121, '', '', 1, 1, 1, '2024-11-13 15:48:41', '2024-11-13 15:48:41');
INSERT INTO `user` VALUES (17, '', '', 2, 10, 2, '', 1731484175, '', '', 1, 1, 1, '2024-11-13 15:49:36', '2024-11-13 15:49:36');
INSERT INTO `user` VALUES (18, '', '', 2, 10, 2, '', 1731484191, '', '', 1, 1, 1, '2024-11-13 15:49:51', '2024-11-13 15:49:52');
INSERT INTO `user` VALUES (19, '', '', 2, 10, 2, '', 1731484296, '', '', 1, 1, 1, '2024-11-13 15:51:37', '2024-11-13 15:51:37');
INSERT INTO `user` VALUES (20, '', '', 2, 10, 2, '', 1731484307, '', '', 1, 1, 1, '2024-11-13 15:51:47', '2024-11-13 15:51:47');
INSERT INTO `user` VALUES (21, '', '', 2, 10, 2, '', 1731484752, '', '', 1, 1, 1, '2024-11-13 15:59:12', '2024-11-13 15:59:12');
INSERT INTO `user` VALUES (22, '', '', 2, 10, 2, '', 1731484771, '', '', 1, 1, 1, '2024-11-13 15:59:31', '2024-11-13 15:59:31');

SET FOREIGN_KEY_CHECKS = 1;
