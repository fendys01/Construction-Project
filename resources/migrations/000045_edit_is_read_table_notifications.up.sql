alter table notifications 
	alter column is_read drop not null,
	alter column is_read drop default;

alter table notifications
	drop column is_read;

alter table notifications
	add column is_read bool not null default false;