CREATE TABLE IF NOT EXISTS auto_type (
    id VARCHAR(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS  auto (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) REFERENCES auto_type (id) NOT NULL,
    availability BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS commission_type (
    id VARCHAR(255) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS rent_threshold (
    auto_type VARCHAR(255) REFERENCES auto_type (id) PRIMARY KEY,
    min_threshold INTEGER NOT NULL,
    max_threshold INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS commission (
    auto_type VARCHAR(255) REFERENCES auto_type (id) NOT NULL,
    type VARCHAR(255) REFERENCES commission_type (id) NOT NULL,
    value INTEGER NOT NULL,
    min_threshold INTEGER
);

CREATE INDEX IF NOT EXISTS idx_commission ON commission (auto_type, type);

CREATE TABLE IF NOT EXISTS auto_rent (
    auto_id VARCHAR(255) REFERENCES auto (id) UNIQUE NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);

insert into auto_type (id) values ('standard');
insert into auto_type (id) values ('special');

insert into auto (id, type, availability) values ('MINI-COOPER-SE', 'standard', true);
insert into auto (id, type, availability) values ('John-Deere-1050K', 'special', true);

insert into commission_type (id) values ('daily');
insert into commission_type (id) values ('agreement');
insert into commission_type (id) values ('weekend');
insert into commission_type (id) values ('penalty');
insert into commission_type (id) values ('insurance');

insert into commission (auto_type, type, value, min_threshold) values ('standard', 'daily', 50, 0);
insert into commission (auto_type, type, value, min_threshold) values ('special', 'daily', 200, 0);
insert into commission (auto_type, type, value, min_threshold) values ('special', 'agreement', 200, 0);
insert into commission (auto_type, type, value, min_threshold) values ('special', 'weekend', 20, 0);
insert into commission (auto_type, type, value, min_threshold) values ('special', 'penalty', 5, 10);
insert into commission (auto_type, type, value, min_threshold) values ('standard', 'insurance', 133, 0);

insert into rent_threshold (auto_type, min_threshold, max_threshold) values ('standard', 0, 1826);
insert into rent_threshold (auto_type, min_threshold, max_threshold) values ('special', 10, 90);
