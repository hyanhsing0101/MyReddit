"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { apiPing } from "@/lib/api";
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
} from "@/lib/auth-storage";

export default function HomeClient() {
  const [loggedIn, setLoggedIn] = useState(false);
  const [pingResult, setPingResult] = useState<string | null>(null);
  const [pingError, setPingError] = useState<string | null>(null);
  const [pingging, setPingging] = useState(false);

  const refreshAuthState = useCallback(() => {
    setLoggedIn(!!getAccessToken());
  }, []);

  useEffect(() => {
    refreshAuthState();
  }, [refreshAuthState]);

  async function handlePing() {
    const token = getAccessToken();
    setPingResult(null);
    setPingError(null);
    if (!token) {
      setPingError("请先登录");
      return;
    }
    setPingging(true);
    try {
      const text = await apiPing(token);
      setPingResult(text);
    } catch (e) {
      setPingError(e instanceof Error ? e.message : "Ping 失败");
    } finally {
      setPingging(false);
    }
  }

  function handleLogout() {
    clearTokens();
    setLoggedIn(false);
    setPingResult(null);
    setPingError(null);
  }

  return (
    <main className="mx-auto flex w-full max-w-lg flex-col gap-8 px-4 py-16">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
          MyReddit
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
          注册 → 登录 → 带 JWT 请求 <code className="rounded bg-zinc-200 px-1 dark:bg-zinc-800">GET /ping</code>
        </p>
      </div>

      <nav className="flex flex-wrap gap-3 text-sm">
        {!loggedIn ? (
          <>
            <Link
              href="/signup"
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              注册
            </Link>
            <Link
              href="/login"
              className="rounded-lg bg-zinc-900 px-4 py-2 text-white dark:bg-zinc-100 dark:text-zinc-900"
            >
              登录
            </Link>
          </>
        ) : (
          <>
            <Link
              href="/post/new"
              className="rounded-lg bg-zinc-900 px-4 py-2 text-white dark:bg-zinc-100 dark:text-zinc-900"
            >
              发帖
            </Link>
            <button
              type="button"
              onClick={handleLogout}
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              退出登录
            </button>
          </>
        )}
      </nav>

      <section className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
        <h2 className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
          鉴权测试
        </h2>
        <button
          type="button"
          onClick={handlePing}
          disabled={pingging}
          className="mt-3 rounded-lg bg-emerald-700 px-4 py-2 text-sm text-white disabled:opacity-50"
        >
          {pingging ? "请求中…" : "调用 GET /ping"}
        </button>
        {pingResult ? (
          <p className="mt-3 text-sm text-emerald-700 dark:text-emerald-400">
            响应：{pingResult}
          </p>
        ) : null}
        {pingError ? (
          <p className="mt-3 text-sm text-red-600 dark:text-red-400">{pingError}</p>
        ) : null}
      </section>

      {loggedIn && getRefreshToken() ? (
        <p className="text-xs text-zinc-500">
          已保存 access / refresh token（localStorage）。后端地址：{" "}
          <code className="rounded bg-zinc-200 px-1 dark:bg-zinc-800">
            {process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://127.0.0.1:8081"}
          </code>
        </p>
      ) : null}
    </main>
  );
}
