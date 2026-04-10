"use client";

import Link from "next/link";
import { useState } from "react";
import {
  API_NEED_LOGIN_CODE,
  API_SUCCESS_CODE,
  apiErrorMessage,
  apiVotePost,
} from "@/lib/api";

/** 小号三角，描边、无填充块面 */
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

type PostVoteControlsProps = {
  postId: number;
  score: number;
  myVote: number | null | undefined;
  accessToken: string | null;
  onUpdated: (patch: { score: number; my_vote: number | null }) => void;
  compact?: boolean;
  /** 封帖等场景下禁止改票 */
  disabled?: boolean;
};

export function PostVoteControls({
  postId,
  score,
  myVote,
  accessToken,
  onUpdated,
  compact = false,
  disabled = false,
}: PostVoteControlsProps) {
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const loggedIn = !!accessToken;
  const upActive = myVote === 1;
  const downActive = myVote === -1;

  async function submit(value: 1 | -1 | 0) {
    setErr(null);
    if (!accessToken || disabled) return;
    setBusy(true);
    try {
      const body = await apiVotePost(accessToken, postId, value);
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
    if (!loggedIn || disabled) return;
    if (myVote === 1) void submit(0);
    else void submit(1);
  }

  function onDown() {
    if (!loggedIn || disabled) return;
    if (myVote === -1) void submit(0);
    else void submit(-1);
  }

  const iconClass = compact ? "h-3 w-3" : "h-3.5 w-3.5";
  const iconBtn =
    "rounded p-0.5 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700 disabled:cursor-not-allowed disabled:opacity-30 dark:hover:bg-zinc-800 dark:hover:text-zinc-200";
  const iconSelUp =
    "text-zinc-900 dark:text-zinc-100";
  const iconSelDown =
    "text-zinc-900 dark:text-zinc-100";

  return (
    <div className="flex flex-col items-start gap-0.5">
      <div className="flex items-center gap-3">
        <span
          className={`tabular-nums text-zinc-500 dark:text-zinc-400 ${compact ? "text-xs" : "text-sm"}`}
        >
          {score}
        </span>
        <div className="flex items-center gap-px" role="group" aria-label="投票">
          <button
            type="button"
            disabled={busy || !loggedIn || disabled}
            onClick={onUp}
            title={
              disabled
                ? "该帖已封禁"
                : loggedIn
                  ? upActive
                    ? "取消"
                    : myVote === -1
                      ? "改为上"
                      : "上票"
                  : "登录后可用"
            }
            className={`${iconBtn} ${upActive ? iconSelUp : ""}`}
            aria-pressed={upActive}
          >
            <IconUp className={iconClass} />
          </button>
          <button
            type="button"
            disabled={busy || !loggedIn || disabled}
            onClick={onDown}
            title={
              disabled
                ? "该帖已封禁"
                : loggedIn
                  ? downActive
                    ? "取消"
                    : myVote === 1
                      ? "改为下"
                      : "下票"
                  : "登录后可用"
            }
            className={`${iconBtn} ${downActive ? iconSelDown : ""}`}
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
