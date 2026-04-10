"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { BoardFavoriteButton } from "@/components/board-favorite-button";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiListBoards,
  type BoardItem,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

export default function BoardsPage() {
  const [boards, setBoards] = useState<BoardItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [authTick, setAuthTick] = useState(0);

  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key === "myreddit_access_token") setAuthTick((t) => t + 1);
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  const load = useCallback(async (p: number) => {
    setLoading(true);
    setError(null);
    try {
      const token = getAccessToken();
      const body = await apiListBoards(p, pageSize, false, token);
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setError(apiErrorMessage(body));
        setBoards([]);
        return;
      }
      setBoards(body.data.list);
      setTotal(body.data.total);
      setPage(body.data.page);
    } catch (e) {
      setError(e instanceof Error ? e.message : "加载失败");
      setBoards([]);
    } finally {
      setLoading(false);
    }
  }, [pageSize]);

  useEffect(() => {
    load(1);
  }, [load, authTick]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <div className="mx-auto max-w-2xl px-4 py-16">
      <div className="mb-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
            板块
          </h1>
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            浏览社区列表（不含系统归档板）
          </p>
        </div>
        <div className="flex flex-wrap gap-3 text-sm">
          {getAccessToken() ? (
            <Link
              href="/favorites"
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              收藏夹
            </Link>
          ) : null}
          <Link
            href="/boards/new"
            className="rounded-lg bg-zinc-900 px-4 py-2 text-white dark:bg-zinc-100 dark:text-zinc-900"
          >
            创建板块
          </Link>
          <Link href="/" className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600">
            首页
          </Link>
        </div>
      </div>

      {loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : boards.length === 0 ? (
        <p className="text-sm text-zinc-500">暂无板块</p>
      ) : (
        <ul className="divide-y divide-zinc-200 rounded-xl border border-zinc-200 dark:divide-zinc-800 dark:border-zinc-800">
          {boards.map((b) => (
            <li key={b.id} className="flex items-stretch gap-2">
              <Link
                href={`/boards/${encodeURIComponent(b.slug)}`}
                className="min-w-0 flex-1 px-4 py-4 hover:bg-zinc-50 dark:hover:bg-zinc-900/50"
              >
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {b.name}
                </span>
                <span className="ml-2 text-sm text-zinc-500">/{b.slug}</span>
                {b.description ? (
                  <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
                    {b.description}
                  </p>
                ) : null}
              </Link>
              <div className="flex shrink-0 items-center pr-3">
                <BoardFavoriteButton
                  boardId={b.id}
                  isSystemSink={b.is_system_sink}
                  isFavorited={!!b.is_favorited}
                  accessToken={getAccessToken()}
                  onUpdated={(next) => {
                    setBoards((prev) =>
                      prev.map((x) =>
                        x.id === b.id ? { ...x, is_favorited: next } : x,
                      ),
                    );
                  }}
                />
              </div>
            </li>
          ))}
        </ul>
      )}

      {!loading && !error && total > pageSize ? (
        <div className="mt-6 flex items-center justify-between text-sm text-zinc-500">
          <span>
            第 {page} / {totalPages} 页 · 共 {total} 个
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => load(page - 1)}
              className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
            >
              上一页
            </button>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => load(page + 1)}
              className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
            >
              下一页
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
