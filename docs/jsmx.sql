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

 Date: 01/11/2024 18:53:10
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
INSERT INTO `aiconfig` VALUES (1, 1, 30, 50, 30, 1, '[]', '[]', '2024-10-21 14:27:15', '2024-10-31 11:16:09');
INSERT INTO `aiconfig` VALUES (2, 2, 30, 50, 30, 1, '[]', '[]', '2024-10-21 14:27:15', '2024-10-31 11:16:11');

-- ----------------------------
-- Table structure for buffer_state
-- ----------------------------
DROP TABLE IF EXISTS `buffer_state`;
CREATE TABLE `buffer_state`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
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
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of buffer_state
-- ----------------------------

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
  `strength` bigint(255) NOT NULL DEFAULT 1,
  `agility` bigint(255) NOT NULL DEFAULT 1,
  `intelligence` bigint(255) NOT NULL DEFAULT 1,
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
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci COMMENT = '角色' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of hero
-- ----------------------------

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
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of login
-- ----------------------------

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
INSERT INTO `monster` VALUES (1, 'm1', 'monster1', 0, 1, 0, 0, 100, 100, 2, 5, 2, 2, 2, 250, 300, 200, 250, 5, 1000, 'monster1', '2024-10-17 18:46:38', '2024-10-25 17:37:02');
INSERT INTO `monster` VALUES (2, 'm2', 'monster2', 0, 2, 0, 1, 150, 120, 4, 10, 3, 3, 3, 250, 300, 200, 250, 5, 1000, 'monster2', '2024-10-25 15:29:50', '2024-10-25 17:37:05');
INSERT INTO `monster` VALUES (3, 'npc1', 'npc1', 1, 100, 0, 0, 10000, 10000, 1000, 5000, 200, 200, 200, 250, 300, 200, 250, 5, 1000, 'npc1', '2024-10-17 18:46:38', '2024-10-25 15:30:16');

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
) ENGINE = InnoDB AUTO_INCREMENT = 60 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of register
-- ----------------------------
INSERT INTO `register` VALUES (1, 1, '', '', '9ea3-7cb1-4152-b598-bb69-a8a8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-30 18:34:07', '2024-10-30 18:34:07');
INSERT INTO `register` VALUES (2, 2, '', '', '1878-6b6a-4f52-8762-4f43-98ab', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-30 18:34:30', '2024-10-30 18:34:30');
INSERT INTO `register` VALUES (3, 3, '', '', 'ee20-4fb8-4b05-9834-c69c-9ba8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 11:57:19', '2024-10-31 11:57:19');
INSERT INTO `register` VALUES (4, 4, '', '', 'd8ef-e0fb-447f-9eb9-7e45-8a8b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:16:36', '2024-10-31 18:16:36');
INSERT INTO `register` VALUES (5, 5, '', '', 'b0f0-890c-43f3-8352-f710-b899', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:16:49', '2024-10-31 18:16:49');
INSERT INTO `register` VALUES (6, 6, '', '', 'b51b-7cdc-41a6-98fe-65a9-89a9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:17:42', '2024-10-31 18:17:42');
INSERT INTO `register` VALUES (7, 7, '', '', '8e7a-03f4-443f-a37d-0355-9989', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:19:48', '2024-10-31 18:19:48');
INSERT INTO `register` VALUES (8, 8, '', '', '582b-c69a-49ee-8e77-1801-aa8a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:29:21', '2024-10-31 18:29:21');
INSERT INTO `register` VALUES (9, 9, '', '', '9e05-98da-47a6-9061-001b-aba9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:29:47', '2024-10-31 18:29:47');
INSERT INTO `register` VALUES (10, 10, '', '', 'd633-278a-4cb8-8ae7-8d32-ab9b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:30:28', '2024-10-31 18:30:28');
INSERT INTO `register` VALUES (11, 11, '', '', '23eb-b1eb-4917-b6b7-96d1-9988', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:30:37', '2024-10-31 18:30:37');
INSERT INTO `register` VALUES (12, 12, '', '', '8d7d-e00c-46a9-83b1-37e8-a8a8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:31:23', '2024-10-31 18:31:23');
INSERT INTO `register` VALUES (13, 13, '', '', 'c00a-e986-40ea-8634-682b-b9a9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:35:03', '2024-10-31 18:35:03');
INSERT INTO `register` VALUES (14, 14, '', '', 'f5fe-c461-4abe-83c1-d99e-9bb8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:35:15', '2024-10-31 18:35:15');
INSERT INTO `register` VALUES (15, 15, '', '', '917a-1d5e-46d9-912b-11ea-a8ab', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:35:45', '2024-10-31 18:35:45');
INSERT INTO `register` VALUES (16, 16, '', '', 'a533-d8df-446a-864e-f5c8-9ba9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:36:03', '2024-10-31 18:36:03');
INSERT INTO `register` VALUES (17, 17, '', '', '0ef2-a98b-43c7-8f19-5b2b-b99b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:44:33', '2024-10-31 18:44:33');
INSERT INTO `register` VALUES (18, 18, '', '', 'c22b-edd0-46d7-aebc-2dcd-989b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-10-31 18:44:36', '2024-10-31 18:44:36');
INSERT INTO `register` VALUES (19, 19, '', '', '87d4-0f22-4ebe-81cb-b803-b899', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 12:13:17', '2024-11-01 12:13:17');
INSERT INTO `register` VALUES (20, 20, '', '', '8e43-1db0-44f5-a619-2617-ab8a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 12:26:02', '2024-11-01 12:26:02');
INSERT INTO `register` VALUES (21, 21, '', '', '6c4a-0cf5-49b2-bfd0-7f9e-8a98', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:20:50', '2024-11-01 14:20:50');
INSERT INTO `register` VALUES (22, 22, '', '', 'f642-72a9-4900-adaf-2347-ba99', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:20:58', '2024-11-01 14:20:58');
INSERT INTO `register` VALUES (23, 23, '', '', '70a5-c84a-48b9-85fa-60bc-a999', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:21:06', '2024-11-01 14:21:06');
INSERT INTO `register` VALUES (24, 24, '', '', '0be5-7b5c-48a9-9e26-5f62-ba88', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:23:00', '2024-11-01 14:23:00');
INSERT INTO `register` VALUES (25, 25, '', '', 'c365-2faf-487a-ae20-5c0c-ba8a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:23:55', '2024-11-01 14:23:55');
INSERT INTO `register` VALUES (26, 26, '', '', 'ad98-a85c-4e40-98f9-9708-a9a8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:26:25', '2024-11-01 14:26:25');
INSERT INTO `register` VALUES (27, 27, '', '', '2f65-ed81-43f3-87f1-4133-ba98', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:27:27', '2024-11-01 14:27:27');
INSERT INTO `register` VALUES (28, 28, '', '', '31a3-09eb-4f91-a193-3104-b89b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:27:45', '2024-11-01 14:27:45');
INSERT INTO `register` VALUES (29, 29, '', '', 'd671-b861-4e15-a371-a3c5-b8b8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:34:58', '2024-11-01 14:34:58');
INSERT INTO `register` VALUES (30, 30, '', '', 'f31c-aa48-4663-85a5-8092-9ab9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:35:04', '2024-11-01 14:35:04');
INSERT INTO `register` VALUES (31, 31, '', '', '593d-9e3b-482b-ab46-9822-bb8a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:35:16', '2024-11-01 14:35:16');
INSERT INTO `register` VALUES (32, 32, '', '', '2bd3-b898-4e32-9fcd-431b-8a99', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:56:33', '2024-11-01 14:56:33');
INSERT INTO `register` VALUES (33, 33, '', '', '0295-0a53-460e-b31a-070f-a989', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:56:40', '2024-11-01 14:56:40');
INSERT INTO `register` VALUES (34, 34, '', '', '01b2-a9d9-46c0-8f35-6f4d-89a8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:57:50', '2024-11-01 14:57:50');
INSERT INTO `register` VALUES (35, 35, '', '', 'f9c0-30fb-45d3-b1d9-4ec7-8a99', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 14:58:12', '2024-11-01 14:58:12');
INSERT INTO `register` VALUES (36, 36, '', '', '39dc-70bb-481a-8257-1b69-9baa', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:05:49', '2024-11-01 15:05:49');
INSERT INTO `register` VALUES (37, 37, '', '', 'e0df-57c1-48e1-b67a-8ae6-bb88', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:06:28', '2024-11-01 15:06:28');
INSERT INTO `register` VALUES (38, 38, '', '', '1a45-23fd-45c2-ae6c-1e6f-bb88', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:09:58', '2024-11-01 15:09:58');
INSERT INTO `register` VALUES (39, 39, '', '', 'a57a-b17d-4e98-bee3-f1d1-9989', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:23:18', '2024-11-01 15:23:18');
INSERT INTO `register` VALUES (40, 40, '', '', '84b9-1ef7-4e0b-8db5-774d-bbb9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:23:45', '2024-11-01 15:23:45');
INSERT INTO `register` VALUES (41, 41, '', '', 'a587-3fbf-4147-89ba-82fd-b8bb', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:25:20', '2024-11-01 15:25:20');
INSERT INTO `register` VALUES (42, 42, '', '', '3707-e36c-4146-9016-f50c-b8b8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:25:37', '2024-11-01 15:25:37');
INSERT INTO `register` VALUES (43, 43, '', '', '7f96-70d2-4401-9189-8d3a-aaba', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:26:19', '2024-11-01 15:26:19');
INSERT INTO `register` VALUES (44, 44, '', '', '2dc2-c9c5-4473-8c28-6e07-89b8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:44:44', '2024-11-01 15:44:44');
INSERT INTO `register` VALUES (45, 45, '', '', '7a7a-f0a4-42e5-9cc0-1205-a8a8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:45:03', '2024-11-01 15:45:03');
INSERT INTO `register` VALUES (46, 46, '', '', '1036-03fc-40a8-960b-dd1e-989b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 15:45:25', '2024-11-01 15:45:25');
INSERT INTO `register` VALUES (47, 47, '', '', '9afd-3ca7-4bc9-8ebd-d134-8b8a', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:02:13', '2024-11-01 16:02:13');
INSERT INTO `register` VALUES (48, 48, '', '', '7979-2ed5-4415-ade1-bf9d-a98b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:02:28', '2024-11-01 16:02:28');
INSERT INTO `register` VALUES (49, 49, '', '', '9942-354b-4d7e-b153-cac3-99a8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:03:37', '2024-11-01 16:03:37');
INSERT INTO `register` VALUES (50, 50, '', '', '272f-67d3-4be1-97b5-fe15-b988', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:03:45', '2024-11-01 16:03:45');
INSERT INTO `register` VALUES (51, 51, '', '', '15c5-6dae-4562-865c-2c55-b8b9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:03:58', '2024-11-01 16:03:58');
INSERT INTO `register` VALUES (52, 52, '', '', '0ce1-e832-487c-9f4e-cae8-99a9', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:05:28', '2024-11-01 16:05:28');
INSERT INTO `register` VALUES (53, 53, '', '', 'a8cf-0ff8-4df9-ade9-42b5-b89b', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:08:31', '2024-11-01 16:08:31');
INSERT INTO `register` VALUES (54, 54, '', '', '5d94-9f9e-4df4-8abd-2096-8bba', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:12:11', '2024-11-01 16:12:11');
INSERT INTO `register` VALUES (55, 55, '', '', '23c1-9126-4bea-9e6b-d049-aaa8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:12:35', '2024-11-01 16:12:35');
INSERT INTO `register` VALUES (56, 56, '', '', '3822-29af-47dc-869a-340b-b898', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:44:38', '2024-11-01 16:44:38');
INSERT INTO `register` VALUES (57, 57, '', '', '5302-2688-4183-beb4-63ea-99ab', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 16:45:00', '2024-11-01 16:45:00');
INSERT INTO `register` VALUES (58, 58, '', '', 'b715-3a0f-4c09-af0e-4f70-bbb8', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 18:27:50', '2024-11-01 18:27:50');
INSERT INTO `register` VALUES (59, 59, '', '', 'fa09-0055-44ca-a08d-4567-8989', '14.1', 'Iphone', 'aagame', 'yyb', 5, '2024-11-01 18:37:31', '2024-11-01 18:37:31');

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
  `reborn` int(11) NOT NULL DEFAULT 60,
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
INSERT INTO `scene_monster_config` VALUES (1, 1, 1, 1000, 60, 15, 140, 0, 50, '2024-10-17 18:47:30', '2024-11-01 18:35:32');
INSERT INTO `scene_monster_config` VALUES (2, 1, 2, 500, 60, 100, 150, 0, 50, '2024-10-25 15:32:54', '2024-11-01 18:35:20');
INSERT INTO `scene_monster_config` VALUES (3, 1, 3, 1, 60, 22, 140, 0, 1, '2024-10-17 18:47:30', '2024-11-01 17:54:46');

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
  `animation` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `damage` bigint(255) NOT NULL DEFAULT 0,
  `mana` bigint(255) NOT NULL DEFAULT 0 COMMENT '消耗',
  `step_time` int(11) NOT NULL COMMENT '飞行速度',
  `cd_time` int(11) NOT NULL DEFAULT 0 COMMENT 'cd间隔',
  `buf_id` int(11) NOT NULL DEFAULT 0,
  `is_range_attack` tinyint(255) NOT NULL DEFAULT 0 COMMENT '是否范围攻击',
  `attack_range` smallint(255) NOT NULL DEFAULT 0 COMMENT '攻击范围，单体为0',
  `create_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
  `update_at` datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0),
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of spell
-- ----------------------------

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
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of user
-- ----------------------------

SET FOREIGN_KEY_CHECKS = 1;
