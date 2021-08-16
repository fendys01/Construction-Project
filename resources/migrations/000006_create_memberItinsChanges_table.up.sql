-- need to copy the latest details before changed by tc / member
create table member_itin_changes (
    id serial PRIMARY KEY,
    member_itin_id int references member_itins(id),
    details json not null,
    changed_by varchar(5), -- owner, tc
    changed_user_id int not null, -- id of members / users(tc)
    created_date timestamptz(0) not NULL
);