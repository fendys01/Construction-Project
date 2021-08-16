create table itin_suggestion_tags(
    itin_sug_id int references itin_suggestions(id) not null,
    tag_id int references tags(id) not null
)