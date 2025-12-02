-- Марки
CREATE TABLE brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

-- Модели
CREATE TABLE models (
    id SERIAL PRIMARY KEY,
    brand_id INTEGER NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL
);

-- Автомобили
CREATE TABLE cars (
    id SERIAL PRIMARY KEY,
    model_id INTEGER NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    year INTEGER CHECK (year >= 1900 AND year <= 2100),
    price_thousands INTEGER CHECK (price_thousands > 0)
);

-- Данные: марки
INSERT INTO brands (name) VALUES
('Toyota'), ('BMW'), ('Mercedes'), ('Audi'), ('Ford'),
('Honda'), ('Volkswagen'), ('Nissan'), ('Hyundai'), ('Kia'),
('Tesla'), ('Chevrolet');

-- Данные: модели
INSERT INTO models (brand_id, name) VALUES
(1, 'Camry'), (1, 'Corolla'), (2, 'X5'), (2, '3 Series'),
(3, 'C-Class'), (3, 'E-Class'), (4, 'A4'), (4, 'Q7'),
(5, 'Focus'), (5, 'Mustang'), (6, 'Civic'), (6, 'Accord'),
(7, 'Golf'), (7, 'Passat'), (8, 'Pathfinder');

-- Данные: автомобили
INSERT INTO cars (model_id, year, price_thousands) VALUES
(1, 2022, 2500), (2, 2021, 1800), (3, 2023, 5200), (4, 2022, 3800),
(5, 2021, 3000), (6, 2023, 4500), (7, 2022, 2900), (8, 2021, 4800),
(9, 2020, 1600), (10, 2023, 3500), (11, 2022, 2100), (12, 2021, 2300),
(13, 2020, 1900), (14, 2022, 2700), (15, 2023, 3200);
