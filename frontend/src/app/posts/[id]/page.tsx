"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CommentVoteControls } from "@/components/comment-vote-controls";
import { PostFavoriteButton } from "@/components/post-favorite-button";
import { PostVoteControls } from "@/components/post-vote-controls";
import {
  API_COMMENT_NOT_EXIST_CODE,
  API_FORBIDDEN_CODE,
  API_INVALID_COMMENT_PARENT_CODE,
  API_PARENT_COMMENT_MISMATCH_CODE,
  API_POST_NOT_EXIST_CODE,
  API_POST_SEALED_CODE,
  API_SUCCESS_CODE,
  apiCreateComment,
  apiDeletePost,
  apiErrorMessage,
  apiGetPost,
  apiListComments,
  apiMePermissions,
  apiSealPost,
  apiUnsealPost,
  tagDisplayLabel,
  type CommentItem,
  type MePermissionsPayload,
  type PostItem,
} from "@/lib/api";
import { getAccessToken } from "@/lib/auth-storage";

function canEditPost(
  post: PostItem,
  me: MePermissionsPayload | null,
): boolean {
  if (!me) return false;
  if (me.is_site_admin) return true;
  if (post.author_id == null) return false;
  return post.author_id === me.user_id;
}

/** 软删：仅作者；无主帖仅站主（与后端一致）。 */
function canDeletePost(
  post: PostItem,
  me: MePermissionsPayload | null,
): boolean {
  if (!me) return false;
  if (post.author_id != null) return post.author_id === me.user_id;
  return me.is_site_admin;
}

type CommentNode = CommentItem & { children: CommentNode[] };

function buildCommentTree(flat: CommentItem[]): CommentNode[] {
  const map = new Map<number, CommentNode>();
  for (const c of flat) {
    map.set(c.id, { ...c, children: [] });
  }
  const roots: CommentNode[] = [];
  for (const c of flat) {
    const node = map.get(c.id)!;
    if (c.parent_id == null) {
      roots.push(node);
    } else {
      const parent = map.get(c.parent_id);
      if (parent) {
        parent.children.push(node);
      } else {
        roots.push(node);
      }
    }
  }
  return roots;
}

