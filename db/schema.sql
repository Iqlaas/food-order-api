CREATE TYPE item_availability AS ENUM ('in_stock', 'out_of_stock');
CREATE TYPE order_status AS ENUM ('received', 'preparing', 'ready', 'completed');

CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    price_cents INT NOT NULL, -- Storing currency as integers (cents) avoids floating-point issues
    availability item_availability DEFAULT 'in_stock',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    total_cents INT NOT NULL,
    status order_status DEFAULT 'received',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id INT REFERENCES menu_items(id),
    quantity INT NOT NULL,
    price_at_purchase_cents INT NOT NULL
);

-- Seed Initial Data for the Single Merchant
INSERT INTO menu_items (name, category, price_cents, availability) VALUES
('Classic Cheeseburger', 'Burgers', 1299, 'in_stock'),
('Truffle Fries', 'Sides', 550, 'in_stock'),
('Iced Matcha Latte', 'Beverages', 650, 'in_stock'),
('Red Velvet Cupcake', 'Desserts', 400, 'out_of_stock');