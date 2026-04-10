"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiGetUserHome,
  API_USER_NOT_EXIST_CODE,
  type UserHomePayload,
} from "@/lib/api";

function previewText(text: string, max = 80) {
  const t = text.replace(/\s+/g, " ").trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max)}…`;
}

export default function UserHomePage() {
  const params = useParams();
  const idParam = params.id;
  const userId =
    typeof idParam === "string"
      ? Number.parseInt(idParam, 10)
      : Array.isArray(idParam)
        ? Number.parseInt(idParam[0] ?? "", 10)
        : NaN;

  const [postPage, setPostPage] = useState(1);
  const [commentPage, setCommentPage] = useState(1);
  const pageSize = 10;

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [data, setData] = useState<UserHomePayload | null>(null);
  const posts = Array.isArray(data?.posts) ? data.posts : [];
  const comments = Array.isArray(data?.comments) ? data.comments : [];

  useEffect(() => {
    if (!Number.isFinite(userId) || userId < 1) {
      setLoading(false);
      setError("无效的用户 ID");
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      setNotFound(false);
      try {
        const body = await apiGetUserHome(
          userId,
          postPage,
          pageSize,
          commentPage,
          pageSize,
        );
        if (cancelled) return;
        if (body.code === API_USER_NOT_EXIST_CODE) {
          setNotFound(true);
          setData(null);
          return;
        }
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setError(apiErrorMessage(body));
          setData(null);
          return;
        }
        setData(body.data);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
          setData(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [userId, postPage, commentPage]);

  const postPages = useMemo(() => {
    if (!data) return 1;
    return Math.max(1, Math.ceil(data.posts_total / data.post_page_size));
  }, [data]);
  const commentPages = useMemo(() => {
    if (!data) return 1;
    return Math.max(1, Math.ceil(data.comments_total / data.comment_page_size));
  }, [data]);

  return (
    <main className="mx-auto w-full max-w-2xl px-4 py-10">
      <p className="mb-6">
        <Link href="/" className="text-sm text-zinc-500 underline">
          ← 返回首页
        </Link>
      </p>

      {loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : notFound ? (
        <p className="text-sm text-zinc-700 dark:text-zinc-300">用户不存在</p>
      ) : error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : data ? (
        <div className="space-y-8">
          <section className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
            <h1 className="text-xl font-semibold text-zinc-900 dark:text-zinc-100">
              用户主页
            </h1>
            <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400">
              {data.username} · ID {data.user_id}
            </p>
          </section>

          <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                最近帖子（{data.posts_total}）
              </h2>
            </div>
            <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
              {posts.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无帖子</p>
              ) : (
                posts.map((p) => (
                  <div key={p.id} className="px-4 py-4">
                    <Link
                      href={`/posts/${p.id}`}
                      className="font-medium text-zinc-900 underline dark:text-zinc-100"
                    >
                      {p.title}
                    </Link>
                    <p className="mt-1 text-xs text-zinc-500">
                      <Link
                        href={`/boards/${encodeURIComponent(p.board_slug)}`}
                        className="underline"
                      >
                        {p.board_name || p.board_slug}
                      </Link>{" "}
                      · 分数 {p.score}
                    </p>
                  </div>
                ))
              )}
            </div>
            {data.posts_total > data.post_page_size ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {data.post_page} / {postPages} 页
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={data.post_page <= 1}
                    onClick={() => setPostPage((n) => Math.max(1, n - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={data.post_page >= postPages}
                    onClick={() => setPostPage((n) => Math.min(postPages, n + 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>

          <section className="rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                最近评论（{data.comments_total}）
              </h2>
            </div>
            <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
              {comments.length === 0 ? (
                <p className="px-4 py-6 text-sm text-zinc-500">暂无评论</p>
              ) : (
                comments.map((c) => (
                  <div key={c.id} className="px-4 py-4">
                    <Link
                      href={`/posts/${c.post_id}`}
                      className="text-sm text-zinc-700 underline dark:text-zinc-300"
                    >
                      {c.post_title || `帖子 ${c.post_id}`}
                    </Link>
                    <p className="mt-1 text-sm text-zinc-700 dark:text-zinc-300">
                      {previewText(c.content)}
                    </p>
                    <p className="mt-1 text-xs text-zinc-500">分数 {c.score}</p>
                  </div>
                ))
              )}
            </div>
            {data.comments_total > data.comment_page_size ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {data.comment_page} / {commentPages} 页
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={data.comment_page <= 1}
                    onClick={() => setCommentPage((n) => Math.max(1, n - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={data.comment_page >= commentPages}
                    onClick={() =>
                      setCommentPage((n) => Math.min(commentPages, n + 1))
                    }
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>
        </div>
      ) : null}
    </main>
  );
}
