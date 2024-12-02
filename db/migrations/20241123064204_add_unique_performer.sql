-- migrate:up
alter table performers
add unique (name);


-- migrate:down
alter table performers
drop constraint performers_name_key;
