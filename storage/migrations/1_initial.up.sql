create type resource_type as enum (
    'jira',
    'clickup'
    );

create table accounts
(
    id serial primary key,
    slack_channel varchar(10) not null,
    props jsonb,
    resource resource_type not null,
    create_at timestamp default now() not null,
    update_at timestamp,
    delete_at timestamp
);

create index resource_type_index
    on accounts (resource);

alter table accounts owner to root;
