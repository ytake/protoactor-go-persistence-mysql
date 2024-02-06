CREATE TABLE `tests`
(
    `id`          varchar(26) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
    `payload`     json                                                  NOT NULL,
    `event_index` bigint                                                 DEFAULT NULL,
    `actor_name`  varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
    `event_name`  varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin  DEFAULT NULL,
    `created_at`  timestamp                                              DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uidx_id` (`id`),
    UNIQUE KEY `uidx_names` (`actor_name`,`event_name`,`event_index`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin
