"use client";

import { useRef, useState } from "react";
import { MarkdownImageUpload } from "@/components/markdown-image-upload";
import { SafeMarkdown } from "@/components/safe-markdown";

type MarkdownEditorProps = {
  value: string;
  onChange: (next: string) => void;
  accessToken: string | null;
  disabled?: boolean;
  rows?: number;
  className?: string;
  placeholder?: string;
  compact?: boolean;
};

export function MarkdownEditor({
  value,
  onChange,
  accessToken,
  disabled,
  rows = 10,
  className = "",
  placeholder,
  compact = false,
}: MarkdownEditorProps) {
  const editorRef = useRef<HTMLTextAreaElement | null>(null);
  const [showPreview, setShowPreview] = useState(!compact);

  function insertSnippet(snippet: string) {
    const el = editorRef.current;
    if (!el) {
      onChange(value + snippet);
      return;
    }
    const start = el.selectionStart ?? value.length;
    const end = el.selectionEnd ?? start;
    const selected = value.slice(start, end);
    const text =
      snippet.indexOf("$TEXT$") >= 0
        ? snippet.replaceAll("$TEXT$", selected || "文本")
        : snippet;
    const next = `${value.slice(0, start)}${text}${value.slice(end)}`;
    onChange(next);
    requestAnimationFrame(() => {
      el.focus();
      const pos = start + text.length;
      el.setSelectionRange(pos, pos);
    });
  }

  const toolbarBtn =
    "rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600";
  const layout = compact
    ? showPreview
      ? "grid grid-cols-1 gap-3"
      : "grid grid-cols-1"
    : `grid gap-4 ${showPreview ? "lg:grid-cols-2" : "grid-cols-1"}`;

  return (
    <div className={`flex flex-col gap-2 ${className}`.trim()}>
      <div className="flex flex-wrap items-center gap-2 rounded-lg border border-zinc-200 bg-zinc-50 p-2 dark:border-zinc-700 dark:bg-zinc-900">
        <button type="button" onClick={() => insertSnippet("## $TEXT$\n\n")} className={toolbarBtn}>
          H2
        </button>
        <button type="button" onClick={() => insertSnippet("**$TEXT$**")} className={toolbarBtn}>
          粗体
        </button>
        <button type="button" onClick={() => insertSnippet("*$TEXT$*")} className={toolbarBtn}>
          斜体
        </button>
        <button type="button" onClick={() => insertSnippet("- $TEXT$\n")} className={toolbarBtn}>
          列表
        </button>
        <button type="button" onClick={() => insertSnippet("> $TEXT$\n")} className={toolbarBtn}>
          引用
        </button>
        <button type="button" onClick={() => insertSnippet("`$TEXT$`")} className={toolbarBtn}>
          行内代码
        </button>
        <button type="button" onClick={() => insertSnippet("\n```\n$TEXT$\n```\n")} className={toolbarBtn}>
          代码块
        </button>
        <button type="button" onClick={() => insertSnippet("[$TEXT$](https://example.com)")} className={toolbarBtn}>
          链接
        </button>
        <button type="button" onClick={() => insertSnippet("\n---\n")} className={toolbarBtn}>
          分割线
        </button>
        <MarkdownImageUpload
          accessToken={accessToken}
          disabled={disabled}
          onInsert={(snip) => insertSnippet(snip)}
        />
        <button
          type="button"
          onClick={() => setShowPreview((v) => !v)}
          className="ml-auto rounded border border-zinc-300 px-2 py-1 text-xs dark:border-zinc-600"
        >
          {showPreview ? "隐藏预览" : "显示预览"}
        </button>
      </div>

      <div className={layout}>
        <textarea
          ref={editorRef}
          className="min-h-[200px] rounded-lg border border-zinc-300 bg-white px-3 py-2 font-mono text-sm dark:border-zinc-600 dark:bg-zinc-900"
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={rows}
          placeholder={placeholder}
          disabled={disabled}
        />
        {showPreview ? (
          <div className="min-h-[160px] rounded-lg border border-zinc-300 bg-white p-3 dark:border-zinc-600 dark:bg-zinc-900">
            <p className="mb-2 text-xs text-zinc-500">实时预览</p>
            <SafeMarkdown markdown={value} className="text-sm" />
          </div>
        ) : null}
      </div>
    </div>
  );
}
