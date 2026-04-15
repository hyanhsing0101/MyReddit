"use client";

import type { Schema } from "hast-util-sanitize";
import ReactMarkdown from "react-markdown";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkGfm from "remark-gfm";

const protocols = defaultSchema.protocols ?? {};
const hrefProto = new Set([
  "http",
  "https",
  "mailto",
  ...((protocols.href as string[] | undefined) ?? []),
]);

const markdownSanitizeSchema: Schema = {
  ...defaultSchema,
  protocols: {
    ...protocols,
    href: [...hrefProto],
    src: ["http", "https"],
  },
};

const mdRootClass =
  "safe-md max-w-none break-words text-inherit [&_a]:text-sky-600 [&_a]:underline dark:[&_a]:text-sky-400 " +
  "[&_blockquote]:my-2 [&_blockquote]:border-l-4 [&_blockquote]:border-zinc-300 [&_blockquote]:pl-3 [&_blockquote]:text-zinc-600 dark:[&_blockquote]:border-zinc-600 dark:[&_blockquote]:text-zinc-400 " +
  "[&_code]:rounded [&_code]:bg-zinc-100 [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-[0.9em] dark:[&_code]:bg-zinc-800 " +
  "[&_h1]:mb-2 [&_h1]:mt-4 [&_h1]:text-xl [&_h1]:font-semibold [&_h2]:mb-2 [&_h2]:mt-3 [&_h2]:text-lg [&_h2]:font-semibold [&_h3]:mb-1 [&_h3]:mt-2 [&_h3]:text-base [&_h3]:font-semibold " +
  "[&_hr]:my-4 [&_hr]:border-zinc-200 dark:[&_hr]:border-zinc-700 " +
  "[&_img]:my-2 [&_img]:max-h-[70vh] [&_img]:max-w-full [&_img]:rounded-md " +
  "[&_li]:my-0.5 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-6 [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-6 " +
  "[&_p]:my-2 [&_pre]:my-2 [&_pre]:max-w-full [&_pre]:overflow-x-auto [&_pre]:rounded-lg [&_pre]:bg-zinc-100 [&_pre]:p-3 dark:[&_pre]:bg-zinc-900 " +
  "[&_pre_code]:bg-transparent [&_pre_code]:p-0 " +
  "[&_table]:my-2 [&_table]:w-full [&_table]:border-collapse [&_td]:border [&_td]:border-zinc-300 [&_td]:px-2 [&_td]:py-1 dark:[&_td]:border-zinc-600 [&_th]:border [&_th]:border-zinc-300 [&_th]:px-2 [&_th]:py-1 dark:[&_th]:border-zinc-600";

type SafeMarkdownProps = {
  markdown: string;
  className?: string;
};

/** Markdown → HTML，经 rehype-sanitize 白名单过滤，禁止脚本与事件属性；图片仅允许 http(s)。 */
export function SafeMarkdown({ markdown, className = "" }: SafeMarkdownProps) {
  return (
    <div className={`${mdRootClass} ${className}`.trim()}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[[rehypeSanitize, markdownSanitizeSchema]]}
        components={{
          a: ({ node: _n, children, ...props }) => (
            <a {...props} target="_blank" rel="noopener noreferrer">
              {children}
            </a>
          ),
          img: ({ node: _n, alt, ...props }) => (
            <img
              {...props}
              alt={alt ?? ""}
              loading="lazy"
              decoding="async"
            />
          ),
          h1: ({ node: _n, children, ...props }) => (
            <h2 {...props}>{children}</h2>
          ),
        }}
      >
        {markdown}
      </ReactMarkdown>
    </div>
  );
}