function CommentBranch({
  postId,
  node,
  depth,
  onReply,
  onCommentVotePatch,
}: {
  postId: number;
  node: CommentNode;
  depth: number;
  onReply: (id: number, name: string) => void;
  onCommentVotePatch: (
    commentId: number,
    patch: { score: number; my_vote: number | null },
  ) => void;
}) {
  const label =
    node.author_username ||
    (node.author_id != null ? `用户 ${node.author_id}` : "匿名");
  return (
    <div
      className={
        depth > 0
          ? "mt-3 border-l-2 border-zinc-200 pl-3 dark:border-zinc-700"
          : ""
      }
    >
      <div className="flex gap-3">
        <CommentVoteControls
          postId={postId}
          commentId={node.id}
          score={node.score ?? 0}
          myVote={node.my_vote ?? null}
          accessToken={getAccessToken()}
          onUpdated={(patch) => onCommentVotePatch(node.id, patch)}
        />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-baseline gap-2 text-sm">
            {node.author_id != null ? (
              <Link
                href={`/users/${node.author_id}`}
                className="font-medium text-zinc-800 underline dark:text-zinc-200"
              >
                {label}
              </Link>
            ) : (
              <span className="font-medium text-zinc-800 dark:text-zinc-200">
                {label}
              </span>
            )}
            <span className="text-xs text-zinc-500">
              {new Date(node.create_time).toLocaleString("zh-CN", {
                month: "numeric",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </span>
          </div>
          <p className="mt-1 whitespace-pre-wrap text-sm text-zinc-700 dark:text-zinc-300">
            {node.content}
          </p>
          <button
            type="button"
            onClick={() => onReply(node.id, label)}
            className="mt-1 text-xs text-zinc-600 underline dark:text-zinc-400"
          >
            回复
          </button>
        </div>
      </div>
      {node.children.map((ch) => (
        <CommentBranch
          key={ch.id}
          postId={postId}
          node={ch}
          depth={depth + 1}
          onReply={onReply}
          onCommentVotePatch={onCommentVotePatch}
        />
      ))}
    </div>
  );
}

export default function PostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const idParam = params.id;
  const id =
    typeof idParam === "string"
      ? Number.parseInt(idParam, 10)
      : Array.isArray(idParam)
        ? Number.parseInt(idParam[0] ?? "", 10)
        : NaN;

  const [post, setPost] = useState<PostItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [notFound, setNotFound] = useState(false);

  const [me, setMe] = useState<MePermissionsPayload | null>(null);
  const [meLoaded, setMeLoaded] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [sealBusy, setSealBusy] = useState(false);
  const [sealError, setSealError] = useState<string | null>(null);

  const [comments, setComments] = useState<CommentItem[]>([]);
  const [commentsTotal, setCommentsTotal] = useState(0);
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [commentsError, setCommentsError] = useState<string | null>(null);
  const [commentBody, setCommentBody] = useState("");
  const [replyTo, setReplyTo] = useState<{ id: number; name: string } | null>(
    null,
  );
  const [commentSubmitting, setCommentSubmitting] = useState(false);
  const [commentSubmitError, setCommentSubmitError] = useState<string | null>(
    null,
  );

  const loadComments = useCallback(async (postId: number) => {
    setCommentsLoading(true);
    setCommentsError(null);
    try {
      const token = getAccessToken();
      const body = await apiListComments(postId, 1, 100, token);
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setCommentsError(apiErrorMessage(body));
        setComments([]);
        setCommentsTotal(0);
        return;
      }
      setComments(Array.isArray(body.data.list) ? body.data.list : []);
      setCommentsTotal(body.data.total);
    } catch (e) {
      setCommentsError(e instanceof Error ? e.message : "评论加载失败");
      setComments([]);
    } finally {
      setCommentsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!Number.isFinite(id) || id < 1) {
      setLoading(false);
      setError("无效的帖子 ID");
      return;
    }

    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      setNotFound(false);
      try {
        const token = getAccessToken();
        const body = await apiGetPost(id, token ?? undefined);
        if (cancelled) return;
        if (body.code === API_POST_NOT_EXIST_CODE) {
          setNotFound(true);
          setPost(null);
          return;
        }
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setError(apiErrorMessage(body));
          setPost(null);
          return;
        }
        setPost(body.data);
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "加载失败");
          setPost(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [id]);

  useEffect(() => {
    if (!post) return;
    loadComments(post.id);
  }, [post, loadComments]);

  useEffect(() => {
    const token = getAccessToken();
    if (!token) {
      setMe(null);
      setMeLoaded(true);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const body = await apiMePermissions(token);
        if (cancelled) return;
        if (body.code === API_SUCCESS_CODE && body.data) {
          setMe(body.data);
        } else {
          setMe(null);
        }
      } catch {
        if (!cancelled) setMe(null);
      } finally {
        if (!cancelled) setMeLoaded(true);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const commentTree = useMemo(
    () => buildCommentTree(comments),
    [comments],
  );

  const patchCommentVote = useCallback(
    (commentId: number, patch: { score: number; my_vote: number | null }) => {
      setComments((prev) =>
        prev.map((c) =>
          c.id === commentId
            ? { ...c, score: patch.score, my_vote: patch.my_vote }
            : c,
        ),
      );
    },
    [],
  );

  const showEdit = useMemo(() => {
    if (!post || !meLoaded) return false;
    return canEditPost(post, me);
  }, [post, me, meLoaded]);

  const showDelete = useMemo(() => {
    if (!post || !meLoaded) return false;
    return canDeletePost(post, me);
  }, [post, me, meLoaded]);

  const isSealed = !!post?.sealed;
  const modActions = post?.moderation_actions;

  async function handleSeal() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    if (!window.confirm("确定封帖？封帖后普通用户不可见正文，且不可评论/投票。"))
      return;
    setSealError(null);
    setSealBusy(true);
    try {
      const body = await apiSealPost(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        const reload = await apiGetPost(post.id, token);
        if (reload.code === API_SUCCESS_CODE && reload.data) {
          setPost(reload.data);
        }
        await loadComments(post.id);
        return;
      }
      setSealError(apiErrorMessage(body));
    } catch (e) {
      setSealError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSealBusy(false);
    }
  }

  async function handleUnseal() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    if (!window.confirm("确定解封该帖？")) return;
    setSealError(null);
    setSealBusy(true);
    try {
      const body = await apiUnsealPost(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        const reload = await apiGetPost(post.id, token);
        if (reload.code === API_SUCCESS_CODE && reload.data) {
          setPost(reload.data);
        }
        await loadComments(post.id);
        return;
      }
      setSealError(apiErrorMessage(body));
    } catch (e) {
      setSealError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSealBusy(false);
    }
  }

  async function handleDelete() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    if (!window.confirm("确定删除该帖？删除后对所有人不可见。")) return;
    setDeleteError(null);
    setDeleteLoading(true);
    try {
      const body = await apiDeletePost(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        router.push("/");
        router.refresh();
        return;
      }
      if (body.code === API_FORBIDDEN_CODE) {
        setDeleteError("无权删除该帖");
        return;
      }
      if (body.code === API_POST_NOT_EXIST_CODE) {
        setNotFound(true);
        setPost(null);
        return;
      }
      setDeleteError(apiErrorMessage(body));
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : "删除失败");
    } finally {
      setDeleteLoading(false);
    }
  }

  async function handleSubmitComment(e: React.FormEvent) {
    e.preventDefault();
    if (!post) return;
    const token = getAccessToken();
    if (!token) {
      setCommentSubmitError("请先登录后再评论");
      return;
    }
    const text = commentBody.trim();
    if (!text) return;
    setCommentSubmitError(null);
    setCommentSubmitting(true);
    try {
      const body = await apiCreateComment(token, post.id, {
        content: text,
        ...(replyTo ? { parent_id: replyTo.id } : {}),
      });
      if (body.code !== API_SUCCESS_CODE) {
        if (body.code === API_POST_NOT_EXIST_CODE) {
          setNotFound(true);
          setPost(null);
          return;
        }
        if (body.code === API_POST_SEALED_CODE) {
          setCommentSubmitError("该帖已封禁，暂不可评论");
          return;
        }
        if (body.code === API_COMMENT_NOT_EXIST_CODE) {
          setCommentSubmitError("父评论不存在，可能已被删除");
          return;
        }
        if (body.code === API_PARENT_COMMENT_MISMATCH_CODE) {
          setCommentSubmitError("父评论不属于当前帖子，请刷新后重试");
          return;
        }
        if (body.code === API_INVALID_COMMENT_PARENT_CODE) {
          setCommentSubmitError("非法父评论参数");
          return;
        }
        setCommentSubmitError(apiErrorMessage(body));
        return;
      }
      setCommentBody("");
      setReplyTo(null);
      await loadComments(post.id);
    } catch (err) {
      setCommentSubmitError(
        err instanceof Error ? err.message : "发表失败",
      );
    } finally {
      setCommentSubmitting(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-10">
      <button
        type="button"
        onClick={() => router.push("/")}
        className="mb-6 text-sm text-zinc-500 underline"
      >
        ← 返回首页
      </button>

      {loading ? (
        <p className="text-sm text-zinc-500">加载中…</p>
      ) : notFound ? (
        <div className="rounded-lg border border-zinc-200 p-6 dark:border-zinc-800">
          <p className="text-zinc-700 dark:text-zinc-300">帖子不存在或已删除。</p>
          <Link href="/" className="mt-4 inline-block text-sm underline">
            回首页
          </Link>
        </div>
      ) : error ? (
        <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
      ) : post ? (
        <>
          <article>
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="flex min-w-0 flex-1 flex-wrap items-start gap-4">
                <PostVoteControls
                  postId={post.id}
                  score={post.score ?? 0}
                  myVote={post.my_vote ?? null}
                  accessToken={getAccessToken()}
                  disabled={isSealed}
                  onUpdated={(patch) => {
                    setPost((prev) =>
                      prev
                        ? {
                            ...prev,
                            score: patch.score,
                            my_vote: patch.my_vote,
                          }
                        : prev,
                    );
                  }}
                />
                <h1 className="min-w-0 flex-1 text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
                  {post.title}
                </h1>
                <PostFavoriteButton
                  postId={post.id}
                  isFavorited={!!post.is_favorited}
                  accessToken={getAccessToken()}
                  onUpdated={(next) => {
                    setPost((prev) =>
                      prev ? { ...prev, is_favorited: next } : prev,
                    );
                  }}
                />
                {showEdit ? (
                  <Link
                    href={`/posts/${post.id}/edit`}
                    className="shrink-0 rounded-lg border border-zinc-300 px-3 py-1.5 text-sm dark:border-zinc-600"
                  >
                    编辑
                  </Link>
                ) : null}
                {modActions?.can_seal ? (
                  <button
                    type="button"
                    disabled={sealBusy}
                    onClick={() => void handleSeal()}
                    className="shrink-0 rounded-lg border border-amber-400 px-3 py-1.5 text-sm text-amber-900 disabled:opacity-50 dark:border-amber-700 dark:text-amber-200"
                  >
                    封帖
                  </button>
                ) : null}
                {modActions?.can_unseal ? (
                  <button
                    type="button"
                    disabled={sealBusy}
                    onClick={() => void handleUnseal()}
                    className="shrink-0 rounded-lg border border-emerald-500 px-3 py-1.5 text-sm text-emerald-800 disabled:opacity-50 dark:border-emerald-700 dark:text-emerald-200"
                  >
                    解封
                  </button>
                ) : null}
              </div>
              {showDelete ? (
                <button
                  type="button"
                  disabled={deleteLoading}
                  onClick={handleDelete}
                  className="shrink-0 rounded-lg border border-red-300 px-3 py-1.5 text-sm text-red-700 disabled:opacity-50 dark:border-red-800 dark:text-red-400"
                >
                  {deleteLoading ? "删除中…" : "删除帖子"}
                </button>
              ) : null}
            </div>
            {sealError ? (
              <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                {sealError}
              </p>
            ) : null}
            {deleteError ? (
              <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                {deleteError}
              </p>
            ) : null}
            <p className="mt-3 text-xs text-zinc-500">
              {post.board_slug ? (
                <>
                  板块：{" "}
                  <Link
                    href={`/boards/${encodeURIComponent(post.board_slug)}`}
                    className="underline"
                  >
                    {post.board_name || post.board_slug}
                  </Link>
                  <span> · </span>
                </>
              ) : null}
              {post.author_id != null ? (
                <Link href={`/users/${post.author_id}`} className="underline">
                  作者 ID {post.author_id}
                </Link>
              ) : (
                "无主帖"
              )}{" "}
              ·{" "}
              {new Date(post.create_time).toLocaleString("zh-CN", {
                year: "numeric",
                month: "numeric",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </p>
            {isSealed ? (
              <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-950 dark:border-amber-900 dark:bg-amber-950/40 dark:text-amber-100">
                该帖已被
                {post.seal_kind === "site"
                  ? "站主"
                  : post.seal_kind === "moderator"
                    ? "版主"
                    : "管理员"}
                封禁
                {post.seal_kind ? `（${post.seal_kind}）` : ""}
                ，部分用户不可见正文；解封后恢复正常展示。
              </div>
            ) : null}
            <div className="mt-8 whitespace-pre-wrap break-words text-sm leading-relaxed text-zinc-800 dark:text-zinc-200">
              {post.content}
            </div>
            {post.tags?.length ? (
              <div className="mt-4 flex flex-wrap gap-2">
                {post.tags.map((tag) => (
                  <span
                    key={tag.id}
                    className="rounded bg-zinc-200 px-2 py-0.5 text-xs text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300"
                  >
                    #{tagDisplayLabel(tag)}
                  </span>
                ))}
              </div>
            ) : null}
          </article>

          <section className="mt-12 rounded-xl border border-zinc-200 dark:border-zinc-800">
            <div className="border-b border-zinc-200 px-4 py-3 dark:border-zinc-800">
              <h2 className="text-sm font-medium text-zinc-800 dark:text-zinc-200">
                评论
                {commentsTotal > 0 ? (
                  <span className="ml-2 font-normal text-zinc-500">
                    （{commentsTotal}）
                  </span>
                ) : null}
              </h2>
            </div>
            <div className="px-4 py-4">
              {commentsLoading ? (
                <p className="text-sm text-zinc-500">加载评论…</p>
              ) : commentsError ? (
                <p className="text-sm text-red-600 dark:text-red-400">
                  {commentsError}
                </p>
              ) : commentTree.length === 0 ? (
                <p className="text-sm text-zinc-500">暂无评论</p>
              ) : (
                <div className="space-y-6">
                  {commentTree.map((node) => (
                    <CommentBranch
                      key={node.id}
                      postId={post.id}
                      node={node}
                      depth={0}
                      onReply={(cid, name) =>
                        setReplyTo({ id: cid, name })
                      }
                      onCommentVotePatch={patchCommentVote}
                    />
                  ))}
                </div>
              )}

              <form
                onSubmit={handleSubmitComment}
                className="mt-8 border-t border-zinc-200 pt-6 dark:border-zinc-800"
              >
                {replyTo ? (
                  <p className="mb-2 text-xs text-zinc-600 dark:text-zinc-400">
                    回复 <strong>{replyTo.name}</strong>
                    <button
                      type="button"
                      className="ml-3 text-zinc-500 underline"
                      onClick={() => setReplyTo(null)}
                    >
                      取消
                    </button>
                  </p>
                ) : null}
                <textarea
                  className="min-h-[100px] w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
                  placeholder={
                    isSealed
                      ? "该帖已封禁，不可评论"
                      : getAccessToken()
                        ? "写评论…"
                        : "登录后可发表评论"
                  }
                  value={commentBody}
                  onChange={(e) => setCommentBody(e.target.value)}
                  disabled={!getAccessToken() || isSealed}
                  rows={4}
                />
                {commentSubmitError ? (
                  <p className="mt-2 text-sm text-red-600 dark:text-red-400">
                    {commentSubmitError}
                  </p>
                ) : null}
                <div className="mt-3 flex flex-wrap gap-3">
                  <button
                    type="submit"
                    disabled={
                      commentSubmitting ||
                      !getAccessToken() ||
                      !commentBody.trim() ||
                      isSealed
                    }
                    className="rounded-lg bg-zinc-900 px-4 py-2 text-sm text-white disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900"
                  >
                    {commentSubmitting ? "发送中…" : "发表评论"}
                  </button>
                  {!getAccessToken() ? (
                    <Link
                      href="/login"
                      className="self-center text-sm text-zinc-600 underline dark:text-zinc-400"
                    >
                      去登录
                    </Link>
                  ) : null}
                </div>
              </form>
            </div>
          </section>
        </>
      ) : null}
    </div>
  );
}
