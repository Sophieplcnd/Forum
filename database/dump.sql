PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "Users" ("ID" INTEGER PRIMARY KEY, "Email" TEXT UNIQUE NOT NULL, "Username" TEXT UNIQUE NOT NULL, "Password" TEXT NOT NULL, SessionID TEXT);
CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    post_id INTEGER,
    content TEXT NOT NULL,
    commentlikes_count INTEGER DEFAULT (0),
    commentdislikes_count INTEGER DEFAULT (0),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(post_id) REFERENCES posts(id)
);
CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS "reactions" (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    post_id INTEGER,
    comment_id INTEGER,
    type INTEGER ,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(post_id) REFERENCES posts(id),
    FOREIGN KEY(comment_id) REFERENCES comments(id)
);
CREATE TABLE IF NOT EXISTS "posts" (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP, "likes_count" INTEGER DEFAULT (0), "dislikes_count" INTEGER DEFAULT (0),
    FOREIGN KEY(user_id) REFERENCES users(id)
);

INSERT INTO categories (name) VALUES ('lifestyle');
INSERT INTO categories (name) VALUES ('news');
INSERT INTO categories (name) VALUES ('gaming');
INSERT INTO categories (name) VALUES ('fashion');
INSERT INTO categories (name) VALUES ('music');
INSERT INTO categories (name) VALUES ('tv-movies');

CREATE TABLE IF NOT EXISTS "categories_posts" (
    id INTEGER PRIMARY KEY,
    category_id INTEGER,
    post_id INTEGER,
    FOREIGN KEY(category_id) REFERENCES categories(id),
    FOREIGN KEY(post_id) REFERENCES posts(id)
);

CREATE TABLE IF NOT EXISTS "postlikes" (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    post_id INTEGER,
    type INTEGER,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(post_id) REFERENCES posts(id)
);



COMMIT;
