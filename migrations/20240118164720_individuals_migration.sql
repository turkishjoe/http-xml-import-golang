-- +goose Up
-- +goose StatementBegin
create table individuals
(
    id         bigint
        constraint individuals_pk
            primary key,
    first_name varchar(300),
    last_name  varchar(300)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table individuals
-- +goose StatementEnd
