/* eslint-disable @next/next/no-img-element */
import type { Metadata } from "next";
import styles from "./_styles/shop.module.css";

export const metadata: Metadata = {
  title: "Antiqueshop Oude dagen | My Portal",
  description:
    "ヨーロッパのアンティークを扱うショップサイトを想定したHTML/CSSポートフォリオです。",
};

export default function ShopCssPage() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles["header-bar"]}>
          <div className={styles["header-title"]}>Antiqushop Oude dagen</div>
          <div>
            <ul className={styles["header-menulist"]}>
              <li className={`${styles.menuitem} ${styles.current}`}>
                <a href="#">Home</a>
              </li>
              <li className={styles.menuitem}>
                <a href="#">AboutUs</a>
              </li>
              <li className={styles.menuitem}>
                <a href="#">Online Store</a>
              </li>
              <li className={styles.menuitem}>
                <a href="#">Contact</a>
              </li>
            </ul>
          </div>
        </div>
      </header>

      <div className={`${styles.eyecatch} ${styles.inner}`}>
        <img src="/shop-css/img/eyecatch01.jpg" alt="店内の様子" />
      </div>

      <main className={styles.main}>
        <div className={styles.shopinfo}>
          <p className={styles["section-title"]}>Antiqueshop Oude dagen</p>
          <div className={styles["shopinfo-text"]}>
            <p>
              ヨーロッパを中心にちょっとお部屋に置きたくなるアンティークをそろえています。
            </p>
            <p>気になる商品がございましたらお気軽にお問い合わせください。</p>
          </div>
        </div>

        <div className={styles.information}>
          <div className={styles["information-title"]}>
            <p className={styles["information-title"]}>Information</p>
          </div>
          <div className={styles["information-list"]}>
            <div className={styles["information-item"]}>
              <p className={styles["information-date"]}>2018/08/01</p>
              <p className={styles["information-text"]}>
                <a href="#">営業時間が変更になりました</a>
              </p>
            </div>
            <div className={styles["information-item"]}>
              <p className={styles["information-date"]}>2018/05/15</p>
              <p className={styles["information-text"]}>
                <a href="#">新商品入荷のお知らせ</a>
              </p>
            </div>
            <div className={styles["information-item"]}>
              <p className={styles["information-date"]}>2018/04/01</p>
              <p className={styles["information-text"]}>
                <a href="#">SHOPオープンのご挨拶</a>
              </p>
            </div>
          </div>
        </div>

        <div className={`${styles.products} ${styles.inner}`}>
          <ul className={styles["products-list"]}>
            <li className={styles["products-item"]}>
              <img src="/shop-css/img/product01.jpg" alt="Clock" />
              <a href="#">
                <p className={styles["products-button"]}>Clock</p>
              </a>
            </li>
            <li className={styles["products-item"]}>
              <img src="/shop-css/img/product03.jpg" alt="Light" />
              <a href="#">
                <p className={styles["products-button"]}>Light</p>
              </a>
            </li>
            <li className={styles["products-item"]}>
              <img src="/shop-css/img/product02.jpg" alt="Book" />
              <a href="#">
                <p className={styles["products-button"]}>Book</p>
              </a>
            </li>
          </ul>
        </div>

        <div className={styles.insta}>
          <p className={styles["section-title"]}>INSTAGRAM</p>
          <div className={styles.instagram}>
            <ul className={styles["insta-list"]}>
              {Array.from({ length: 8 }, (_, i) => i + 1).map((n) => (
                <li key={n} className={styles["insta-item"]}>
                  <img
                    src={`/shop-css/img/insta0${n}.jpg`}
                    alt={`Instagram投稿 ${n}`}
                  />
                </li>
              ))}
            </ul>
          </div>
          <div className={styles.morebutton}>
            <a href="#">
              <span>more</span>
            </a>
          </div>
        </div>

        <div className={styles.access}>
          <p className={styles["section-title"]}>ACCESS</p>
          <div className={styles["access-info"]}>
            <iframe
              className={styles["access-map"]}
              title="ショップの地図"
              src="https://www.google.com/maps/embed?pb=!1m18!1m12!1m3!1d3745.2072688593034!2d139.66505719368703!3d35.66188569151479!2m3!1f0!2f0!3f0!3m2!1i1024!2i768!4f13.1!3m3!1m2!1s0x6018f36b9a296133%3A0x66c43a9f356d5e5d!2z5LiL5YyX5rKi6aeF!5e0!3m2!1sja!2sjp!4v1586594699073!5m2!1sja!2sjp"
              width={250}
              height={250}
              style={{ border: 0 }}
              allowFullScreen
              tabIndex={0}
            />
            <div className={styles["access-text"]}>
              <p>〒155-0031 東京都世田谷区北沢２丁目２４−２</p>
              <p>
                <span>OPEN</span>10:00～18:00
              </p>
              <p>
                <span>TEL</span>03-1234-5678
              </p>
              <p>
                <span>Email</span>anqique_oude_dagen@example.com
              </p>
            </div>
          </div>
        </div>

        <div className={styles.sns}>
          <ul className={styles["sns-list"]}>
            <li>
              <a
                href="https://twitter.com"
                target="_blank"
                rel="noreferrer"
                className={`${styles["sns-item"]} fa fa-twitter`}
                aria-hidden="true"
              ></a>
            </li>
            <li>
              <a
                href="https://facebook.com"
                target="_blank"
                rel="noreferrer"
                className={`${styles["sns-item"]} fa fa-facebook`}
                aria-hidden="true"
              ></a>
            </li>
            <li>
              <a
                href="https://instagram.com"
                target="_blank"
                rel="noreferrer"
                className={`${styles["sns-item"]} fa fa-instagram`}
                aria-hidden="true"
              ></a>
            </li>
          </ul>
        </div>
      </main>

      <footer className={styles.footer}>
        <p className={styles["footer-copyright"]}>
          Antiqueshop Oude dagen Co.,Ltd 2018-2020
        </p>
      </footer>
    </div>
  );
}
