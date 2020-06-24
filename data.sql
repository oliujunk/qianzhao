/*
 Navicat Premium Data Transfer

 Source Server         : data
 Source Server Type    : SQLite
 Source Server Version : 3030001
 Source Schema         : main

 Target Server Type    : SQLite
 Target Server Version : 3030001
 File Encoding         : 65001

 Date: 23/06/2020 09:08:08
*/

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for config
-- ----------------------------
DROP TABLE IF EXISTS "config";
CREATE TABLE "config" (
  "id" INTEGER NOT NULL,
  "device_name" text(255),
  "device_code" text(5) NOT NULL,
  "item_code" text(4) NOT NULL,
  "serial_number" text(4) NOT NULL,
  "element_name" text(255),
  "element_num" text(255),
  "element_code" text(255),
  "longitude" text(8),
  "latitude" text(8),
  "elevation" text(8),
  "software_version" text(8),
  "IP" text(15),
  "mask" text(15),
  "gateway" text(15),
  "http_port" integer(5),
  "ftp_port" integer(5),
  "command_port" integer(5),
  "management_ip" text(15),
  "management_port" integer(5),
  "sntp_ip" text(15),
  "sample" integer(2),
  PRIMARY KEY ("id")
);

-- ----------------------------
-- Records of config
-- ----------------------------
INSERT INTO "config" VALUES (1, 789, 88889, 'X232', '0000', '-/-/-/-/-/-/-/-/-/-/-/-/-/-/-/-', '101/102/100/100/100/100/100/100/100/100/100/100/100/100/100/100', '2321/2322/-/-/-/-/-/-/-/-/-/-/-/-/-/-', NULL, NULL, NULL, NULL, '192.168.1.254', NULL, NULL, 80, NULL, NULL, NULL, NULL, NULL, NULL);

-- ----------------------------
-- Table structure for data
-- ----------------------------
DROP TABLE IF EXISTS "data";
CREATE TABLE "data" (
  "timestamp" integer(20) NOT NULL,
  "e1" integer(6) NOT NULL,
  "e2" integer(6) NOT NULL,
  "e3" integer(6) NOT NULL,
  "e4" integer(6) NOT NULL,
  "e5" integer(6) NOT NULL,
  "e6" integer(6) NOT NULL,
  "e7" integer(6) NOT NULL,
  "e8" integer(6) NOT NULL,
  "e9" integer(6) NOT NULL,
  "e10" integer(6) NOT NULL,
  "e11" integer(6) NOT NULL,
  "e12" integer(6) NOT NULL,
  "e13" integer(6) NOT NULL,
  "e14" integer(6) NOT NULL,
  "e15" integer(6) NOT NULL,
  "e16" integer(6) NOT NULL
);

-- ----------------------------
-- Records of data
-- ----------------------------

-- ----------------------------
-- Table structure for element
-- ----------------------------
DROP TABLE IF EXISTS "element";
CREATE TABLE "element" (
  "index" text(3) NOT NULL,
  "name" text(10) NOT NULL,
  "unit" text(10) NOT NULL,
  "min" integer(10) NOT NULL,
  "max" integer(10) NOT NULL,
  "prec" real(10) NOT NULL,
  "decimal" integer(1)
);

