import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Dockerイメージに .next/standalone だけをコピーして動かすための出力形式
  output: "standalone",
  // featureは /demo-portal 配下へ移動した。旧URL（/<feature-slug>）へのリンク・ブックマークを新URLへ転送する
  async redirects() {
    return [
      ...["book-database", "simple-ledger", "simple-cms", "form-builder"].map((slug) => ({
        source: `/${slug}/:path*`,
        destination: `/demo-portal/${slug}/:path*`,
        permanent: true,
      })),
      // shop-css は public/shop-css/ の画像と同じパスを使うため、ページ本体のみを転送する
      // （redirectsはpublicより先に評価されるので、:path* にすると画像まで転送されてしまう）
      { source: "/shop-css", destination: "/demo-portal/shop-css", permanent: true },
    ];
  },
};

export default nextConfig;
