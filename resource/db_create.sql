
CREATE TABLE IF NOT EXISTS fotadevicelogs (
    processid TEXT NOT NULL,
    deviceid TEXT NOT NULL,
    fileid TEXT NOT NULL,
    loglevel TEXT CHECK (loglevel IN ('INFO', 'WARN', 'ERROR', 'DEBUG')) NOT NULL,
    status TEXT NOT NULL,
    createdby TEXT NOT NULL, 
    errormessage TEXT, 
    metadata JSONB DEFAULT '{}',
    createdat BIGINT DEFAULT CAST(
        extract(
            epoch FROM NOW()
        ) * 1000 AS BIGINT
    ) NOT NULL,
    PRIMARY KEY (processid, deviceid, fileid, createdat)
);



CREATE TABLE IF NOT EXISTS devicefotalogs (
    processid TEXT NOT NULL,
    deviceid TEXT NOT NULL,
    fileid TEXT NOT NULL,
    loglevel TEXT CHECK (loglevel IN ('INFO', 'WARN', 'ERROR', 'DEBUG')) NOT NULL,
    status TEXT NOT NULL,
    createdby TEXT NOT NULL ,
    errormessage TEXT, 
    createdat BIGINT DEFAULT CAST(
        extract(
            epoch FROM NOW()
        ) * 1000 AS BIGINT
    ) NOT NULL,
    PRIMARY KEY (processid, fileid, deviceid, createdat)
);

-- table with client

CREATE TABLE IF NOT EXISTS fotadevicelogs (
    processid TEXT NOT NULL,
    processname TEXT NOT NULL,
    deviceid TEXT NOT NULL,
    fileid TEXT NOT NULL,
    loglevel TEXT NOT NULL,
    status TEXT NOT NULL,
    errormesage TEXT, 
    metadata JSONB DEFAULT '{}',
    createdby TEXT NOT NULL, 
    createdat BIGINT DEFAULT CAST(
        extract(
            epoch FROM NOW()
        ) * 1000 AS BIGINT
    ) NOT NULL,
    PRIMARY KEY (processid, deviceid, fileid , processname, createdat)
);