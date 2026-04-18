import { useCallback, useRef, useState } from "react";
import { copyToClipboard } from "@/utils/clipboard";

export function CopyButton({
  text,
  toast = false,
  timeout = 1_000,
}: {
  text: string;
  toast?: boolean;
  timeout?: number;
}) {
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(null);

  const handleClick = useCallback(() => {
    if (copied) return;
    void copyToClipboard(text, toast);
    if (timeout > 0) {
      setCopied(true);
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setCopied(false), timeout);
    }
  }, [text, toast, timeout, copied]);

  if (!text) return null;
  return (
    <button
      type="button"
      className="inline-flex items-center justify-center w-4 h-4 text-muted hover:text-foreground cursor-pointer transition-colors"
      aria-label={copied ? "已复制" : "复制"}
      onClick={handleClick}
    >
      <span className={`${copied ? "icon-[mdi--check] text-green-500" : "icon-[mdi--content-copy]"} w-3.5 h-3.5`} />
    </button>
  );
}

export function Copyable({
  text,
  children,
  toast = true,
}: {
  text: string;
  children: React.ReactNode;
  toast?: boolean;
}) {
  if (!text || text === "-") return <>{children}</>;
  return (
    <button
      type="button"
      className="cursor-pointer bg-transparent p-0 border-0 text-left font-inherit"
      title="点击复制"
      onClick={() => void copyToClipboard(text, toast)}
    >
      {children}
    </button>
  );
}
