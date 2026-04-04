"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiListPosts,
  apiPing,
  type PostItem,
} from "@/lib/api";
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
} from "@/lib/auth-storage";

function previewText(text: string, max = 120) {
  const t = text.replace(/\s+/g, " ").trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max)}…`;
}

export default function HomeClient() {
  const [loggedIn, setLoggedIn] = useState(false);
  const [pingResult, setPingResult] = useState<string | null>(null);
  const [pingError, setPingError] = useState<string | null>(null);
  const [pingging, setPingging] = useState(false);

  const [posts, setPosts] = useState<PostItem[]>([]);
  const [listPage, setListPage] = useState(1);
  const [listTotal, setListTotal] = useState(0);
  const [listPageSize] = useState(10);
  const [listLoading, setListLoading] = useState(true);
  const [listError, setListError] = useState<string | null>(null);

  const refreshAuthState = useCallback(() => {
    setLoggedIn(!!getAccessToken());
  }, []);

  useEffect(() => {
    refreshAuthState();
  }, [refreshAuthState]);

  const loadPosts = useCallback(async (page: number) => {
    setListLoading(true);
    setListError(null);
    try {
      const body = await apiListPosts(page, listPageSize);
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setListError(apiErrorMessage(body));
        setPosts([]);
        return;
      }
      setPosts(body.data.list);
      setListTotal(body.data.total);
      setListPage(body.data.page);
    } catch (e) {
      setListError(e instanceof Error ? e.message : "加载失败");
      setPosts([]);
    } finally {
      setListLoading(false);
    }
  }, [listPageSize]);

  useEffect(() => {
    loadPosts(1);
  }, [loadPosts]);

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

  const totalPages = Math.max(1, Math.ceil(listTotal / listPageSize));
  const canPrev = listPage > 1;
  const canNext = listPage < totalPages;

  return (
    <main className="mx-auto flex w-full max-w-2xl flex-col gap-8 px-4 py-16">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
          MyReddit
        </h1>
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
          最新帖子 · 登录后可发帖与 Ping
        </p>
      </div>

      <nav className="flex flex-wrap gap-3 text-sm">
        {!loggedIn ? (
          <>
            <Link
              href="/boards"
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              板块
            </Link>
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
              href="/boards"
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              板块
            </Link>
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

      <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
        <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
          <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
            帖子
          </h2>
        </div>
        <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
          {listLoading ? (
            <p className="px-4 py-8 text-center text-sm text-zinc-500">
              加载中…
            </p>
          ) : listError ? (
            <p className="px-4 py-8 text-center text-sm text-red-600 dark:text-red-400">
              {listError}
            </p>
          ) : posts.length === 0 ? (
            <p className="px-4 py-8 text-center text-sm text-zinc-500">
              暂无帖子
            </p>
          ) : (
            posts.map((post) => (
              <article key={post.id} className="px-4 py-4">
                <h3 className="font-medium text-zinc-900 dark:text-zinc-100">
                  <Link
                    href={`/posts/${post.id}`}
                    className="hover:underline"
                  >
                    {post.title}
                  </Link>
                </h3>
                <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
                  {previewText(post.content)}
                </p>
                <p className="mt-2 text-xs text-zinc-500">
                  {post.board_slug ? (
                    <>
                      <Link
                        href={`/boards/${encodeURIComponent(post.board_slug)}`}
                        className="text-zinc-700 underline dark:text-zinc-300"
                      >
                        {post.board_name || post.board_slug}
                      </Link>
                      <span> · </span>
                    </>
                  ) : null}
                  {post.author_id != null
                    ? `作者 ID ${post.author_id}`
                    : "无主帖"}{" "}
                  ·{" "}
                  {new Date(post.create_time).toLocaleString("zh-CN", {
                    month: "numeric",
                    day: "numeric",
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </p>
              </article>
            ))
          )}
        </div>
        {!listLoading && !listError && listTotal > listPageSize ? (
          <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
            <span className="text-zinc-500">
              第 {listPage} / {totalPages} 页 · 共 {listTotal} 条
            </span>
            <div className="flex gap-2">
              <button
                type="button"
                disabled={!canPrev}
                onClick={() => loadPosts(listPage - 1)}
                className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
              >
                上一页
              </button>
              <button
                type="button"
                disabled={!canNext}
                onClick={() => loadPosts(listPage + 1)}
                className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
              >
                下一页
              </button>
            </div>
          </div>
        ) : null}
      </section>

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
