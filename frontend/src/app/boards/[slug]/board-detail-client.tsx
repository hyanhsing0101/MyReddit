"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { BoardFavoriteButton } from "@/components/board-favorite-button";
import { PostFavoriteButton } from "@/components/post-favorite-button";
import { PostVoteControls } from "@/components/post-vote-controls";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiGetBoardBySlug,
  apiListPosts,
  type BoardItem,
  type PostItem,
  type PostSort,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

function previewText(text: string, max = 100) {
  const t = text.replace(/\s+/g, " ").trim();
  if (t.length <= max) return t;
  return `${t.slice(0, max)}…`;
}

export default function BoardDetailClient() {
  const params = useParams();
  const router = useRouter();
  const slugRaw = params.slug;
  const slug =
    typeof slugRaw === "string"
      ? slugRaw
      : Array.isArray(slugRaw)
        ? slugRaw[0] ?? ""
        : "";

  const [board, setBoard] = useState<BoardItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [posts, setPosts] = useState<PostItem[]>([]);
  const [postPage, setPostPage] = useState(1);
  const [postSort, setPostSort] = useState<PostSort>("new");
  const [postTotal, setPostTotal] = useState(0);
  const postPageSize = 10;
  const [postsLoading, setPostsLoading] = useState(false);
  const [postsError, setPostsError] = useState<string | null>(null);
  const [authTick, setAuthTick] = useState(0);

  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key === "myreddit_access_token") setAuthTick((t) => t + 1);
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  useEffect(() => {
    if (!slug) {
      setLoading(false);
      setError("无效的板块地址");
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const body = await apiGetBoardBySlug(slug, getAccessToken());
        if (cancelled) return;
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setError(apiErrorMessage(body));
          setBoard(null);
          return;
        }
        setBoard(body.data);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
          setBoard(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [slug, authTick]);

  useEffect(() => {
    setPostPage(1);
  }, [slug]);

  const loadPosts = useCallback(
    async (boardId: number, page: number) => {
      setPostsLoading(true);
      setPostsError(null);
      try {
        const token = getAccessToken();
        const body = await apiListPosts(
          page,
          postPageSize,
          boardId,
          token,
          postSort,
        );
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setPostsError(apiErrorMessage(body));
          setPosts([]);
          return;
        }
        setPosts(body.data.list);
        setPostTotal(body.data.total);
        setPostPage(body.data.page);
      } catch (e) {
        setPostsError(e instanceof Error ? e.message : "帖子加载失败");
        setPosts([]);
      } finally {
        setPostsLoading(false);
      }
    },
    [postPageSize, postSort],
  );

  useEffect(() => {
    if (!board?.id) return;
    void loadPosts(board.id, postPage);
  }, [board?.id, postPage, authTick, postSort, loadPosts]);

  const totalPostPages = Math.max(1, Math.ceil(postTotal / postPageSize));

  return (
    <div className="mx-auto max-w-2xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push("/boards")}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回板块列表
      </button>

      {loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : error ? (
        <div className="rounded-lg border border-zinc-200 p-6 dark:border-zinc-800">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
          <Link href="/boards" className="mt-4 inline-block text-sm underline">
            回列表
          </Link>
        </div>
      ) : board ? (
        <article>
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
                {board.name}
              </h1>
              <p className="mt-2 text-sm text-zinc-500">
                slug：
                <code className="rounded bg-zinc-200 px-1 dark:bg-zinc-800">
                  {board.slug}
                </code>
                {board.is_system_sink ? (
                  <span className="ml-2 text-amber-700 dark:text-amber-400">
                    （系统板）
                  </span>
                ) : null}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <BoardFavoriteButton
                boardId={board.id}
                isSystemSink={board.is_system_sink}
                isFavorited={!!board.is_favorited}
                accessToken={getAccessToken()}
                onUpdated={(next) =>
                  setBoard((prev) =>
                    prev ? { ...prev, is_favorited: next } : prev,
                  )
                }
              />
              {!board.is_system_sink ? (
                <Link
                  href={`/post/new?board_id=${board.id}`}
                  className="rounded-lg bg-zinc-900 px-4 py-2 text-sm text-white dark:bg-zinc-100 dark:text-zinc-900"
                >
                  在本板发帖
                </Link>
              ) : null}
            </div>
          </div>
          {board.description ? (
            <p className="mt-6 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
              {board.description}
            </p>
          ) : (
            <p className="mt-6 text-sm text-zinc-500">暂无描述</p>
          )}

          <section className="mt-10 rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="flex flex-col gap-3 border-b border-zinc-200 px-4 py-3 dark:border-zinc-800 sm:flex-row sm:items-center sm:justify-between">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                本板帖子
              </h2>
              <div className="flex flex-wrap gap-2">
                {(
                  [
                    { id: "new" as const, label: "最新" },
                    { id: "hot" as const, label: "热门" },
                    { id: "top" as const, label: "高分" },
                  ] as const
                ).map(({ id, label }) => (
                  <button
                    key={id}
                    type="button"
                    onClick={() => {
                      setPostSort(id);
                      setPostPage(1);
                    }}
                    className={
                      postSort === id
                        ? "rounded-lg border border-zinc-900 bg-zinc-900 px-3 py-1 text-xs text-white dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900"
                        : "rounded-lg border border-zinc-300 px-3 py-1 text-xs dark:border-zinc-600"
                    }
                  >
                    {label}
                  </button>
                ))}
              </div>
            </div>
            <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
              {postsLoading ? (
                <p className="px-4 py-6 text-center text-sm text-zinc-500">
                  加载帖子…
                </p>
              ) : postsError ? (
                <p className="px-4 py-6 text-center text-sm text-red-600 dark:text-red-400">
                  {postsError}
                </p>
              ) : posts.length === 0 ? (
                <p className="px-4 py-6 text-center text-sm text-zinc-500">
                  本板暂无帖子
                </p>
              ) : (
                posts.map((post) => (
                  <div
                    key={post.id}
                    className="flex gap-3 px-4 py-4 max-sm:flex-col max-sm:gap-2"
                  >
                    <div className="flex shrink-0 flex-col items-center gap-2">
                      <PostVoteControls
                        postId={post.id}
                        score={post.score ?? 0}
                        myVote={post.my_vote ?? null}
                        accessToken={getAccessToken()}
                        compact
                        onUpdated={(patch) => {
                          setPosts((prev) =>
                            prev.map((p) =>
                              p.id === post.id
                                ? {
                                    ...p,
                                    score: patch.score,
                                    my_vote: patch.my_vote,
                                  }
                                : p,
                            ),
                          );
                        }}
                      />
                      <PostFavoriteButton
                        postId={post.id}
                        isFavorited={!!post.is_favorited}
                        accessToken={getAccessToken()}
                        onUpdated={(next) => {
                          setPosts((prev) =>
                            prev.map((p) =>
                              p.id === post.id
                                ? { ...p, is_favorited: next }
                                : p,
                            ),
                          );
                        }}
                      />
                    </div>
                    <div className="min-w-0 flex-1">
                      <Link
                        href={`/posts/${post.id}`}
                        className="font-medium text-zinc-900 hover:underline dark:text-zinc-100"
                      >
                        {post.title}
                      </Link>
                      <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
                        {previewText(post.content)}
                      </p>
                    </div>
                  </div>
                ))
              )}
            </div>
            {!postsLoading &&
            !postsError &&
            postTotal > postPageSize ? (
              <div className="flex items-center justify-between border-t border-zinc-200 px-4 py-3 text-sm dark:border-zinc-800">
                <span className="text-zinc-500">
                  第 {postPage} / {totalPostPages} 页 · 共 {postTotal} 条
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    disabled={postPage <= 1 || !board}
                    onClick={() => setPostPage((p) => Math.max(1, p - 1))}
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    上一页
                  </button>
                  <button
                    type="button"
                    disabled={postPage >= totalPostPages || !board}
                    onClick={() =>
                      setPostPage((p) => Math.min(totalPostPages, p + 1))
                    }
                    className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
                  >
                    下一页
                  </button>
                </div>
              </div>
            ) : null}
          </section>

          <p className="mt-8">
            <Link href="/" className="text-sm text-zinc-500 underline">
              首页
            </Link>
          </p>
        </article>
      ) : null}
    </div>
  );
}
