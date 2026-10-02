import type { Metadata } from "next";
import { DashboardView } from "./_components/DashboardView";

export const metadata: Metadata = {
  title: "複式簿記シンプル家計簿 | My Portal",
  description: "複式簿記ベースの家計簿デモアプリのダッシュボード",
};

export default function SimpleLedgerDashboardPage() {
  return <DashboardView />;
}
