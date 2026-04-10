"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { BoardFavoriteButton } from "@/components/board-favorite-button";
import { PostFavoriteButton } from "@/components/post-favorite-button";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiListFavoriteBoards,
  apiListFavoritePosts,
  type BoardFavoriteRow,
  type PostFavoriteRow,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

function previewText(text: string, max = 90) {
  const t = text.replace(/\s+/g, " ").trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max)}…`;
}

type FavoriteTab = "posts" | "boards";

export default function FavoritesPage() {
  const token = getAccessToken();
  const [tab, setTab] = useState<FavoriteTab>("posts");

  const [postRows, setPostRows] = useState<PostFavoriteRow[]>([]);
  const [postTotal, setPostTotal] = useState(0);
  const [postPage, setPostPage] = useState(1);
  const postPageSize = 10;

  const [boardRows, setBoardRows] = useState<BoardFavoriteRow[]>([]);
  const [boardTotal, setBoardTotal] = useState(0);
  const [boardPage, setBoardPage] = useState(1);
  const boardPageSize = 10;

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadPosts = useCallback(
    async (p: number) => {
      if (!token) return;
      const body = await apiListFavoritePosts(token, p, postPageSize);
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        throw new Error(apiErrorMessage(body));
      }
      setPostRows(body.data.list ?? []);
      setPostTotal(body.data.total ?? 0);
      setPostPage(body.data.page ?? p);
    },
    [token],
  );

  const loadBoards = useCallback(
    async (p: number) => {
      if (!token) return;
      const body = await apiListFavoriteBoards(token, p, boardPageSize);
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        throw new Error(apiErrorMessage(body));
      }
      setBoardRows(body.data.list ?? []);
      setBoardTotal(body.data.total ?? 0);
      setBoardPage(body.data.page ?? p);
    },
    [token],
  );

  useEffect(() => {
    if (!token) {
      setLoading(false);
      setError("请先登录");
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        await Promise.all([loadPosts(1), loadBoards(1)]);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [token, loadPosts, loadBoards]);

  const postTotalPages = useMemo(
    () => Math.max(1, Math.ceil(postTotal / postPageSize)),
    [postTotal],
  );
  const boardTotalPages = useMemo(
    () => Math.max(1, Math.ceil(boardTotal / boardPageSize)),
    [boardTotal],
  );

  return (
    <main className="mx-auto w-full max-w-2xl px-4 py-16">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
            收藏夹
          </h1>
          <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
            收藏帖子与订阅板块统一查看
          </p>
        </div>
        <Link
          href="/"
          className="rounded-lg border border-zinc-300 px-4 py-2 text-sm dark:border-zinc-600"
        >
          回首页
        </Link>
      </div>

      {!token ? (
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          请先 <Link href="/login" className="underline">登录</Link>
        </p>
      ) : loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : (
        <>
          <div className="mb-4 flex gap-2">
            <button
              type="button"
              onClick={() => setTab("posts")}
              className={
                tab === "posts"
                  ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1.5 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                  : "rounded-lg border border-zinc-300 px-3 py-1.5 text-xs dark:border-zinc-600"
              }
            >
              收藏帖子（{postTotal}）
            </button>
            <button
              type="button"
              onClick={() => setTab("boards")}
              className={
                tab === "boards"
                  ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1.5 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                  : "rounded-lg border border-zinc-300 px-3 py-1.5 text-xs dark:border-zinc-600"
              }
            >
              订阅板块（{boardTotal}）
            </button>
          </div>

          {tab === "posts" ? (
            <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
              <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
                {postRows.length === 0 ? (
                  <p className="px-4 py-6 text-sm text-zinc-500">暂无收藏帖子</p>
                ) : (
                  postRows.map((p) => (
                    <article key={p.id} className="flex gap-3 px-4 py-4">
                      <div className="min-w-0 flex-1">
                        <h2 className="font-medium text-zinc-900 dark:text-zinc-100">
                          <Link href={`/posts/${p.id}`} className="hover:underline">
                            {p.title}
                          </Link>
                        </h2>
                        <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
                          {previewText(p.content)}
                        </p>
                        <p className="mt-2 text-xs text-zinc-500">
                          收藏于 {new Date(p.favorited_at).toLocaleString("zh-CN")}
                        </p>
                      </div>
                      <PostFavoriteButton
                        postId={p.id}
                        isFavorited
                        accessToken={token}
                        onUpdated={(next) => {
                          if (!next) {
                            setPostRows((prev) => prev.filter((x) => x.id !== p.id));
                            setPostTotal((n) => Math.max(0, n - 1));
                          }
                        }}
                      />
                    </article>
                  ))
                )}
              </div>
              {postTotal > postPageSize ? (
                <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                  <span className="text-zinc-500">
                    第 {postPage} / {postTotalPages} 页
                  </span>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      disabled={postPage <= 1}
                      onClick={() => void loadPosts(postPage - 1)}
                      className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                    >
                      上一页
                    </button>
                    <button
                      type="button"
                      disabled={postPage >= postTotalPages}
                      onClick={() => void loadPosts(postPage + 1)}
                      className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                    >
                      下一页
                    </button>
                  </div>
                </div>
              ) : null}
            </section>
          ) : (
            <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
              <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
                {boardRows.length === 0 ? (
                  <p className="px-4 py-6 text-sm text-zinc-500">暂无订阅板块</p>
                ) : (
                  boardRows.map((b) => (
                    <article key={b.id} className="flex items-center gap-3 px-4 py-4">
                      <div className="min-w-0 flex-1">
                        <Link
                          href={`/boards/${encodeURIComponent(b.slug)}`}
                          className="font-medium text-zinc-900 underline dark:text-zinc-100"
                        >
                          {b.name}
                        </Link>
                        <p className="mt-1 text-xs text-zinc-500">
                          /{b.slug} · 订阅于{" "}
                          {new Date(b.favorited_at).toLocaleString("zh-CN")}
                        </p>
                      </div>
                      <BoardFavoriteButton
                        boardId={b.id}
                        isSystemSink={b.is_system_sink}
                        isFavorited
                        accessToken={token}
                        onUpdated={(next) => {
                          if (!next) {
                            setBoardRows((prev) => prev.filter((x) => x.id !== b.id));
                            setBoardTotal((n) => Math.max(0, n - 1));
                          }
                        }}
                      />
                    </article>
                  ))
                )}
              </div>
              {boardTotal > boardPageSize ? (
                <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                  <span className="text-zinc-500">
                    第 {boardPage} / {boardTotalPages} 页
                  </span>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      disabled={boardPage <= 1}
                      onClick={() => void loadBoards(boardPage - 1)}
                      className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                    >
                      上一页
                    </button>
                    <button
                      type="button"
                      disabled={boardPage >= boardTotalPages}
                      onClick={() => void loadBoards(boardPage + 1)}
                      className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                    >
                      下一页
                    </button>
                  </div>
                </div>
              ) : null}
            </section>
          )}
        </>
      )}
    </main>
  );
}
