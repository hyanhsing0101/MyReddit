"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { MarkdownEditor } from "@/components/markdown-editor";
import { SafeMarkdown } from "@/components/safe-markdown";
import { CommentVoteControls } from "@/components/comment-vote-controls";
import { PostFavoriteButton } from "@/components/post-favorite-button";
import { PostVoteControls } from "@/components/post-vote-controls";
import {
  API_CANNOT_REPORT_OWN_COMMENT_CODE,
  API_CANNOT_REPORT_OWN_POST_CODE,
  API_COMMENT_NOT_EXIST_CODE,
  API_CANNOT_APPEAL_UNSEALED_POST_CODE,
  API_DUPLICATE_COMMENT_REPORT_CODE,
  API_DUPLICATE_POST_REPORT_CODE,
  API_FORBIDDEN_CODE,
  API_INVALID_COMMENT_PARENT_CODE,
  API_PARENT_COMMENT_MISMATCH_CODE,
  API_POST_COMMENTS_LOCKED_CODE,
  API_POST_APPEAL_NOT_EXIST_CODE,
  API_POST_NOT_EXIST_CODE,
  API_POST_SEALED_CODE,
  API_SUCCESS_CODE,
  apiCreateComment,
  apiCreateCommentReport,
  apiCreatePostReport,
  apiDeletePost,
  apiErrorMessage,
  apiGetMyPostAppeal,
  apiGetPost,
  apiLockPostComments,
  apiListComments,
  apiMePermissions,
  apiPinPost,
  apiSealPost,
  apiUnlockPostComments,
  apiUnpinPost,
  apiUnsealPost,
  apiUpsertMyPostAppeal,
  tagDisplayLabel,
  type CommentItem,
  type MePermissionsPayload,
  type PostAppealItem,
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

/** 软删：作者、站主、板块版主可删（与后端一致）。 */
function canDeletePost(
  post: PostItem,
  me: MePermissionsPayload | null,
): boolean {
  if (!me) return false;
  if (me.is_site_admin) return true;
  if (me.moderated_board_ids?.includes(post.board_id)) return true;
  if (post.author_id != null) return post.author_id === me.user_id;
  return false;
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

function FlagReportIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
      className="h-4 w-4"
      aria-hidden
    >
      <path d="M5 21V3h9l.4 2H19v9h-6l-.4-2H7v9H5zm2-11h6.2l.4 2H17V8h-6.2l-.4-2H7v4z" />
    </svg>
  );
}