-- ----------------------------
-- Records of element
-- ----------------------------
INSERT INTO "element" VALUES (100, '未定义', '-', 0, 10, 1.0, 0);
INSERT INTO "element" VALUES (101, '大气温度', '℃', -50, 100, 0.1, 1);
INSERT INTO "element" VALUES (102, '大气湿度', '%RH', 0, 100, 0.1, 1);
INSERT INTO "element" VALUES (103, '模拟气压', 'hPa', 500, 1500, 0.1, 1);
INSERT INTO "element" VALUES (104, '雨量', 'mm', 0, 100, 0.1, 1);
INSERT INTO "element" VALUES (105, '简易总辐射', 'W/m2', 0, 1500, 1.0, 0);
INSERT INTO "element" VALUES (106, '土壤温度', '℃', -50, 100, 0.1, 1);
INSERT INTO "element" VALUES (107, '土壤湿度', '%RH', 0, 100, 0.1, 1);
INSERT INTO "element" VALUES (108, '风速', 'm/s', 0, 70, 0.1, 1);
INSERT INTO "element" VALUES (109, '风向', '°', 0, 360, 1.0, 0);
INSERT INTO "element" VALUES (110, '蒸发', 'mm', 0, 1000, 0.1, 1);
INSERT INTO "element" VALUES (111, '雪量', 'mm', 0, 1000, 1.0, 0);
INSERT INTO "element" VALUES (112, '照度', 'Lux', 0, 200000, 10.0, 0);
INSERT INTO "element" VALUES (113, '日照时数', 'h', 0, 24, 0.1, 1);
INSERT INTO "element" VALUES (114, '光合', 'W/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (115, '雨量累计', 'mm', 0, 6000, 0.1, 1);
INSERT INTO "element" VALUES (116, '辐射累计', 'MJ/m2', 0, 1000, 0.01, 2);
INSERT INTO "element" VALUES (117, '有无雨雪', '', 0, 1, 1.0, 0);
INSERT INTO "element" VALUES (118, '噪声', 'dB', 0, 100, 0.1, 1);
INSERT INTO "element" VALUES (119, '水位', 'cm', 0, 10000, 0.1, 1);
INSERT INTO "element" VALUES (120, '二氧化碳', 'PPM', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (121, '曝辐量', 'cal/cm2', 0, 3, 1.0, 0);
INSERT INTO "element" VALUES (122, '液位', 'mm', 0, 1000, 0.1, 1);
INSERT INTO "element" VALUES (123, '光合有效辐射', 'W/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (124, '电压', 'V', 0, 15, 0.001, 3);
INSERT INTO "element" VALUES (125, '紫外线', 'mW/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (126, '粉尘', 'ug/m3', 0, 4, 0.1, 1);
INSERT INTO "element" VALUES (127, '数字气压', 'hPa', 10, 1100, 0.1, 1);
INSERT INTO "element" VALUES (142, '电流', 'mA', 0, 5000, 0.1, 1);
INSERT INTO "element" VALUES (129, '最大风速', 'm/s', 0, 70, 0.1, 1);
INSERT INTO "element" VALUES (130, '平均风速', 'm/s', 0, 70, 0.1, 1);
INSERT INTO "element" VALUES (131, '经度', '°', -180, 180, 0.1, 1);
INSERT INTO "element" VALUES (132, '纬度', '°', -90, 90, 0.1, 1);
INSERT INTO "element" VALUES (133, '海拔高度', 'm', -2000, 9000, 0.1, 1);
INSERT INTO "element" VALUES (134, 'TBQ总辐射', 'W/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (135, '直接辐射', 'W/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (136, '散射辐射', 'W/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (138, '紫外辐射', 'W/m2', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (139, '贴片温度', '℃', -50, 100, 0.1, 1);
INSERT INTO "element" VALUES (140, '露点温度', '℃', -50, 40, 0.1, 1);
INSERT INTO "element" VALUES (141, '一氧化碳', 'PPM', 0, 1000, 1.0, 0);
INSERT INTO "element" VALUES (128, 'pH值', '', 0, 14, 0.01, 2);
INSERT INTO "element" VALUES (143, '超声波风速', 'm/s', 0, 60, 0.01, 2);
INSERT INTO "element" VALUES (144, '水温', '℃', -50, 100, 0.1, 1);
INSERT INTO "element" VALUES (145, 'PM2.5', 'ug/m3', 0, 10, 0.1, 1);
INSERT INTO "element" VALUES (146, 'PM10', 'ug/m3', 0, 10, 0.1, 1);
INSERT INTO "element" VALUES (152, '能见度', ' m', 0, 50000, 0.1, 1);
INSERT INTO "element" VALUES (147, 0.1, ' ', 0, 10000, 0.1, 1);
INSERT INTO "element" VALUES (148, 0.01, ' ', 0, 10000, 0.01, 2);
INSERT INTO "element" VALUES (149, 0.001, ' ', 0, 10000, 0.001, 3);
INSERT INTO "element" VALUES (150, 'PM2.5', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (151, 'PM10', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (153, '负氧离子', '个', 0, 5000, 10.0, 0);
INSERT INTO "element" VALUES (154, '盐分', 'mg/L', 0, 15000, 1.0, 0);
INSERT INTO "element" VALUES (155, '电导率', 'mS/cm', 0, 20, 0.01, 2);
INSERT INTO "element" VALUES (156, 'SO2', 'ug/m3', 0, 20000, 1.0, 0);
INSERT INTO "element" VALUES (157, 'CO', 'mg/m3', 0, 2000, 0.01, 2);
INSERT INTO "element" VALUES (158, 'NO2', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (159, 'O3', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (160, '管道流量', 'm3/s', 0, 1000, 0.01, 2);
INSERT INTO "element" VALUES (161, '流速', 'L/min', 0, 10000, 1.0, 0);
INSERT INTO "element" VALUES (162, '管道压力', 'KPa', 0, 10000, 1.0, 0);
INSERT INTO "element" VALUES (163, '温差', '℃', -150, 150, 0.1, 1);
INSERT INTO "element" VALUES (164, '溶解氧', 'mg/L', 0, 20, 0.01, 2);
INSERT INTO "element" VALUES (165, '溶解氧差', 'mg/L', -20, 20, 0.01, 2);
INSERT INTO "element" VALUES (166, '氨氮', 'mg/L', 0, 2000, 0.1, 1);
INSERT INTO "element" VALUES (167, '负氧离子', '个', 0, 50000, 1.0, 0);
INSERT INTO "element" VALUES (168, 'TSP', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (169, '水位', 'm', 0, 300, 0.01, 2);
INSERT INTO "element" VALUES (170, '浊度', ' npu', 0, 20, 0.1, 1);
INSERT INTO "element" VALUES (171, 0.016, ' ', 0, 1000, 0.016, 3);
INSERT INTO "element" VALUES (172, 0.019, ' ', 0, 1000, 0.019, 3);
INSERT INTO "element" VALUES (173, 0.2, ' ', 0, 1000, 0.2, 1);
INSERT INTO "element" VALUES (174, 0.3, ' ', 0, 1000, 0.3, 1);
INSERT INTO "element" VALUES (175, 0.5, ' ', 0, 1000, 0.5, 1);
INSERT INTO "element" VALUES (176, 0.7, ' ', 0, 1000, 0.7, 1);
INSERT INTO "element" VALUES (177, 0.8, ' ', 0, 1000, 0.8, 1);
INSERT INTO "element" VALUES (178, 0.9, ' ', 0, 1000, 0.9, 1);
INSERT INTO "element" VALUES (179, 0.022, ' ', 0, 1000, 0.022, 3);
INSERT INTO "element" VALUES (180, 0.025, ' ', 0, 1000, 0.025, 3);
INSERT INTO "element" VALUES (181, 'SO2', 'ppb', 0, 20000, 1.0, 0);
INSERT INTO "element" VALUES (182, 'CO', 'ppb', 0, 20000, 1.0, 0);
INSERT INTO "element" VALUES (183, 'NO2', 'ppb', 0, 20000, 1.0, 0);
INSERT INTO "element" VALUES (184, 'O3', 'ppb', 0, 20000, 1.0, 0);
INSERT INTO "element" VALUES (185, 'NH3', 'ppm', 0, 5000, 0.01, 2);
INSERT INTO "element" VALUES (186, 'H2S', 'ppm', 0, 5000, 0.01, 2);
INSERT INTO "element" VALUES (187, '氧气', '%Vol', 0, 1000, 0.1, 1);
INSERT INTO "element" VALUES (188, '声波液位', 'mm', 0, 30000, 0.4, 1);
INSERT INTO "element" VALUES (189, '5min-PM25', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (191, '水位', 'm', 0, 10000, 0.1, 1);
INSERT INTO "element" VALUES (190, '5min-PM10', 'ug/m3', 0, 2000, 1.0, 0);
INSERT INTO "element" VALUES (192, '归零', 0, 0, 0, 0.0, 0);
INSERT INTO "element" VALUES (193, 'CO', 'mg/m3', 0, 2000, 0.1, 1);
INSERT INTO "element" VALUES (194, 'CH4', '%LEL', 0, 100, 0.1, 1);
INSERT INTO "element" VALUES (195, 'CH4S', 'mg/m3', 0, 2000, 0.001, 3);
INSERT INTO "element" VALUES (196, '能见度', 'm', 0, 60000, 1.0, 0);
INSERT INTO "element" VALUES (197, 0.105, '℃', -500, 10000, 0.105, 3);
INSERT INTO "element" VALUES (198, '土壤盐分', '', 0, 10000, 1.0, 0);
INSERT INTO "element" VALUES (199, '累积流量', 'm3', 0, 60000, 0.01, 2);
INSERT INTO "element" VALUES (200, '菌落总数', 'CFU', 0, 800, 1.0, 0);
INSERT INTO "element" VALUES (201, '土壤热通量', 'w/m2', -500, 500, 0.1, 1);
INSERT INTO "element" VALUES (202, '液位', 'mm', 0, 5000, 1.0, 0);
INSERT INTO "element" VALUES (203, '液位', 'mm', 0, 20000, 1.0, 0);
INSERT INTO "element" VALUES (204, '氨氮', 'ppm', 0, 300, 0.1, 1);
INSERT INTO "element" VALUES (205, '1.2PM', 'ug', 0, 1000, 1.2, 1);
INSERT INTO "element" VALUES (206, '1.5PM', 'ug', 0, 1000, 1.5, 1);
INSERT INTO "element" VALUES (207, '毫米汞柱', 'mmHg', 0, 2000, 0.1, 1);
INSERT INTO "element" VALUES (208, '照度', 'Lux', 0, 200000, 1.0, 0);
INSERT INTO "element" VALUES (209, '电压', 'V', 0, 100000, 0.1, 1);
INSERT INTO "element" VALUES (210, '电流', 'A', 0, 100000, 0.1, 1);
INSERT INTO "element" VALUES (211, 'ORP', 'mV', -10000, 10000, 0.01, 2);
INSERT INTO "element" VALUES (212, '电压', 'V', 0, 10000, 1.0, 0);
INSERT INTO "element" VALUES (213, '电流', 'A', 0, 10000, 1.0, 0);
INSERT INTO "element" VALUES (214, '功率', 'W', 0, 10000, 1.0, 0);
INSERT INTO "element" VALUES (215, '流量', 'm3/h', 0, 10000, 0.01, 2);
INSERT INTO "element" VALUES (216, '累计流量', 'm3', 0, 30000, 1.0, 0);

PRAGMA foreign_keys = true;
