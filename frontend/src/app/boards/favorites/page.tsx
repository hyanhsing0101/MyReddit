"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { BoardFavoriteButton } from "@/components/board-favorite-button";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiListFavoriteBoards,
  type BoardFavoriteRow,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

export default function BoardFavoritesPage() {
  const [rows, setRows] = useState<BoardFavoriteRow[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(
    async (p: number) => {
      const token = getAccessToken();
      if (!token) {
        setLoading(false);
        setError("请先登录");
        setRows([]);
        return;
      }
      setLoading(true);
      setError(null);
      try {
        const body = await apiListFavoriteBoards(token, p, pageSize);
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setError(apiErrorMessage(body));
          setRows([]);
          return;
        }
        setRows(body.data.list);
        setTotal(body.data.total);
        setPage(body.data.page);
      } catch (e) {
        setError(e instanceof Error ? e.message : "加载失败");
        setRows([]);
      } finally {
        setLoading(false);
      }
    },
    [pageSize],
  );

  useEffect(() => {
    void load(1);
  }, [load]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const token = getAccessToken();

  return (
    <div className="mx-auto max-w-2xl px-4 py-16">
      <div className="mb-8 flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
            收藏的板块
          </h1>
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            按收藏时间倒序
          </p>
        </div>
        <Link
          href="/favorites"
          className="rounded-lg border border-zinc-300 px-4 py-2 text-sm dark:border-zinc-600"
        >
          回收藏夹
        </Link>
      </div>

      {!token ? (
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          请先{" "}
          <Link href="/login" className="underline">
            登录
          </Link>
        </p>
      ) : loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : rows.length === 0 ? (
        <p className="text-sm text-zinc-500">暂无收藏，可在板块列表或详情页点击「收藏」。</p>
      ) : (
        <ul className="divide-y divide-zinc-200 rounded-xl border border-zinc-200 dark:divide-zinc-800 dark:border-zinc-800">
          {rows.map((b) => (
            <li key={b.id} className="flex items-stretch gap-2">
              <Link
                href={`/boards/${encodeURIComponent(b.slug)}`}
                className="min-w-0 flex-1 px-4 py-4 hover:bg-zinc-50 dark:hover:bg-zinc-900/50"
              >
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {b.name}
                </span>
                <span className="ml-2 text-sm text-zinc-500">/{b.slug}</span>
                <p className="mt-1 text-xs text-zinc-500">
                  收藏于 {new Date(b.favorited_at).toLocaleString()}
                </p>
              </Link>
              <div className="flex shrink-0 items-center pr-3">
                <BoardFavoriteButton
                  boardId={b.id}
                  isSystemSink={b.is_system_sink}
                  isFavorited
                  accessToken={token}
                  onUpdated={(next) => {
                    if (!next) {
                      setRows((prev) => prev.filter((x) => x.id !== b.id));
                      setTotal((t) => Math.max(0, t - 1));
                    }
                  }}
                />
              </div>
            </li>
          ))}
        </ul>
      )}

      {!loading && !error && token && total > pageSize ? (
        <div className="mt-6 flex items-center justify-between text-sm text-zinc-500">
          <span>
            第 {page} / {totalPages} 页 · 共 {total} 个
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => void load(page - 1)}
              className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
            >
              上一页
            </button>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => void load(page + 1)}
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
