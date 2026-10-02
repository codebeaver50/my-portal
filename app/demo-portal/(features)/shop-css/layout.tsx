export default function ShopCssLayout({ children }: LayoutProps<"/demo-portal/shop-css">) {
  return (
    <>
      {/* destyle.css(CDN)は非layer化CSSのためTailwindのlayer化ユーティリティより優先されmainのflexを壊す。preflightで代替 */}
      {/* eslint-disable-next-line @next/next/no-page-custom-font -- App Router route-scoped stylesheet, not pages/_document */}
      <link
        rel="stylesheet"
        href="https://fonts.googleapis.com/css2?family=Montserrat:wght@400;500;700&display=swap"
      />
      {children}
    </>
  );
}
