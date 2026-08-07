DROP PROCEDURE IF EXISTS schedulerGenerateDailySchedule;

CREATE PROCEDURE schedulerGenerateDailySchedule(
    IN req_date DATE,
    IN start_hour INT,
    IN end_hour INT,
    IN interval_minutes INT,
    IN provider_id INT
)
BEGIN
    DECLARE cur_hour INT DEFAULT start_hour;
    DECLARE cur_minute INT DEFAULT 0;
    DECLARE now_ts DATETIME DEFAULT NOW();

    -- Delete existing open slots for this provider on this date
    DELETE FROM scheduler
    WHERE caldateof = req_date
      AND calphysician = provider_id
      AND caltype = 'open'
      AND calstatus = 'open';

    -- Generate new slots
    WHILE cur_hour < end_hour DO
        SET cur_minute = 0;
        WHILE cur_minute < 60 DO
            INSERT INTO scheduler (
                created_at, updated_at,
                caldateof, calcreated,
                caltype, calhour, calminute, calduration,
                calfacility, calroom,
                calphysician, calpatient,
                calcptcode, calstatus,
                calprenote, calpostnote,
                calmark, calgroupid, calgroupmembers,
                calrecurnote, calrecurid,
                calappttemplate, calattendees,
                user
            ) VALUES (
                now_ts, now_ts,
                req_date, now_ts,
                'open', cur_hour, cur_minute, interval_minutes,
                NULL, NULL,
                provider_id, 0,
                NULL, 'open',
                '', NULL,
                0, 0, NULL,
                NULL, 0,
                0, NULL,
                0
            );
            SET cur_minute = cur_minute + interval_minutes;
        END WHILE;
        SET cur_hour = cur_hour + 1;
    END WHILE;
END;
