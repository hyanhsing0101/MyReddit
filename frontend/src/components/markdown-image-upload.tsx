"use client";

import { useRef, useState } from "react";
import {
  API_NEED_LOGIN_CODE,
  API_SUCCESS_CODE,
  API_UPLOAD_DISABLED_CODE,
  API_UPLOAD_INVALID_IMAGE_CODE,
  API_UPLOAD_TOO_LARGE_CODE,
  apiErrorMessage,
  apiUploadImage,
} from "@/lib/api";

type MarkdownImageUploadProps = {
  accessToken: string | null;
  disabled?: boolean;
  onInsert: (markdownSnippet: string) => void;
};

/** 上传图片并在光标处插入 `![](url)`（需后端开启 upload 并配置 public_url）。 */
export function MarkdownImageUpload({
  accessToken,
  disabled,
  onInsert,
}: MarkdownImageUploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);

  async function onPick(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!file || disabled || busy) return;
    const token = accessToken;
    if (!token) {
      setMsg("请先登录");
      return;
    }
    setBusy(true);
    setMsg(null);
    try {
      const body = await apiUploadImage(token, file);
      if (body.code === API_SUCCESS_CODE && body.data?.url) {
        const snippet = `\n![](${body.data.url})\n`;
        onInsert(snippet);
        return;
      }
      if (body.code === API_NEED_LOGIN_CODE) {
        setMsg("请先登录");
        return;
      }
      if (body.code === API_UPLOAD_DISABLED_CODE) {
        setMsg("服务器未开启图床或未配置 public_url");
        return;
      }
      if (body.code === API_UPLOAD_TOO_LARGE_CODE) {
        setMsg("图片过大");
        return;
      }
      if (body.code === API_UPLOAD_INVALID_IMAGE_CODE) {
        setMsg("仅支持 PNG / JPEG / GIF / WebP");
        return;
      }
      setMsg(apiErrorMessage(body));
    } catch (err) {
      setMsg(err instanceof Error ? err.message : "上传失败");
    } finally {
      setBusy(false);
    }
  }

  return (
    <span className="inline-flex flex-col gap-0.5">
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg,image/gif,image/webp"
        className="hidden"
        onChange={(e) => void onPick(e)}
      />
      <button
        type="button"
        disabled={disabled || busy}
        onClick={() => inputRef.current?.click()}
        className="rounded border border-zinc-300 px-2 py-1 text-xs text-zinc-700 disabled:opacity-50 dark:border-zinc-600 dark:text-zinc-200"
      >
        {busy ? "上传中…" : "插入图片"}
      </button>
      {msg ? <span className="text-xs text-amber-700 dark:text-amber-400">{msg}</span> : null}
    </span>
  );
}
