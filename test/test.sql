-- password = 12345678
INSERT INTO accounts (balance, identified, name, phone, password) VALUES 
(0, FALSE, '0', '0', '$2a$10$W1uTjnpz.h/hbfWuRhO04ekfs6FffeMsIbtFpxLiFhE6eMgW7oMUi'),
(1000000, FALSE, '1', '1', '$2a$10$W1uTjnpz.h/hbfWuRhO04ekfs6FffeMsIbtFpxLiFhE6eMgW7oMUi'),
(10000000, TRUE, '2', '2', '$2a$10$W1uTjnpz.h/hbfWuRhO04ekfs6FffeMsIbtFpxLiFhE6eMgW7oMUi');

INSERT INTO transactions (acc_id, amount) VALUES 
(1, 1000000),
(2, 10000000);