INSERT INTO members (email, password, name)
VALUES ('shakya1221@gmail.com', '$2a$10$8cvP4Nv3LdR3J303AQ7NIOSnb1rQaNU/iyo65Gcv/oFSTyP03UodK', 'admin')
ON CONFLICT (email) DO NOTHING;