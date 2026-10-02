CREATE TABLE form_builder_forms (
    id INT NOT NULL AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id)
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- フォームの項目定義。options は select/radio/checkbox の選択肢（文字列のJSON配列）。
-- layout_row が同じ項目は同じ行に横並びで表示し、行の中の並び順は sort_order で決める。
-- sort_order はフォーム全体での表示順（行ごとに左から右）で、layout_row の昇順と矛盾しない。
CREATE TABLE form_builder_form_fields (
    id INT NOT NULL AUTO_INCREMENT,
    form_id INT NOT NULL,
    label VARCHAR(100) NOT NULL,
    type ENUM('text', 'textarea', 'number', 'email', 'date', 'select', 'radio', 'checkbox') NOT NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    options JSON NOT NULL,
    layout_row INT NOT NULL DEFAULT 0,
    sort_order INT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    KEY form_builder_form_fields_form_id_idx (form_id),
    CONSTRAINT form_builder_form_fields_form_id_fkey
        FOREIGN KEY (form_id) REFERENCES form_builder_forms (id)
        ON DELETE CASCADE ON UPDATE CASCADE
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- フォームへの回答。data は { "<field_id>": 値 } 形式のJSONオブジェクト。
-- 項目ごとに列を持たないため、フォーム定義を変更してもスキーマ変更は不要。
CREATE TABLE form_builder_form_records (
    id INT NOT NULL AUTO_INCREMENT,
    form_id INT NOT NULL,
    data JSON NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    KEY form_builder_form_records_form_id_created_at_idx (form_id, created_at),
    CONSTRAINT form_builder_form_records_form_id_fkey
        FOREIGN KEY (form_id) REFERENCES form_builder_forms (id)
        ON DELETE CASCADE ON UPDATE CASCADE
) DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- サンプルフォーム（デモ用）
INSERT INTO form_builder_forms (id, title, description, created_at, updated_at) VALUES
    (1, '勉強会 参加申込フォーム', '社外向け技術勉強会の参加申込を受け付けるフォームです。', DATE_SUB(NOW(3), INTERVAL 30 DAY), DATE_SUB(NOW(3), INTERVAL 30 DAY)),
    (2, '社内ランチ満足度アンケート', '社員食堂のメニュー改善のためのアンケートです。', DATE_SUB(NOW(3), INTERVAL 20 DAY), DATE_SUB(NOW(3), INTERVAL 20 DAY)),
    (3, '備品購入申請', '業務で必要な備品の購入を申請するフォームです。', DATE_SUB(NOW(3), INTERVAL 10 DAY), DATE_SUB(NOW(3), INTERVAL 10 DAY));

INSERT INTO form_builder_form_fields (id, form_id, label, type, required, options, layout_row, sort_order) VALUES
    (1, 1, '氏名', 'text', TRUE, JSON_ARRAY(), 0, 1),
    (2, 1, 'メールアドレス', 'email', TRUE, JSON_ARRAY(), 0, 2),
    (3, 1, '参加希望日', 'date', TRUE, JSON_ARRAY(), 1, 3),
    (4, 1, '参加方法', 'radio', TRUE, JSON_ARRAY('会場参加', 'オンライン参加'), 1, 4),
    (5, 1, '興味のあるテーマ', 'checkbox', FALSE, JSON_ARRAY('Webフロントエンド', 'バックエンド', 'インフラ', 'AI・機械学習'), 2, 5),
    (6, 1, '質問・要望', 'textarea', FALSE, JSON_ARRAY(), 3, 6),
    (7, 2, '部署', 'select', TRUE, JSON_ARRAY('開発部', '営業部', '管理部', 'デザイン部'), 0, 1),
    (8, 2, '満足度（1〜5）', 'number', TRUE, JSON_ARRAY(), 0, 2),
    (9, 2, 'よく利用するメニュー', 'checkbox', FALSE, JSON_ARRAY('日替わり定食', 'カレー', '麺類', 'サラダ'), 1, 3),
    (10, 2, 'コメント', 'textarea', FALSE, JSON_ARRAY(), 2, 4),
    (11, 3, '申請者', 'text', TRUE, JSON_ARRAY(), 0, 1),
    (12, 3, '品名', 'text', TRUE, JSON_ARRAY(), 1, 2),
    (13, 3, '数量', 'number', TRUE, JSON_ARRAY(), 1, 3),
    (14, 3, '希望納期', 'date', FALSE, JSON_ARRAY(), 1, 4),
    (15, 3, '購入理由', 'textarea', TRUE, JSON_ARRAY(), 2, 5);

-- サンプル回答（実行日からの相対日付で常に「直近」のデータに見えるようにする）
INSERT INTO form_builder_form_records (form_id, data, created_at) VALUES
    (1, JSON_OBJECT('1', '山田 太郎', '2', 'taro.yamada@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 14 DAY), '%Y-%m-%d'), '4', '会場参加', '5', JSON_ARRAY('Webフロントエンド', 'バックエンド'), '6', '懇親会はありますか？'), DATE_SUB(NOW(3), INTERVAL 28 DAY)),
    (1, JSON_OBJECT('1', '佐藤 花子', '2', 'hanako.sato@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 14 DAY), '%Y-%m-%d'), '4', 'オンライン参加', '5', JSON_ARRAY('AI・機械学習')), DATE_SUB(NOW(3), INTERVAL 25 DAY)),
    (1, JSON_OBJECT('1', '鈴木 一郎', '2', 'ichiro.suzuki@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 21 DAY), '%Y-%m-%d'), '4', '会場参加', '5', JSON_ARRAY('インフラ'), '6', '駐車場の有無を教えてください。'), DATE_SUB(NOW(3), INTERVAL 21 DAY)),
    (1, JSON_OBJECT('1', '高橋 美咲', '2', 'misaki.takahashi@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 14 DAY), '%Y-%m-%d'), '4', 'オンライン参加'), DATE_SUB(NOW(3), INTERVAL 18 DAY)),
    (1, JSON_OBJECT('1', '田中 健', '2', 'ken.tanaka@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 21 DAY), '%Y-%m-%d'), '4', '会場参加', '5', JSON_ARRAY('バックエンド', 'インフラ', 'AI・機械学習')), DATE_SUB(NOW(3), INTERVAL 12 DAY)),
    (1, JSON_OBJECT('1', '伊藤 さくら', '2', 'sakura.ito@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 21 DAY), '%Y-%m-%d'), '4', 'オンライン参加', '5', JSON_ARRAY('Webフロントエンド'), '6', '資料は後日共有されますか？'), DATE_SUB(NOW(3), INTERVAL 7 DAY)),
    (1, JSON_OBJECT('1', '渡辺 翔', '2', 'sho.watanabe@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 14 DAY), '%Y-%m-%d'), '4', '会場参加'), DATE_SUB(NOW(3), INTERVAL 3 DAY)),
    (1, JSON_OBJECT('1', '中村 結衣', '2', 'yui.nakamura@example.com', '3', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 21 DAY), '%Y-%m-%d'), '4', 'オンライン参加', '5', JSON_ARRAY('AI・機械学習', 'Webフロントエンド')), DATE_SUB(NOW(3), INTERVAL 1 DAY)),
    (2, JSON_OBJECT('7', '開発部', '8', 4, '9', JSON_ARRAY('日替わり定食', 'カレー'), '10', 'カレーの辛さを選べると嬉しいです。'), DATE_SUB(NOW(3), INTERVAL 19 DAY)),
    (2, JSON_OBJECT('7', '営業部', '8', 3, '9', JSON_ARRAY('麺類')), DATE_SUB(NOW(3), INTERVAL 17 DAY)),
    (2, JSON_OBJECT('7', 'デザイン部', '8', 5, '9', JSON_ARRAY('サラダ', '日替わり定食'), '10', 'サラダの種類が増えて満足しています。'), DATE_SUB(NOW(3), INTERVAL 14 DAY)),
    (2, JSON_OBJECT('7', '管理部', '8', 2, '10', '昼休みの混雑が気になります。'), DATE_SUB(NOW(3), INTERVAL 9 DAY)),
    (2, JSON_OBJECT('7', '開発部', '8', 4, '9', JSON_ARRAY('カレー', '麺類')), DATE_SUB(NOW(3), INTERVAL 5 DAY)),
    (2, JSON_OBJECT('7', '営業部', '8', 5, '9', JSON_ARRAY('日替わり定食')), DATE_SUB(NOW(3), INTERVAL 2 DAY)),
    (3, JSON_OBJECT('11', '山田 太郎', '12', '27インチモニター', '13', 1, '14', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 7 DAY), '%Y-%m-%d'), '15', 'デュアルディスプレイで作業効率を上げるため。'), DATE_SUB(NOW(3), INTERVAL 8 DAY)),
    (3, JSON_OBJECT('11', '佐藤 花子', '12', 'USB-Cハブ', '13', 2, '15', '会議室のプロジェクター接続用。'), DATE_SUB(NOW(3), INTERVAL 4 DAY)),
    (3, JSON_OBJECT('11', '鈴木 一郎', '12', 'ホワイトボードマーカー（黒）', '13', 10, '14', DATE_FORMAT(DATE_ADD(CURDATE(), INTERVAL 3 DAY), '%Y-%m-%d'), '15', '在庫切れのため。'), DATE_SUB(NOW(3), INTERVAL 1 DAY));
