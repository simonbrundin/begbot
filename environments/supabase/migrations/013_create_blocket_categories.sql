-- Create blocket_categories table
CREATE TABLE IF NOT EXISTS blocket_categories (
    id SERIAL PRIMARY KEY,
    blocket_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id VARCHAR(50),
    llm_category VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_blocket_categories_llm_category ON blocket_categories(llm_category);

-- Seed with common categories based on LLM categories
INSERT INTO blocket_categories (blocket_id, name, llm_category) VALUES
-- Telefoner och tillbehör
('2.93.3217.39', 'Mobiltelefoner', 'phone'),
('2.93.3217.42', 'Smartwatch och aktivitetsarmband', 'watch'),
('2.93.3217.40', 'Telefontillbehör', 'accessory'),

-- Datorer
('2.93.3215', 'Datorer', 'computer'),
('2.93.3215.42', 'Laptops', 'computer'),
('2.93.3215.43', 'Stationära datorer', 'computer'),
('2.93.3215.44', 'MacBooks', 'computer'),
('2.93.3215.45', 'PC-komponenter', 'component'),

-- Tablet
('2.93.3217.38', 'Surfplattor', 'tablet'),
('2.93.3217.41', 'Surflattetstillbehör', 'accessory'),

-- TV, Ljud och Bild
('2.93.3906', 'Ljud och bild', 'headphones'),
('2.93.3906.55', 'Hörlurar och headset', 'headphones'),
('2.93.3906.56', 'Högtalare', 'headphones'),
('2.93.3906.57', 'TV', 'other'),

-- Gaming
('2.93.3905', 'TV-spel och spelkonsoler', 'other'),
('2.93.3905.63', 'PlayStation', 'other'),
('2.93.3905.64', 'Xbox', 'other'),
('2.93.3905.65', 'Nintendo', 'other'),

-- Övriga kategorier
('2.93.3217', 'Telefoner och tillbehör', 'accessory'),
('2.93.3904', 'Foto och video', 'other')
ON CONFLICT (blocket_id) DO NOTHING;