function CommentBranch({
  postId,
  node,
  depth,
  onReply,
  onCommentVotePatch,
  me,
  meLoaded,
  onOpenCommentReport,
}: {
  postId: number;
  node: CommentNode;
  depth: number;
  onReply: (id: number, name: string) => void;
  onCommentVotePatch: (
    commentId: number,
    patch: { score: number; my_vote: number | null },
  ) => void;
  me: MePermissionsPayload | null;
  meLoaded: boolean;
  onOpenCommentReport: (commentId: number) => void;
}) {
  const label =
    node.author_username ||
    (node.author_id != null ? `用户 ${node.author_id}` : "匿名");
  const showCommentReportBtn =
    meLoaded &&
    !!me &&
    (node.author_id == null || node.author_id !== me.user_id);
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
          <div className="flex w-full min-w-0 flex-wrap items-baseline gap-2 text-sm">
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
            {showCommentReportBtn ? (
              <button
                type="button"
                title="举报该评论"
                aria-label="举报该评论"
                onClick={() => onOpenCommentReport(node.id)}
                className="ms-auto inline-flex shrink-0 items-center justify-center rounded p-1 text-zinc-400 hover:bg-zinc-100 hover:text-rose-600 dark:hover:bg-zinc-800 dark:hover:text-rose-400"
              >
                <FlagReportIcon />
              </button>
            ) : null}
          </div>
          <SafeMarkdown
            markdown={node.content}
            className="mt-1 text-sm text-zinc-700 dark:text-zinc-300"
          />
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
          me={me}
          meLoaded={meLoaded}
          onOpenCommentReport={onOpenCommentReport}
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
  const [modBusy, setModBusy] = useState(false);
  const [postReportOpen, setPostReportOpen] = useState(false);
  const [reportReason, setReportReason] = useState("");
  const [reportDetail, setReportDetail] = useState("");
  const [reportBusy, setReportBusy] = useState(false);
  const [reportError, setReportError] = useState<string | null>(null);
  const [reportOk, setReportOk] = useState<string | null>(null);

  const [commentReportId, setCommentReportId] = useState<number | null>(null);
  const [cReportReason, setCReportReason] = useState("");
  const [cReportDetail, setCReportDetail] = useState("");
  const [cReportBusy, setCReportBusy] = useState(false);
  const [cReportError, setCReportError] = useState<string | null>(null);
  const [cReportOk, setCReportOk] = useState<string | null>(null);
  const [appeal, setAppeal] = useState<PostAppealItem | null>(null);
  const [appealReason, setAppealReason] = useState("");
  const [appealTitle, setAppealTitle] = useState("");
  const [appealContent, setAppealContent] = useState("");
  const [appealReply, setAppealReply] = useState("");
  const [appealBusy, setAppealBusy] = useState(false);
  const [appealError, setAppealError] = useState<string | null>(null);
  const [appealOk, setAppealOk] = useState<string | null>(null);

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

  const showReport = useMemo(() => {
    if (!post || !meLoaded || !me) return false;
    if (post.author_id == null) return true;
    return post.author_id !== me.user_id;
  }, [post, me, meLoaded]);

  const showAppeal = useMemo(() => {
    if (!post || !meLoaded || !me) return false;
    if (!post.sealed) return false;
    return post.author_id != null && post.author_id === me.user_id;
  }, [post, me, meLoaded]);

  useEffect(() => {
    if (!post || !showAppeal) {
      setAppeal(null);
      setAppealError(null);
      return;
    }
    const token = getAccessToken();
    if (!token) return;
    let cancelled = false;
    (async () => {
      try {
        const body = await apiGetMyPostAppeal(token, post.id);
        if (cancelled) return;
        if (body.code === API_POST_APPEAL_NOT_EXIST_CODE) {
          setAppeal(null);
          setAppealReason("");
          setAppealTitle(post.title);
          setAppealContent(post.content);
          setAppealReply("");
          return;
        }
        if (body.code !== API_SUCCESS_CODE || !body.data) {
          setAppeal(null);
          return;
        }
        setAppeal(body.data);
        setAppealReason(body.data.reason ?? "");
        setAppealTitle(body.data.requested_title ?? post.title);
        setAppealContent(body.data.requested_content ?? post.content);
        setAppealReply(body.data.user_reply ?? "");
      } catch {
        if (!cancelled) setAppeal(null);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [post, showAppeal]);

  const isSealed = !!post?.sealed;
  const commentsLocked = !!post?.comments_locked;
  const isPinned = !!post?.pinned;
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

  async function handleLockComments() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    if (!window.confirm("确定锁定评论？锁定后普通用户不可新增评论。")) return;
    setSealError(null);
    setModBusy(true);
    try {
      const body = await apiLockPostComments(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        const reload = await apiGetPost(post.id, token);
        if (reload.code === API_SUCCESS_CODE && reload.data) setPost(reload.data);
        return;
      }
      setSealError(apiErrorMessage(body));
    } catch (e) {
      setSealError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setModBusy(false);
    }
  }

  async function handleUnlockComments() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    if (!window.confirm("确定解除评论锁定？")) return;
    setSealError(null);
    setModBusy(true);
    try {
      const body = await apiUnlockPostComments(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        const reload = await apiGetPost(post.id, token);
        if (reload.code === API_SUCCESS_CODE && reload.data) setPost(reload.data);
        return;
      }
      setSealError(apiErrorMessage(body));
    } catch (e) {
      setSealError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setModBusy(false);
    }
  }

  async function handlePin() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    setSealError(null);
    setModBusy(true);
    try {
      const body = await apiPinPost(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        const reload = await apiGetPost(post.id, token);
        if (reload.code === API_SUCCESS_CODE && reload.data) setPost(reload.data);
        return;
      }
      setSealError(apiErrorMessage(body));
    } catch (e) {
      setSealError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setModBusy(false);
    }
  }

  async function handleUnpin() {
    if (!post) return;
    const token = getAccessToken();
    if (!token) return;
    setSealError(null);
    setModBusy(true);
    try {
      const body = await apiUnpinPost(token, post.id);
      if (body.code === API_SUCCESS_CODE) {
        const reload = await apiGetPost(post.id, token);
        if (reload.code === API_SUCCESS_CODE && reload.data) setPost(reload.data);
        return;
      }
      setSealError(apiErrorMessage(body));
    } catch (e) {
      setSealError(e instanceof Error ? e.message : "操作失败");
    } finally {
      setModBusy(false);
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
        if (body.code === API_POST_COMMENTS_LOCKED_CODE) {
          setCommentSubmitError("该帖已锁定评论，暂不可评论");
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

  async function handleReportPost(e: React.FormEvent) {
    e.preventDefault();
    if (!post) return;
    const token = getAccessToken();
    if (!token) {
      setReportError("请先登录后举报");
      return;
    }
    const reason = reportReason.trim();
    const detail = reportDetail.trim();
    if (!reason) {
      setReportError("请填写举报原因");
      return;
    }
    setReportBusy(true);
    setReportError(null);
    setReportOk(null);
    try {
      const body = await apiCreatePostReport(token, post.id, { reason, detail });
      if (body.code !== API_SUCCESS_CODE) {
        if (body.code === API_CANNOT_REPORT_OWN_POST_CODE) {
          setReportError("不能举报自己的帖子");
          return;
        }
        if (body.code === API_DUPLICATE_POST_REPORT_CODE) {
          setReportError("你已有一条待处理举报，请等待版主处理");
          return;
        }
        if (body.code === API_POST_NOT_EXIST_CODE) {
          setNotFound(true);
          setPost(null);
          return;
        }
        setReportError(apiErrorMessage(body));
        return;
      }
      setReportReason("");
      setReportDetail("");
      setReportOk("举报已提交，感谢反馈。");
    } catch (err) {
      setReportError(err instanceof Error ? err.message : "举报失败");
    } finally {
      setReportBusy(false);
    }
  }

  async function handleCommentReportSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!post || commentReportId == null) return;
    const token = getAccessToken();
    if (!token) {
      setCReportError("请先登录后举报");
      return;
    }
    const reason = cReportReason.trim();
    const detail = cReportDetail.trim();
    if (!reason) {
      setCReportError("请填写举报原因");
      return;
    }
    setCReportBusy(true);
    setCReportError(null);
    setCReportOk(null);
    try {
      const body = await apiCreateCommentReport(
        token,
        post.id,
        commentReportId,
        { reason, detail },
      );
      if (body.code !== API_SUCCESS_CODE) {
        if (body.code === API_CANNOT_REPORT_OWN_COMMENT_CODE) {
          setCReportError("不能举报自己的评论");
          return;
        }
        if (body.code === API_DUPLICATE_COMMENT_REPORT_CODE) {
          setCReportError("你已有一条待处理举报，请等待版主处理");
          return;
        }
        if (body.code === API_COMMENT_NOT_EXIST_CODE) {
          setCReportError("评论不存在或已删除");
          return;
        }
        if (body.code === API_POST_NOT_EXIST_CODE) {
          setNotFound(true);
          setPost(null);
          return;
        }
        setCReportError(apiErrorMessage(body));
        return;
      }
      setCReportReason("");
      setCReportDetail("");
      setCReportOk("举报已提交，感谢反馈。");
    } catch (err) {
      setCReportError(err instanceof Error ? err.message : "举报失败");
    } finally {
      setCReportBusy(false);
    }
  }

  async function handleAppealSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!post) return;
    const token = getAccessToken();
    if (!token) {
      setAppealError("请先登录后再提交申诉");
      return;
    }
    const reason = appealReason.trim();
    const title = appealTitle.trim();
    const content = appealContent.trim();
    const userReply = appealReply.trim();
    if (!reason || !title || !content) {
      setAppealError("请填写申诉说明与修改后的标题/正文");
      return;
    }
    setAppealBusy(true);
    setAppealError(null);
    setAppealOk(null);
    try {
      const body = await apiUpsertMyPostAppeal(token, post.id, {
        reason,
        requested_title: title,
        requested_content: content,
        user_reply: userReply,
      });
      if (body.code !== API_SUCCESS_CODE) {
        if (body.code === API_CANNOT_APPEAL_UNSEALED_POST_CODE) {
          setAppealError("帖子未被封禁，无需申诉");
          return;
        }
        if (body.code === API_FORBIDDEN_CODE) {
          setAppealError("你不是该帖作者，无法申诉");
          return;
        }
        if (body.code === API_POST_NOT_EXIST_CODE) {
          setNotFound(true);
          setPost(null);
          return;
        }
        setAppealError(apiErrorMessage(body));
        return;
      }
      setAppealOk("申诉已提交/已更新，请等待版主处理。");
      const reload = await apiGetMyPostAppeal(token, post.id);
      if (reload.code === API_SUCCESS_CODE && reload.data) {
        setAppeal(reload.data);
      }
    } catch (err) {
      setAppealError(err instanceof Error ? err.message : "申诉失败");
    } finally {
      setAppealBusy(false);
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
                {showReport ? (
                  <button
                    type="button"
                    title="举报该帖"
                    aria-label="举报该帖"
                    onClick={() => {
                      setPostReportOpen(true);
                      setReportError(null);
                      setReportOk(null);
                    }}
                    className="inline-flex shrink-0 items-center justify-center rounded-lg border border-zinc-300 p-2 text-zinc-500 hover:border-rose-300 hover:text-rose-600 dark:border-zinc-600 dark:hover:border-rose-800 dark:hover:text-rose-400"
                  >
                    <FlagReportIcon />
                  </button>
                ) : null}
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
                {modActions?.can_lock_comments ? (
                  <button
                    type="button"
                    disabled={modBusy}
                    onClick={() => void handleLockComments()}
                    className="shrink-0 rounded-lg border border-amber-500 px-3 py-1.5 text-sm text-amber-900 disabled:opacity-50 dark:border-amber-700 dark:text-amber-200"
                  >
                    锁评
                  </button>
                ) : null}
                {modActions?.can_unlock_comments ? (
                  <button
                    type="button"
                    disabled={modBusy}
                    onClick={() => void handleUnlockComments()}
                    className="shrink-0 rounded-lg border border-emerald-500 px-3 py-1.5 text-sm text-emerald-800 disabled:opacity-50 dark:border-emerald-700 dark:text-emerald-200"
                  >
                    解锁评论
                  </button>
                ) : null}
                {modActions?.can_pin ? (
                  <button
                    type="button"
                    disabled={modBusy}
                    onClick={() => void handlePin()}
                    className="shrink-0 rounded-lg border border-indigo-500 px-3 py-1.5 text-sm text-indigo-800 disabled:opacity-50 dark:border-indigo-700 dark:text-indigo-200"
                  >
                    置顶
                  </button>
                ) : null}
                {modActions?.can_unpin ? (
                  <button
                    type="button"
                    disabled={modBusy}
                    onClick={() => void handleUnpin()}
                    className="shrink-0 rounded-lg border border-slate-500 px-3 py-1.5 text-sm text-slate-800 disabled:opacity-50 dark:border-slate-600 dark:text-slate-200"
                  >
                    取消置顶
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
            {commentsLocked ? (
              <div className="mt-3 rounded-lg border border-zinc-300 bg-zinc-50 px-4 py-3 text-sm text-zinc-800 dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-200">
                本帖评论已锁定，暂不可新增评论。
              </div>
            ) : null}
            {isPinned ? (
              <div className="mt-3 rounded-lg border border-indigo-200 bg-indigo-50 px-4 py-2 text-sm text-indigo-900 dark:border-indigo-900 dark:bg-indigo-950/40 dark:text-indigo-100">
                该帖已置顶。
              </div>
            ) : null}
            {showAppeal ? (
              <section className="mt-6 rounded-xl border border-blue-200 bg-blue-50/60 p-4 dark:border-blue-900 dark:bg-blue-950/30">
                <h2 className="text-sm font-medium text-blue-950 dark:text-blue-100">
                  封帖申诉（作者）
                </h2>
                {appeal ? (
                  <p className="mt-2 text-xs text-blue-900/80 dark:text-blue-200/80">
                    当前状态：
                    {appeal.status === "open"
                      ? "待处理"
                      : appeal.status === "in_review"
                        ? "处理中"
                        : appeal.status === "approved"
                          ? "已通过"
                          : "已驳回"}
                    （申诉 #{appeal.id}）
                  </p>
                ) : (
                  <p className="mt-2 text-xs text-blue-900/80 dark:text-blue-200/80">
                    你还没有提交申诉；请填写修改后的标题与正文，并说明申诉理由。
                  </p>
                )}
                {appeal &&
                (appeal.status === "approved" || appeal.status === "rejected") ? (
                  <div className="mt-3 space-y-2 text-sm text-blue-950 dark:text-blue-100">
                    {appeal.moderator_reply ? (
                      <p className="whitespace-pre-wrap">
                        <span className="font-medium">版主回复：</span>
                        {appeal.moderator_reply}
                      </p>
                    ) : null}
                    <p className="text-xs text-blue-900/70 dark:text-blue-200/70">
                      更新时间：
                      {new Date(appeal.update_time).toLocaleString("zh-CN")}
                    </p>
                  </div>
                ) : null}
                <form onSubmit={handleAppealSubmit} className="mt-3 space-y-3">
                  <div>
                    <p className="mb-1 text-xs text-blue-900/70 dark:text-blue-200/70">
                      申诉说明（必填）
                    </p>
                    <MarkdownEditor
                      value={appealReason}
                      onChange={(v) => setAppealReason(v.slice(0, 500))}
                      accessToken={getAccessToken()}
                      rows={4}
                      compact
                      placeholder="说明为何应解封、做了哪些整改"
                      disabled={
                        appealBusy ||
                        (!!appeal &&
                          (appeal.status === "approved" || appeal.status === "rejected"))
                      }
                    />
                  </div>
                  <input
                    value={appealTitle}
                    onChange={(e) => setAppealTitle(e.target.value)}
                    maxLength={300}
                    placeholder="修改后的标题（必填）"
                    disabled={
                      appealBusy ||
                      (!!appeal &&
                        (appeal.status === "approved" || appeal.status === "rejected"))
                    }
                    className="w-full rounded-lg border border-blue-200 bg-white px-3 py-2 text-sm dark:border-blue-900 dark:bg-zinc-900"
                  />
                  <div>
                    <p className="mb-1 text-xs text-blue-900/70 dark:text-blue-200/70">
                      修改后的正文（必填）
                    </p>
                    <MarkdownEditor
                      value={appealContent}
                      onChange={(v) => setAppealContent(v.slice(0, 20000))}
                      accessToken={getAccessToken()}
                      rows={8}
                      placeholder="修改后的正文（支持 Markdown）"
                      disabled={
                        appealBusy ||
                        (!!appeal &&
                          (appeal.status === "approved" || appeal.status === "rejected"))
                      }
                    />
                  </div>
                  <div>
                    <p className="mb-1 text-xs text-blue-900/70 dark:text-blue-200/70">
                      给版主的补充回复（可选）
                    </p>
                    <MarkdownEditor
                      value={appealReply}
                      onChange={(v) => setAppealReply(v.slice(0, 2000))}
                      accessToken={getAccessToken()}
                      rows={3}
                      compact
                      placeholder="用于追问 / 补充材料"
                      disabled={
                        appealBusy ||
                        (!!appeal &&
                          (appeal.status === "approved" || appeal.status === "rejected"))
                      }
                    />
                  </div>
                  {appealError ? (
                    <p className="text-sm text-red-600 dark:text-red-400">{appealError}</p>
                  ) : null}
                  {appealOk ? (
                    <p className="text-sm text-emerald-700 dark:text-emerald-400">{appealOk}</p>
                  ) : null}
                  <button
                    type="submit"
                    disabled={
                      appealBusy ||
                      (!!appeal &&
                        (appeal.status === "approved" || appeal.status === "rejected"))
                    }
                    className="rounded-lg border border-blue-400 bg-blue-600 px-4 py-2 text-sm text-white disabled:opacity-50 dark:border-blue-800 dark:bg-blue-700"
                  >
                    {appealBusy
                      ? "提交中…"
                      : appeal &&
                          (appeal.status === "open" || appeal.status === "in_review")
                        ? "更新申诉"
                        : "提交申诉"}
                  </button>
                </form>
              </section>
            ) : null}
            <SafeMarkdown
              markdown={post.content}
              className="mt-8 text-sm leading-relaxed text-zinc-800 dark:text-zinc-200"
            />
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
                      me={me}
                      meLoaded={meLoaded}
                      onOpenCommentReport={(cid) => {
                        setCommentReportId(cid);
                        setCReportReason("");
                        setCReportDetail("");
                        setCReportError(null);
                        setCReportOk(null);
                      }}
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
                <MarkdownEditor
                  value={commentBody}
                  onChange={setCommentBody}
                  accessToken={getAccessToken()}
                  compact
                  rows={5}
                  placeholder={
                    isSealed
                      ? "该帖已封禁，不可评论"
                      : commentsLocked
                        ? "该帖已锁评，不可评论"
                      : getAccessToken()
                        ? "写评论…（支持 Markdown）"
                        : "登录后可发表评论"
                  }
                  disabled={!getAccessToken() || isSealed || commentsLocked}
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
                      isSealed ||
                      commentsLocked
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

          {postReportOpen && showReport ? (
            <div
              className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
              role="dialog"
              aria-modal="true"
              aria-labelledby="post-report-title"
            >
              <div className="max-h-[90vh] w-full max-w-md overflow-y-auto rounded-xl border border-zinc-200 bg-white p-4 shadow-xl dark:border-zinc-700 dark:bg-zinc-900">
                <div className="flex items-start justify-between gap-2">
                  <h2
                    id="post-report-title"
                    className="text-sm font-medium text-zinc-800 dark:text-zinc-200"
                  >
                    举报帖子
                  </h2>
                  <button
                    type="button"
                    className="rounded px-2 py-1 text-xs text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                    onClick={() => {
                      setPostReportOpen(false);
                      setReportError(null);
                      setReportOk(null);
                    }}
                  >
                    关闭
                  </button>
                </div>
                <form onSubmit={handleReportPost} className="mt-3 space-y-3">
                  <input
                    value={reportReason}
                    onChange={(e) => setReportReason(e.target.value)}
                    maxLength={120}
                    placeholder="举报原因（必填，例如：广告、辱骂、人身攻击）"
                    className="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
                  />
                  <textarea
                    value={reportDetail}
                    onChange={(e) => setReportDetail(e.target.value)}
                    maxLength={1000}
                    rows={3}
                    placeholder="补充说明（可选）"
                    className="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
                  />
                  {reportError ? (
                    <p className="text-sm text-red-600 dark:text-red-400">{reportError}</p>
                  ) : null}
                  {reportOk ? (
                    <p className="text-sm text-emerald-700 dark:text-emerald-400">{reportOk}</p>
                  ) : null}
                  <button
                    type="submit"
                    disabled={reportBusy || !reportReason.trim()}
                    className="rounded-lg border border-rose-300 px-4 py-2 text-sm text-rose-700 disabled:opacity-50 dark:border-rose-800 dark:text-rose-300"
                  >
                    {reportBusy ? "提交中…" : "提交举报"}
                  </button>
                </form>
              </div>
            </div>
          ) : null}

          {commentReportId != null && post ? (
            <div
              className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
              role="dialog"
              aria-modal="true"
              aria-labelledby="comment-report-title"
            >
              <div className="max-h-[90vh] w-full max-w-md overflow-y-auto rounded-xl border border-zinc-200 bg-white p-4 shadow-xl dark:border-zinc-700 dark:bg-zinc-900">
                <div className="flex items-start justify-between gap-2">
                  <h2
                    id="comment-report-title"
                    className="text-sm font-medium text-zinc-800 dark:text-zinc-200"
                  >
                    举报评论 #{commentReportId}
                  </h2>
                  <button
                    type="button"
                    className="rounded px-2 py-1 text-xs text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                    onClick={() => {
                      setCommentReportId(null);
                      setCReportError(null);
                      setCReportOk(null);
                    }}
                  >
                    关闭
                  </button>
                </div>
                <form onSubmit={handleCommentReportSubmit} className="mt-3 space-y-3">
                  <input
                    value={cReportReason}
                    onChange={(e) => setCReportReason(e.target.value)}
                    maxLength={120}
                    placeholder="举报原因（必填）"
                    className="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
                  />
                  <textarea
                    value={cReportDetail}
                    onChange={(e) => setCReportDetail(e.target.value)}
                    maxLength={1000}
                    rows={3}
                    placeholder="补充说明（可选）"
                    className="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-600 dark:bg-zinc-900"
                  />
                  {cReportError ? (
                    <p className="text-sm text-red-600 dark:text-red-400">{cReportError}</p>
                  ) : null}
                  {cReportOk ? (
                    <p className="text-sm text-emerald-700 dark:text-emerald-400">{cReportOk}</p>
                  ) : null}
                  <button
                    type="submit"
                    disabled={cReportBusy || !cReportReason.trim()}
                    className="rounded-lg border border-rose-300 px-4 py-2 text-sm text-rose-700 disabled:opacity-50 dark:border-rose-800 dark:text-rose-300"
                  >
                    {cReportBusy ? "提交中…" : "提交举报"}
                  </button>
                </form>
              </div>
            </div>
          ) : null}
        </>
      ) : null}
    </div>
  );
}
