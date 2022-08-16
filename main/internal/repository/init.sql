create table if not exists cities
(
    city_id bigint       not null primary key,
    name    varchar(128) not null
);

create table if not exists interests
(
    interest_id bigint       not null primary key,
    name        varchar(256) not null
);

create table if not exists profile
(
    profile_id bigint      not null primary key AUTO_INCREMENT,
    name       varchar(32) not null default '',
    surname    varchar(64) not null default '',
    city_id    bigint      not null default 0,
    birth_date datetime    not null default CURRENT_TIMESTAMP
);

create index profile_name_surname_idx on profile (name, surname);

create table if not exists profile_interests
(
    profile_id  bigint not null,
    interest_id bigint not null,
    primary key (profile_id, interest_id),
    foreign key (profile_id) references profile (profile_id),
    foreign key (interest_id) references interests (interest_id)
);

create table if not exists friends
(
    profile_id       bigint not null,
    other_profile_id bigint not null,
    primary key (profile_id, other_profile_id),
    foreign key (profile_id) references profile (profile_id),
    foreign key (other_profile_id) references profile (profile_id)
);

create table if not exists friendship_application
(
    profile_id       bigint not null,
    other_profile_id bigint not null,
    primary key (profile_id, other_profile_id),
    foreign key (profile_id) references profile (profile_id),
    foreign key (other_profile_id) references profile (profile_id)
);

create table if not exists logins
(
    login         varchar(32)   not null,
    password_hash varbinary(60) not null,
    profile_id    bigint        not null primary key,
    foreign key (profile_id) references profile (profile_id)
);

insert into cities (city_id, name)
values (0, 'undefined'),
       (1, 'Москва'),
       (2, 'Санкт-Петербург'),
       (3, 'Новосибирск'),
       (4, 'Екатеринбург'),
       (5, 'Нижний Новгород'),
       (6, 'Казань'),
       (7, 'Самара'),
       (8, 'Челябинск'),
       (9, 'Омск'),
       (10, 'Ростов-на-Дону'),
       (11, 'Уфа'),
       (12, 'Красноярск'),
       (13, 'Пермь'),
       (14, 'Волгоград'),
       (15, 'Воронеж'),
       (16, 'Саратов'),
       (17, 'Краснодар'),
       (18, 'Тольятти'),
       (19, 'Тюмень'),
       (20, 'Ижевск'),
       (21, 'Барнаул'),
       (22, 'Ульяновск'),
       (23, 'Иркутск'),
       (24, 'Владивосток'),
       (25, 'Ярославль'),
       (26, 'Хабаровск'),
       (27, 'Махачкала'),
       (28, 'Оренбург'),
       (29, 'Томск'),
       (30, 'Новокузнецк')
ON DUPLICATE KEY UPDATE city_id=city_id;

insert into interests (interest_id, name)
values (1, 'Футбол'),
       (2, 'Пиво'),
       (3, 'Зависать с друзьями'),
       (4, 'Книги'),
       (5, 'Рыбалка'),
       (6, 'Мода'),
       (7, 'Красота'),
       (8, 'Фитнес'),
       (9, 'Правильное питание'),
       (10, 'Автомобили'),
       (11, 'Сериалы'),
       (12, 'Кино'),
       (13, 'Музыка'),
       (14, 'Изобразительное искусство'),
       (15, 'Наука'),
       (16, 'Семья'),
       (17, 'Отношения'),
       (18, 'Путешествия'),
       (19, 'Домашние животные')
ON DUPLICATE KEY UPDATE interest_id=interest_id;

create table if not exists subscriptions
(
    subscriber_id bigint not null,
    publisher_id  bigint not null,
    foreign key (subscriber_id) references profile (profile_id),
    foreign key (publisher_id) references profile (profile_id),
    primary key (subscriber_id, publisher_id)
);

create index subscriptions_publisher_id_idx on subscriptions (publisher_id);

create table if not exists updates
(
    update_id    bigint    not null primary key AUTO_INCREMENT,
    publisher_id bigint    not null,
    ts           timestamp not null default CURRENT_TIMESTAMP,
    text         text      not null,
    foreign key (publisher_id) references profile (profile_id)
);
