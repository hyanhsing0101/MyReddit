"use client";

import Link from "next/link";
import { useState } from "react";
import {
  API_NEED_LOGIN_CODE,
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiVoteComment,
} from "@/lib/api";

function IconUp({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      aria-hidden
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinejoin="round"
    >
      <path d="M12 6 19 18H5L12 6z" />
    </svg>
  );
}

function IconDown({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      className={className}
      aria-hidden
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinejoin="round"
    >
      <path d="M12 18 5 6h14l-7 12z" />
    </svg>
  );
}

type CommentVoteControlsProps = {
  postId: number;
  commentId: number;
  score: number;
  myVote: number | null | undefined;
  accessToken: string | null;
  onUpdated: (patch: { score: number; my_vote: number | null }) => void;
  compact?: boolean;
};

export function CommentVoteControls({
  postId,
  commentId,
  score,
  myVote,
  accessToken,
  onUpdated,
  compact = true,
}: CommentVoteControlsProps) {
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const loggedIn = !!accessToken;
  const upActive = myVote === 1;
  const downActive = myVote === -1;

  async function submit(value: 1 | -1 | 0) {
    setErr(null);
    if (!accessToken) return;
    setBusy(true);
    try {
      const body = await apiVoteComment(
        accessToken,
        postId,
        commentId,
        value,
      );
      if (body.code === API_NEED_LOGIN_CODE) {
        setErr("请重新登录");
        return;
      }
      if (body.code !== API_SUCCESS_CODE || !body.data) {
        setErr(apiErrorMessage(body));
        return;
      }
      onUpdated({
        score: body.data.score,
        my_vote: body.data.my_vote,
      });
    } catch (e) {
      setErr(e instanceof Error ? e.message : "投票失败");
    } finally {
      setBusy(false);
    }
  }

  function onUp() {
    if (!loggedIn) return;
    if (myVote === 1) void submit(0);
    else void submit(1);
  }

  function onDown() {
    if (!loggedIn) return;
    if (myVote === -1) void submit(0);
    else void submit(-1);
  }

  const iconClass = compact ? "h-3 w-3" : "h-3.5 w-3.5";
  const iconBtn =
    "rounded p-0.5 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700 disabled:cursor-not-allowed disabled:opacity-30 dark:hover:bg-zinc-800 dark:hover:text-zinc-200";
  const iconSel = "text-zinc-900 dark:text-zinc-100";

  return (
    <div className="flex flex-col items-start gap-0.5">
      <div className="flex items-center gap-2">
        <span
          className={`tabular-nums text-zinc-500 dark:text-zinc-400 ${compact ? "text-xs" : "text-sm"}`}
        >
          {score}
        </span>
        <div className="flex items-center gap-px" role="group" aria-label="评论投票">
          <button
            type="button"
            disabled={busy || !loggedIn}
            onClick={onUp}
            title={
              loggedIn
                ? upActive
                  ? "取消"
                  : myVote === -1
                    ? "改为上"
                    : "上票"
                : "登录后可用"
            }
            className={`${iconBtn} ${upActive ? iconSel : ""}`}
            aria-pressed={upActive}
          >
            <IconUp className={iconClass} />
          </button>
          <button
            type="button"
            disabled={busy || !loggedIn}
            onClick={onDown}
            title={
              loggedIn
                ? downActive
                  ? "取消"
                  : myVote === 1
                    ? "改为下"
                    : "下票"
                : "登录后可用"
            }
            className={`${iconBtn} ${downActive ? iconSel : ""}`}
            aria-pressed={downActive}
          >
            <IconDown className={iconClass} />
          </button>
        </div>
      </div>
      {!loggedIn && !compact ? (
        <Link
          href="/login"
          className="text-[11px] text-zinc-400 underline-offset-2 hover:text-zinc-600 hover:underline dark:hover:text-zinc-300"
        >
          登录
        </Link>
      ) : null}
      {err ? (
        <p className="text-[11px] text-red-500 dark:text-red-400">{err}</p>
      ) : null}
    </div>
  );
}
