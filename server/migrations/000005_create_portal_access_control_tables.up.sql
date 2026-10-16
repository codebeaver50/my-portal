-- ポータル共通のアクセス管理（RBAC）。feature横断のため、テーブル名はfeature-slugではなく portal_ を付ける。
-- ユーザーにロールを割り当て（多対多）、ロールごとにリソース（feature等）への権限レベルを持たせる。

-- ログインするユーザー。password_hash は bcrypt。
-- failed_login_count / locked_until は連続してログインに失敗したときのアカウントロックに使う。
CREATE TABLE portal_users (
    id INT NOT NULL AUTO_INCREMENT,
    login_id VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    failed_login_count INT NOT NULL DEFAULT 0,
    locked_until DATETIME(3) NULL,
    last_login_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    UNIQUE KEY portal_users_login_id_key (login_id)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ロール。code はプログラムから参照する固定の識別子（例: guest は未ログインの訪問者に適用する）。
-- is_system のロール（admin / guest）は削除できない。
CREATE TABLE portal_roles (
    id INT NOT NULL AUTO_INCREMENT,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    UNIQUE KEY portal_roles_code_key (code)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE portal_user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (user_id, role_id),
    KEY portal_user_roles_role_id_idx (role_id),
    CONSTRAINT portal_user_roles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES portal_users (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT portal_user_roles_role_id_fkey
        FOREIGN KEY (role_id) REFERENCES portal_roles (id)
        ON DELETE CASCADE ON UPDATE CASCADE
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ロールごとのリソースへの権限。resource はfeature-slug（例: book-database）または access-control（アクセス管理画面）。
-- write は read を含む。行がないリソースは権限なし（none）として扱う。
CREATE TABLE portal_role_permissions (
    role_id INT NOT NULL,
    resource VARCHAR(50) NOT NULL,
    level ENUM('read', 'write') NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (role_id, resource),
    CONSTRAINT portal_role_permissions_role_id_fkey
        FOREIGN KEY (role_id) REFERENCES portal_roles (id)
        ON DELETE CASCADE ON UPDATE CASCADE
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- ログインセッション。Cookieにはランダムなトークンを渡し、DBにはそのSHA-256（token_hash）だけを保存する
-- （DBの内容が漏れてもセッションを乗っ取られないようにするため）。行を削除すればそのセッションは即座に無効になる。
CREATE TABLE portal_sessions (
    token_hash CHAR(64) NOT NULL,
    user_id INT NOT NULL,
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    ip_address VARCHAR(45) NOT NULL DEFAULT '',
    expires_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (token_hash),
    KEY portal_sessions_user_id_idx (user_id),
    CONSTRAINT portal_sessions_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES portal_users (id)
        ON DELETE CASCADE ON UPDATE CASCADE
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- シードデータ: 組み込みロールと、訪問者が試せるデモ用アカウント（パスワードはいずれも demo-pass）
INSERT INTO portal_roles (id, code, name, description, is_system) VALUES
    (1, 'admin', '管理者', 'すべてのfeatureの読み書きと、ユーザー・ロールの管理ができる', TRUE),
    (2, 'editor', '編集者', 'すべてのfeatureの読み書きができる', FALSE),
    (3, 'viewer', '閲覧者', 'すべてのfeatureを閲覧できる', FALSE),
    (4, 'guest', 'ゲスト', 'ログインしていない訪問者に適用される', TRUE);

INSERT INTO portal_role_permissions (role_id, resource, level) VALUES
    (1, 'book-database', 'write'),
    (1, 'simple-ledger', 'write'),
    (1, 'simple-cms', 'write'),
    (1, 'form-builder', 'write'),
    (1, 'shop-css', 'read'),
    (1, 'access-control', 'write'),
    (2, 'book-database', 'write'),
    (2, 'simple-ledger', 'write'),
    (2, 'simple-cms', 'write'),
    (2, 'form-builder', 'write'),
    (2, 'shop-css', 'read'),
    (3, 'book-database', 'read'),
    (3, 'simple-ledger', 'read'),
    (3, 'simple-cms', 'read'),
    (3, 'form-builder', 'read'),
    (3, 'shop-css', 'read'),
    (3, 'access-control', 'read'),
    (4, 'book-database', 'read'),
    (4, 'simple-ledger', 'read'),
    (4, 'simple-cms', 'read'),
    (4, 'form-builder', 'read'),
    (4, 'shop-css', 'read');

INSERT INTO portal_users (id, login_id, display_name, password_hash) VALUES
    (1, 'demo-admin', 'デモ管理者', '$2a$12$K1MZqu9dWN4XtOlkkzQN/ucsAWqPU6zH8LCzGOBS7e5kT5gdbQM.O'),
    (2, 'demo-editor', 'デモ編集者', '$2a$12$K1MZqu9dWN4XtOlkkzQN/ucsAWqPU6zH8LCzGOBS7e5kT5gdbQM.O'),
    (3, 'demo-guest', 'デモゲスト', '$2a$12$K1MZqu9dWN4XtOlkkzQN/ucsAWqPU6zH8LCzGOBS7e5kT5gdbQM.O');

INSERT INTO portal_user_roles (user_id, role_id) VALUES
    (1, 1),
    (2, 2),
    (3, 3);
