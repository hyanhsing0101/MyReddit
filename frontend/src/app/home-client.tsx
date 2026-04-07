"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { PostFavoriteButton } from "@/components/post-favorite-button";
import { PostVoteControls } from "@/components/post-vote-controls";
import {
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiListPosts,
  apiPing,
  apiSearch,
  tagDisplayLabel,
  type BoardItem,
  type PostItem,
  type SearchScope,
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

  const [searchQ, setSearchQ] = useState("");
  const [searchScope, setSearchScope] = useState<SearchScope>("all");
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  const [searchPosts, setSearchPosts] = useState<PostItem[]>([]);
  const [searchBoards, setSearchBoards] = useState<BoardItem[]>([]);
  const [searched, setSearched] = useState(false);

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
      const token = getAccessToken();
      const body = await apiListPosts(page, listPageSize, undefined, token);
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
    void loadPosts(listPage);
  }, [listPage, loggedIn, loadPosts]);

  async function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    const q = searchQ.trim();
    if (!q) {
      setSearchError("请输入关键词");
      setSearchPosts([]);
      setSearchBoards([]);
      setSearched(false);
      return;
    }
    setSearching(true);
    setSearchError(null);
    try {
      const body = await apiSearch(q, searchScope, 20, 10);
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setSearchError(apiErrorMessage(body));
        setSearchPosts([]);
        setSearchBoards([]);
        setSearched(true);
        return;
      }
      setSearchPosts(body.data.posts ?? []);
      setSearchBoards(body.data.boards ?? []);
      setSearched(true);
    } catch (e) {
      setSearchError(e instanceof Error ? e.message : "搜索失败");
      setSearchPosts([]);
      setSearchBoards([]);
      setSearched(true);
    } finally {
      setSearching(false);
    }
  }

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
              href="/boards/favorites"
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              收藏板块
            </Link>
            <Link
              href="/posts/favorites"
              className="rounded-lg border border-zinc-300 px-4 py-2 dark:border-zinc-600"
            >
              收藏帖子
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

      <section className="rounded-xl border border-zinc-200 p-4 dark:border-zinc-800">
        <h2 className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
          搜索
        </h2>
        <form onSubmit={handleSearch} className="mt-3 flex flex-wrap gap-2">
          <input
            value={searchQ}
            onChange={(e) => setSearchQ(e.target.value)}
            placeholder="搜标题、正文、板块（FTS 精准）"
            className="min-w-[220px] flex-1 rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
          />
          <select
            value={searchScope}
            onChange={(e) => setSearchScope(e.target.value as SearchScope)}
            className="rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
          >
            <option value="all">全局</option>
            <option value="posts">仅帖子</option>
            <option value="boards">仅板块</option>
          </select>
          <button
            type="submit"
            disabled={searching}
            className="rounded-lg bg-zinc-900 px-4 py-2 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
          >
            {searching ? "搜索中…" : "搜索"}
          </button>
        </form>
        {searchError ? (
          <p className="mt-2 text-sm text-red-600 dark:text-red-400">{searchError}</p>
        ) : null}
        {searched ? (
          <div className="mt-4 space-y-4">
            {searchScope !== "boards" ? (
              <div>
                <p className="text-xs text-zinc-500">帖子结果（{searchPosts.length}）</p>
                {searchPosts.length === 0 ? (
                  <p className="mt-1 text-sm text-zinc-500">无</p>
                ) : (
                  <ul className="mt-1 space-y-1 text-sm">
                    {searchPosts.map((p) => (
                      <li key={p.id}>
                        <Link href={`/posts/${p.id}`} className="underline">
                          {p.title}
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ) : null}
            {searchScope !== "posts" ? (
              <div>
                <p className="text-xs text-zinc-500">板块结果（{searchBoards.length}）</p>
                {searchBoards.length === 0 ? (
                  <p className="mt-1 text-sm text-zinc-500">无</p>
                ) : (
                  <ul className="mt-1 space-y-1 text-sm">
                    {searchBoards.map((b) => (
                      <li key={b.id}>
                        <Link href={`/boards/${encodeURIComponent(b.slug)}`} className="underline">
                          {b.name} ({b.slug})
                        </Link>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            ) : null}
          </div>
        ) : null}
      </section>

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
              <article
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
                            ? {
                                ...p,
                                is_favorited: next,
                              }
                            : p,
                        ),
                      );
                    }}
                  />
                </div>
                <div className="min-w-0 flex-1">
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
                {post.tags?.length ? (
                  <p className="mt-2 flex flex-wrap gap-1 text-xs">
                    {post.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="rounded bg-zinc-200 px-2 py-0.5 text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300"
                      >
                        #{tagDisplayLabel(tag)}
                      </span>
                    ))}
                  </p>
                ) : null}
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
                </div>
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
                onClick={() => setListPage((p) => Math.max(1, p - 1))}
                className="rounded border border-zinc-300 px-3 py-1 disabled:opacity-40 dark:border-zinc-600"
              >
                上一页
              </button>
              <button
                type="button"
                disabled={!canNext}
                onClick={() =>
                  setListPage((p) => Math.min(totalPages, p + 1))
                }
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
