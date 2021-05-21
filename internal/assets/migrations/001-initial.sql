-- +migrate Up
create table users
(
    id      bigserial,
    address text not null,
    amount  bigint,
    denom   text,
    UNIQUE (address, denom),
    PRIMARY KEY (id)
);

create table transfers
(
    id      bigserial,
    address text not null,
    amount  bigint,
    denom   text,
    status  text default 'not_sent',
    user_id bigint references users (id),
    PRIMARY KEY (id)
);

-- +migrate Down
drop table users cascade;
drop table transfers cascade;