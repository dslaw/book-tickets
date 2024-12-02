-- migrate:up
-- NB: Events are soft-deleted, so omitting an `on delete cascade` clause.
alter table event_performers
add constraint event_performers_event_id_fkey
    foreign key (event_id)
    references events (id);

alter table event_performers
add constraint event_performers_performer_id_fkey
    foreign key (performer_id)
    references performers (id)
    on delete cascade;

-- migrate:down
alter table event_performers
drop constraint event_performers_performer_id_fkey;

alter table event_performers
drop constraint event_performers_event_id_fkey;
