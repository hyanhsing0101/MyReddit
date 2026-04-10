import type { Metadata } from "next";
import Link from "next/link";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "MyReddit",
  description: "MyReddit 前端",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <header className="sticky top-0 z-10 border-b border-zinc-200 bg-white/90 backdrop-blur dark:border-zinc-800 dark:bg-zinc-950/90">
          <nav className="mx-auto flex max-w-2xl flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3 text-sm">
            <Link
              href="/"
              className="font-medium text-zinc-900 dark:text-zinc-100"
            >
              首页
            </Link>
            <Link
              href="/boards"
              className="text-zinc-700 hover:underline dark:text-zinc-300"
            >
              板块
            </Link>
            <Link
              href="/favorites"
              className="text-zinc-700 hover:underline dark:text-zinc-300"
            >
              收藏夹
            </Link>
            <span className="hidden flex-1 sm:block" aria-hidden />
            <Link
              href="/login"
              className="text-zinc-600 hover:underline dark:text-zinc-400"
            >
              登录
            </Link>
          </nav>
        </header>
        <div className="flex-1">{children}</div>
      </body>
    </html>
  );
}
